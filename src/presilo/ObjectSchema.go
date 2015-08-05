package presilo

import (
  "encoding/json"
)

type ObjectSchema struct {

  Schema
  Properties []Schema
  Required []string `json:"required"`
}

func NewObjectSchema(contents map[string]interface{}) (*ObjectSchema, error) {

  var ret *ObjectSchema
  var err error

  ret = new(ObjectSchema)
  err = json.Unmarshal(contents, &contents)

  if(err != nil) {
    return nil, rr
  }

  return ret, nil
}
