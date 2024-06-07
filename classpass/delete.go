package classpass

import (
	"context"

	"github.com/Ventilateur/helia-nails/core/models"
)

func (c *Classpass) Delete(_ context.Context, appointment models.Appointment) error {
	return c.svc.Events.Delete(appointment.Employee.Classpass.GoogleCalendarId, appointment.Ids.Classpass).Do()
}
