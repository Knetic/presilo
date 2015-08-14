package presilo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func ParseSchemaFile(path string) (TypeSchema, error) {

	var context *SchemaParseContext
	var contentsBytes []byte
	var name string
	var err error

	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	name = filepath.Base(path)

	contentsBytes, err = ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	context = NewSchemaParseContext()
	return ParseSchema(contentsBytes, name, context)
}

func ParseSchema(contentsBytes []byte, defaultTitle string, context *SchemaParseContext) (TypeSchema, error) {

	var schema TypeSchema
	var contents map[string]*json.RawMessage
	var schemaRef string
	var schemaType string
	var present bool
	var err error

	err = json.Unmarshal(contentsBytes, &contents)
	if err != nil {
		return nil, err
	}

	// if this is a reference schema, simply return that exact schema, and do no other processing.
	schemaRef, err = getJsonString(contents, "$ref")
	if(err != nil) {
		return nil, err
	}

	if(len(schemaRef) > 0) {

		schema, present = context.SchemaDefinitions[schemaRef]

		if !present {
			errorMsg := fmt.Sprintf("Schema ref '%s' could not be resolved.", schemaRef)
			return nil, errors.New(errorMsg)
		}

		return schema, nil
	}

	// figure out type
	// TODO: allow "null" as a secondary type.
	schemaType, err = getJsonString(contents, "type")
	if err != nil {
		return nil, err
	}

	if(len(schemaType) <= 0) {
		return nil, errors.New("Schema could not be parsed, type was not specified")
	}

	switch schemaType {

	case "boolean":
		schema, err = NewBooleanSchema(contentsBytes, context)

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

	if err != nil {
		return nil, err
	}

	if len(schema.GetTitle()) == 0 {
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

	if schema.GetSchemaType() == SCHEMATYPE_OBJECT {
		return recurseObjectSchema(schema.(*ObjectSchema), schemas)
	}
	if schema.GetSchemaType() == SCHEMATYPE_ARRAY {
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

func getJsonString(source map[string]*json.RawMessage, key string) (string, error) {

	var ret string
	var retBytes []byte
	var message *json.RawMessage
	var err error
	var present bool

	message, present = source[key]
	if(!present) {
		return "", nil
	}

	retBytes, err = message.MarshalJSON()
	if(err != nil) {
		return "", err
	}

	ret = string(retBytes)
	return strings.Replace(ret, "\"", "", -1), nil
}
