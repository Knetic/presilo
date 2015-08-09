package presilo

import (
  "encoding/json"
)

/*
  A schema which describes an integer.
*/
type IntegerSchema struct {

  Schema
  Minimum *int `json:"minimum"`
  Maximum *int `json:"maximum"`
}

/*
  Creates a new integer schema from a byte slice that can be interpreted as json.
*/
func NewIntegerSchema(contents []byte) (IntegerSchema, error) {

  var ret IntegerSchema
  var err error

  err = json.Unmarshal(contents, &ret)
  if(err != nil) {
    return ret, err
  }

  return ret, nil
}
