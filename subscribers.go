package statuscast

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	api "statuscast-go/internal/statuscast"
)

// Add creates a new subscriber.
func (sc *SubscribersClient) Add(ctx context.Context, req AddSubscriberRequest, opts ...RequestOption) (*Subscriber, *Response, error) {
	body := api.APIV4SubscriberPostReq{}
	body.EmailAddress.SetTo(req.Email)
	if req.Phone != "" {
		body.PhoneNumber.SetTo(req.Phone)
	}

	useEmail := false
	useSMS := false
	useWebhook := false
	for _, ch := range req.Channels {
		switch ch {
		case ChannelEmail:
			useEmail = true
		case ChannelSMS:
			useSMS = true
		case ChannelWebhook:
			useWebhook = true
		}
	}
	if len(req.Channels) == 0 {
		useEmail = true
	}
	body.UseEmail.SetTo(useEmail)
	body.UseSMS.SetTo(useSMS)
	body.UseWebhook.SetTo(useWebhook)

	if len(req.Groups) > 0 {
		ids, err := stringsToInt32Slice(req.Groups)
		if err != nil {
			return nil, nil, err
		}
		body.Groups = api.OptNilInt32Array{Value: ids, Set: true}
	}
	if len(req.Components) > 0 {
		ids, err := stringsToInt32Slice(req.Components)
		if err != nil {
			return nil, nil, err
		}
		body.Components = api.OptNilInt32Array{Value: ids, Set: true}
	}

	res, err := sc.c.ogen.APIV4SubscriberPost(ctx, api.OptAPIV4SubscriberPostReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4SubscriberPostOK:
		return mapSubscriberPostOK(r), &Response{}, nil
	case *api.APIV4SubscriberPostUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// BulkImport imports subscribers from CSV data. The CSV must have an "email" header column.
// Per-row failures are accumulated in the result rather than returned as an error.
func (sc *SubscribersClient) BulkImport(ctx context.Context, csvData []byte, opts ...RequestOption) (*BulkImportResult, *Response, error) {
	r := csv.NewReader(bytes.NewReader(csvData))

	header, err := r.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("csv header: %w", err)
	}

	emailIdx := -1
	for i, h := range header {
		if strings.EqualFold(strings.TrimSpace(h), "email") {
			emailIdx = i
			break
		}
	}
	if emailIdx == -1 {
		return nil, nil, fmt.Errorf("csv missing 'email' column")
	}

	var result BulkImportResult
	for rowNum := 1; ; rowNum++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkImportError{Row: rowNum, Message: err.Error()})
			continue
		}

		email := strings.TrimSpace(record[emailIdx])
		if email == "" {
			result.Skipped++
			continue
		}

		body := api.APIV4SubscriberPostReq{}
		body.EmailAddress.SetTo(email)
		body.UseEmail.SetTo(true)

		res, apiErr := sc.c.ogen.APIV4SubscriberPost(ctx, api.OptAPIV4SubscriberPostReq{Set: true, Value: body})
		if apiErr != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkImportError{Row: rowNum, Email: email, Message: apiErr.Error()})
			continue
		}
		switch res.(type) {
		case *api.APIV4SubscriberPostOK:
			result.Imported++
		default:
			result.Failed++
			result.Errors = append(result.Errors, BulkImportError{Row: rowNum, Email: email, Message: "unexpected API response"})
		}
	}

	return &result, &Response{}, nil
}

