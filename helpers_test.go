package statuscast

import (
	"testing"
)

// ─── idToInt32 ────────────────────────────────────────────────────────────────

func TestIDToInt32_Valid(t *testing.T) {
	cases := []struct {
		input    string
		expected int32
	}{
		{"0", 0},
		{"1", 1},
		{"42", 42},
		{"-1", -1},
		{"2147483647", 2147483647},  // math.MaxInt32
		{"-2147483648", -2147483648}, // math.MinInt32
	}
	for _, tc := range cases {
		got, err := idToInt32(tc.input)
		if err != nil {
			t.Errorf("idToInt32(%q): unexpected error: %v", tc.input, err)
		}
		if got != tc.expected {
			t.Errorf("idToInt32(%q) = %d; want %d", tc.input, got, tc.expected)
		}
	}
}

func TestIDToInt32_Invalid(t *testing.T) {
	cases := []string{"", "abc", "1.5", "0x1F", "  42"}
	for _, input := range cases {
		_, err := idToInt32(input)
		if err == nil {
			t.Errorf("idToInt32(%q): expected error, got nil", input)
		}
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Errorf("idToInt32(%q): expected *APIError, got %T", input, err)
		}
		if apiErr.Message == "" {
			t.Errorf("idToInt32(%q): APIError.Message should not be empty", input)
		}
	}
}

// ─── int32ToID ────────────────────────────────────────────────────────────────

