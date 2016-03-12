package presilo

import (
	"fmt"
	"regexp"
	"strings"
)

func GeneratePython(schema *ObjectSchema, module string, tabstyle string) string {

	var ret *BufferedFormatString

	ret = NewBufferedFormatString(tabstyle)

	generatePythonImports(schema, ret)
	ret.Printfln("")
	generatePythonSignature(schema, ret)
	ret.Printfln("")
	generatePythonConstructor(schema, ret)
	ret.Printfln("")
	generatePythonDeserializer(schema, ret)
	ret.Printfln("")
	generatePythonSerializer(schema, ret)
	ret.Printfln("")
	generatePythonFunctions(schema, ret)
	ret.Printfln("")

	return ret.String()
}

func ValidatePythonModule(module string) bool {

	// ref: https://www.python.org/dev/peps/pep-0008/#package-and-module-names
	pattern := "^[a-z0-9_]+$"
	matched, err := regexp.MatchString(pattern, module)
	return err == nil && matched
}

func generatePythonImports(schema *ObjectSchema, buffer *BufferedFormatString) {

	buffer.Printfln("import string")
	buffer.Printfln("import json")

	if containsRegexpMatch(schema) {
		buffer.Printfln("import re")
	}
}

func generatePythonSignature(schema *ObjectSchema, buffer *BufferedFormatString) {

	var description string

	description = schema.GetDescription()

	if len(description) != 0 {
		buffer.Printfln("'''\n%s\n'''\n", schema.GetDescription())
	}

	buffer.Printfln("class %s(object):", ToCamelCase(schema.Title))
	buffer.AddIndentation(1)
}

func generatePythonConstructor(schema *ObjectSchema, buffer *BufferedFormatString) {

	var declarations, setters []string
	var propertyName string
	var toWrite string

	if(len(schema.RequiredProperties) <= 0) {
		return
	}

	buffer.Print("\ndef __init__(self")

	// required properties
	for _, propertyName = range schema.RequiredProperties {

		propertyName = ToSnakeCase(propertyName)
		declarations = append(declarations, propertyName)

		toWrite = fmt.Sprintf("\nself.set_%s(%s)", ToSnakeCase(propertyName), propertyName)
		setters = append(setters, toWrite)
	}

	buffer.Print(", ")
	buffer.Print(strings.Join(declarations, ", "))
	buffer.Print("):")

	// use setters
	buffer.AddIndentation(1)

	for _, setter := range setters {
		buffer.Print(setter)
	}

	buffer.AddIndentation(-1)
}

func generatePythonDeserializer(schema *ObjectSchema, buffer *BufferedFormatString) {

	var property TypeSchema
	var ctorArguments []string
	var argument string
	var className string
	var propertyName, casedPropertyName string

	className = ToCamelCase(schema.GetTitle())

	buffer.Printf("\n@staticmethod")
	buffer.Printf("\ndef deserialize_from(map):")
	buffer.AddIndentation(1)

	// use constructor
	buffer.Printf("\nret = %s(", className)

	for _, propertyName = range schema.RequiredProperties {

		argument = fmt.Sprintf("map[\"%s\"]", ToJavaCase(propertyName))
		ctorArguments = append(ctorArguments, argument)
	}

	buffer.Printf("%s)", strings.Join(ctorArguments, ", "))

	// misc setters
	buffer.Printf("\n")
	for _, propertyName = range schema.GetOrderedPropertyNames() {

		property = schema.Properties[propertyName]
		propertyName = ToJavaCase(property.GetTitle())
		casedPropertyName = fmt.Sprintf("map[\"%s\"]", propertyName)

		// if it's already set, skip it.
		if arrayContainsString(ctorArguments, casedPropertyName) {
			continue
		}

		// if it's constrained, use the setter
		if property.HasConstraints() {

			casedPropertyName = ToJavaCase(propertyName)
			buffer.Printf("\nret.set_%s(%s)", ToSnakeCase(propertyName), casedPropertyName)
			continue
		}

		// otherwise set.
		buffer.Printf("\nret.%s = %s", propertyName, casedPropertyName)
	}

	buffer.Printf("\nreturn ret")
	buffer.AddIndentation(-1)
	buffer.Printf("\n")
}

func generatePythonSerializer(schema *ObjectSchema, buffer *BufferedFormatString) {


	buffer.Printf("\ndef to_json(self):")
	buffer.AddIndentation(1)
	buffer.Printf("\nreturn json.dumps(self, default=lambda o: o.__dict__, sort_keys=True, indent=4)")
	buffer.AddIndentation(-1)
}

