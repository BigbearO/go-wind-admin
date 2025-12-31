package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/tx7do/go-crud/entgo/mixin"
)

// MembershipPosition holds the schema definition for the MembershipPosition entity.
type MembershipPosition struct {
	ent.Schema
}

func (MembershipPosition) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "sys_membership_positions",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("用户与岗位关联表"),
	}
}

// Fields of the MembershipPosition.
func (MembershipPosition) Fields() []ent.Field {
	return []ent.Field{
		// 关联到 membership（必填）
		field.Uint32("membership_id").
			Comment("membership ID").
			Nillable(),

		// 关联到岗位（必填）
		field.Uint32("position_id").
			Comment("岗位ID").
			Nillable(),

		field.Bool("is_primary").
			Comment("是否主岗位").
			Optional().
			Nillable().
			Default(false),

		field.Time("start_at").
			Comment("生效时间").
			Optional().
			Nillable(),

		field.Time("end_at").
			Comment("失效时间").
			Optional().
			Nillable(),

		field.Time("assigned_at").
			Comment("岗位分配时间").
			Optional().
			Nillable(),

		field.Uint32("assigned_by").
			Comment("分配者用户 ID").
			Optional().
			Nillable(),

		field.Enum("status").
			NamedValues(
				"Probation", "PROBATION",
				"Active", "ACTIVE",
				"Leave", "LEAVE",
				"Terminated", "TERMINATED",
				"Expired", "EXPIRED",
			).
			Default("ACTIVE").
			Comment("岗位状态"),
	}
}

// Mixin of the MembershipPosition.
func (MembershipPosition) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID{},
		mixin.Remark{},
	}
}

// Indexes of the MembershipPosition.
func (MembershipPosition) Indexes() []ent.Index {
	return []ent.Index{
		// 在 tenant 维度上保证唯一性，避免全局记录冲突
		index.Fields("tenant_id", "membership_id", "position_id").
			Unique().
			StorageKey("uix_sys_membership_position_tenant_mem_pos"),

		index.Fields("membership_id").StorageKey("idx_sys_membership_position_membership_id"),
		index.Fields("position_id").StorageKey("idx_sys_membership_position_position_id"),
	}
}
