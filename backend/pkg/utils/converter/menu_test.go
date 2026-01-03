package converter

import (
	"testing"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
)

func TestMenuConverter_Convert(t *testing.T) {
	c := NewMenuPermissionConverter()

	tests := []struct {
		name     string
		fullPath string
		typ      adminV1.Menu_Type
		want     string
	}{
		{"catalog no action", "/users", adminV1.Menu_CATALOG, "users"},
		{"menu view", "/users", adminV1.Menu_MENU, "users:view"},
		{"admin v1 prefix", "/foo", adminV1.Menu_EMBEDDED, "foo:embed"},
		{"api v2 prefix", "/admin/settings", adminV1.Menu_LINK, "admin:settings:link"},
		{"no version catalog", "/users/profile", adminV1.Menu_CATALOG, "users:profile"},
		{"button no action", "admin/button", adminV1.Menu_BUTTON, "admin:button"},
		{"leading and trailing slashes", "/orders/", adminV1.Menu_MENU, "orders:view"},
		{"complex path keep later version", "/admin/v2/inner", adminV1.Menu_MENU, "admin:v2:inner:view"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.ConvertCode(tt.fullPath, tt.typ)
			if got != tt.want {
				t.Fatalf("ConvertCode(%q, %v) = %q, want %q", tt.fullPath, tt.typ, got, tt.want)
			}
		})
	}
}

func TestMenuConverter_typeToAction(t *testing.T) {
	c := NewMenuPermissionConverter()

	tests := []struct {
		typ  adminV1.Menu_Type
		want string
	}{
		{adminV1.Menu_CATALOG, ""},
		{adminV1.Menu_MENU, "view"},
		{adminV1.Menu_BUTTON, ""},
		{adminV1.Menu_EMBEDDED, "embed"},
		{adminV1.Menu_LINK, "link"},
		{adminV1.Menu_Type(999), ""},
	}

	for _, tt := range tests {
		got := c.typeToAction(tt.typ)
		if got != tt.want {
			t.Fatalf("typeToAction(%v) = %q, want %q", tt.typ, got, tt.want)
		}
	}
}
