package treatwell

import (
	"fmt"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
)

func (tw *Treatwell) GetWorkingHours(employee models.Employee, date time.Time) ([]models.TimeSlot, error) {
	var timeSlots []models.TimeSlot

	employeeInfo, ok := tw.employeeInfo[employee.Treatwell.Id]
	if !ok {
		return nil, fmt.Errorf("unknown employee %s", employee.Name)
	}

	for _, workingHour := range employeeInfo.WorkingHours {
		if workingHour.Date == date.Format(time.DateOnly) {
			for _, slot := range workingHour.TimeSlots {
				timeSlots = append(timeSlots, models.TimeSlot{
					TimeFrom: slot.TimeFrom,
					TimeTo:   slot.TimeTo,
				})
			}
		}
	}

	return timeSlots, nil
}
