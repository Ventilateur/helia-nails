package treatwell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/mapping"
	twmodels "github.com/Ventilateur/helia-nails/treatwell/models"
	"github.com/Ventilateur/helia-nails/utils"
)

func (tw *Treatwell) Update(appointment models.Appointment) error {
	twAppointment, err := doRequestWithResponse[twmodels.Appointment](
		tw,
		http.MethodGet,
		apiURL+"/venue/"+tw.venueID+"/appointment/"+appointment.OriginalID+".json",
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to get Treatwell appointment to update: %w", err)
	}

	twAppointment.AppointmentDate = appointment.StartTime.Format(time.DateOnly)
	twAppointment.StartTime = fmt.Sprintf("%02d:%02d", appointment.StartTime.Hour(), appointment.StartTime.Minute())
	twAppointment.EndTime = fmt.Sprintf("%02d:%02d", appointment.EndTime.Hour(), appointment.EndTime.Minute())
	twAppointment.EmployeeId = mapping.EmployeeTreatwellIDMap[appointment.Employee]

	payload, err := json.Marshal(twAppointment)
	if err != nil {
		return fmt.Errorf("%s: %w", utils.ErrUnmarshalJSON, err)
	}

	return doRequestWithoutResponse(
		tw,
		http.MethodPut,
		apiURL+"/venue/"+tw.venueID+"/appointment/"+appointment.OriginalID+".json",
		bytes.NewBuffer(payload),
		nil,
	)
}
