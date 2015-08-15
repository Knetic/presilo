package presilo

import (
	"encoding/json"
)

/*
  A schema which describes an integer.
*/
type BooleanSchema struct {
	Schema
}

/*
  Creates a new integer schema from a byte slice that can be interpreted as json.
*/
func NewBooleanSchema(contents []byte, context *SchemaParseContext) (*BooleanSchema, error) {

	var ret *BooleanSchema
	var err error

	ret = new(BooleanSchema)
	ret.typeCode = SCHEMATYPE_BOOLEAN

	err = json.Unmarshal(contents, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (this *BooleanSchema) HasConstraints() bool {
	return false
}
