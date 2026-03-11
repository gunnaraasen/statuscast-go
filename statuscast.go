// Package statuscast provides a Go SDK for managing StatusCast status pages,
// incidents, components, subscribers, and notifications.
//
// # Quick Start
//
//	client, err := statuscast.New(
//		statuscast.WithAPIKey("your-api-key"),
//		statuscast.WithBaseURL("https://app.statuscast.com"), // optional override
//	)
//
//	// Create and immediately post an incident
//	incident, err := client.Incidents.Create(ctx, statuscast.CreateIncidentRequest{
//		Title:      "Database connectivity degraded",
//		Components: []string{"comp_db_primary", "comp_db_replica"},
//		Status:     statuscast.StatusInvestigating,
//		Notify:     true,
//	})
package statuscast

import (
	"context"
	"net/http"
	"time"

	api "statuscast-go/internal/statuscast"
)

// ─── Client ──────────────────────────────────────────────────────────────────

// Client is the top-level StatusCast SDK client. All domain operations are
// available as fields so callers can import only what they need and IDE
// auto-complete surfaces all capabilities naturally.
//
//	client.Incidents.Create(...)
//	client.Components.Update(...)
//	client.Subscribers.Add(...)
type Client struct {
	Components    *ComponentsClient
	Incidents     *IncidentsClient
	Subscribers   *SubscribersClient
	Groups        *GroupsClient
	Notifications *NotificationsClient
	Reports       *ReportsClient
	Access        *AccessClient

	ogen    *api.Client
	http    *http.Client
	baseURL string
	apiKey  string
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithAPIKey sets the API key used to authenticate requests.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = key }
}

// WithBaseURL overrides the default API base URL. Useful for on-premise
// deployments or testing against a staging environment.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient replaces the default HTTP client. Use this to set custom
// timeouts, proxy settings, or transport-level retry logic.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.http = h }
}

// New creates a configured Client. Returns an error if required options
// (e.g. API key) are missing.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		baseURL: "https://app.statuscast.com",
		http:    &http.Client{Timeout: 30 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	if c.apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	bs := bearerSource{key: c.apiKey}
	ogenClient, err := api.NewClient(c.baseURL, bs, api.WithClient(c.http))
	if err != nil {
		return nil, err
	}
	c.ogen = ogenClient
	c.Components = &ComponentsClient{c}
	c.Incidents = &IncidentsClient{c}
	c.Subscribers = &SubscribersClient{c}
	c.Groups = &GroupsClient{c}
	c.Notifications = &NotificationsClient{c}
	c.Reports = &ReportsClient{c}
	c.Access = &AccessClient{c}
	return c, nil
}

// bearerSource implements api.SecuritySource using a static API key.
type bearerSource struct{ key string }

func (b bearerSource) Bearer(_ context.Context, _ api.OperationName) (api.Bearer, error) {
	return api.Bearer{Token: b.key}, nil
}

// ─── Shared Types ─────────────────────────────────────────────────────────────

// Response wraps the standard http.Response to expose rate limits and metadata.
type Response struct {
	*http.Response
	RequestID string
}

// ComponentStatus represents the operational state of a component.
type ComponentStatus string

const (
	StatusOperational      ComponentStatus = "operational"
	StatusDegradedPerf     ComponentStatus = "degraded_performance"
	StatusPartialOutage    ComponentStatus = "partial_outage"
	StatusMajorOutage      ComponentStatus = "major_outage"
	StatusUnderMaintenance ComponentStatus = "under_maintenance"
)

// IncidentStatus represents the lifecycle stage of an incident.
type IncidentStatus string

const (
	StatusInvestigating IncidentStatus = "investigating"
	StatusIdentified    IncidentStatus = "identified"
	StatusMonitoring    IncidentStatus = "monitoring"
	StatusResolved      IncidentStatus = "resolved"
)

// IncidentPostType categorizes the post visually.
type IncidentPostType string

const (
	PostTypeOutage      IncidentPostType = "outage"
	PostTypeMaintenance IncidentPostType = "maintenance"
	PostTypeInfo        IncidentPostType = "info"
)

// NotificationChannel controls which delivery channels a notification uses.
type NotificationChannel string

const (
	ChannelEmail   NotificationChannel = "email"
	ChannelSMS     NotificationChannel = "sms"
	ChannelSlack   NotificationChannel = "slack"
	ChannelTeams   NotificationChannel = "teams"
	ChannelWebhook NotificationChannel = "webhook"
)

// ComponentType describes the monitoring source of a component.
type ComponentType string

const (
	ComponentTypeNative  ComponentType = "native"
	ComponentTypeBeacon  ComponentType = "beacon"
	ComponentTypeThirdPt ComponentType = "third_party"
)

// Pagination holds paging parameters for list operations.
type Pagination struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}

// PagedResult wraps list responses with cursor metadata.
type PagedResult[T any] struct {
	Items      []T `json:"items"`
	TotalCount int `json:"total_count"`
	Page       int `json:"page"`
	TotalPages int `json:"total_pages"`
}

// RequestOption allows for per-request customization like custom headers.
type RequestOption func(*http.Request)

