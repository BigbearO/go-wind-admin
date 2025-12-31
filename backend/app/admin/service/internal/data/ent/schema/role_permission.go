package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/tx7do/go-crud/entgo/mixin"
)

// RolePermission holds the schema definition for the RolePermission entity.
type RolePermission struct {
	ent.Schema
}

func (RolePermission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "sys_role_permissions",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("角色 - 权限多对多关联表"),
	}
}

// Fields of the RolePermission.
func (RolePermission) Fields() []ent.Field {
	return []ent.Field{

		field.Uint32("role_id").
			Comment("API资源ID（关联sys_api_resources.id）").
			Nillable(),

		field.Uint32("permission_id").
			Comment("权限ID（关联sys_permissions.id）").
			Nillable(),
	}
}

// Mixin of the RolePermission.
func (RolePermission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID{},
	}
}

// Indexes of the RolePermission.
func (RolePermission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id", "permission_id").
			Unique().
			StorageKey("uix_perm_role"),

		index.Fields("tenant_id", "role_id").
			Unique().
			StorageKey("uix_perm_role_tenant"),
	}
}