// Get retrieves a subscriber by ID.
func (sc *SubscribersClient) Get(ctx context.Context, id string, opts ...RequestOption) (*Subscriber, *Response, error) {
	intID, err := idToInt32(id)
	if err != nil {
		return nil, nil, err
	}
	res, err := sc.c.ogen.APIV4SubscriberIDGet(ctx, api.APIV4SubscriberIDGetParams{ID: intID})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4SubscriberIDGetOK:
		return mapSubscriberIDGetOK(r), &Response{}, nil
	case *api.APIV4SubscriberIDGetUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// List returns all subscribers, optionally scoped to a group.
func (sc *SubscribersClient) List(ctx context.Context, groupID string, page Pagination, opts ...RequestOption) (*PagedResult[Subscriber], *Response, error) {
	body := api.APIV4SubscribersSearchPostReq{}
	if groupID != "" {
		gid, err := idToInt32(groupID)
		if err != nil {
			return nil, nil, err
		}
		body.GroupId.SetTo(gid)
	}

	res, err := sc.c.ogen.APIV4SubscribersSearchPost(ctx, api.OptAPIV4SubscribersSearchPostReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4SubscribersSearchPostOK:
		items := make([]Subscriber, 0)
		if r.Items.Set && !r.Items.Null {
			for _, item := range r.Items.Value {
				items = append(items, *mapSubscriberSearchItem(&item))
			}
		}
		totalPages := 1
		if r.Pages.Set && r.Pages.Value > 0 {
			totalPages = int(r.Pages.Value)
		}
		result := &PagedResult[Subscriber]{
			Items:      items,
			TotalCount: int(r.TotalItems.Value),
			Page:       int(r.Page.Value),
			TotalPages: totalPages,
		}
		return result, &Response{}, nil
	case *api.APIV4SubscribersSearchPostUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// Update patches a subscriber record.
func (sc *SubscribersClient) Update(ctx context.Context, id string, req UpdateSubscriberRequest, opts ...RequestOption) (*Subscriber, *Response, error) {
	intID, err := idToInt32(id)
	if err != nil {
		return nil, nil, err
	}
	body := api.APIV4SubscriberPutReq{}
	body.ID.SetTo(intID)

	if len(req.Groups) > 0 {
		ids, err := stringsToInt32Slice(req.Groups)
		if err != nil {
			return nil, nil, err
		}
		body.Groups = api.OptNilInt32Array{Value: ids, Set: true}
	}
	if len(req.Components) > 0 {
		ids, err := stringsToInt32Slice(req.Components)
		if err != nil {
			return nil, nil, err
		}
		body.Components = api.OptNilInt32Array{Value: ids, Set: true}
	}

	for _, ch := range req.Channels {
		switch ch {
		case ChannelEmail:
			body.UseEmail.SetTo(true)
		case ChannelSMS:
			body.UseSMS.SetTo(true)
		case ChannelWebhook:
			body.UseWebhook.SetTo(true)
		}
	}

	res, err := sc.c.ogen.APIV4SubscriberPut(ctx, api.OptAPIV4SubscriberPutReq{Set: true, Value: body})
	if err != nil {
		return nil, nil, err
	}
	switch r := res.(type) {
	case *api.APIV4SubscriberPutOK:
		return mapSubscriberPutOK(r), &Response{}, nil
	case *api.APIV4SubscriberPutUnauthorized:
		return nil, nil, ErrUnauthorized
	default:
		return nil, nil, unexpectedResponse(res)
	}
}

// Remove unsubscribes and deletes a subscriber by ID.
func (sc *SubscribersClient) Remove(ctx context.Context, id string, opts ...RequestOption) (*Response, error) {
	intID, err := idToInt32(id)
	if err != nil {
		return nil, err
	}
	res, err := sc.c.ogen.APIV4SubscriberIDDelete(ctx, api.APIV4SubscriberIDDeleteParams{ID: intID})
	if err != nil {
		return nil, err
	}
	switch res.(type) {
	case *api.APIV4SubscriberIDDeleteOK:
		return &Response{}, nil
	case *api.APIV4SubscriberIDDeleteUnauthorized:
		return nil, ErrUnauthorized
	default:
		return nil, unexpectedResponse(res)
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func stringsToInt32Slice(strs []string) ([]int32, error) {
	ids := make([]int32, 0, len(strs))
	for _, s := range strs {
		n, err := idToInt32(s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, n)
	}
	return ids, nil
}

func mapSubscriberChannels(useEmail, useSMS, useWebhook bool) []NotificationChannel {
	var channels []NotificationChannel
	if useEmail {
		channels = append(channels, ChannelEmail)
	}
	if useSMS {
		channels = append(channels, ChannelSMS)
	}
	if useWebhook {
		channels = append(channels, ChannelWebhook)
	}
	return channels
}

func mapSubscriberPostOK(r *api.APIV4SubscriberPostOK) *Subscriber {
	s := &Subscriber{
		ID:    optInt32ToID(r.ID),
		Email: optNilStringVal(r.EmailAddress),
		Phone: optNilStringVal(r.PhoneNumber),
	}
	if r.WhenCreated.Set && !r.WhenCreated.Null {
		s.CreatedAt = r.WhenCreated.Value
	}
	s.Channels = mapSubscriberChannels(
		r.SubscribeToIncidentPosts.Value,
		r.SmsSubscribeToIncidentPosts.Value,
		false,
	)
	return s
}

func mapSubscriberIDGetOK(r *api.APIV4SubscriberIDGetOK) *Subscriber {
	s := &Subscriber{
		ID:    optInt32ToID(r.ID),
		Email: optNilStringVal(r.EmailAddress),
		Phone: optNilStringVal(r.PhoneNumber),
	}
	if r.WhenCreated.Set && !r.WhenCreated.Null {
		s.CreatedAt = r.WhenCreated.Value
	}
	if r.Components.Set && !r.Components.Null {
		for _, c := range r.Components.Value {
			s.Components = append(s.Components, optInt32ToID(c.ID))
		}
	}
	if r.Groups.Set && !r.Groups.Null {
		for _, g := range r.Groups.Value {
			s.Groups = append(s.Groups, optInt32ToID(g.ID))
		}
	}
	return s
}

func mapSubscriberPutOK(r *api.APIV4SubscriberPutOK) *Subscriber {
	s := &Subscriber{
		ID:    optInt32ToID(r.ID),
		Email: optNilStringVal(r.EmailAddress),
		Phone: optNilStringVal(r.PhoneNumber),
	}
	if r.WhenCreated.Set && !r.WhenCreated.Null {
		s.CreatedAt = r.WhenCreated.Value
	}
	if r.Components.Set && !r.Components.Null {
		for _, c := range r.Components.Value {
			s.Components = append(s.Components, optInt32ToID(c.ID))
		}
	}
	if r.Groups.Set && !r.Groups.Null {
		for _, g := range r.Groups.Value {
			s.Groups = append(s.Groups, optInt32ToID(g.ID))
		}
	}
	return s
}

func mapSubscriberSearchItem(item *api.APIV4SubscribersSearchPostOKItemsItem) *Subscriber {
	s := &Subscriber{
		ID:    optInt32ToID(item.ID),
		Email: optNilStringVal(item.EmailAddress),
		Phone: optNilStringVal(item.PhoneNumber),
	}
	if item.WhenCreated.Set && !item.WhenCreated.Null {
		s.CreatedAt = item.WhenCreated.Value
	}
	if item.Components.Set && !item.Components.Null {
		for _, c := range item.Components.Value {
			s.Components = append(s.Components, optInt32ToID(c.ID))
		}
	}
	if item.Groups.Set && !item.Groups.Null {
		for _, g := range item.Groups.Value {
			s.Groups = append(s.Groups, optInt32ToID(g.ID))
		}
	}
	return s
}