// ─── Components ───────────────────────────────────────────────────────────────

// Component represents a service or infrastructure asset tracked on the status page.
type Component struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      ComponentStatus `json:"status"`
	Type        ComponentType   `json:"type"`
	ParentID    string          `json:"parent_id"` // empty if root component
	Children    []string        `json:"children"`  // IDs of immediate sub-components
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// CreateComponentRequest defines inputs for creating a new component.
type CreateComponentRequest struct {
	Name          string          `json:"name"`
	Description   string          `json:"description,omitempty"`
	Type          ComponentType   `json:"type,omitempty"`
	ParentID      string          `json:"parent_id,omitempty"` // omit to create a root component
	InitialStatus ComponentStatus `json:"initial_status,omitempty"`
}

// UpdateComponentRequest defines fields that may be changed on an existing component.
// Zero values are ignored — only supplied fields are patched.
type UpdateComponentRequest struct {
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Status      *ComponentStatus `json:"status,omitempty"`
}

// ComponentsClient groups all component management operations.
type ComponentsClient struct{ c *Client }

// SetStatus is a convenience wrapper around Update for the common case of
// changing a component's operational status without touching other fields.
func (cc *ComponentsClient) SetStatus(ctx context.Context, id string, status ComponentStatus, opts ...RequestOption) (*Component, *Response, error) {
	return cc.Update(ctx, id, UpdateComponentRequest{Status: &status}, opts...)
}

// ─── Incidents ────────────────────────────────────────────────────────────────

