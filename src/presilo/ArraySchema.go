package presilo

import (
  "encoding/json"
)

/*
  A schema which describes an array.
*/
type ArraySchema struct {

  Schema

  // TODO: Items []TypeSchema

  // TODO: MaxItems *int `json:"maxItems"`
  // TODO: MinItems *int `json:"minItems"`
  // TODO: UniqueItems *bool `json:"uniqueItems"`
}

/*
  Creates a new integer schema from a byte slice that can be interpreted as json.
*/
func NewArraySchema(contents []byte) (*ArraySchema, error) {

  var ret *ArraySchema
  var err error

  ret = new(ArraySchema)
  ret.typeCode = SCHEMATYPE_ARRAY

  err = json.Unmarshal(contents, &ret)
  if(err != nil) {
    return ret, err
  }

  return ret, nil
}

func (this *ArraySchema) HasConstraints() bool {
  return false
}
