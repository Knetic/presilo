package presilo

import (
	"encoding/json"
)

/*
  A schema which describes a number, which may include floating point numbers.
*/
type NumberSchema struct {
	Schema
	Minimum          *float64 `json:"minimum"`
	Maximum          *float64 `json:"maximum"`
	ExclusiveMinimum *bool    `json:"exclusiveMinimum"`
	ExclusiveMaximum *bool    `json:"exclusiveMaximum"`
	MultipleOf       *float64 `json:"multipleOf"`
	Enum						*[]float64 `json:"enum"`
}

/*
  Creates a new integer schema from a byte slice that can be interpreted as json.
*/
func NewNumberSchema(contents []byte, context *SchemaParseContext) (*NumberSchema, error) {

	var ret *NumberSchema
	var err error

	ret = new(NumberSchema)
	ret.typeCode = SCHEMATYPE_NUMBER

	err = json.Unmarshal(contents, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (this *NumberSchema) HasConstraints() bool {
	return this.Minimum != nil ||
				this.Maximum != nil ||
				this.MultipleOf != nil ||
				this.Enum != nil
}

func (this *NumberSchema) HasMinimum() bool {
	return this.Minimum != nil
}

func (this *NumberSchema) HasMaximum() bool {
	return this.Maximum != nil
}

func (this *NumberSchema) HasMultiple() bool {
	return this.MultipleOf != nil
}

func (this *NumberSchema) HasEnum() bool {
	return this.Enum != nil
}

func (this *NumberSchema) GetMinimum() interface{} {
	return *this.Minimum
}

func (this *NumberSchema) GetMaximum() interface{} {
	return *this.Maximum
}

func (this *NumberSchema) GetMultiple() interface{} {
	return *this.MultipleOf
}

func (this *NumberSchema) GetEnum() []interface{} {

	var ret []interface{}
	var enumValues []float64
	var length int

	length = len(*this.Enum)
	ret = make([]interface{}, length)
	enumValues = *this.Enum

	for i := 0; i < length; i++ {
		ret[i] = enumValues[i]
	}
	return ret
}

func (this *NumberSchema) IsExclusiveMaximum() bool {
	return this.ExclusiveMaximum != nil
}

func (this *NumberSchema) IsExclusiveMinimum() bool {
	return this.ExclusiveMinimum != nil
}

func (this *NumberSchema) GetConstraintFormat() string {
	return "%f"
}
