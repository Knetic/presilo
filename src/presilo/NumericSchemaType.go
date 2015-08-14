package presilo

type NumericSchemaType interface {

	TypeSchema
	HasConstraints() bool
	HasMinimum() bool
	HasMaximum() bool
	HasMultiple() bool
	HasEnum() bool
	GetMinimum() interface{}
	GetMaximum() interface{}
	GetMultiple() interface{}
	GetEnum() []interface{}
	IsExclusiveMaximum() bool
	IsExclusiveMinimum() bool
	GetConstraintFormat() string
}
