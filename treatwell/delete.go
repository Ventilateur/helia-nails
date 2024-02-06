package treatwell

import (
	"net/http"
	"strings"
)

func (tw *Treatwell) Delete(id string) error {
	return doRequestWithoutResponse(
		tw,
		http.MethodPost,
		apiURL+"/venue/"+tw.venueID+"/appointment/"+id+"/cancel",
		strings.NewReader(`{"notifyConsumer":false,"cancellationReason":"CC","requestRefund":false,"platform":"DESKTOP","includeFutureRecurrences":null}`),
		nil,
	)
}
