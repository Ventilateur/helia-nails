package core

import (
	"fmt"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	googlecalendar "github.com/Ventilateur/helia-nails/googlecalendar"
	"github.com/Ventilateur/helia-nails/treatwell"
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
	twAppointments, err := s.tw.ListAppointments(from, to)
	if err != nil {
		return err
	}

	ggEvents, err := s.gc.List(calendarID, from, to)
	if err != nil {
		return err
	}

	for id, appointment := range twAppointments {
		// Ignore to avoid duplication
		if appointment.Source == exceptSource {
			continue
		}

		if event, ok := ggEvents[id]; ok {
			if needUpdate(appointment, event) {
				// if the TW appointment is already on GG and needs to be updated
				err = s.gc.Update(calendarID, id, appointment)
				if err != nil {
					return err
				}
			}
		} else {
			// if the TW appointment is not on GG and needs to be added
			err = s.gc.Book(calendarID, appointment)
			if err != nil {
				return err
			}
		}
	}

	for id, event := range ggEvents {
		if _, ok := twAppointments[id]; !ok && event.Source == models.SourceTreatwell {
			// If the GG is marked as TW source but doesn't exist in TW, then delete it (case when an appointment is deleted)
			err = s.gc.DeleteAppointment(calendarID, id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Sync) GoogleCalendarToTreatwell(calendarID string, from time.Time, to time.Time) error {
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

	for id, event := range ggEvents {
		if event.Source == models.SourceTreatwell {
			continue
		}

		if appointment, ok := twAppointments[id]; ok {
			if needUpdate(appointment, event) {
				// if the GG event is already on TW and needs to be updated
				// TODO
			}
		} else {
			// if the GG event is not on TW and needs to be added
			err = s.tw.BookAnonymously(event)
			if err != nil {
				return fmt.Errorf("failed to book Treatwell from event %s: %w", event.Id, err)
			}
		}
	}

	return nil
}

func needUpdate(a1, a2 models.Appointment) bool {
	return a1.StartTime.Round(time.Minute) != a2.StartTime.Round(time.Minute) ||
		a1.EndTime.Round(time.Minute) != a2.EndTime.Round(time.Minute)
}

//func (s *Sync) GoogleCalendarToTreatwell(from time.Time, to time.Time) error {
//	twAppointments, err := s.tw.ListAppointments(from, to)
//	if err != nil {
//		return err
//	}
//
//	for _, employee := range []string{"Tee", "Minette", "Jade", "Chloé"} {
//		appointments, err := s.gc.List(employee, from, to)
//		if err != nil {
//			return err
//		}
//
//		for _, appointment := range appointments {
//			if appointment.Source == models.SourceTreatwell {
//				continue
//			}
//
//			for _, twAppointment := range twAppointments {
//				if twAppointment.Employee == appointment.Employee {
//					if twAppointment.
//				}
//			}
//		}
//	}
//}
