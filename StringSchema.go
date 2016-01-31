package presilo

import (
	"encoding/json"
)

/*
  A schema which describes an integer.
*/
type StringSchema struct {
	Schema
	MaxLength     *int      `json:"maxLength"`
	MinLength     *int      `json:"minLength"`
	Pattern       *string   `json:"pattern"`
	MaxByteLength *int      `json:"maxByteLength"`
	MinByteLength *int      `json:"minByteLength"`
	Enum          *[]string `json:"enum"`
}

func NewStringSchema() *StringSchema {

	ret := new(StringSchema)
	ret.typeCode = SCHEMATYPE_STRING
	return ret
}

/*
  Creates a new integer schema from a byte slice that can be interpreted as json.
*/
func ParseStringSchema(contents []byte, context *SchemaParseContext) (*StringSchema, error) {

	var ret *StringSchema
	var err error

	ret = NewStringSchema()

	err = json.Unmarshal(contents, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (this *StringSchema) HasConstraints() bool {
	return this.Enum != nil ||
		this.MinLength != nil ||
		this.MaxLength != nil ||
		this.Pattern != nil ||
		this.MaxByteLength != nil ||
		this.MinByteLength != nil
}

func (this *StringSchema) HasEnum() bool {
	return this.Enum != nil
}

func (this *StringSchema) GetEnum() []interface{} {

	var ret []interface{}
	var enumValues []string
	var length int

	length = len(*this.Enum)
	ret = make([]interface{}, length)
	enumValues = *this.Enum

	for i := 0; i < length; i++ {
		ret[i] = enumValues[i]
	}
	return ret
}
