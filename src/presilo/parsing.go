package presilo

import (

  "path/filepath"
  "io/ioutil"
  "strings"
  "fmt"
  "errors"
  "encoding/json"
)

func ParseSchemaFile(path string) (TypeSchema, error) {

    var contentsBytes []byte
    var contents map[string]*json.RawMessage
    var err error

    path, err = filepath.Abs(path)
    if(err != nil) {
      return nil, err
    }

    contentsBytes, err = ioutil.ReadFile(path)
    if(err != nil) {
      return nil, err
    }

    err = json.Unmarshal(contentsBytes, &contents)
    if(err != nil) {
      return nil, err
    }

    return Parse(contentsBytes, contents)
}

func Parse(contentsBytes []byte, contents map[string]*json.RawMessage) (TypeSchema, error) {

  var schemaTypeRaw *json.RawMessage
  var schemaTypeBytes []byte
  var schemaType string
  var present bool
  var err error

  schemaTypeRaw, present = contents["type"]
  if(!present) {
    return nil, errors.New("Type was not specified")
  }
  if(schemaTypeRaw == nil) {
    return nil, errors.New("Schema could not be parsed, type was not specified")
  }

  schemaTypeBytes, err = schemaTypeRaw.MarshalJSON()
  if(err != nil) {
    return nil, err
  }

  schemaType = string(schemaTypeBytes)
  schemaType = strings.Replace(schemaType, "\"", "", -1)

  switch(schemaType) {

    case "integer": return NewIntegerSchema(contentsBytes)
  }

  errorMsg := fmt.Sprintf("Unrecognized schema type: '%s'", schemaType)
  return nil, errors.New(errorMsg)
}
