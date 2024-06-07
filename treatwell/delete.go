package treatwell

import (
	"context"
	"net/http"
	"strings"

	"github.com/Ventilateur/helia-nails/core/models"
)

func (tw *Treatwell) Delete1(id string) error {
	return doRequestWithoutResponse(
		tw,
		http.MethodPost,
		apiURL+"/venue/"+tw.venueID+"/appointment/"+id+"/cancel",
		strings.NewReader(`{"notifyConsumer":false,"cancellationReason":"CC","requestRefund":false,"platform":"DESKTOP","includeFutureRecurrences":null}`),
		nil,
	)
}

func (tw *Treatwell) Delete(_ context.Context, appointment models.Appointment) error {
	return doRequestWithoutResponse(
		tw,
		http.MethodPost,
		apiURL+"/venue/"+tw.venueID+"/appointment/"+appointment.Id(models.SourceTreatwell)+"/cancel",
		strings.NewReader(`{"notifyConsumer":false,"cancellationReason":"CC","requestRefund":false,"platform":"DESKTOP","includeFutureRecurrences":null}`),
		nil,
	)
}
