package statuscast

import (
	"context"
	"errors"
)

// List returns all teams. Teams are not supported by StatusCast API v4.
func (tc *TeamsClient) List(ctx context.Context, page Pagination, opts ...RequestOption) (*PagedResult[Team], *Response, error) {
	return nil, nil, errors.New("not supported by StatusCast API v4")
}
