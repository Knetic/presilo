package presilo

import (

  "path/filepath"
  "io/ioutil"
  "strings"
  "fmt"
  "errors"
  "encoding/json"
)

// TODO: Need parse context, which can tie together external schema files as well as schema-local definitions.
/*
  Contains parsing context, such as the currently-defined schemas by ID, and schema-local definitions.
*/
type SchemaParseContext struct {
}

func ParseSchemaFile(path string) (TypeSchema, error) {

    var contentsBytes []byte
    var name string
    var err error

    path, err = filepath.Abs(path)
    if(err != nil) {
      return nil, err
    }

    name = filepath.Base(path)

    contentsBytes, err = ioutil.ReadFile(path)
    if(err != nil) {
      return nil, err
    }

    return ParseSchema(contentsBytes, name)
}

func ParseSchema(contentsBytes []byte, defaultTitle string) (TypeSchema, error) {

  var schema TypeSchema
  var schemaTypeRaw *json.RawMessage
  var contents map[string]*json.RawMessage
  var schemaTypeBytes []byte
  var schemaType string
  var present bool
  var err error

  err = json.Unmarshal(contentsBytes, &contents)
  if(err != nil) {
    return nil, err
  }

  // TODO: see if '$ref' is defined, and if so, use that definition.
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

  case "integer":
    schema, err = NewIntegerSchema(contentsBytes)

  case "number":
    schema, err = NewNumberSchema(contentsBytes)

  case "string":
      schema, err = NewStringSchema(contentsBytes)

  case "array":
    schema, err = NewArraySchema(contentsBytes)

  case "object":
    schema, err = NewObjectSchema(contentsBytes)

  default:
    errorMsg := fmt.Sprintf("Unrecognized schema type: '%s'", schemaType)
    return nil, errors.New(errorMsg)
  }

  if(err != nil) {
    return nil, err
  }

  if(len(schema.GetTitle()) == 0) {
    schema.SetTitle(defaultTitle)
  }

  return schema, nil
}

/*
  Recurses the properties of the given [root],
  adding all sub-schemas to the given [schemas].
*/
func RecurseObjectSchemas(schema TypeSchema, schemas []*ObjectSchema) []*ObjectSchema {

  if(schema.GetSchemaType() == SCHEMATYPE_OBJECT) {
    return recurseObjectSchema(schema.(*ObjectSchema), schemas)
  }
  if(schema.GetSchemaType() == SCHEMATYPE_ARRAY) {
    return RecurseObjectSchemas(schema.(*ArraySchema).Items, schemas)
  }

  return schemas
}

func recurseObjectSchema(schema *ObjectSchema, schemas []*ObjectSchema) []*ObjectSchema {

  schemas = append(schemas, schema)

  for _, property := range schema.Properties {
    schemas = RecurseObjectSchemas(property, schemas)
  }

  return schemas
}
