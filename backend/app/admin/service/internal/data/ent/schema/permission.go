package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type Permission struct {
	ent.Schema
}

func (Permission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "sys_permissions",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("权限表（资源/路由/菜单/数据权限 等）"),
	}
}

func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Nillable().
			Comment("名称"),

		field.String("code").
			Optional().
			Nillable().
			Comment("唯一编码（租户范围内唯一，便于引用/导入）"),

		field.String("path").
			Optional().
			Nillable().
			Comment("路径/路由，如 `/api/users` 或 菜单路径"),

		field.String("resource").
			Optional().
			Nillable().
			Comment("资源标识（如 API 资源名）"),

		field.String("method").
			Optional().
			Nillable().
			Comment("HTTP 方法/动作，如 GET/POST（可选）"),

		field.Int32("sort_order").
			Optional().
			Nillable().
			Default(0).
			Comment("排序序号"),

		field.Enum("type").
			NamedValues(
				"Api", "API",
				"Menu", "MENU",
				"Button", "BUTTON",
				"Page", "PAGE",
				"Data", "DATA",
				"Other", "OTHER",
			).
			Nillable().
			Default("API").
			Comment("权限类型"),
	}
}

func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.Remark{},
		mixin.SwitchStatus{},
		mixin.TenantID{},
		mixin.Tree[Permission]{},
	}
}

func (Permission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id").StorageKey("idx_perm_tenant_id"),
		index.Fields("parent_id").StorageKey("idx_perm_parent_id"),
		// tenant + code 唯一，便于按租户内查找/引用
		index.Fields("tenant_id", "code").
			Unique().
			StorageKey("uix_perm_tenant_code"),
	}
}