func TestInt32ToID(t *testing.T) {
	cases := []struct {
		input    int32
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{-1, "-1"},
		{2147483647, "2147483647"},
	}
	for _, tc := range cases {
		got := int32ToID(tc.input)
		if got != tc.expected {
			t.Errorf("int32ToID(%d) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

// ─── mapAPIComponentStatus ────────────────────────────────────────────────────

func TestMapAPIComponentStatus(t *testing.T) {
	cases := []struct {
		input    string
		expected ComponentStatus
	}{
		{"Available", StatusOperational},
		{"DegradedPerformance", StatusDegradedPerf},
		{"Unavailable", StatusMajorOutage},
		{"Maintenance", StatusUnderMaintenance},
		{"Unknown", StatusOperational},  // default
		{"", StatusOperational},          // default
		{"informational", StatusOperational}, // case-sensitive default
	}
	for _, tc := range cases {
		got := mapAPIComponentStatus(tc.input)
		if got != tc.expected {
			t.Errorf("mapAPIComponentStatus(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

// ─── mapFacadeComponentStatus ─────────────────────────────────────────────────

func TestMapFacadeComponentStatus(t *testing.T) {
	cases := []struct {
		input    ComponentStatus
		expected string
	}{
		{StatusOperational, "Available"},
		{StatusDegradedPerf, "DegradedPerformance"},
		{StatusPartialOutage, "Unavailable"},
		{StatusMajorOutage, "Unavailable"},
		{StatusUnderMaintenance, "Maintenance"},
		{"", "Available"},             // default
		{"unknown_status", "Available"}, // default
	}
	for _, tc := range cases {
		got := mapFacadeComponentStatus(tc.input)
		if got != tc.expected {
			t.Errorf("mapFacadeComponentStatus(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

// Both StatusPartialOutage and StatusMajorOutage map to "Unavailable".
func TestMapFacadeComponentStatus_OutageVariants(t *testing.T) {
	if got := mapFacadeComponentStatus(StatusPartialOutage); got != "Unavailable" {
		t.Errorf("StatusPartialOutage mapped to %q; want Unavailable", got)
	}
	if got := mapFacadeComponentStatus(StatusMajorOutage); got != "Unavailable" {
		t.Errorf("StatusMajorOutage mapped to %q; want Unavailable", got)
	}
}

// ─── mapAPIIncidentStatus ─────────────────────────────────────────────────────

func TestMapAPIIncidentStatus(t *testing.T) {
	cases := []struct {
		input    string
		expected IncidentStatus
	}{
		{"Investigating", StatusInvestigating},
		{"Identified", StatusIdentified},
		{"Monitoring", StatusMonitoring},
		{"Closed", StatusResolved},
		{"Unknown", StatusInvestigating}, // default
		{"", StatusInvestigating},         // default
	}
	for _, tc := range cases {
		got := mapAPIIncidentStatus(tc.input)
		if got != tc.expected {
			t.Errorf("mapAPIIncidentStatus(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

// ─── mapFacadeIncidentStatus ──────────────────────────────────────────────────

func TestMapFacadeIncidentStatus(t *testing.T) {
	cases := []struct {
		input    IncidentStatus
		expected string
	}{
		{StatusInvestigating, "Investigating"},
		{StatusIdentified, "Identified"},
		{StatusMonitoring, "Monitoring"},
		{StatusResolved, "Closed"},
		{"", "Investigating"},      // default
		{"other", "Investigating"}, // default
	}
	for _, tc := range cases {
		got := mapFacadeIncidentStatus(tc.input)
		if got != tc.expected {
			t.Errorf("mapFacadeIncidentStatus(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

// ─── mapFacadePostType ────────────────────────────────────────────────────────

func TestMapFacadePostType(t *testing.T) {
	cases := []struct {
		input    IncidentPostType
		expected string
	}{
		{PostTypeOutage, "ServiceUnavailable"},
		{PostTypeMaintenance, "ScheduledMaintenance"},
		{PostTypeInfo, "Informational"},
		{"", "ServiceUnavailable"},      // default (outage)
		{"other", "ServiceUnavailable"}, // default
	}
	for _, tc := range cases {
		got := mapFacadePostType(tc.input)
		if got != tc.expected {
			t.Errorf("mapFacadePostType(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

// ─── mapFacadeRole ────────────────────────────────────────────────────────────

func TestMapFacadeRole(t *testing.T) {
	cases := []struct {
		input    Role
		expected string
	}{
		{RoleEmployee, "Employee"},
		{RoleManager, "Manager"},
		{RoleAdministrator, "Administrator"},
		{RoleCompanyAdministrator, "Administrator"},
		{"", "Employee"},      // default
		{"other", "Employee"}, // default
	}
	for _, tc := range cases {
		got := mapFacadeRole(tc.input)
		if got != tc.expected {
			t.Errorf("mapFacadeRole(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

// Both administrator roles map to "Administrator".
func TestMapFacadeRole_AdministratorVariants(t *testing.T) {
	if got := mapFacadeRole(RoleAdministrator); got != "Administrator" {
		t.Errorf("RoleAdministrator mapped to %q; want Administrator", got)
	}
	if got := mapFacadeRole(RoleCompanyAdministrator); got != "Administrator" {
		t.Errorf("RoleCompanyAdministrator mapped to %q; want Administrator", got)
	}
}

// ─── stringsToInt32Slice ──────────────────────────────────────────────────────

func TestStringsToInt32Slice_Valid(t *testing.T) {
	got, err := stringsToInt32Slice([]string{"1", "2", "3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []int32{1, 2, 3}
	if len(got) != len(expected) {
		t.Fatalf("len = %d; want %d", len(got), len(expected))
	}
	for i, v := range got {
		if v != expected[i] {
			t.Errorf("[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

func TestStringsToInt32Slice_Empty(t *testing.T) {
	got, err := stringsToInt32Slice([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestStringsToInt32Slice_InvalidEntry(t *testing.T) {
	_, err := stringsToInt32Slice([]string{"1", "not-a-number", "3"})
	if err == nil {
		t.Fatal("expected error for non-numeric entry, got nil")
	}
}

func TestStringsToInt32Slice_PreservesOrder(t *testing.T) {
	input := []string{"100", "200", "50"}
	got, err := stringsToInt32Slice(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []int32{100, 200, 50}
	for i, v := range got {
		if v != expected[i] {
			t.Errorf("[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

// ─── unexpectedResponse ───────────────────────────────────────────────────────

func TestUnexpectedResponse(t *testing.T) {
	type someType struct{}
	err := unexpectedResponse(&someType{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Message == "" {
		t.Error("APIError.Message should not be empty")
	}
}
