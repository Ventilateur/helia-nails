package treatwell

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	twmodels "github.com/Ventilateur/helia-nails/treatwell/models"
	"github.com/Ventilateur/helia-nails/utils"
)

func (tw *Treatwell) Update(_ context.Context, appointment models.Appointment) error {
	twAppointment, err := doRequestWithResponse[twmodels.Appointment](
		tw,
		http.MethodGet,
		apiURL+"/venue/"+tw.venueID+"/appointment/"+appointment.Id(models.SourceTreatwell)+".json",
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to get Treatwell appointment to update: %w", err)
	}

	twAppointment.AppointmentDate = appointment.StartTime.Format(time.DateOnly)
	twAppointment.StartTime = appointment.TreatwellStartTime()
	twAppointment.EndTime = appointment.TreatwellEndTime()
	twAppointment.EmployeeId = appointment.Employee.Treatwell.Id

	payload, err := json.Marshal(twAppointment)
	if err != nil {
		return fmt.Errorf("%s: %w", utils.ErrUnmarshalJSON, err)
	}

	return doRequestWithoutResponse(
		tw,
		http.MethodPut,
		apiURL+"/venue/"+tw.venueID+"/appointment/"+appointment.Id(models.SourceTreatwell)+".json",
		bytes.NewBuffer(payload),
		nil,
	)
}
