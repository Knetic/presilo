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

    var context *SchemaParseContext
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

    context = NewSchemaParseContext()
    return ParseSchema(contentsBytes, name, context)
}

func ParseSchema(contentsBytes []byte, defaultTitle string, context *SchemaParseContext) (TypeSchema, error) {

  var schema TypeSchema
  var schemaTypeRaw, schemaRefRaw *json.RawMessage
  var contents map[string]*json.RawMessage
  var schemaTypeBytes, schemaRefBytes []byte
  var schemaRef string
  var schemaType string
  var present bool
  var err error

  err = json.Unmarshal(contentsBytes, &contents)
  if(err != nil) {
    return nil, err
  }

  // if this is a reference schema, simply return that exact schema, and do no other processing.
  schemaRefRaw, present = contents["$ref"]
  if(present) {

    schemaRefBytes, err = schemaRefRaw.MarshalJSON()
    if(err != nil) {
      return nil, err
    }

    schemaRef = string(schemaRefBytes)

    schema, present = context.SchemaDefinitions[schemaRef]
    if(!present) {

      errorMsg := fmt.Sprintf("Schema ref '%s' could not be resolved.", schemaRef)
      return nil, errors.New(errorMsg)
    }

    return schema, nil
  }

  // figure out type
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
    schema, err = NewIntegerSchema(contentsBytes, context)

  case "number":
    schema, err = NewNumberSchema(contentsBytes, context)

  case "string":
    schema, err = NewStringSchema(contentsBytes, context)

  case "array":
    schema, err = NewArraySchema(contentsBytes, context)

  case "object":
    schema, err = NewObjectSchema(contentsBytes, context)

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

  context.SchemaDefinitions[schema.GetID()] = schema
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
