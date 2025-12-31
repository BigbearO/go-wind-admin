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
		schema.Comment("权限核心表"),
	}
}

func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Nillable().
			Comment("权限名称（如：删除用户）"),

		field.String("code").
			Optional().
			Nillable().
			Comment("权限唯一编码（如：user.delete）"),

		field.String("path").
			Optional().
			Nillable().
			Comment("树路径，如：/1/10/"),

		field.String("module").
			Comment("所属业务模块（如：用户管理/订单管理）").
			Optional().
			Nillable(),

		field.Int32("sort_order").
			Optional().
			Nillable().
			Default(0).
			Comment("排序序号"),

		field.Enum("type").
			NamedValues(
				"Catalog", "CATALOG",
				"Menu", "MENU",
				"Page", "PAGE",
				"Button", "BUTTON",
				"Api", "API",
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

		index.Fields("module", "type").
			StorageKey("idx_perm_module_type"),
	}
}
