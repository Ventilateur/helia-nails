package core

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	googlecalendar "github.com/Ventilateur/helia-nails/googlecalendar"
	"github.com/Ventilateur/helia-nails/treatwell"
	"github.com/Ventilateur/helia-nails/utils"
)

type Sync struct {
	tw *treatwell.Treatwell
	gc *googlecalendar.GoogleCalendar
}

func New(tw *treatwell.Treatwell, ga *googlecalendar.GoogleCalendar) *Sync {
	return &Sync{
		tw: tw,
		gc: ga,
	}
}

func (s *Sync) TreatwellToGoogleCalendar(calendarID string, from time.Time, to time.Time, exceptSource models.Source) error {
	slog.Info("Syncing Treatwell to Google Calendar...")

	twAppointments, err := s.tw.ListAppointments(from, to)
	if err != nil {
		return err
	}

	ggEvents, err := s.gc.List(calendarID, from, to)
	if err != nil {
		return err
	}

	for _, twAppointment := range utils.MapToOrderedSlice(twAppointments) {
		// Ignore to avoid duplication
		if twAppointment.Source == exceptSource {
			slog.Info(fmt.Sprintf("Ignore: %s", twAppointment))
			continue
		}

		if ggEvent, ok := ggEvents[twAppointment.Id]; ok {
			if needUpdate(twAppointment, ggEvent) {
				// if the TW appointment is already on GG and needs to be updated
				err = s.gc.Update(calendarID, ggEvent.OriginalID, twAppointment)
				if err != nil {
					return err
				}
				slog.Info(fmt.Sprintf("Update: %s to %s", ggEvent, twAppointment))
			} else {
				slog.Info(fmt.Sprintf("Keep: %s", ggEvent))
			}
		} else {
			// if the TW appointment is not on GG and needs to be added
			err = s.gc.Book(calendarID, twAppointment)
			if err != nil {
				return err
			}
			slog.Info(fmt.Sprintf("Add: %s", twAppointment))
		}
	}

	for _, event := range utils.MapToOrderedSlice(ggEvents) {
		if _, ok := twAppointments[event.Id]; !ok && event.Source == models.SourceTreatwell {
			// If the GG is marked as TW source but doesn't exist in TW, then delete it (case when an appointment is deleted)
			err = s.gc.DeleteAppointment(calendarID, event.OriginalID)
			if err != nil {
				return fmt.Errorf("failed to delete event %s: %w", event, err)
			}
			slog.Info(fmt.Sprintf("Delete: %s", event.String()))
		}
	}

	return nil
}

func (s *Sync) GoogleCalendarToTreatwell(calendarID string, from time.Time, to time.Time) error {
	slog.Info("Syncing Google Calendar to Treatwell...")

	err := s.tw.Preload(from, to)
	if err != nil {
		return fmt.Errorf("failed to preload Treatwell info: %w", err)
	}

	twAppointments, err := s.tw.ListAppointments(from, to)
	if err != nil {
		return err
	}

	ggEvents, err := s.gc.List(calendarID, from, to)
	if err != nil {
		return err
	}

	for _, event := range utils.MapToOrderedSlice(ggEvents) {
		if event.Source == models.SourceTreatwell {
			slog.Info(fmt.Sprintf("Ignore: %s", event.String()))
			continue
		}

		if appointment, ok := twAppointments[event.Id]; ok {
			if needUpdate(appointment, event) {
				// if the GG event is already on TW and needs to be updated
				// TODO
				slog.Info(fmt.Sprintf("Update: %s to %s", appointment, event))
			} else {
				slog.Info(fmt.Sprintf("Keep: %s", appointment))
			}
		} else {
			// if the GG event is not on TW and needs to be added
			err = s.tw.BookAnonymously(event)
			if err != nil {
				return fmt.Errorf("failed to book Treatwell from event %s: %w", event.Id, err)
			}
			slog.Info(fmt.Sprintf("Add: %s", event))
		}
	}

	return nil
}

func needUpdate(a1, a2 models.Appointment) bool {
	return a1.StartTime.Round(time.Minute) != a2.StartTime.Round(time.Minute) ||
		a1.EndTime.Round(time.Minute) != a2.EndTime.Round(time.Minute)
}
