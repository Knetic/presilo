package presilo

import (
	"encoding/json"
)

/*
  A schema which describes an integer.
*/
type IntegerSchema struct {
	Schema
	Minimum          *int  `json:"minimum"`
	Maximum          *int  `json:"maximum"`
	ExclusiveMaximum *bool `json:"exclusiveMaximum"`
	ExclusiveMinimum *bool `json:"exclusiveMinimum"`
	MultipleOf       *int  `json:"multipleOf"`
	Enum						 *[]int `json:"enum"`
}

/*
  Creates a new integer schema from a byte slice that can be interpreted as json.
*/
func NewIntegerSchema(contents []byte, context *SchemaParseContext) (*IntegerSchema, error) {

	var ret *IntegerSchema
	var err error

	ret = new(IntegerSchema)

	ret.typeCode = SCHEMATYPE_INTEGER
	err = json.Unmarshal(contents, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (this *IntegerSchema) HasConstraints() bool {
	return this.Minimum != nil ||
				this.Maximum != nil ||
				this.MultipleOf != nil ||
				this.Enum != nil
}

func (this *IntegerSchema) HasMinimum() bool {
	return this.Minimum != nil
}

func (this *IntegerSchema) HasMaximum() bool {
	return this.Maximum != nil
}

func (this *IntegerSchema) HasEnum() bool {
	return this.Enum != nil
}

func (this *IntegerSchema) HasMultiple() bool {
	return this.MultipleOf != nil
}

func (this *IntegerSchema) GetMinimum() interface{} {
	return *this.Minimum
}

func (this *IntegerSchema) GetMaximum() interface{} {
	return *this.Maximum
}

func (this *IntegerSchema) GetMultiple() interface{} {
	return *this.MultipleOf
}

func (this *IntegerSchema) GetEnum() []interface{} {

	var ret []interface{}
	var enumValues []int
	var length int

	length = len(*this.Enum)
	ret = make([]interface{}, length)
	enumValues = *this.Enum

	for i := 0; i < length; i++ {
		ret[i] = enumValues[i]
	}
	return ret
}

func (this *IntegerSchema) IsExclusiveMaximum() bool {
	return this.ExclusiveMaximum != nil
}

func (this *IntegerSchema) IsExclusiveMinimum() bool {
	return this.ExclusiveMinimum != nil
}

func (this *IntegerSchema) GetConstraintFormat() string {
	return "%d"
}
