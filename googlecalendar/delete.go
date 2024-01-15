package googlecalendar

func (c *GoogleCalendar) DeleteAppointment(calendarID string, id string) error {
	return c.svc.Events.Delete(calendarID, id).Do()
}
