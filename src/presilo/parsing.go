package presilo

import (
  "encoding/json"
)

func Parse(contents map[string]interface{}) (Schema, error) {

  var ret Schema
}

func createSchemaByType(contents map[string]interface{}) (Schema, error) {

  var schemaType string
  var err error

  schemaType, err = contents["type"]
  if(err != nil) {
    return nil, err
  }

  switch(schemaType) {

  case "object": return NewObjectSchema(contents)
  }

  return nil, nil
}
