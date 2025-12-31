package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/tx7do/go-crud/entgo/mixin"
)

// MembershipOrgUnit 表示 membership 与组织单元的关联（多对多/冗余关系）
type MembershipOrgUnit struct {
	ent.Schema
}

func (MembershipOrgUnit) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "sys_membership_org_units",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("成员与组织单元关联表"),
	}
}

func (MembershipOrgUnit) Fields() []ent.Field {
	return []ent.Field{
		// 关联到 membership（必填）
		field.Uint32("membership_id").
			Comment("membership ID").
			Nillable(),

		// 关联到组织单元（必填）
		field.Uint32("org_unit_id").
			Comment("组织单元 ID").
			Nillable(),

		// 可选：在该组织内的岗位/角色引用（冗余）
		field.Uint32("position_id").
			Optional().
			Nillable().
			Comment("岗位 ID（可选，冗余）"),
		field.Uint32("role_id").
			Optional().
			Nillable().
			Comment("角色 ID（可选，冗余）"),

		// 可选：该关联的生效/结束时间
		field.Time("start_at").
			Optional().
			Nillable().
			Comment("生效时间（UTC）"),
		field.Time("end_at").
			Optional().
			Nillable().
			Comment("结束时间（UTC）"),

		// 分配审计字段（记录分配时刻与分配者）
		field.Time("assigned_at").
			Optional().
			Nillable().
			Comment("分配时间（UTC）"),
		field.Uint32("assigned_by").
			Optional().
			Nillable().
			Comment("分配者用户 ID"),

		// 标记当前是否为主要所属（用于单一场景）
		field.Bool("is_primary").
			Default(false).
			Nillable().
			Comment("是否为主所属"),

		field.Enum("status").
			Comment("关联状态").
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

func (MembershipOrgUnit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID{},
		mixin.Remark{},
	}
}

func (MembershipOrgUnit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "membership_id", "org_unit_id").
			Unique().
			StorageKey("uix_membership_org_unit_tenant_mem_org"),

		index.Fields("membership_id").StorageKey("idx_mou_membership_id"),
		index.Fields("org_unit_id").StorageKey("idx_mou_org_unit_id"),
	}
}