func generatePythonFunctions(schema *ObjectSchema, buffer *BufferedFormatString) {

	var subschema TypeSchema
	var propertyName, snakeName, description string

	for _, propertyName = range schema.GetOrderedPropertyNames() {

		subschema = schema.Properties[propertyName]
		snakeName = ToSnakeCase(propertyName)
		description = subschema.GetDescription()

		// getter
		buffer.Printf("\ndef get_%s(self):", snakeName)
		buffer.AddIndentation(1)

		if len(description) != 0 {
			buffer.Print("\n'''")
			buffer.AddIndentation(1)

			buffer.Printf("\nGets %s, defined as:\n%s\n", snakeName, description)

			buffer.AddIndentation(-1)
			buffer.Print("\n'''")
		}

		buffer.Printf("\nreturn self.%s\n", snakeName)
		buffer.AddIndentation(-1)

		// setter
		buffer.Printf("\ndef set_%s(self, value):", snakeName)
		buffer.AddIndentation(1)

		if len(description) != 0 {
			buffer.Print("\n'''")
			buffer.AddIndentation(1)

			buffer.Printf("\nSets %s, defined as:\n%s\n", snakeName, description)

			buffer.AddIndentation(-1)
			buffer.Print("\n'''")
		}

		switch subschema.GetSchemaType() {
		case SCHEMATYPE_STRING:
			generatePythonStringSetter(subschema.(*StringSchema), buffer)
		case SCHEMATYPE_INTEGER:
			fallthrough
		case SCHEMATYPE_NUMBER:
			generatePythonNumericSetter(subschema.(NumericSchemaType), buffer)
		case SCHEMATYPE_OBJECT:
			generatePythonObjectSetter(subschema.(*ObjectSchema), buffer)
		case SCHEMATYPE_ARRAY:
			generatePythonArraySetter(subschema.(*ArraySchema), buffer)
		}

		buffer.Printf("\nself.%s = value\n", snakeName)
		buffer.AddIndentation(-1)
	}
}

func generatePythonStringSetter(schema *StringSchema, buffer *BufferedFormatString) {

	if !schema.Nullable {
		generatePythonNullCheck(buffer)
	}

	if schema.MinLength != nil {
		generatePythonRangeCheck(*schema.MinLength, "len(value)", "was shorter than allowable minimum", "%d", false, "<", "", buffer)
	}

	if schema.MaxLength != nil {
		generatePythonRangeCheck(*schema.MaxLength, "len(value)", "was longer than allowable maximum", "%d", false, ">", "", buffer)
	}

	if schema.HasEnum() {
		generatePythonEnumCheck(schema, buffer, schema.GetEnum(), "\"", "\"")
	}

	if schema.Pattern != nil {

		buffer.Printf("\nif(not re.match(\"%s\", value)):", *schema.Pattern)
		buffer.AddIndentation(1)

		buffer.Printf("\nraise ValueError(\"Value '\" + value + \"' did not match pattern '%s'\")", *schema.Pattern)

		buffer.AddIndentation(-1)
	}
}

func generatePythonNumericSetter(schema NumericSchemaType, buffer *BufferedFormatString) {

	if schema.HasMinimum() {
		generatePythonRangeCheck(schema.GetMinimum(), "value", "is under the allowable minimum", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<", buffer)
	}

	if schema.HasMaximum() {
		generatePythonRangeCheck(schema.GetMaximum(), "value", "is over the allowable maximum", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">", buffer)
	}

	if schema.HasEnum() {
		generatePythonEnumCheck(schema, buffer, schema.GetEnum(), "", "")
	}

	if schema.HasMultiple() {

		buffer.Printf("\n\t\tif(value %% %f != 0):\n\t", schema.GetMultiple())
		buffer.AddIndentation(1)

		buffer.Printf("\nraise ValueError.new(\"Property '\" + value + \"' was not a multiple of %v\")", schema.GetMultiple())

		buffer.AddIndentation(-1)
	}
}

func generatePythonObjectSetter(schema *ObjectSchema, buffer *BufferedFormatString) {

	if !schema.Nullable {
		generatePythonNullCheck(buffer)
	}
}

func generatePythonArraySetter(schema *ArraySchema, buffer *BufferedFormatString) {

	if !schema.Nullable {
		generatePythonNullCheck(buffer)
	}

	if schema.MinItems != nil {
		generatePythonRangeCheck(*schema.MinItems, "len(value)", "does not have enough items", "%d", false, "<", "", buffer)
	}

	if schema.MaxItems != nil {
		generatePythonRangeCheck(*schema.MaxItems, "len(value)", "does not have enough items", "%d", false, ">", "", buffer)
	}
}

func generatePythonRangeCheck(value interface{}, reference, message, format string, exclusive bool, comparator, exclusiveComparator string, buffer *BufferedFormatString) {

	var compareString string

	if exclusive {
		compareString = exclusiveComparator
	} else {
		compareString = comparator
	}

	buffer.Printf("\nif(%s %s "+format+"):", reference, compareString, value)
	buffer.AddIndentation(1)

	buffer.Printf("\nraise ValueError(\"Property '\"+ value +\"' %s.\")\n", message)

	buffer.AddIndentation(-1)
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generatePythonEnumCheck(schema TypeSchema, buffer *BufferedFormatString, enumValues []interface{}, prefix string, postfix string) {

	var stringValues []string
	var length int

	length = len(enumValues)

	if length <= 0 {
		return
	}

	// convert enum values to strings
	for _, enum := range enumValues {
		stringValues = append(stringValues, fmt.Sprintf("%s%v%s", prefix, enum, postfix))
	}

	// write array of valid values
	buffer.Printf("\nvalidValues = [%s]\n", strings.Join(stringValues, ","))

	// compare
	buffer.Print("\nif(value not in validValues):")
	buffer.AddIndentation(1)

	buffer.Print("\nraise ValueError(\"Given value '\"+value+\"' was not found in list of acceptable values\")\n")

	buffer.AddIndentation(-1)
}

func generatePythonNullCheck(buffer *BufferedFormatString) {

	buffer.Print("\nif(value == None):")
	buffer.AddIndentation(1)
	buffer.Print("\nraise ValueError(\"Cannot set property to null value\")")
	buffer.AddIndentation(-1)
}
