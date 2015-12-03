package presilo

import (
	"fmt"
	"strings"
)

func GenerateRuby(schema *ObjectSchema, module string, tabstyle string) string {

	var buffer *BufferedFormatString

	buffer = NewBufferedFormatString(tabstyle)
	buffer.Printf("module %s\n", ToCamelCase(module))
	buffer.AddIndentation(1)

	generateRubySignature(schema, buffer)
	buffer.Print("\n")
	generateRubyConstructor(schema, buffer)
	buffer.Print("\n")
	generateRubySerializers(schema, buffer)
	buffer.Print("\n")
	generateRubyFunctions(schema, buffer)

	buffer.AddIndentation(-1)
	buffer.Print("\nend")
	buffer.AddIndentation(-1)
	buffer.Print("\nend")

	return buffer.String()
}

func generateRubySignature(schema *ObjectSchema, buffer *BufferedFormatString) {

	var subschema TypeSchema
	var readers, accessors []string
	var propertyName string
	var toWrite string

	buffer.Printf("\nclass %s", ToCamelCase(schema.Title))
	buffer.AddIndentation(1)

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

		buffer.Print("\nattr_reader ")
		buffer.AddIndentation(6)
		buffer.Print(strings.Join(readers, ",\n"))
		buffer.AddIndentation(-6)
	}

	if(len(accessors) > 0) {

		buffer.Print("\nattr_accessor ")
		buffer.AddIndentation(7)
		buffer.Print(strings.Join(accessors, ",\n"))
		buffer.AddIndentation(-7)
	}
}

func generateRubyConstructor(schema *ObjectSchema, buffer *BufferedFormatString) {

	var declarations []string
	var propertyName string

	buffer.Print("\ndef initialize(")

	for _, propertyName = range schema.RequiredProperties {

		propertyName = ToSnakeCase(propertyName)
		declarations = append(declarations, propertyName)
	}

	buffer.Printf("%s)\n", strings.Join(declarations, ","))
	buffer.AddIndentation(1)

	for _, propertyName = range schema.RequiredProperties {
		buffer.Printf("\nset_%s(%s)", propertyName, propertyName)
	}

	buffer.AddIndentation(-1)
	buffer.Print("\nend\n")
}

func generateRubySerializers(schema *ObjectSchema, buffer *BufferedFormatString) {

	var title string

	title = ToCamelCase(schema.GetTitle())

	// serialize
	buffer.Printf("\n# Serializes and returns a hash of this %s.", title)
	buffer.Print("\ndef to_hash()")
	buffer.AddIndentation(1)

	buffer.Print("\nret = {}")
	buffer.Print("\ninstance_variables.each {|field|")
	buffer.AddIndentation(1)
	buffer.Print("\nfield_name = field.to_s().delete(\"@\")")
	buffer.Print("\nfield_value = instance_variable_get(field)")
	buffer.Print("\n\nif field_value.methods.include? 'to_hash'")
	buffer.AddIndentation(1)
	buffer.Print("\nret[field_name] = field_value.to_hash()")
	buffer.Print("\nnext")
	buffer.AddIndentation(-1)
	buffer.Print("\nend")

	buffer.Print("\nret[field_name] = field_value")

	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
	buffer.Print("\nreturn ret")
	buffer.AddIndentation(-1)
	buffer.Print("\nend")

	// deserialize
	/*
	buffer.Printf("\n# Deserializes and returns a new %s from the given hash.", title)
	buffer.Print("\ndef self.from_hash(hash)")
	buffer.AddIndentation(1)

	buffer.Printf("\nret = %s.new()", title)
	buffer.Print("\nreturn ret")

	buffer.AddIndentation(-1)
	buffer.Print("\nend")
	*/
}

func generateRubyFunctions(schema *ObjectSchema, buffer *BufferedFormatString) {

	var subschema TypeSchema
	var propertyName, snakeName, description string

	for propertyName, subschema = range schema.Properties {

		snakeName = ToSnakeCase(propertyName)
		description = subschema.GetDescription()
		description = strings.Replace(description, "\n", "\n# ", -1)

		// getter
		if(len(description) > 0) {
			buffer.Printf("\n# Gets the value of %s, which is defined as:\n# %s", snakeName, description)
		}

		buffer.Printf("\ndef get_%s()", snakeName)
		buffer.AddIndentation(1)

		buffer.Printf("\nreturn @%s", snakeName)

		buffer.AddIndentation(-1)
		buffer.Print("\nend\n")

		// setter
		if(len(description) > 0) {
			buffer.Printf("\n# Sets the value of %s, which is defined as:\n# %s", snakeName, description)
		}

		buffer.Printf("\ndef set_%s(%s)", snakeName, snakeName)
		buffer.AddIndentation(1)

		switch subschema.GetSchemaType() {
		case SCHEMATYPE_STRING:
			generateRubyStringSetter(subschema.(*StringSchema), buffer)
		case SCHEMATYPE_INTEGER:
			fallthrough
		case SCHEMATYPE_NUMBER:
			generateRubyNumericSetter(subschema.(NumericSchemaType), buffer)
		case SCHEMATYPE_OBJECT:
			generateRubyObjectSetter(subschema.(*ObjectSchema), buffer)
		case SCHEMATYPE_ARRAY:
			generateRubyArraySetter(subschema.(*ArraySchema), buffer)
		}

		buffer.Printf("\n@%s = value", snakeName)
		buffer.AddIndentation(-1)
		buffer.Print("\nend\n")
	}
}

