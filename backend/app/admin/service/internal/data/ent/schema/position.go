package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/tx7do/go-crud/entgo/mixin"
)

// Position holds the schema definition for the Position entity.
type Position struct {
	ent.Schema
}

func (Position) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "sys_positions",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("职位表"),
	}
}

// Fields of the Position.
func (Position) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("职位名称").
			NotEmpty().
			Optional().
			Nillable(),

		field.String("code").
			Comment("唯一编码").
			NotEmpty().
			Optional().
			Nillable(),

		field.Uint32("org_unit_id").
			Comment("所属组织单元ID").
			Nillable(),

		field.Uint32("reports_to_position_id").
			Comment("汇报关系").
			Optional().
			Nillable(),

		field.String("description").
			Comment("职位描述").
			Optional().
			Nillable(),

		field.String("job_family").
			Comment("职类/序列").
			Optional().
			Nillable(),

		field.String("job_grade").
			Comment("职级").
			Optional().
			Nillable(),

		field.Int32("level").
			Comment("数值化职级").
			Optional().
			Nillable(),

		field.Uint32("headcount").
			Comment("编制人数").
			Default(0).
			Optional().
			Nillable(),

		field.Bool("is_key_position").
			Comment("是否关键岗位").
			Default(false).
			Optional().
			Nillable(),

		field.Enum("type").
			NamedValues(
				"Regular", "REGULAR",
				"Manager", "MANAGER",
				"Lead", "LEAD",
				"Intern", "INTERN",
				"Contract", "CONTRACT",
				"Other", "OTHER",
			).
			Default("REGULAR").
			Comment("岗位类型"),

		field.Time("start_at").
			Comment("生效时间（UTC）").
			Optional().
			Nillable(),

		field.Time("end_at").
			Comment("结束有效期（UTC）").
			Optional().
			Nillable(),
	}
}

// Mixin of the Position.
func (Position) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.SortOrder{},
		mixin.Remark{},
		mixin.Tree[Position]{},
		mixin.TenantID{},
		mixin.SwitchStatus{},
	}
}

// Indexes of the Position.
func (Position) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique().StorageKey("idx_sys_position_code"),
		index.Fields("tenant_id", "code").Unique().StorageKey("uix_sys_position_tenant_code"),
		index.Fields("name").StorageKey("idx_sys_position_name"),
		index.Fields("org_unit_id").StorageKey("idx_sys_position_org_unit_id"),
		index.Fields("type").StorageKey("idx_sys_position_type"),
	}
}
