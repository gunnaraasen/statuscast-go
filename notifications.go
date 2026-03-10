package statuscast

import (
	"context"
	"errors"

	api "statuscast-go/internal/statuscast"
)

// CreateTemplate creates a new branded notification template.
func (nc *NotificationsClient) CreateTemplate(ctx context.Context, t NotificationTemplate, opts ...RequestOption) (*NotificationTemplate, *Response, error) {
	body := api.APIV4ContenttemplatePostReq{}
	if t.Subject != "" {
		body.Subject.SetTo(t.Subject)
	}
	if t.Body != "" {
		body.Contents.SetTo(t.Body)
	}

	res, err := nc.c.ogen.APIV4ContenttemplatePost(ctx, api.OptAPIV4ContenttemplatePostReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4ContenttemplatePostOK:
		return mapContentTemplatePostOK(r), &Response{}, nil
	case *api.APIV4ContenttemplatePostUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// UpdateTemplate replaces a template's content.
func (nc *NotificationsClient) UpdateTemplate(ctx context.Context, id string, t NotificationTemplate, opts ...RequestOption) (*NotificationTemplate, *Response, error) {
	intID, err := idToInt32(id)
	if err != nil {
		return nil, nil, err
	}
	body := api.APIV4ContenttemplatePutReq{}
	body.ID.SetTo(intID)
	if t.Subject != "" {
		body.Subject.SetTo(t.Subject)
	}
	if t.Body != "" {
		body.Contents.SetTo(t.Body)
	}

	res, err := nc.c.ogen.APIV4ContenttemplatePut(ctx, api.OptAPIV4ContenttemplatePutReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4ContenttemplatePutOK:
		return mapContentTemplatePutOK(r), &Response{}, nil
	case *api.APIV4ContenttemplatePutUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// ListTemplates returns all saved notification templates.
func (nc *NotificationsClient) ListTemplates(ctx context.Context, page Pagination, opts ...RequestOption) (*PagedResult[NotificationTemplate], *Response, error) {
	res, err := nc.c.ogen.APIV4ContenttemplateGet(ctx)
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4ContenttemplateGetOKApplicationJSON:
		items := make([]NotificationTemplate, 0, len(*r))
		for _, item := range *r {
			items = append(items, NotificationTemplate{
				ID:      optInt32ToID(item.ID),
				Subject: optNilStringVal(item.Subject),
				Body:    optNilStringVal(item.Contents),
			})
		}
		result := &PagedResult[NotificationTemplate]{
			Items:      items,
			TotalCount: len(items),
			Page:       1,
			TotalPages: 1,
		}
		return result, &Response{}, nil
	case *api.APIV4ContenttemplateGetUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// GetLog is not supported by the StatusCast API v4.
func (nc *NotificationsClient) GetLog(ctx context.Context, id string, opts ...RequestOption) (*NotificationLog, *Response, error) {
	return nil, nil, errors.New("not supported by StatusCast API v4")
}

// ListLogs is not supported by the StatusCast API v4.
func (nc *NotificationsClient) ListLogs(ctx context.Context, incidentID string, page Pagination, opts ...RequestOption) (*PagedResult[NotificationLog], *Response, error) {
	return nil, nil, errors.New("not supported by StatusCast API v4")
}

// ─── mapping helpers ─────────────────────────────────────────────────────────

func mapContentTemplatePostOK(r *api.APIV4ContenttemplatePostOK) *NotificationTemplate {
	return &NotificationTemplate{
		ID:      optInt32ToID(r.ID),
		Subject: optNilStringVal(r.Subject),
		Body:    optNilStringVal(r.Contents),
	}
}

func mapContentTemplatePutOK(r *api.APIV4ContenttemplatePutOK) *NotificationTemplate {
	return &NotificationTemplate{
		ID:      optInt32ToID(r.ID),
		Subject: optNilStringVal(r.Subject),
		Body:    optNilStringVal(r.Contents),
	}
}
