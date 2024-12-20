package planity

import (
	"context"
	"fmt"
	"time"

	coremodels "github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/planity/models"
	"github.com/coder/websocket/wsjson"
)

const (
	blockTitle = "__BLOCK__"
)

func (p *Planity) Book(ctx context.Context, appointment coremodels.Appointment) error {
	reqId := p.nextReqId()
	req := models.NewBookRequest(
		appointment.Employee.Planity.Id,
		appointment.Service.Planity.Id,
		appointment.StartTime,
		appointment.EndTime,
		fmt.Sprintf("[%s] %s", appointment.Source, appointment.ClientName),
		appointment.CustomNotes(),
	)
	req.Desc.RequestId = reqId

	if err := wsjson.Write(ctx, p.wsConn, req); err != nil {
		return fmt.Errorf("failed to send book request: %w", err)
	}

	if _, err := waitForMessage[models.Message[models.GenericResponse]](
		p.messages,
		func(m *models.Message[models.GenericResponse]) bool {
			return m.Desc.RequestId == reqId && m.Desc.Body.Status == "ok"
		},
	); err != nil {
		return fmt.Errorf("failed to wait for book response: %w", err)
	}

	return nil
}

func (p *Planity) Block(ctx context.Context, employee coremodels.Employee, from, to time.Time) error {
	blockers, err := p.list(ctx, employee, from, to, false)
	if err != nil {
		return err
	}

	for _, block := range blockers {
		if block.StartTime.Truncate(time.Minute).Equal(from.Truncate(time.Minute)) &&
			block.EndTime.Truncate(time.Minute).Equal(to.Truncate(time.Minute)) {
			return nil
		}
	}

	reqId := p.nextReqId()
	req := models.NewBookRequest(
		employee.Planity.Id,
		"",
		from,
		to,
		blockTitle,
		"",
	)
	req.Desc.RequestId = reqId

	if err := wsjson.Write(ctx, p.wsConn, req); err != nil {
		return fmt.Errorf("failed to send book request: %w", err)
	}

	if _, err := waitForMessage[models.Message[models.GenericResponse]](
		p.messages,
		func(m *models.Message[models.GenericResponse]) bool {
			return m.Desc.RequestId == reqId && m.Desc.Body.Status == "ok"
		},
	); err != nil {
		return fmt.Errorf("failed to wait for book response: %w", err)
	}

	return nil
}
