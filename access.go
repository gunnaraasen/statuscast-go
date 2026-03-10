package statuscast

import (
	"context"
	"errors"

	"github.com/google/uuid"

	api "statuscast-go/internal/statuscast"
)

// InviteUser sends an invitation to a new team member.
func (ac *AccessClient) InviteUser(ctx context.Context, req InviteUserRequest, opts ...RequestOption) (*AdminUser, *Response, error) {
	role := api.APIV4UserPostReqRole(mapFacadeRole(req.Role))
	body := api.APIV4UserPostReq{
		UserName: req.Email,
		FullName: req.Name,
		Role:     api.OptAPIV4UserPostReqRole{Value: role, Set: true},
	}

	res, err := ac.c.ogen.APIV4UserPost(ctx, api.OptAPIV4UserPostReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4UserPostOK:
		return mapAdminUserPostOK(r), &Response{}, nil
	case *api.APIV4UserPostUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// UpdateRole changes a user's permission level.
func (ac *AccessClient) UpdateRole(ctx context.Context, userID string, role Role, opts ...RequestOption) (*AdminUser, *Response, error) {
	intID, err := idToInt32(userID)
	if err != nil {
		return nil, nil, err
	}
	apiRole := api.APIV4UserPutReqRole(mapFacadeRole(role))
	body := api.APIV4UserPutReq{
		Role: api.OptAPIV4UserPutReqRole{Value: apiRole, Set: true},
	}
	body.ID.SetTo(intID)

	res, err := ac.c.ogen.APIV4UserPut(ctx, api.OptAPIV4UserPutReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4UserPutOK:
		return mapAdminUserPutOK(r), &Response{}, nil
	case *api.APIV4UserPutUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// RemoveUser revokes a team member's access.
// The userID must be a UUID string (as returned in the AdminUser.ID field from InviteUser or ListUsers).
func (ac *AccessClient) RemoveUser(ctx context.Context, userID string, opts ...RequestOption) (*Response, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, &APIError{Message: "RemoveUser requires a UUID user ID (use the ID returned by InviteUser or ListUsers)"}
	}
	res, err := ac.c.ogen.APIV4UserDelete(ctx, api.APIV4UserDeleteParams{ID: uid, ReassignId: uuid.UUID{}})
	if err != nil {
		return nil, err
	}
	switch res.(type) {
	case *api.APIV4UserDeleteOK:
		return &Response{}, nil
	case *api.APIV4UserDeleteUnauthorized:
		return nil, ErrUnauthorized
	default:
		return nil, unexpectedResponse(res)
	}
}

// ListUsers returns all admin users.
func (ac *AccessClient) ListUsers(ctx context.Context, page Pagination, opts ...RequestOption) (*PagedResult[AdminUser], *Response, error) {
	body := api.APIV4UsersPostReq{}
	if page.Page > 0 {
		body.PageNumber.SetTo(int32(page.Page))
	}
	if page.PerPage > 0 {
		body.PageSize.SetTo(int32(page.PerPage))
	}

	res, err := ac.c.ogen.APIV4UsersPost(ctx, api.OptAPIV4UsersPostReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4UsersPostOK:
		items := make([]AdminUser, 0)
		if r.Items.Set && !r.Items.Null {
			for _, item := range r.Items.Value {
				items = append(items, *mapAdminUserListItem(&item))
			}
		}
		totalPages := 1
		if r.Pages.Set && r.Pages.Value > 0 {
			totalPages = int(r.Pages.Value)
		}
		result := &PagedResult[AdminUser]{
			Items:      items,
			TotalCount: int(r.TotalItems.Value),
			Page:       int(r.Page.Value),
			TotalPages: totalPages,
		}
		return result, &Response{}, nil
	case *api.APIV4UsersPostUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// SetPageVisibility is not supported by the StatusCast API v4.
func (ac *AccessClient) SetPageVisibility(ctx context.Context, visibility PageVisibility, opts ...RequestOption) (*Response, error) {
	return nil, errors.New("not supported by StatusCast API v4")
}

// ─── mapping helpers ─────────────────────────────────────────────────────────

func mapAdminUserPostOK(r *api.APIV4UserPostOK) *AdminUser {
	u := &AdminUser{
		Email: optNilStringVal(r.Email),
		Name:  optNilStringVal(r.FullName),
	}
	// Use UUID as the ID when available (needed for RemoveUser).
	if r.UserId.Set {
		u.ID = r.UserId.Value.String()
	} else {
		u.ID = optInt32ToID(r.ID)
	}
	return u
}

func mapAdminUserPutOK(r *api.APIV4UserPutOK) *AdminUser {
	u := &AdminUser{
		Email: optNilStringVal(r.Email),
		Name:  optNilStringVal(r.FullName),
	}
	if r.UserId.Set {
		u.ID = r.UserId.Value.String()
	} else {
		u.ID = optInt32ToID(r.ID)
	}
	return u
}

func mapAdminUserListItem(item *api.APIV4UsersPostOKItemsItem) *AdminUser {
	u := &AdminUser{
		Email: optNilStringVal(item.Email),
		Name:  optNilStringVal(item.FullName),
	}
	if item.UserId.Set {
		u.ID = item.UserId.Value.String()
	} else {
		u.ID = optInt32ToID(item.ID)
	}
	return u
}
