package statuscast

import (
	"fmt"
	"strconv"

	api "statuscast-go/internal/statuscast"
)

// idToInt32 converts a string ID to int32 or returns an APIError.
func idToInt32(id string) (int32, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return 0, &APIError{Message: fmt.Sprintf("invalid ID %q: %v", id, err)}
	}
	return int32(n), nil
}

// int32ToID converts an int32 ID to string.
func int32ToID(id int32) string {
	return strconv.Itoa(int(id))
}

// optInt32ToID converts OptInt32 to string ID; returns "" if not set.
func optInt32ToID(id api.OptInt32) string {
	if id.Set {
		return int32ToID(id.Value)
	}
	return ""
}

// optNilStringVal returns the string value of an OptNilString, or "" if unset/null.
func optNilStringVal(s api.OptNilString) string {
	return s.Value
}

// mapAPIComponentStatus maps the API component status string to a ComponentStatus.
func mapAPIComponentStatus(s string) ComponentStatus {
	switch s {
	case "Available":
		return StatusOperational
	case "DegradedPerformance":
		return StatusDegradedPerf
	case "Unavailable":
		return StatusMajorOutage
	case "Maintenance":
		return StatusUnderMaintenance
	default:
		return StatusOperational
	}
}

// mapFacadeComponentStatus maps a ComponentStatus to the API status string.
func mapFacadeComponentStatus(s ComponentStatus) string {
	switch s {
	case StatusOperational:
		return "Available"
	case StatusDegradedPerf:
		return "DegradedPerformance"
	case StatusPartialOutage:
		return "Unavailable"
	case StatusMajorOutage:
		return "Unavailable"
	case StatusUnderMaintenance:
		return "Maintenance"
	default:
		return "Available"
	}
}

// mapAPIIncidentStatus maps the API post type string to an IncidentStatus.
func mapAPIIncidentStatus(s string) IncidentStatus {
	switch s {
	case "Investigating":
		return StatusInvestigating
	case "Identified":
		return StatusIdentified
	case "Monitoring":
		return StatusMonitoring
	case "Closed":
		return StatusResolved
	default:
		return StatusInvestigating
	}
}

// mapFacadeIncidentStatus maps an IncidentStatus to the API post type string.
func mapFacadeIncidentStatus(s IncidentStatus) string {
	switch s {
	case StatusInvestigating:
		return "Investigating"
	case StatusIdentified:
		return "Identified"
	case StatusMonitoring:
		return "Monitoring"
	case StatusResolved:
		return "Closed"
	default:
		return "Investigating"
	}
}

// mapFacadePostType maps an IncidentPostType to the API incident type string.
func mapFacadePostType(p IncidentPostType) string {
	switch p {
	case PostTypeMaintenance:
		return "ScheduledMaintenance"
	case PostTypeInfo:
		return "Informational"
	default: // PostTypeOutage
		return "ServiceUnavailable"
	}
}

// mapFacadeRole maps a Role to the API user role string.
func mapFacadeRole(r Role) string {
	switch r {
	case RoleManager:
		return "Manager"
	case RoleAdministrator, RoleCompanyAdministrator:
		return "Administrator"
	default: // RoleEmployee
		return "Employee"
	}
}

// unexpectedResponse returns an APIError for unexpected response types.
func unexpectedResponse(res any) error {
	return &APIError{Message: fmt.Sprintf("unexpected response type: %T", res)}
}
