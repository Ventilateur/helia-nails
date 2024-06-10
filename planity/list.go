package planity

import (
	"context"
	"fmt"
	"time"

	coremodels "github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/planity/models"
	"github.com/Ventilateur/helia-nails/utils"
	"nhooyr.io/websocket/wsjson"
)

func (p *Planity) List(ctx context.Context, employee coremodels.Employee, from, to time.Time) ([]coremodels.Appointment, error) {
	return p.list(ctx, employee, from, to, true)
}

func (p *Planity) list(ctx context.Context, employee coremodels.Employee, from, to time.Time, excludeBlocks bool) ([]coremodels.Appointment, error) {
	var appointments []coremodels.Appointment

	err := wsjson.Write(ctx, p.wsConn, models.NewGetCalendarRequest(p.nextReqId(), employee.Planity.Id, from, to))
	if err != nil {
		return nil, fmt.Errorf("failed to send get calendar request: %w", err)
	}

	msg, err := waitForMessage[models.Message[models.GetCalendarResponse]](
		p.messages,
		func(m *models.Message[models.GetCalendarResponse]) bool {
			return m.Type == "d" && m.Desc.Body.D != nil && m.Desc.Body.EmployeeId() == employee.Planity.Id
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for calendar response: %w", err)
	}

	for id, appointment := range msg.Desc.Body.D {
		start, end, err := parseTimes(appointment.Start, appointment.Duration)
		if err != nil {
			return nil, err
		}

		// Ignore appointments that were updated to other employee, cf. "rf" field.
		// Ignore appointments that were deleted ("dat" field not null).
		if (appointment.Rf == "" || appointment.Rf == employee.Planity.Id) &&
			appointment.DeletedAt == nil &&
			(appointment.Title != blockTitle) == excludeBlocks {

			source, alternateId := utils.ParseCustomID(appointment.Notes)
			if source == "" {
				source = coremodels.SourcePlanity
			}

			appointments = append(appointments, coremodels.Appointment{
				Source: source,
				Ids: coremodels.AppointmentIds{
					Planity: id,
					Treatwell: func() string {
						if source == coremodels.SourceTreatwell {
							return alternateId
						}
						return ""
					}(),
					Classpass: func() string {
						if source == coremodels.SourceClassPass {
							return alternateId
						}
						return ""
					}(),
				},
				Employee:   p.config.GetEmployee(coremodels.SourcePlanity, employee.Planity.Id),
				Service:    p.config.GetService(coremodels.SourcePlanity, appointment.ServiceId, ""),
				StartTime:  start,
				EndTime:    end,
				ClientName: appointment.Client.Name,
				Notes:      appointment.Notes,
			})
		}
	}

	return appointments, nil
}

func parseTimes(startStr string, duration int64) (start time.Time, end time.Time, err error) {
	start, err = time.Parse(utils.PlanityTimeFormat, startStr)
	if err != nil {
		return start, end, fmt.Errorf("failed to parse start time %s: %w", startStr, err)
	}
	end = start.Add(time.Duration(duration) * time.Minute)

	return start, end, nil
}
