package presilo

import (
	"fmt"
	"strings"
	"bytes"
)

func GenerateRuby(schema *ObjectSchema, module string) string {

	var ret bytes.Buffer
	var toWrite string

	toWrite = fmt.Sprintf("module %s\n\n", ToCamelCase(module))
	ret.WriteString(toWrite)

	ret.WriteString(generateRubySignature(schema))
	ret.WriteString("\n")
	ret.WriteString(generateRubyConstructor(schema))
	ret.WriteString("\n")
	ret.WriteString(generateRubyFunctions(schema))
	ret.WriteString("\nend\nend\n")

	return ret.String()
}

func generateRubySignature(schema *ObjectSchema) string {

	var ret bytes.Buffer
	var subschema TypeSchema
	var readers, accessors []string
	var propertyName string
	var toWrite string

	toWrite = fmt.Sprintf("class %s\n", ToCamelCase(schema.Title))
	ret.WriteString(toWrite)

	for propertyName, subschema = range schema.Properties {

		propertyName = ToSnakeCase(propertyName)

		if(subschema.HasConstraints()) {
			toWrite = fmt.Sprintf(":%s", propertyName)
			readers = append(readers, toWrite)

		} else {

			toWrite = fmt.Sprintf(":%s", propertyName)
			accessors = append(accessors, toWrite)
		}
	}

	if(len(readers) > 0) {
		ret.WriteString("\n\tattr_reader ")
		ret.WriteString(strings.Join(readers, ",\n\t\t\t\t\t\t\t"))
	}

	if(len(accessors) > 0) {
		ret.WriteString("\n\tattr_accessor ")
		ret.WriteString(strings.Join(accessors, ",\n\t\t\t\t\t\t\t\t")) // god. Ruby.
	}

	return ret.String()
}

func generateRubyConstructor(schema *ObjectSchema) string {

	var ret bytes.Buffer
	var declarations, setters []string
	var propertyName string
	var toWrite string

	ret.WriteString("\n\tdef initialize(")

	for _, propertyName = range schema.RequiredProperties {

		propertyName = ToSnakeCase(propertyName)

		declarations = append(declarations, propertyName)

		toWrite = fmt.Sprintf("\n\t\tset_%s(%s)", propertyName, propertyName)
		setters = append(setters, toWrite)
	}

	toWrite = strings.Join(declarations, ",")
	ret.WriteString(toWrite)
	ret.WriteString(")")

	for _, setter := range setters {
		ret.WriteString(setter)
	}

	ret.WriteString("\n\tend\n")
	return ret.String()
}

func generateRubyFunctions(schema *ObjectSchema) string {

	var ret bytes.Buffer
	var subschema TypeSchema
	var toWrite string
	var propertyName, snakeName string

	for propertyName, subschema = range schema.Properties {

		snakeName = ToSnakeCase(propertyName)

		// setter
		toWrite = fmt.Sprintf("\n\tdef set_%s(%s)", snakeName, snakeName)
		ret.WriteString(toWrite)

		switch subschema.GetSchemaType() {
		case SCHEMATYPE_BOOLEAN:
			toWrite = ""
		case SCHEMATYPE_STRING:
			toWrite = generateRubyStringSetter(subschema.(*StringSchema))
		case SCHEMATYPE_INTEGER:
			fallthrough
		case SCHEMATYPE_NUMBER:
			toWrite = generateRubyNumericSetter(subschema.(NumericSchemaType))
		case SCHEMATYPE_OBJECT:
			toWrite = generateRubyObjectSetter(subschema.(*ObjectSchema))
		case SCHEMATYPE_ARRAY:
			toWrite = generateRubyArraySetter(subschema.(*ArraySchema))
		}

		ret.WriteString(toWrite)

		toWrite = fmt.Sprintf("\n\t\t@%s = value", snakeName)
		ret.WriteString(toWrite)

		ret.WriteString("\n\tend\n")
	}

	return ret.String()
}

