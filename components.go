package statuscast

import (
	"context"

	api "statuscast-go/internal/statuscast"
)

// Create adds a new component (or sub-component) to the status page.
func (cc *ComponentsClient) Create(ctx context.Context, req CreateComponentRequest, opts ...RequestOption) (*Component, *Response, error) {
	body := api.APIV4ComponentPostReq{}
	body.Name.SetTo(req.Name)
	if req.Description != "" {
		body.Description.SetTo(req.Description)
	}
	if req.ParentID != "" {
		pid, err := idToInt32(req.ParentID)
		if err != nil {
			return nil, nil, err
		}
		body.ParentId.SetTo(pid)
	}

	res, err := cc.c.ogen.APIV4ComponentPost(ctx, api.OptAPIV4ComponentPostReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4ComponentPostOK:
		return mapComponentPostOK(r), &Response{}, nil
	case *api.APIV4ComponentPostUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// Get retrieves a single component by ID.
func (cc *ComponentsClient) Get(ctx context.Context, id string, opts ...RequestOption) (*Component, *Response, error) {
	intID, err := idToInt32(id)
	if err != nil {
		return nil, nil, err
	}
	res, err := cc.c.ogen.APIV4ComponentIDGet(ctx, api.APIV4ComponentIDGetParams{ID: intID})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4ComponentIDGetOK:
		return mapComponentIDGetOK(r), &Response{}, nil
	case *api.APIV4ComponentIDGetUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// List returns all components, optionally filtered to direct children of parentID.
// Pass an empty parentID to list only root components.
func (cc *ComponentsClient) List(ctx context.Context, parentID string, page Pagination, opts ...RequestOption) (*PagedResult[Component], *Response, error) {
	res, err := cc.c.ogen.APIV4ComponentsGet(ctx)
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4ComponentsGetOKApplicationJSON:
		items := []Component{}
		for _, item := range *r {
			comp := mapComponentsGetOKItem(&item)
			if parentID != "" {
				if optInt32ToID(item.ParentId) != parentID {
					continue
				}
			}
			items = append(items, *comp)
		}
		result := &PagedResult[Component]{
			Items:      items,
			TotalCount: len(items),
			Page:       1,
			TotalPages: 1,
		}
		return result, &Response{}, nil
	case *api.APIV4ComponentsGetUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// Update applies a partial patch to a component.
func (cc *ComponentsClient) Update(ctx context.Context, id string, req UpdateComponentRequest, opts ...RequestOption) (*Component, *Response, error) {
	intID, err := idToInt32(id)
	if err != nil {
		return nil, nil, err
	}
	body := api.APIV4ComponentPutReq{}
	body.ID.SetTo(intID)
	if req.Name != nil {
		body.Name.SetTo(*req.Name)
	}
	if req.Description != nil {
		body.Description.SetTo(*req.Description)
	}

	res, err := cc.c.ogen.APIV4ComponentPut(ctx, api.OptAPIV4ComponentPutReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4ComponentPutOK:
		return mapComponentPutOK(r), &Response{}, nil
	case *api.APIV4ComponentPutUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// Delete removes a component. Deleting a parent automatically removes all
// sub-components; use with care.
func (cc *ComponentsClient) Delete(ctx context.Context, id string, opts ...RequestOption) (*Response, error) {
	intID, err := idToInt32(id)
	if err != nil {
		return nil, err
	}
	res, err := cc.c.ogen.APIV4ComponentDelete(ctx, api.APIV4ComponentDeleteParams{ID: intID})
	if err != nil {
		return nil, err
	}
	switch res.(type) {
	case *api.APIV4ComponentDeleteOK:
		return &Response{}, nil
	case *api.APIV4ComponentDeleteUnauthorized:
		return nil, ErrUnauthorized
	default:
		return nil, unexpectedResponse(res)
	}
}

// ─── mapping helpers ─────────────────────────────────────────────────────────

func mapComponentPostOK(r *api.APIV4ComponentPostOK) *Component {
	return &Component{
		ID:          optInt32ToID(r.ID),
		Name:        optNilStringVal(r.Name),
		Description: optNilStringVal(r.Description),
		Status:      mapAPIComponentStatus(string(r.Status.Value)),
		ParentID:    optInt32ToID(r.ParentId),
	}
}

func mapComponentIDGetOK(r *api.APIV4ComponentIDGetOK) *Component {
	return &Component{
		ID:          optInt32ToID(r.ID),
		Name:        optNilStringVal(r.Name),
		Description: optNilStringVal(r.Description),
		Status:      mapAPIComponentStatus(string(r.Status.Value)),
		ParentID:    optInt32ToID(r.ParentId),
	}
}

func mapComponentPutOK(r *api.APIV4ComponentPutOK) *Component {
	return &Component{
		ID:          optInt32ToID(r.ID),
		Name:        optNilStringVal(r.Name),
		Description: optNilStringVal(r.Description),
		Status:      mapAPIComponentStatus(string(r.Status.Value)),
		ParentID:    optInt32ToID(r.ParentId),
	}
}

func mapComponentsGetOKItem(item *api.APIV4ComponentsGetOKItem) *Component {
	return &Component{
		ID:          optInt32ToID(item.ID),
		Name:        optNilStringVal(item.Name),
		Description: optNilStringVal(item.Description),
		Status:      mapAPIComponentStatus(string(item.Status.Value)),
		ParentID:    optInt32ToID(item.ParentId),
	}
}

