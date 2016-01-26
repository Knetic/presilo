package presilo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Parses (and returns) the schema from the given [path].
func ParseSchemaFile(path string) (TypeSchema, *SchemaParseContext, error) {

	var sourceFile *os.File
	var name string
	var err error

	path, err = filepath.Abs(path)
	if err != nil {
		return nil, nil, err
	}

	name = filepath.Base(path)

	sourceFile, err = os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer sourceFile.Close()

	return ParseSchemaStream(sourceFile, name)
}

/*
	Parses (and returns) the schema from the given [reader].
	Once this method ends, all returned schemas ought to be properly populated and resolved.
*/
func ParseSchemaStream(reader io.Reader, defaultTitle string) (TypeSchema, *SchemaParseContext, error) {

	var context *SchemaParseContext

	context = NewSchemaParseContext()
	schema, err := ParseSchemaStreamContinue(reader, defaultTitle, context)
	if(err != nil) {
		return nil, nil, err
	}

	err = LinkSchemas(context)
	return schema, context, err
}

/*
	Parses the given [reader] into the given [context], returning the resulatant schema as well as any error encountered.
	This is primarily useful for when you want to parse multiple schemas that are formatted normally, but are not presented in
	one reader.
	This might happen when parsing multiple files that reference each other, drawing from multiple input sources, or when
	schemas are given in some other form, like an array.

	After using this method to parse all required schemas, you must call LinkSchemas() to resolve any outstanding unresolved schema references.
*/
func ParseSchemaStreamContinue(reader io.Reader, defaultTitle string, context *SchemaParseContext) (TypeSchema, error) {

	var buffer bytes.Buffer

	buffer.ReadFrom(reader)
	return ParseSchema(buffer.Bytes(), defaultTitle, context)
}

func ParseSchema(contentsBytes []byte, defaultTitle string, context *SchemaParseContext) (TypeSchema, error) {

	var schema TypeSchema
	var contents map[string]*json.RawMessage
	var schemaRef string
	var schemaType string
	var present, nullable bool
	var err error

	err = json.Unmarshal(contentsBytes, &contents)
	if err != nil {
		return nil, err
	}

	// if this is a reference schema, simply return that exact schema, and do no other processing.
	schemaRef, err = getJsonString(contents, "$ref")
	if err != nil {
		return nil, err
	}

	if len(schemaRef) > 0 {

		schema, present = context.SchemaDefinitions[schemaRef]

		if !present {
			return NewUnresolvedSchema(schemaRef), nil
		}

		return schema, nil
	}

	// if there are definitions, parse them and add them now
	parseDefinitions(contents, context)

	// figure out type
	schemaType, nullable, err = parseSchemaType(contents)
	if len(schemaType) <= 0 {
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

	schema.SetNullable(nullable)

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

func parseSchemaType(contents map[string]*json.RawMessage) (string, bool, error) {

	var typeMessage *json.RawMessage
	var schemaTypes []string
	var typeBytes []byte
	var schemaType string
	var present bool
	var err error

	typeMessage, present = contents["type"]
	if !present {
		return "", false, nil
	}

	typeBytes, err = typeMessage.MarshalJSON()
	if err != nil {
		return "", false, err
	}

	// array?
	err = json.Unmarshal(typeBytes, &schemaTypes)
	if err == nil {

		// check to make sure that length is exactly two, we only support a type and null.
		if len(schemaTypes) != 2 || (schemaTypes[0] != "null" && schemaTypes[1] != "null") {
			return "", false, errors.New("Multi-type schemas must only contain a single type and 'null'")
		}

		if schemaTypes[0] != "null" {
			schemaType = schemaTypes[0]
		} else {
			schemaType = schemaTypes[1]
		}

		return schemaType, true, nil
	}

	// must be single string value?
	err = json.Unmarshal(typeBytes, &schemaType)
	if err != nil {
		return "", false, errors.New("Schema type must be a string, or array of strings")
	}

	// some other type (like a number), ditch it.
	return schemaType, false, nil
}

/*
	Parses any definitions present in the given [contents], and
*/
func parseDefinitions(contents map[string]*json.RawMessage, context *SchemaParseContext) {

	var rawDefinitions *json.RawMessage
	var definitionBytes []byte
	var definitions map[string]*json.RawMessage
	var schema TypeSchema
	var schemaText []byte
	var present bool
	var err error

	rawDefinitions, present = contents["definitions"]
	if !present {
		return
	}

	definitionBytes, err = rawDefinitions.MarshalJSON()
	if err != nil {
		return
	}

	err = json.Unmarshal(definitionBytes, &definitions)
	if err != nil {
		return
	}

	for definitionKey, definitionValue := range definitions {

		schemaText, err = definitionValue.MarshalJSON()

		schema, err = ParseSchema(schemaText, definitionKey, context)
		if err != nil {
			fmt.Printf("Unable to load definition '%s': %s\n", definitionKey, err)
			return
		}

		definitionKey = fmt.Sprintf("#/definitions/%s", definitionKey)
		context.SchemaDefinitions[definitionKey] = schema
	}
}

/*
	"Links" any remaining unresolved schema references together.
	Returns an error if there are any schema references which cannot be resolved.
*/
func LinkSchemas(context *SchemaParseContext) error {

	var schema TypeSchema
	var err error

	for _, schema = range context.SchemaDefinitions {

		err = linkSchema(schema, context)
		if(err != nil) {
			return err
		}
	}

	return nil
}

// If the given [schema] is an ObjectSchema, this runs through all its properties and replaces any unresolved references.
// If there are references which cannot be resolved, and error is returned.
func linkSchema(schema TypeSchema, context *SchemaParseContext) error {

	var objectSchema *ObjectSchema
	var subschema, replacement TypeSchema
	var propertyName, refID string
	var found bool

	if(schema.GetSchemaType() != SCHEMATYPE_OBJECT) {
		return nil
	}

	objectSchema = schema.(*ObjectSchema)

	for propertyName, subschema = range objectSchema.Properties {

		if(subschema.GetSchemaType() != SCHEMATYPE_UNRESOLVED) {
			continue
		}

		refID = subschema.GetID()
		replacement, found = context.SchemaDefinitions[refID]
		if(!found) {
			errorMsg := fmt.Sprintf("Schema ref '%s' could not be resolved.", refID)
			return errors.New(errorMsg)
		}

		objectSchema.Properties[propertyName] = replacement
	}

	return nil
}

func getJsonString(source map[string]*json.RawMessage, key string) (string, error) {

	var ret string
	var retBytes []byte
	var message *json.RawMessage
	var err error
	var present bool

	message, present = source[key]
	if !present {
		return "", nil
	}

	retBytes, err = message.MarshalJSON()
	if err != nil {
		return "", err
	}

	ret = string(retBytes)
	return strings.Replace(ret, "\"", "", -1), nil
}
