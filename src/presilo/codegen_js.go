package presilo

import (
	"bytes"
  "fmt"
  "strings"
)

/*
  Generates valid JS code for a given schema.
*/
func GenerateJS(schema *ObjectSchema, module string) string {

	var ret bytes.Buffer

	ret.WriteString(generateJSConstructor(schema, module))
	ret.WriteString("\n")
	ret.WriteString(generateJSFunctions(schema, module))
	ret.WriteString("\n")

	return ret.String()
}

func generateJSConstructor(schema *ObjectSchema, module string) string {

  var ret bytes.Buffer
  var parameterNames []string
  var toWrite, propertyName, parameterName string

  // generate list of property names
  for _, propertyName = range schema.RequiredProperties {

		propertyName = ToJavaCase(propertyName)
		parameterNames = append(parameterNames, propertyName)
	}

  // write constructor signature
	toWrite = fmt.Sprintf("\n/*\n%s\n*/\n", schema.Description)
	ret.WriteString(toWrite)

  toWrite = fmt.Sprintf("%s.%s = function(", module, schema.Title)
  ret.WriteString(toWrite)

  ret.WriteString(strings.Join(parameterNames, ","))
  ret.WriteString(")\n{")

  // body
  for _, parameterName = range parameterNames {

    toWrite = fmt.Sprintf("\n\tthis.set%s(%s)", ToCamelCase(parameterName), parameterName)
    ret.WriteString(toWrite)
  }

  ret.WriteString("\n}\n\n")

  return ret.String()
}

func generateJSFunctions(schema *ObjectSchema, module string) string {

  var ret bytes.Buffer
	//var subschema TypeSchema
	var propertyName, propertyNameCamel, propertyNameJava, schemaName, toWrite string

	schemaName = ToCamelCase(schema.Title)

	for propertyName, _ = range schema.Properties {

		propertyNameCamel = ToCamelCase(propertyName)
		propertyNameJava = ToJavaCase(propertyName)

		toWrite = fmt.Sprintf("%s.%s.prototype.set%s = function(%s)", module, schemaName, propertyNameCamel, propertyNameJava)
		ret.WriteString(toWrite)

		// TODO: type checks
		// TODO: constraints
		toWrite = fmt.Sprintf("\n{\n\tthis.%s = %s\n}\n", propertyNameJava, propertyNameJava)
		ret.WriteString(toWrite)
	}
  return ret.String()
}

/*
	Returns checks appropriate for verifying an object's type.
*/
func generateJSObjectSetter(schema *ObjectSchema) string {

	var ret bytes.Buffer

	// TODO: type check
	return ret.String()
}

/*
	Returns checks appropriate for verifying a numeric value and its constraints.
*/
func generateJSNumericSetter(schema NumericSchemaType) string {

	var ret bytes.Buffer

	// TODO: type check
	// TODO: min/max check
	// TODO: enum check
	// TODO: multiple check
	return ret.String()
}

/*
	Returns checks appropriate for verifying a string value and its constraints.
*/
func generateJSStringSetter(schema *StringSchema) string {

	var ret bytes.Buffer

	// TODO: type check
	// TODO: length check
	// TODO: pattern check
	// TODO: enum check
	return ret.String()
}

/*
	Returns checks appropriate for verifying an array value and its constraints.
*/
func generateJSArraySetter(schema *ArraySchema) string {

	var ret bytes.Buffer

	// TODO: type check
	// TODO: value uniformity check
	// TODO: length checks
	return ret.String()
}

/*
	Generates code which throws an error if the given [parameter]'s type name is not equal to the given [typeName]
*/
func generateJSTypeCheck(parameter string, typeName string) string {

	var ret bytes.Buffer
	return ret.String
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generateJSEnumCheck(parameter string, validValues []interface{}, stringFormat string) string {

	var ret bytes.Buffer
	return ret.String
}
