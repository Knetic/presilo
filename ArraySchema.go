package presilo

import (
	"encoding/json"
	"errors"
)

/*
  A schema which describes an array.
*/
type ArraySchema struct {
	Schema

	// TODO: item schema should be able to have more than one type in it.
	// mixed types should be automatically set as subclasses of a common ancestor, in languages that support it
	Items TypeSchema

	MaxItems    *int  `json:"maxItems"`
	MinItems    *int  `json:"minItems"`
	UniqueItems *bool `json:"uniqueItems"`

	RawItems *json.RawMessage `json:"items"`
}

/*
  Creates a new integer schema from a byte slice that can be interpreted as json.
*/
func NewArraySchema(contents []byte, context *SchemaParseContext) (*ArraySchema, error) {

	var ret *ArraySchema
	var err error

	ret = new(ArraySchema)
	ret.typeCode = SCHEMATYPE_ARRAY

	err = json.Unmarshal(contents, &ret)
	if err != nil {
		return ret, err
	}

	if ret.RawItems == nil {
		return nil, errors.New("Array specified, but no item type given.")
	}

	ret.Items, err = ParseSchema(*ret.RawItems, "", context)
	return ret, err
}

func (this *ArraySchema) HasConstraints() bool {
	return this.MaxItems != nil || this.MinItems != nil || this.UniqueItems != nil
}
