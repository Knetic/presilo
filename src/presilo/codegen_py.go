package presilo

import (
	"fmt"
	"strings"
	"bytes"
)

func GeneratePython(schema *ObjectSchema, module string) string {

	var ret bytes.Buffer

	ret.WriteString(generatePythonImports(schema))
	ret.WriteString("\n")
	ret.WriteString(generatePythonSignature(schema))
	ret.WriteString("\n")
	ret.WriteString(generatePythonConstructor(schema))
	ret.WriteString("\n")
	ret.WriteString(generatePythonFunctions(schema))
	ret.WriteString("\n")

	return ret.String()
}

func generatePythonImports(schema *ObjectSchema) string {

	var ret bytes.Buffer

	ret.WriteString("import string\n")

	if containsRegexpMatch(schema) {
		ret.WriteString("import re\n")
	}

	return ret.String()
}

func generatePythonSignature(schema *ObjectSchema) string {

	return fmt.Sprintf("class %s(Object):", ToCamelCase(schema.Title))
}

func generatePythonConstructor(schema *ObjectSchema) string {

	var ret bytes.Buffer
	var declarations, setters []string
	var propertyName string
	var toWrite string

	ret.WriteString("\n\tdef __init__(")

	// required properties
	for _, propertyName = range schema.RequiredProperties {

		propertyName = ToSnakeCase(propertyName)

		declarations = append(declarations, propertyName)

		toWrite = fmt.Sprintf("\n\t\tself.set_%s(%s)", propertyName, propertyName)
		setters = append(setters, toWrite)
	}

	toWrite = strings.Join(declarations, ",")
	ret.WriteString(toWrite)
	ret.WriteString("):")

	for _, setter := range setters {
		ret.WriteString(setter)
	}

	return ret.String()
}

func generatePythonFunctions(schema *ObjectSchema) string {

	var ret bytes.Buffer
	var subschema TypeSchema
	var toWrite string
	var propertyName, snakeName string

	for propertyName, subschema = range schema.Properties {

		snakeName = ToSnakeCase(propertyName)

		// setter
		toWrite = fmt.Sprintf("\n\tdef set_%s(%s):", snakeName, snakeName)
		ret.WriteString(toWrite)

		switch subschema.GetSchemaType() {
		case SCHEMATYPE_BOOLEAN:
			toWrite = ""
		case SCHEMATYPE_STRING:
			toWrite = generatePythonStringSetter(subschema.(*StringSchema))
		case SCHEMATYPE_INTEGER:
			fallthrough
		case SCHEMATYPE_NUMBER:
			toWrite = generatePythonNumericSetter(subschema.(NumericSchemaType))
		case SCHEMATYPE_OBJECT:
			toWrite = generatePythonObjectSetter(subschema.(*ObjectSchema))
		case SCHEMATYPE_ARRAY:
			toWrite = generatePythonArraySetter(subschema.(*ArraySchema))
		}

		ret.WriteString(toWrite)

		toWrite = fmt.Sprintf("\n\t\tself.%s = value\n", snakeName)
		ret.WriteString(toWrite)
	}

	return ret.String()
}

func generatePythonStringSetter(schema *StringSchema) string {

	var ret bytes.Buffer
	var toWrite string

	ret.WriteString(generatePythonNullCheck())

	if schema.MinLength != nil {
		ret.WriteString(generatePythonRangeCheck(*schema.MinLength, "value.length", "was shorter than allowable minimum", "%d", false, "<", ""))
	}

	if schema.MaxLength != nil {
		ret.WriteString(generatePythonRangeCheck(*schema.MaxLength, "value.length", "was longer than allowable maximum", "%d", false, ">", ""))
	}

	if schema.Pattern != nil {

		toWrite = fmt.Sprintf("\n\t\tif(not re.match(\"%s\", value)):", *schema.Pattern)
		ret.WriteString(toWrite)

		toWrite = fmt.Sprintf("\n\t\t\traise ValueError(\"Value '\" value \"' did not match pattern '%s'\")", *schema.Pattern)
		ret.WriteString(toWrite)
	}
	return ret.String()
}

func generatePythonNumericSetter(schema NumericSchemaType) string {

	var ret bytes.Buffer
	var toWrite string

	if schema.HasMinimum() {
		ret.WriteString(generatePythonRangeCheck(schema.GetMinimum(), "value", "is under the allowable minimum", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<"))
	}

	if schema.HasMaximum() {
		ret.WriteString(generatePythonRangeCheck(schema.GetMaximum(), "value", "is over the allowable maximum", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">"))
	}

	if schema.HasEnum() {
		ret.WriteString(generatePythonEnumCheck(schema, schema.GetEnum(), "", ""))
	}

	if schema.HasMultiple() {

		toWrite = fmt.Sprintf("\n\t\tif(value %% %f != 0):\n\t", schema.GetMultiple())
		ret.WriteString(toWrite)

		toWrite = fmt.Sprintf("\n\t\t\traise ValueError.new(\"Property '\" + value + \"' was not a multiple of %v\")", schema.GetMultiple())
		ret.WriteString(toWrite)

	}
	return ret.String()
}

func generatePythonObjectSetter(schema *ObjectSchema) string {

	return generatePythonNullCheck()
}

func generatePythonArraySetter(schema *ArraySchema) string {

	var ret bytes.Buffer

	ret.WriteString(generatePythonNullCheck())

	if schema.MinItems != nil {
		ret.WriteString(generatePythonRangeCheck(*schema.MinItems, "len(value)", "does not have enough items", "%d", false, "<", ""))
	}

	if schema.MaxItems != nil {
		ret.WriteString(generatePythonRangeCheck(*schema.MaxItems, "len(value)", "does not have enough items", "%d", false, ">", ""))
	}

	return ret.String()
}

func generatePythonRangeCheck(value interface{}, reference, message, format string, exclusive bool, comparator, exclusiveComparator string) string {

	var ret bytes.Buffer
	var toWrite, compareString string

	if exclusive {
		compareString = exclusiveComparator
	} else {
		compareString = comparator
	}

	toWrite = "\n\t\tif(" + reference + " " + compareString + " " + format + "):"
	toWrite = fmt.Sprintf(toWrite, value)
	ret.WriteString(toWrite)

	toWrite = fmt.Sprintf("\n\t\t\traise ValueError(\"Property '\"+ value +\"' %s.\")\n", message)
	ret.WriteString(toWrite)

	return ret.String()
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generatePythonEnumCheck(schema TypeSchema, enumValues []interface{}, prefix string, postfix string) string {

	var ret bytes.Buffer
	var stringValues []string
	var constraint string
	var length int

	length = len(enumValues)

	if length <= 0 {
		return ""
	}

	// convert enum values to strings
	for _, enum := range enumValues {
		stringValues = append(stringValues, fmt.Sprintf("%v", enum))
	}

	// write array of valid values
	constraint = fmt.Sprintf("\n\t\tvalidValues = [%s]\n", strings.Join(stringValues, ","))
	ret.WriteString(constraint)

	// compare
	ret.WriteString("\n\t\tif(value not in validValues):")
	ret.WriteString("\n\t\t\traise ValueError(\"Given value '\"+value+\"' was not found in list of acceptable values\")\n")

	return ret.String()
}

func generatePythonNullCheck() string {

	var ret bytes.Buffer

	ret.WriteString("\n\t\tif(value == None):")
	ret.WriteString("\n\t\t\traise ValueError(\"Cannot set property to null value\")")

	return ret.String()
}
