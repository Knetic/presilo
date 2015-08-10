package presilo

import (
  "encoding/json"
)

/*
  A schema which describes a number, which may include floating point numbers.
*/
type NumberSchema struct {

  Schema
  Minimum *float64 `json:"minimum"`
  Maximum *float64 `json:"maximum"`
  // TODO: ExclusiveMinimum *bool `json:"exclusiveMinimum"`
  // TODO: ExclusiveMaximum *bool `json:"exclusiveMaximum"`
  // TODO: MultipleOf *float64 `json:"multipleOf"`
}

/*
  Creates a new integer schema from a byte slice that can be interpreted as json.
*/
func NewNumberSchema(contents []byte) (*NumberSchema, error) {

  var ret *NumberSchema
  var err error

  ret = new(NumberSchema)
  ret.typeCode = SCHEMATYPE_NUMBER

  err = json.Unmarshal(contents, &ret)
  if(err != nil) {
    return ret, err
  }

  return ret, nil
}

func (this *NumberSchema) HasConstraints() bool {
  return false
}
