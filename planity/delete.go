package planity

import (
	"context"
	"fmt"

	coremodels "github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/planity/models"
	"github.com/coder/websocket/wsjson"
)

func (p *Planity) Delete(ctx context.Context, appointment coremodels.Appointment) error {
	reqId := p.nextReqId()
	req := models.NewDeleteRequest(
		reqId,
		p.authInfo.UserId,
		appointment.Employee.Planity.Id,
		appointment.Ids.Planity,
	)

	if err := wsjson.Write(ctx, p.wsConn, req); err != nil {
		return fmt.Errorf("failed to send delete request: %w", err)
	}

	if _, err := waitForMessage[models.Message[models.GenericResponse]](
		p.messages,
		func(m *models.Message[models.GenericResponse]) bool {
			return m.Desc.RequestId == reqId && m.Desc.Body.Status == "ok"
		},
	); err != nil {
		return fmt.Errorf("failed to wait for delete response: %w", err)
	}

	return nil
}
