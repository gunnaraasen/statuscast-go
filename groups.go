package statuscast

import (
	"context"

	api "statuscast-go/internal/statuscast"
)

// List returns all groups.
func (gc *GroupsClient) List(ctx context.Context, page Pagination, opts ...RequestOption) (*PagedResult[Group], *Response, error) {
	body := api.APIV4GroupsPostReq{}
	if page.Page > 0 {
		body.PageNumber.SetTo(int32(page.Page))
	}
	if page.PerPage > 0 {
		body.PageSize.SetTo(int32(page.PerPage))
	}

	res, err := gc.c.ogen.APIV4GroupsPost(ctx, api.OptAPIV4GroupsPostReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4GroupsPostOK:
		items := make([]Group, 0)
		if r.Items.Set && !r.Items.Null {
			for _, item := range r.Items.Value {
				items = append(items, Group{
					ID:        optInt32ToID(item.ID),
					Name:      optNilStringVal(item.Name),
					CreatedAt: item.DateCreated.Value,
				})
			}
		}
		totalPages := 1
		if r.Pages.Set && r.Pages.Value > 0 {
			totalPages = int(r.Pages.Value)
		}
		result := &PagedResult[Group]{
			Items:      items,
			TotalCount: int(r.TotalItems.Value),
			Page:       int(r.Page.Value),
			TotalPages: totalPages,
		}
		return result, &Response{}, nil
	case *api.APIV4GroupsPostUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}
