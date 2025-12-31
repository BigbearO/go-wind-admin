package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

type OrgUnit struct {
	ent.Schema
}

func (OrgUnit) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "sys_org_units",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("组织单元表"),
	}
}

func (OrgUnit) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Nillable().
			Comment("名称"),

		field.String("code").
			Optional().
			Nillable().
			Comment("唯一编码（可用于导入/识别）"),

		field.String("path").
			Optional().
			Nillable().
			Comment("树路径，如：/1/10/"),

		field.Int32("sort_order").
			Optional().
			Nillable().
			Default(0).
			Comment("排序序号"),

		field.Uint32("leader_id").
			Optional().
			Nillable().
			Comment("负责人用户ID"),

		field.Enum("type").
			NamedValues(
				"Company", "COMPANY",
				"Division", "DIVISION",
				"Department", "DEPARTMENT",
				"Team", "TEAM",
				"Project", "PROJECT",
				"Committee", "COMMITTEE",
				"Region", "REGION",
				"Other", "OTHER",
			).
			Nillable().
			Default("DEPARTMENT").
			Comment("组织类型"),

		field.Strings("business_scopes").
			Optional().
			Comment("组织的业务范围/服务条线"),

		field.String("external_id").
			Optional().
			Nillable().
			Comment("外部系统ID"),

		field.Bool("is_legal_entity").
			Optional().
			Nillable().
			Default(false).
			Comment("是否为法定主体"),

		field.String("registration_number").
			Optional().
			Nillable().
			Comment("注册号/统一社会信用代码"),

		field.String("tax_id").
			Optional().
			Nillable().
			Comment("税号"),

		field.Uint32("legal_entity_org_id").
			Optional().
			Nillable().
			Comment("关联的法定主体组织ID"),

		field.String("address").
			Optional().
			Nillable().
			Comment("详细地址"),

		field.String("phone").
			Optional().
			Nillable().
			Comment("联系电话"),

		field.String("email").
			Optional().
			Nillable().
			Comment("联系邮箱"),

		field.String("timezone").
			Optional().
			Nillable().
			Comment("时区"),

		field.String("country").
			Optional().
			Nillable().
			Comment("国家/地区代码"),

		field.Float("latitude").
			Optional().
			Nillable().
			Comment("纬度"),

		field.Float("longitude").
			Optional().
			Nillable().
			Comment("经度"),

		field.Time("start_at").
			Comment("生效时间（UTC）").
			Optional().
			Nillable(),

		field.Time("end_at").
			Comment("结束有效期（UTC）").
			Optional().
			Nillable(),

		field.Uint32("contact_user_id").
			Optional().
			Nillable().
			Comment("业务联系人用户ID"),

		field.Strings("permission_tags").
			Optional().
			Comment("与权限/角色映射的标签"),
	}
}

func (OrgUnit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.AutoIncrementId{},
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.SwitchStatus{},
		mixin.TenantID{},
		mixin.Remark{},
		mixin.Description{},
		mixin.Tree[OrgUnit]{},
	}
}

func (OrgUnit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id").StorageKey("idx_org_tenant_id"),
		//index.Fields("parent_id").StorageKey("idx_org_parent_id"),
		// tenant + code 唯一，便于按业务租户内查找
		index.Fields("tenant_id", "code").
			Unique().
			StorageKey("uix_org_tenant_code"),
	}
}
