package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/tx7do/go-crud/entgo/mixin"
)

// tenant_id 为 NULL 或者 0 时代表全局或个人系统成员。
// 建议使用 0 代表全局/平台，避免 NULL 带来的三值逻辑复杂性。

// Membership holds the schema definition for the Membership entity.
type Membership struct {
	ent.Schema
}

func (Membership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "sys_memberships",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("成员关联表。存储用户在不同租户或系统空间的身份、状态及角色。tenant_id 为 NULL 时代表全局或个人系统成员。"),
	}
}

// Fields of the Membership.
func (Membership) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("user_id").
			Comment("用户ID").
			Nillable(),

		// 冗余单一场景字段（互斥：与独立关联表二选一）
		field.Uint32("org_unit_id").
			Optional().
			Nillable().
			Comment("组织架构ID（单一，冗余/互斥）"),
		field.Uint32("position_id").
			Optional().
			Nillable().
			Comment("职位ID（单一，冗余/互斥）"),
		field.Uint32("role_id").
			Optional().
			Nillable().
			Comment("角色ID（单一，冗余/互斥）"),

		field.Bool("is_primary").
			Default(false).
			Nillable().
			Comment("是否主身份"),

		field.Time("start_at").
			Comment("生效时间").
			Optional().
			Nillable(),

		field.Time("end_at").
			Comment("失效时间").
			Optional().
			Nillable(),

		field.Time("assigned_at").
			Comment("分配时间（UTC）").
			Optional().
			Nillable(),
		field.Uint32("assigned_by").
			Comment("分配者用户ID").
			Optional().
			Nillable(),

		field.Enum("status").
			Comment("状态").
			Optional().
			Nillable().
			Default("ACTIVE").
			NamedValues(
				"Active", "ACTIVE",
				"Pending", "PENDING",
				"Inactive", "INACTIVE",
				"Suspended", "SUSPENDED",
				"Expired", "EXPIRED",
			),
	}
}

// Mixin of the Membership.
func (Membership) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID{},
		mixin.Remark{},
	}
}

// Indexes of the Membership.
func (Membership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "user_id").Unique().StorageKey("uix_sys_membership_tenant_user"),

		index.Fields("user_id").StorageKey("idx_membership_user_id"),
		index.Fields("tenant_id").StorageKey("idx_membership_tenant_id"),

		index.Fields("org_unit_id").StorageKey("idx_membership_org_unit_id"),
		index.Fields("position_id").StorageKey("idx_membership_position_id"),
		index.Fields("role_id").StorageKey("idx_membership_role_id"),
	}
}