func generateRubyStringSetter(schema *StringSchema) string {

	var ret bytes.Buffer
	var toWrite string

	ret.WriteString(generateRubyNullCheck())

	if schema.MinLength != nil {
		ret.WriteString(generateRubyRangeCheck(*schema.MinLength, "value.length", "was shorter than allowable minimum", "%d", false, "<", ""))
	}

	if schema.MaxLength != nil {
		ret.WriteString(generateRubyRangeCheck(*schema.MaxLength, "value.length", "was longer than allowable maximum", "%d", false, ">", ""))
	}

	if schema.Pattern != nil {

		toWrite = fmt.Sprintf("\n\t\tif(value =~ /%s/)\n", *schema.Pattern)
		ret.WriteString(toWrite)

		toWrite = fmt.Sprintf("\n\t\t\traise StandardError.new(\"Value '#{value}' did not match pattern '%s'\")", *schema.Pattern)
		ret.WriteString(toWrite)

		ret.WriteString("\n\t\tend")
	}
	return ret.String()
}

func generateRubyNumericSetter(schema NumericSchemaType) string {

	var ret bytes.Buffer
	var toWrite string

	if schema.HasMinimum() {
		ret.WriteString(generateRubyRangeCheck(schema.GetMinimum(), "value", "is under the allowable minimum", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<"))
	}

	if schema.HasMaximum() {
		ret.WriteString(generateRubyRangeCheck(schema.GetMaximum(), "value", "is over the allowable maximum", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">"))
	}

	if schema.HasEnum() {
		ret.WriteString(generateRubyEnumCheck(schema, schema.GetEnum(), "", ""))
	}

	if schema.HasMultiple() {

		toWrite = fmt.Sprintf("\n\tif(value %% %f != 0)\n\t", schema.GetMultiple())
		ret.WriteString(toWrite)

		toWrite = fmt.Sprintf("\n\t\traise StandardError.new(\"Property '#{value}' was not a multiple of %v\")", schema.GetMultiple())
		ret.WriteString(toWrite)

		ret.WriteString("\n\tend\n")
	}
	return ret.String()
}

func generateRubyObjectSetter(schema *ObjectSchema) string {

	return generateRubyNullCheck()
}

func generateRubyArraySetter(schema *ArraySchema) string {

	var ret bytes.Buffer

	ret.WriteString(generateRubyNullCheck())

	if schema.MinItems != nil {
		ret.WriteString(generateRubyRangeCheck(*schema.MinItems, "value.Length", "does not have enough items", "%d", false, "<", ""))
	}

	if schema.MaxItems != nil {
		ret.WriteString(generateRubyRangeCheck(*schema.MaxItems, "value.Length", "does not have enough items", "%d", false, ">", ""))
	}

	return ret.String()
}

func generateRubyRangeCheck(value interface{}, reference, message, format string, exclusive bool, comparator, exclusiveComparator string) string {

	var ret bytes.Buffer
	var toWrite, compareString string

	if exclusive {
		compareString = exclusiveComparator
	} else {
		compareString = comparator
	}

	toWrite = "\n\t\tif(" + reference + " " + compareString + " " + format + ")"
	toWrite = fmt.Sprintf(toWrite, value)
	ret.WriteString(toWrite)

	toWrite = fmt.Sprintf("\n\t\t\traise StandardError.new(\"Property '#{value}' %s.\")\n\t\tend\n", message)
	ret.WriteString(toWrite)

	return ret.String()
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generateRubyEnumCheck(schema TypeSchema, enumValues []interface{}, prefix string, postfix string) string {

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
	ret.WriteString("\n\t\tunless(validValues.include?(value))")
	ret.WriteString("\n\t\t\traise StandardError.new(\"Given value '#{value}' was not found in list of acceptable values\")\n")
	ret.WriteString("\t\tend\n")

	return ret.String()
}

func generateRubyNullCheck() string {

	var ret bytes.Buffer

	ret.WriteString("\n\t\tif(value == nil)")
	ret.WriteString("\n\t\t\traise StandardError.new(\"Cannot set property to null value\")")
	ret.WriteString("\n\t\tend\n")

	return ret.String()
}
