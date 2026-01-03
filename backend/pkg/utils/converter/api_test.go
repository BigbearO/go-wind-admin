package converter

import (
	"testing"
)

func TestApiPermissionConverter_ConvertByPath(t *testing.T) {
	c := NewApiPermissionConverter()

	cases := []struct {
		name   string
		method string
		path   string
		want   string
	}{
		{"get list users", "GET", "/v1/users", "users:list"},
		{"get single user", "GET", "/v1/users/{id}", "users:get"},
		{"create user", "POST", "/v1/users", "users:create"},
		{"update user", "PUT", "/v1/users/{id}", "users:update"},
		{"delete user", "DELETE", "/v1/users/{id}", "users:delete"},
		{"nested admin settings", "GET", "/api/v1/admin/settings", "admin:settings:list"},
		{"hyphen group", "GET", "/v1/user-groups", "user-groups:list"},
		{"get task by typeNames", "GET", "/admin/v1/tasks:type-names", "tasks:type-names:list"},
		{"walk route", "GET", "/admin/v1/api-resources/walk-route", "api-resources:walk-route:list"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := c.ConvertCodeByPath(tc.method, tc.path)
			if got != tc.want {
				t.Fatalf("ConvertCodeByPath(%q, %q) = %q, want %q", tc.method, tc.path, got, tc.want)
			}
		})
	}
}

func TestApiPermissionConverter_ConvertByOperationID(t *testing.T) {
	c := NewApiPermissionConverter()

	cases := []struct {
		name string
		op   string
		want string
	}{
		{"rpc with service and name", "TaskService_ListTaskTypeName", "task:task-type-name:list"},
		{"rpc without Service suffix", "Task_ListTaskTypeName", "task:task-type-name:list"},
		{"get walk route data", "ApiResourceService_GetWalkRouteData", "api-resource:walk-route-data:get"},
		{"rpc without name", "Task_List", "task:list"},
		{"invalid rpc format", "GetStatus", ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := c.ConvertCodeByOperationID(tc.op)
			if got != tc.want {
				t.Fatalf("ConvertCodeByOperationID(%q) = %q, want %q", tc.op, got, tc.want)
			}
		})
	}
}

func TestStripVersionPrefix(t *testing.T) {
	c := NewApiPermissionConverter()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"root slash", "/", ""},
		{"no leading slash v1/users", "v1/users", "users"},
		{"v1 users", "/v1/users", "users"},
		{"api v1 admin", "/api/v1/admin/settings", "admin/settings"},
		{"v2 only", "/v2", ""},
		{"v10 without slash", "v10", ""},
		{"api v10", "/api/v10", ""},
		{"double slash after version", "/api/v1//admin", "admin"},
		{"double slash after version", "/api/v1.9/admin", "admin"},
		{"no version path", "/foo/bar", "foo/bar"},
		{"no version path", "/admin/v1/tasks:type-names", "tasks:type-names"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := c.stripVersionPrefix(tc.in)
			if got != tc.want {
				t.Fatalf("stripVersionPrefix(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
