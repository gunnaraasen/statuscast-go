package statuscast

import (
	"context"
	"time"

	api "statuscast-go/internal/statuscast"
)

// Uptime returns uptime percentage for each component over the given window.
func (rc *ReportsClient) Uptime(ctx context.Context, since, until time.Time, opts ...RequestOption) ([]UptimeReport, *Response, error) {
	res, err := rc.c.ogen.APIV4ComponentsUptimeGet(ctx, api.APIV4ComponentsUptimeGetParams{})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4ComponentsUptimeGetOK:
		report := UptimeReport{
			Uptime:      r.Uptime.Value,
			WindowStart: r.Start.Value,
			WindowEnd:   r.End.Value,
		}
		return []UptimeReport{report}, &Response{}, nil
	case *api.APIV4ComponentsUptimeGetUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// IncidentSummary returns MTTD/MTTR analytics for incidents in the given window.
// MTTD is the mean time from incident start to detection (DateCreated - StartDate).
// MTTR is the mean time from detection to resolution (EndDate - DateCreated).
func (rc *ReportsClient) IncidentSummary(ctx context.Context, since, until time.Time, opts ...RequestOption) (*IncidentSummaryReport, *Response, error) {
	body := api.APIV4IncidentsPostReq{}
	body.PageSize.SetTo(1000)
	body.StartDateAfter = api.OptNilDateTime{Value: since, Set: true}
	body.EndDateBefore = api.OptNilDateTime{Value: until, Set: true}

	res, err := rc.c.ogen.APIV4IncidentsPost(ctx, api.OptAPIV4IncidentsPostReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	r, ok := res.(*api.APIV4IncidentsPostOK)
	if !ok {
		if _, unauth := res.(*api.APIV4IncidentsPostUnauthorized); unauth {
			return nil, nil, ErrUnauthorized
		}
		return nil, nil, unexpectedResponse(res)
	}

	report := &IncidentSummaryReport{
		Since:       since,
		Until:       until,
		ByComponent: make(map[string]int),
	}
	if !r.Items.Set || r.Items.Null {
		return report, &Response{}, nil
	}

	var totalDetect, totalResolve time.Duration
	var detectCount, resolveCount int

	for _, item := range r.Items.Value {
		report.TotalIncidents++
		created := item.DateCreated.Value

		// MTTD: time from incident start to detection (StartDate < DateCreated).
		if item.StartDate.Set && !item.StartDate.Null && created.After(item.StartDate.Value) {
			totalDetect += created.Sub(item.StartDate.Value)
			detectCount++
		}

		// MTTR: time from detection to resolution.
		if item.EndDate.Set && !item.EndDate.Null {
			totalResolve += item.EndDate.Value.Sub(created)
			resolveCount++
		}

		// Count by affected component.
		if item.AffectedComponents.Set && !item.AffectedComponents.Null {
			for _, c := range item.AffectedComponents.Value {
				if c.ComponentId.Set && !c.ComponentId.Null {
					report.ByComponent[int32ToID(c.ComponentId.Value)]++
				}
			}
		}
	}

	if detectCount > 0 {
		report.MeanTimeToDetect = totalDetect / time.Duration(detectCount)
	}
	if resolveCount > 0 {
		report.MeanTimeToResolve = totalResolve / time.Duration(resolveCount)
	}

	return report, &Response{}, nil
}
