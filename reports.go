package statuscast

import (
	"context"
	"errors"
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

// IncidentSummary is not supported by the StatusCast API v4.
func (rc *ReportsClient) IncidentSummary(ctx context.Context, since, until time.Time, opts ...RequestOption) (*IncidentSummaryReport, *Response, error) {
	return nil, nil, errors.New("not supported by StatusCast API v4")
}

// ListRCAs is not supported by the StatusCast API v4.
func (rc *ReportsClient) ListRCAs(ctx context.Context, page Pagination, opts ...RequestOption) (*PagedResult[RootCauseAnalysis], *Response, error) {
	return nil, nil, errors.New("not supported by StatusCast API v4")
}