// Incident represents a service disruption event with a full audit trail.
type Incident struct {
	ID             string             `json:"id"`
	Title          string             `json:"title"`
	Status         IncidentStatus     `json:"status"`
	PostType       IncidentPostType   `json:"post_type"`
	Components     []string           `json:"components"` // component IDs affected
	CustomFields   map[string]any     `json:"custom_fields,omitempty"`
	Updates        []IncidentUpdate   `json:"updates"`
	ScheduledStart *time.Time         `json:"scheduled_start,omitempty"`
	ScheduledEnd   *time.Time         `json:"scheduled_end,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	ResolvedAt     *time.Time         `json:"resolved_at,omitempty"` // nil if unresolved
}

// IncidentUpdate is a single timestamped message posted to an incident timeline.
type IncidentUpdate struct {
	ID        string         `json:"id"`
	Message   string         `json:"message"`
	Status    IncidentStatus `json:"status"`
	PostedBy  string         `json:"posted_by"`
	CreatedAt time.Time      `json:"created_at"`
}

// CreateIncidentRequest defines inputs for opening a new incident.
type CreateIncidentRequest struct {
	Title          string                `json:"title"`
	Message        string                `json:"message"`          // initial update body; supports Markdown
	Components     []string              `json:"components"`       // component IDs to affect
	Status         IncidentStatus        `json:"status,omitempty"` // defaults to StatusInvestigating
	PostType       IncidentPostType      `json:"post_type,omitempty"`
	TemplateID     string                `json:"template_id,omitempty"` // optional: pre-populate from a saved template
	CustomFields   map[string]any        `json:"custom_fields,omitempty"`
	ScheduledStart *time.Time            `json:"scheduled_start,omitempty"`
	ScheduledEnd   *time.Time            `json:"scheduled_end,omitempty"`
	Notify         bool                  `json:"notify"`             // send subscriber notifications immediately
	Channels       []NotificationChannel `json:"channels,omitempty"` // override default channels for this incident
}

// UpdateIncidentRequest posts a new timeline update and/or advances the status.
// Either Message or Status (or both) must be set.
type UpdateIncidentRequest struct {
	Message      string         `json:"message,omitempty"` // new update body; supports Markdown
	Status       IncidentStatus `json:"status,omitempty"`  // set to advance the lifecycle stage
	Notify       bool           `json:"notify"`            // send subscriber notifications for this update
	CustomFields map[string]any `json:"custom_fields,omitempty"`
}

// ResolveRequest closes an incident with a final message.
type ResolveRequest struct {
	Message string `json:"message"` // resolution summary shown to subscribers
	Notify  bool   `json:"notify"`
}

// IncidentsClient groups all incident lifecycle operations.
type IncidentsClient struct{ c *Client }

// IncidentFilter narrows list results. All fields are optional.
type IncidentFilter struct {
	ActiveOnly   bool       `url:"active_only,omitempty"`
	ComponentIDs []string   `url:"component_ids,omitempty"`
	Since        *time.Time `url:"since,omitempty"`
	Until        *time.Time `url:"until,omitempty"`
}

// Resolve is a convenience wrapper that advances an incident to StatusResolved
// and optionally notifies all subscribers in a single call.
func (ic *IncidentsClient) Resolve(ctx context.Context, incidentID string, req ResolveRequest, opts ...RequestOption) (*Incident, *Response, error) {
	return ic.update(ctx, incidentID, UpdateIncidentRequest{
		Message: req.Message,
		Status:  StatusResolved,
		Notify:  req.Notify,
	}, opts...)
}

// ─── Subscribers ─────────────────────────────────────────────────────────────

// Subscriber is a user who receives status page notifications.
type Subscriber struct {
	ID         string                `json:"id"`
	Email      string                `json:"email"`
	Phone      string                `json:"phone,omitempty"`      // E.164 format, e.g. "+15551234567"
	Groups     []string              `json:"groups,omitempty"`     // group IDs
	Components []string              `json:"components,omitempty"` // component IDs subscribed to; empty = all
	Channels   []NotificationChannel `json:"channels"`
	CreatedAt  time.Time             `json:"created_at"`
}

// AddSubscriberRequest adds a single subscriber.
type AddSubscriberRequest struct {
	Email      string                `json:"email"`
	Phone      string                `json:"phone,omitempty"`      // optional, required if SMS channel enabled
	Groups     []string              `json:"groups,omitempty"`     // assign to groups immediately
	Components []string              `json:"components,omitempty"` // omit to subscribe to all components
	Channels   []NotificationChannel `json:"channels,omitempty"`   // defaults to email only
}

// BulkImportResult summarises a bulk subscriber import operation.
type BulkImportResult struct {
	Imported int               `json:"imported"`
	Skipped  int               `json:"skipped"` // duplicates
	Failed   int               `json:"failed"`
	Errors   []BulkImportError `json:"errors,omitempty"`
}

// BulkImportError describes a single row failure during bulk import.
type BulkImportError struct {
	Row     int    `json:"row"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// SubscribersClient groups all subscriber management operations.
type SubscribersClient struct{ c *Client }

// UpdateSubscriberRequest changes a subscriber's groups, components, or channels.
type UpdateSubscriberRequest struct {
	Groups     []string              `json:"groups,omitempty"`
	Components []string              `json:"components,omitempty"`
	Channels   []NotificationChannel `json:"channels,omitempty"`
}

// ─── Notifications ────────────────────────────────────────────────────────────

// NotificationTemplate defines branded email / channel templates.
type NotificationTemplate struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Channel   NotificationChannel `json:"channel"`
	Subject   string              `json:"subject,omitempty"` // for email templates; may include macros e.g. {{.IncidentTitle}}
	Body      string              `json:"body"`              // HTML or Markdown body
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// NotificationsClient groups notification template and audit log operations.
type NotificationsClient struct{ c *Client }

// ─── Reports ─────────────────────────────────────────────────────────────────

// UptimeReport holds aggregate uptime percentages over a time window.
type UptimeReport struct {
	ComponentID   string    `json:"component_id"`
	ComponentName string    `json:"component_name"`
	Uptime        float64   `json:"uptime"` // percentage, e.g. 99.97
	WindowStart   time.Time `json:"window_start"`
	WindowEnd     time.Time `json:"window_end"`
}

// IncidentSummaryReport provides MTTD/MTTR analytics across incidents.
type IncidentSummaryReport struct {
	TotalIncidents    int            `json:"total_incidents"`
	MeanTimeToDetect  time.Duration  `json:"mean_time_to_detect"`
	MeanTimeToResolve time.Duration  `json:"mean_time_to_resolve"`
	ByComponent       map[string]int `json:"by_component"` // componentID → incident count
	Since             time.Time      `json:"since"`
	Until             time.Time      `json:"until"`
}

// ReportsClient exposes uptime, MTTR, and RCA reporting.
type ReportsClient struct{ c *Client }

// ─── Access Control ───────────────────────────────────────────────────────────

// Role represents the permission level of an administrative user.
type Role string

const (
	RoleEmployee             Role = "employee"
	RoleManager              Role = "manager"
	RoleAdministrator        Role = "administrator"
	RoleCompanyAdministrator Role = "company_administrator"
)

// AdminUser is an internal team member with dashboard access.
type AdminUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// InviteUserRequest sends an invitation email to a new admin user.
type InviteUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  Role   `json:"role"`
}

// AccessClient manages admin users, roles, and page visibility.
type AccessClient struct{ c *Client }

// ─── Groups ───────────────────────────────────────────────────────────────────

// Group represents a logical grouping of components or users.
type Group struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GroupsClient groups all group management operations.
type GroupsClient struct{ c *Client }

// ─── Errors ───────────────────────────────────────────────────────────────────

// APIError is returned when the StatusCast API responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Code       string // machine-readable error code from the API
	Message    string // human-readable description
}

func (e *APIError) Error() string {
	return e.Message
}

// Sentinel errors for common failure modes. Callers can use errors.Is for
// clean, type-safe error handling.
var (
	ErrMissingAPIKey  = &APIError{Code: "missing_api_key", Message: "an API key is required; use WithAPIKey()"}
	ErrNotFound       = &APIError{StatusCode: 404, Code: "not_found", Message: "resource not found"}
	ErrUnauthorized   = &APIError{StatusCode: 401, Code: "unauthorized", Message: "invalid or expired API key"}
	ErrRateLimited    = &APIError{StatusCode: 429, Code: "rate_limited", Message: "too many requests; back off and retry"}
	ErrIncidentClosed = &APIError{Code: "incident_closed", Message: "cannot update a resolved incident"}
)