func generateRubyStringSetter(schema *StringSchema, buffer *BufferedFormatString) {

	generateRubyNullCheck(buffer)

	if schema.MinLength != nil {
		generateRubyRangeCheck(*schema.MinLength, "value.length", "was shorter than allowable minimum", "%d", false, "<", "", buffer)
	}

	if schema.MaxLength != nil {
		generateRubyRangeCheck(*schema.MaxLength, "value.length", "was longer than allowable maximum", "%d", false, ">", "", buffer)
	}

	if schema.HasEnum() {
		generateRubyEnumCheck(schema, buffer, schema.GetEnum(), "'", "'")
	}

	if schema.Pattern != nil {

		buffer.Printf("\nif(value =~ /%s/)\n", *schema.Pattern)
		buffer.AddIndentation(1)

		buffer.Printf("\nraise StandardError.new(\"Value '#{value}' did not match pattern '%s'\")", *schema.Pattern)

		buffer.AddIndentation(-1)
		buffer.Print("\nend")
	}
}

func generateRubyNumericSetter(schema NumericSchemaType, buffer *BufferedFormatString) {

	if schema.HasMinimum() {
		generateRubyRangeCheck(schema.GetMinimum(), "value", "is under the allowable minimum", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<", buffer)
	}

	if schema.HasMaximum() {
		generateRubyRangeCheck(schema.GetMaximum(), "value", "is over the allowable maximum", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">", buffer)
	}

	if schema.HasEnum() {
		generateRubyEnumCheck(schema, buffer, schema.GetEnum(), "", "")
	}

	if schema.HasMultiple() {

		buffer.Printf("\nif(value %% %f != 0)", schema.GetMultiple())
		buffer.AddIndentation(1)

		buffer.Printf("\nraise StandardError.new(\"Property '#{value}' was not a multiple of %v\")", schema.GetMultiple())

		buffer.AddIndentation(-1)
		buffer.Print("\nend\n")
	}
}

func generateRubyObjectSetter(schema *ObjectSchema, buffer *BufferedFormatString) {

	generateRubyNullCheck(buffer)
}

func generateRubyArraySetter(schema *ArraySchema, buffer *BufferedFormatString) {

	generateRubyNullCheck(buffer)

	if schema.MinItems != nil {
		generateRubyRangeCheck(*schema.MinItems, "value.Length", "does not have enough items", "%d", false, "<", "", buffer)
	}

	if schema.MaxItems != nil {
		generateRubyRangeCheck(*schema.MaxItems, "value.Length", "does not have enough items", "%d", false, ">", "", buffer)
	}
}

func generateRubyRangeCheck(value interface{}, reference, message, format string, exclusive bool, comparator, exclusiveComparator string, buffer *BufferedFormatString) {

	var compareString string

	if exclusive {
		compareString = exclusiveComparator
	} else {
		compareString = comparator
	}

	buffer.Printf("\nif(%s %s " + format + ")", value, compareString)
	buffer.AddIndentation(1)

	buffer.Printf("\nraise StandardError.new(\"Property '#{value}' %s.\")", message)

	buffer.AddIndentation(-1)
	buffer.Print("\nend\n")
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generateRubyEnumCheck(schema TypeSchema, buffer *BufferedFormatString, enumValues []interface{}, prefix string, postfix string) {

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
	buffer.Print("\nunless(validValues.include?(value))")
	buffer.AddIndentation(1)

	buffer.Print("\nraise StandardError.new(\"Given value '#{value}' was not found in list of acceptable values\")")

	buffer.AddIndentation(-1)
	buffer.Print("\nend\n")
}

func generateRubyNullCheck(buffer *BufferedFormatString) {

	buffer.Print("\nif(value == nil)")
	buffer.AddIndentation(1)

	buffer.Print("\nraise StandardError.new(\"Cannot set property to null value\")")

	buffer.AddIndentation(-1)
	buffer.Print("\nend\n")
}
