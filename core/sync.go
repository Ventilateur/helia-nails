package core

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Ventilateur/helia-nails/config"
	"github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/treatwell"
	"github.com/Ventilateur/helia-nails/utils"
)

type Platform interface {
	List(ctx context.Context, employee models.Employee, from time.Time, to time.Time) ([]models.Appointment, error)
	Book(ctx context.Context, appointment models.Appointment) error
	Update(ctx context.Context, appointment models.Appointment) error
	Delete(ctx context.Context, appointment models.Appointment) error
	Block(ctx context.Context, employee models.Employee, from, to time.Time) error
	Name() models.Source
}

func Sync(ctx context.Context, src Platform, dest Platform, employee models.Employee, from time.Time, to time.Time) error {
	slog.Info(fmt.Sprintf("Syncing [%s] %s -> %s...", employee.Name, src.Name(), dest.Name()))

	srcAppointments, err := src.List(ctx, employee, from, to)
	if err != nil {
		return err
	}

	destAppointments, err := dest.List(ctx, employee, from, to)
	if err != nil {
		return err
	}

	for _, srcAppt := range srcAppointments {
		if destAppt, ok := findAppointment(destAppointments, srcAppt); ok {
			if needUpdate(srcAppt, destAppt) {
				if srcAppt.Source != dest.Name() {
					tmp := destAppt
					tmp.Employee = srcAppt.Employee
					tmp.StartTime = srcAppt.StartTime
					tmp.EndTime = srcAppt.EndTime
					if err = dest.Update(ctx, tmp); err != nil {
						return err
					}
					slog.Info(fmt.Sprintf("Update %s -> %s: %v to %v", src.Name(), dest.Name(), destAppt, srcAppt))
				} else {
					tmp := srcAppt
					tmp.Employee = destAppt.Employee
					tmp.StartTime = destAppt.StartTime
					tmp.EndTime = destAppt.EndTime
					if err = src.Update(ctx, tmp); err != nil {
						return err
					}
					slog.Info(fmt.Sprintf("Update %s -> %s: %v", dest.Name(), src.Name(), tmp))
				}
			} else {
				//slog.Info(fmt.Sprintf("Keep: %v", srcAppt))
			}
		} else {
			// if the source appointment is not found on destination
			if srcAppt.Source != dest.Name() {
				// if the appointment doesn't come from the destination platform, then add it on destination platform
				if err := dest.Book(ctx, srcAppt); err != nil {
					return err
				}
				slog.Info(fmt.Sprintf("Add in %s: %v", dest.Name(), srcAppt))
			} else {
				// if the appointment comes from the destination platform, remove it from the source platform
				if err := src.Delete(ctx, srcAppt); err != nil {
					return err
				}
				slog.Info(fmt.Sprintf("Delete in %s: %v", src.Name(), srcAppt))
			}
		}
	}

	for _, destAppt := range destAppointments {
		if _, ok := findAppointment(srcAppointments, destAppt); !ok {
			if destAppt.Source == dest.Name() {
				if err := src.Book(ctx, destAppt); err != nil {
					return err
				}
				slog.Info(fmt.Sprintf("Book in %s: %v", src.Name(), destAppt))
			} else {
				if err := dest.Delete(ctx, destAppt); err != nil {
					return err
				}
				slog.Info(fmt.Sprintf("Delete in %s: %v", dest.Name(), destAppt))
			}
		}
	}

	return nil
}

func needUpdate(a1, a2 models.Appointment) bool {
	timeChanges := !a1.StartTime.Round(time.Minute).Equal(a2.StartTime.Round(time.Minute)) ||
		!a1.EndTime.Round(time.Minute).Equal(a2.EndTime.Round(time.Minute))
	employeeChanges := a1.Employee.Name != a2.Employee.Name

	return timeChanges || employeeChanges
}

func findAppointment(appts []models.Appointment, appt models.Appointment) (models.Appointment, bool) {
	for _, each := range appts {
		if equalIgnoreEmpty(each.Ids.Treatwell, appt.Ids.Treatwell) ||
			equalIgnoreEmpty(each.Ids.Classpass, appt.Ids.Classpass) ||
			equalIgnoreEmpty(each.Ids.Planity, appt.Ids.Planity) {
			return each, true
		}
	}
	return models.Appointment{}, false
}

func equalIgnoreEmpty(s1, s2 string) bool {
	if s1 == "" || s2 == "" {
		return false
	}
	return s1 == s2
}

func SyncWorkingHours(cfg *config.Config, tw *treatwell.Treatwell, employee models.Employee, from time.Time, to time.Time, otherPlatform Platform) error {
	ctx := context.Background()
	date := from

	for date.Before(to) {
		openTime, closeTime, err := utils.ParseTimes(
			fmt.Sprintf("%sT%s:00+00:00", date.Format(time.DateOnly), cfg.OpenTime),
			fmt.Sprintf("%sT%s:00+00:00", date.Format(time.DateOnly), cfg.CloseTime),
		)
		if err != nil {
			return fmt.Errorf("failed to parse open/close times [%s, %s]: %w", cfg.OpenTime, cfg.CloseTime, err)
		}

		timeSlots, err := tw.GetWorkingHours(employee, date)
		if err != nil {
			return fmt.Errorf("failed to get working hours of %s on %s: %w", employee.Name, date.Format(time.DateOnly), err)
		}

		if len(timeSlots) == 0 {
			// The employee doesn't work that day -> Block from 10:15 to 19:15
			if err := otherPlatform.Block(ctx, employee,
				utils.TimeWithLocation(openTime), utils.TimeWithLocation(closeTime),
			); err != nil {
				return fmt.Errorf("failed to block date %s: %w", date.Format(time.DateOnly), err)
			}
		} else {
			// The employee works that day but may start late or finish early
			slot := timeSlots[0]
			slotFrom, slotTo, err := utils.ParseTimes(
				fmt.Sprintf("%sT%s:00+00:00", date.Format(time.DateOnly), slot.TimeFrom),
				fmt.Sprintf("%sT%s:00+00:00", date.Format(time.DateOnly), slot.TimeTo),
			)
			if err != nil {
				return fmt.Errorf("failed to parse time slots %v: %w", slot, err)
			}

			if slot.TimeFrom > cfg.OpenTime {
				// Block from open time (e.g. "10:15") to TimeFrom
				if err := otherPlatform.Block(ctx, employee,
					utils.TimeWithLocation(openTime), utils.TimeWithLocation(slotFrom),
				); err != nil {
					return fmt.Errorf("failed to block date %s from 10:15 to %s: %w", date.Format(time.DateOnly), slot.TimeFrom, err)
				}
			}
			if slot.TimeTo < cfg.CloseTime {
				// Block from TimeTo to close time (e.g. "19:15")
				if err := otherPlatform.Block(ctx, employee,
					utils.TimeWithLocation(slotTo), utils.TimeWithLocation(closeTime),
				); err != nil {
					return fmt.Errorf("failed to block date %s from %s to 19:15: %w", date.Format(time.DateOnly), slot.TimeTo, err)
				}
			}
		}

		date = date.Add(24 * time.Hour)
	}

	return nil
}
