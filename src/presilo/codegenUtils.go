package presilo

import (
  "strings"
)

/*
	Returns true if any string property of the given schema contains a pattern match
*/
func containsRegexpMatch(schema *ObjectSchema) bool {

	var schemaType SchemaType

	for _, property := range schema.Properties {

		schemaType = property.GetSchemaType()

		if(schemaType == SCHEMATYPE_STRING && property.(*StringSchema).Pattern != nil) {
			return true
		}
	}

	return false
}

/*
	Returns true if any string property of the given schema contains a pattern match
*/
func containsNumberMod(schema *ObjectSchema) bool {

	var schemaType SchemaType

	for _, property := range schema.Properties {

		schemaType = property.GetSchemaType()

		if(schemaType == SCHEMATYPE_NUMBER && property.(*NumberSchema).MultipleOf != nil) {
			return true
		}
	}

	return false
}

/*
	Returns a string with double-quotes properly escaped
*/
func sanitizeQuotedString(target string) string {

	return strings.Replace(target, "\"", "\\\"", -1)
}
