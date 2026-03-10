package statuscast

import (
	"context"
	"errors"

	api "statuscast-go/internal/statuscast"
)

// Create opens a new incident and optionally notifies subscribers.
func (ic *IncidentsClient) Create(ctx context.Context, req CreateIncidentRequest, opts ...RequestOption) (*Incident, *Response, error) {
	incidentType := api.APIV4IncidentPostReqIncidentType(mapFacadePostType(req.PostType))
	postType := api.APIV4IncidentPostReqPostsItemPostType(mapFacadeIncidentStatus(req.Status))
	if req.Status == "" {
		postType = api.APIV4IncidentPostReqPostsItemPostTypeInvestigating
	}

	post := api.APIV4IncidentPostReqPostsItem{}
	post.Text.SetTo(req.Message)
	post.PostType = api.OptAPIV4IncidentPostReqPostsItemPostType{
		Value: postType,
		Set:   true,
	}
	post.IsPublished.SetTo(true)

	body := api.APIV4IncidentPostReq{
		IncidentType:      api.OptAPIV4IncidentPostReqIncidentType{Value: incidentType, Set: true},
		SendNotifications: api.OptBool{Value: req.Notify, Set: true},
	}
	body.Title.SetTo(req.Title)
	body.Posts = api.OptNilAPIV4IncidentPostReqPostsItemArray{
		Value: []api.APIV4IncidentPostReqPostsItem{post},
		Set:   true,
	}

	if len(req.Components) > 0 {
		ids := make([]int32, 0, len(req.Components))
		for _, cid := range req.Components {
			n, err := idToInt32(cid)
			if err != nil {
				return nil, nil, err
			}
			ids = append(ids, n)
		}
		body.AffectedComponents = api.OptNilInt32Array{Value: ids, Set: true}
	}

	res, err := ic.c.ogen.APIV4IncidentPost(ctx, api.OptAPIV4IncidentPostReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4IncidentPostOK:
		return mapIncidentPostOK(r), &Response{}, nil
	case *api.APIV4IncidentPostUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// Get retrieves a single incident with its full update timeline.
func (ic *IncidentsClient) Get(ctx context.Context, id string, opts ...RequestOption) (*Incident, *Response, error) {
	intID, err := idToInt32(id)
	if err != nil {
		return nil, nil, err
	}
	res, err := ic.c.ogen.APIV4IncidentIDGet(ctx, api.APIV4IncidentIDGetParams{ID: intID})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4IncidentIDGetOK:
		return mapIncidentIDGetOK(r), &Response{}, nil
	case *api.APIV4IncidentIDGetUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// List returns incidents, optionally filtered by active-only or time range.
func (ic *IncidentsClient) List(ctx context.Context, filter IncidentFilter, page Pagination, opts ...RequestOption) (*PagedResult[Incident], *Response, error) {
	body := api.APIV4IncidentsPostReq{}
	if page.Page > 0 {
		body.PageNumber.SetTo(int32(page.Page))
	}
	if page.PerPage > 0 {
		body.PageSize.SetTo(int32(page.PerPage))
	}
	if filter.ActiveOnly {
		body.IncidentStatusBy = api.OptAPIV4IncidentsPostReqIncidentStatusBy{
			Value: "OpenStarted",
			Set:   true,
		}
	}

	res, err := ic.c.ogen.APIV4IncidentsPost(ctx, api.OptAPIV4IncidentsPostReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4IncidentsPostOK:
		items := make([]Incident, 0)
		if r.Items.Set && !r.Items.Null {
			for _, item := range r.Items.Value {
				items = append(items, *mapIncidentsPostOKItem(&item))
			}
		}
		totalPages := 1
		if r.Pages.Set && r.Pages.Value > 0 {
			totalPages = int(r.Pages.Value)
		}
		result := &PagedResult[Incident]{
			Items:      items,
			TotalCount: int(r.TotalItems.Value),
			Page:       int(r.Page.Value),
			TotalPages: totalPages,
		}
		return result, &Response{}, nil
	case *api.APIV4IncidentsPostUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// PostUpdate appends a new message to an existing incident's timeline.
func (ic *IncidentsClient) PostUpdate(ctx context.Context, incidentID string, req UpdateIncidentRequest, opts ...RequestOption) (*IncidentUpdate, *Response, error) {
	incident, resp, err := ic.update(ctx, incidentID, req, opts...)
	if err != nil {
		return nil, nil, err
	}
	// Return the last update from the updated incident.
	if len(incident.Updates) > 0 {
		last := incident.Updates[len(incident.Updates)-1]
		return &last, resp, nil
	}
	return &IncidentUpdate{}, resp, nil
}

func (ic *IncidentsClient) update(ctx context.Context, id string, req UpdateIncidentRequest, opts ...RequestOption) (*Incident, *Response, error) {
	intID, err := idToInt32(id)
	if err != nil {
		return nil, nil, err
	}
	postType := api.APIV4IncidentPutReqPostsItemPostType(mapFacadeIncidentStatus(req.Status))

	post := api.APIV4IncidentPutReqPostsItem{}
	post.Text.SetTo(req.Message)
	post.PostType = api.OptAPIV4IncidentPutReqPostsItemPostType{
		Value: postType,
		Set:   true,
	}
	post.IsPublished.SetTo(true)

	body := api.APIV4IncidentPutReq{
		SendNotifications: api.OptBool{Value: req.Notify, Set: true},
	}
	body.ID.SetTo(intID)
	body.Posts = api.OptNilAPIV4IncidentPutReqPostsItemArray{
		Value: []api.APIV4IncidentPutReqPostsItem{post},
		Set:   true,
	}

	res, err := ic.c.ogen.APIV4IncidentPut(ctx, api.OptAPIV4IncidentPutReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4IncidentPutOK:
		return mapIncidentPutOK(r), &Response{}, nil
	case *api.APIV4IncidentPutUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// FileRCA is not supported by the StatusCast API v4.
func (ic *IncidentsClient) FileRCA(ctx context.Context, incidentID string, req RCARequest, opts ...RequestOption) (*RootCauseAnalysis, *Response, error) {
	return nil, nil, errors.New("not supported by StatusCast API v4")
}

// ─── mapping helpers ─────────────────────────────────────────────────────────

func mapIncidentPostOK(r *api.APIV4IncidentPostOK) *Incident {
	inc := &Incident{
		ID:        optInt32ToID(r.ID),
		Title:     optNilStringVal(r.Title),
		CreatedAt: r.DateCreated.Value,
	}
	if r.EndDate.Set && !r.EndDate.Null {
		t := r.EndDate.Value
		inc.ResolvedAt = &t
	}
	inc.Updates, inc.Status = mapIncidentPostOKPosts(r.Posts)
	inc.Components = mapIncidentPostOKAffectedComponents(r.AffectedComponents)
	return inc
}

func mapIncidentPostOKPosts(posts api.OptNilAPIV4IncidentPostOKPostsItemArray) ([]IncidentUpdate, IncidentStatus) {
	status := StatusInvestigating
	updates := []IncidentUpdate{}
	if !posts.Set || posts.Null {
		return updates, status
	}
	for _, p := range posts.Value {
		update := IncidentUpdate{
			Message:   optNilStringVal(p.Text),
			Status:    mapAPIIncidentStatus(string(p.PostType.Value)),
			CreatedAt: p.Date.Value,
		}
		if p.ID.Set && !p.ID.Null {
			update.ID = int32ToID(p.ID.Value)
		}
		updates = append(updates, update)
	}
	if len(updates) > 0 {
		status = updates[len(updates)-1].Status
	}
	return updates, status
}

func mapIncidentPostOKAffectedComponents(comps api.OptNilAPIV4IncidentPostOKAffectedComponentsItemArray) []string {
	if !comps.Set || comps.Null {
		return nil
	}
	ids := make([]string, 0, len(comps.Value))
	for _, c := range comps.Value {
		if c.ComponentId.Set && !c.ComponentId.Null {
			ids = append(ids, int32ToID(c.ComponentId.Value))
		}
	}
	return ids
}

func mapIncidentIDGetOK(r *api.APIV4IncidentIDGetOK) *Incident {
	inc := &Incident{
		ID:        optInt32ToID(r.ID),
		Title:     optNilStringVal(r.Title),
		CreatedAt: r.DateCreated.Value,
	}
	if r.EndDate.Set && !r.EndDate.Null {
		t := r.EndDate.Value
		inc.ResolvedAt = &t
	}
	// Map posts to updates
	if r.Posts.Set && !r.Posts.Null {
		for _, p := range r.Posts.Value {
			update := IncidentUpdate{
				Message:   optNilStringVal(p.Text),
				Status:    mapAPIIncidentStatus(string(p.PostType.Value)),
				CreatedAt: p.Date.Value,
			}
			if p.ID.Set && !p.ID.Null {
				update.ID = int32ToID(p.ID.Value)
			}
			inc.Updates = append(inc.Updates, update)
		}
		if len(inc.Updates) > 0 {
			inc.Status = inc.Updates[len(inc.Updates)-1].Status
		}
	}
	// Map affected components
	if r.AffectedComponents.Set && !r.AffectedComponents.Null {
		for _, c := range r.AffectedComponents.Value {
			if c.ComponentId.Set && !c.ComponentId.Null {
				inc.Components = append(inc.Components, int32ToID(c.ComponentId.Value))
			}
		}
	}
	return inc
}

func mapIncidentPutOK(r *api.APIV4IncidentPutOK) *Incident {
	inc := &Incident{
		ID:        optInt32ToID(r.ID),
		Title:     optNilStringVal(r.Title),
		CreatedAt: r.DateCreated.Value,
	}
	if r.EndDate.Set && !r.EndDate.Null {
		t := r.EndDate.Value
		inc.ResolvedAt = &t
	}
	if r.Posts.Set && !r.Posts.Null {
		for _, p := range r.Posts.Value {
			update := IncidentUpdate{
				Message:   optNilStringVal(p.Text),
				Status:    mapAPIIncidentStatus(string(p.PostType.Value)),
				CreatedAt: p.Date.Value,
			}
			if p.ID.Set && !p.ID.Null {
				update.ID = int32ToID(p.ID.Value)
			}
			inc.Updates = append(inc.Updates, update)
		}
		if len(inc.Updates) > 0 {
			inc.Status = inc.Updates[len(inc.Updates)-1].Status
		}
	}
	if r.AffectedComponents.Set && !r.AffectedComponents.Null {
		for _, c := range r.AffectedComponents.Value {
			if c.ComponentId.Set && !c.ComponentId.Null {
				inc.Components = append(inc.Components, int32ToID(c.ComponentId.Value))
			}
		}
	}
	return inc
}

func mapIncidentsPostOKItem(item *api.APIV4IncidentsPostOKItemsItem) *Incident {
	inc := &Incident{
		ID:        optInt32ToID(item.ID),
		Title:     optNilStringVal(item.Title),
		CreatedAt: item.DateCreated.Value,
	}
	if item.EndDate.Set && !item.EndDate.Null {
		t := item.EndDate.Value
		inc.ResolvedAt = &t
	}
	if item.Posts.Set && !item.Posts.Null {
		for _, p := range item.Posts.Value {
			update := IncidentUpdate{
				Message:   optNilStringVal(p.Text),
				Status:    mapAPIIncidentStatus(string(p.PostType.Value)),
				CreatedAt: p.Date.Value,
			}
			if p.ID.Set && !p.ID.Null {
				update.ID = int32ToID(p.ID.Value)
			}
			inc.Updates = append(inc.Updates, update)
		}
		if len(inc.Updates) > 0 {
			inc.Status = inc.Updates[len(inc.Updates)-1].Status
		}
	}
	if item.AffectedComponents.Set && !item.AffectedComponents.Null {
		for _, c := range item.AffectedComponents.Value {
			if c.ComponentId.Set && !c.ComponentId.Null {
				inc.Components = append(inc.Components, int32ToID(c.ComponentId.Value))
			}
		}
	}
	return inc
}
