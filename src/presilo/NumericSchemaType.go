package presilo

type NumericSchemaType interface {

  HasConstraints() bool
  HasMinimum() bool
  HasMaximum() bool
  HasMultiple() bool
  GetMinimum() interface{}
  GetMaximum() interface{}
  GetMultiple() interface{}
  IsExclusiveMaximum() bool
  IsExclusiveMinimum() bool
  GetConstraintFormat() string
}
