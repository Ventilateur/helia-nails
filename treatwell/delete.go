package treatwell

import "net/http"

func (tw *Treatwell) Delete(id string) error {
	return doRequestWithoutResponse(
		tw,
		http.MethodPost,
		apiURL+"/venue/"+tw.venueID+"/appointments/"+id+"/cancel",
		nil,
		nil,
	)
}
