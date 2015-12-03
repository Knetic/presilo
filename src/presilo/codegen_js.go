package presilo

import (
	"strings"
)

/*
  Generates valid JS code for a given schema.
*/
func GenerateJS(schema *ObjectSchema, module string, tabstyle string) string {

	var buffer *BufferedFormatString

	buffer = NewBufferedFormatString(tabstyle)

	generateJSModuleCheck(buffer, module)
	buffer.Print("\n")
	generateJSConstructor(schema, buffer, module)
	buffer.Print("\n")
	generateJSFunctions(schema, buffer, module)
	buffer.Print("\n")

	return buffer.String()
}

func generateJSModuleCheck(buffer *BufferedFormatString, module string) {

	// check for undefined, first.
	buffer.Printf("\nif(typeof(%s) === \"undefined\")\n{", module)
	buffer.AddIndentation(1)

	buffer.Printf("\n%s = {}", module)

	buffer.AddIndentation(-1)
	buffer.Print("\n}")

	// then null check.
	buffer.Print("\nelse\n{")
	buffer.AddIndentation(1)

	buffer.Printf("\n%s = %s || {}", module, module)

	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
}

func generateJSConstructor(schema *ObjectSchema, buffer *BufferedFormatString, module string) {

	var parameterNames []string
	var propertyName, parameterName string

	// generate list of property names
	for _, propertyName = range schema.RequiredProperties {

		propertyName = ToJavaCase(propertyName)
		parameterNames = append(parameterNames, propertyName)
	}

	// write constructor signature
	buffer.Printf("\n/*\n%s\n*/\n", schema.Description)

	buffer.Printf("\n%s.%s = function(", module, schema.Title)

	buffer.Print(strings.Join(parameterNames, ","))
	buffer.Print(")\n{")

	buffer.AddIndentation(1)

	// body
	for _, parameterName = range parameterNames {
		buffer.Printf("\nthis.set%s(%s)", ToCamelCase(parameterName), parameterName)
	}

	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
}

func generateJSFunctions(schema *ObjectSchema, buffer *BufferedFormatString, module string) {

	var subschema TypeSchema
	var propertyName, propertyNameCamel, propertyNameJava, schemaName string

	schemaName = ToCamelCase(schema.Title)

	for propertyName, subschema = range schema.Properties {

		propertyNameCamel = ToCamelCase(propertyName)
		propertyNameJava = ToJavaCase(propertyName)

		buffer.Printf("\n%s.%s.prototype.set%s = function(value)\n{", module, schemaName, propertyNameCamel)
		buffer.AddIndentation(1)

		// undefined check
		buffer.Printf("\nif(typeof(value) === 'undefined')\n{")
		buffer.AddIndentation(1)

		buffer.Printf("\nthrow new ReferenceError(\"Cannot set property '%s', no value given\")", propertyNameJava)

		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")

		switch subschema.GetSchemaType() {
		case SCHEMATYPE_STRING:
			generateJSStringSetter(subschema.(*StringSchema), buffer)
		case SCHEMATYPE_INTEGER:
			fallthrough
		case SCHEMATYPE_NUMBER:
			generateJSNumericSetter(subschema.(NumericSchemaType), buffer)
		case SCHEMATYPE_OBJECT:
			generateJSObjectSetter(subschema.(*ObjectSchema), buffer)
		case SCHEMATYPE_ARRAY:
			generateJSArraySetter(subschema.(*ArraySchema), buffer)
		}

		buffer.Printf("\nthis.%s = value;", propertyNameJava)
		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}
}

/*
	Returns checks appropriate for verifying an object's type.
*/
func generateJSObjectSetter(schema *ObjectSchema, buffer *BufferedFormatString) {

	generateJSTypeCheck(schema, buffer)
}

/*
	Returns checks appropriate for verifying a numeric value and its constraints.
*/
func generateJSNumericSetter(schema NumericSchemaType, buffer *BufferedFormatString) {

	generateJSTypeCheck(schema, buffer)

	if schema.HasMinimum() {
		generateJSRangeCheck(schema.GetMinimum(), "value", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<", buffer)
	}

	if schema.HasMaximum() {
		generateJSRangeCheck(schema.GetMaximum(), "value", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">", buffer)
	}

	if schema.HasEnum() {
		generateJSEnumCheck(schema, buffer, schema.GetEnum(), "", "")
	}

	if schema.HasMultiple() {

		buffer.Printf("\nif(value %% %f != 0)\n{", schema.GetMultiple())
		buffer.AddIndentation(1)

		buffer.Printf("\nthrow new Error(\"Property '\"+value+\"' was not a multiple of %s\")", schema.GetMultiple())

		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}
}

/*
	Returns checks appropriate for verifying a string value and its constraints.
*/
func generateJSStringSetter(schema *StringSchema, buffer *BufferedFormatString) {

	generateJSTypeCheck(schema, buffer)

	if schema.MinLength != nil {
		generateJSRangeCheck(*schema.MinLength, "value.length", "%d", false, "<", "", buffer)
	}

	if schema.MaxLength != nil {
		generateJSRangeCheck(*schema.MaxLength, "value.length", "%d", false, ">", "", buffer)
	}

	if schema.Pattern != nil {

		buffer.Printf("\nvar regex = new RegExp(\"%s\")", *schema.Pattern)
		buffer.Printf("\nif(!regex.test(value))\n{")
		buffer.AddIndentation(1)

		buffer.Printf("\nthrow new Error(\"Property '\"+value+\"' did not match pattern '%s'\")", *schema.Pattern)

		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}

	if schema.Enum != nil {
		generateJSEnumCheck(schema, buffer, schema.GetEnum(), "\"", "\"")
	}
}

/*
	Returns checks appropriate for verifying an array value and its constraints.
*/
func generateJSArraySetter(schema *ArraySchema, buffer *BufferedFormatString) {

	generateJSTypeCheck(schema, buffer)
	// TODO: value uniformity check

	if schema.MinItems != nil {
		generateJSRangeCheck(*schema.MinItems, "value.length", "%d", false, "<", "", buffer)
	}

	if schema.MaxItems != nil {
		generateJSRangeCheck(*schema.MaxItems, "value.length", "%d", false, ">", "", buffer)
	}
}

func generateJSRangeCheck(value interface{}, reference string, format string, exclusive bool, comparator, exclusiveComparator string, buffer *BufferedFormatString) {

	var compareString string

	if exclusive {
		compareString = exclusiveComparator
	} else {
		compareString = comparator
	}

	buffer.Printf("\nif(%s %s "+format+")\n{", reference, compareString, value)
	buffer.AddIndentation(1)

	buffer.Printf("\nthrow new RangeError(\"Property '\"+value+\"' was out of range.\")")
	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
}

/*
	Generates code which throws an error if the given [parameter]'s type name is not equal to the given [typeName]
*/
func generateJSTypeCheck(schema TypeSchema, buffer *BufferedFormatString) {

	var schemaType SchemaType
	var expectedType string
	var shouldWriteCtorCheck bool

	schemaType = schema.GetSchemaType()
	expectedType = getJSTypeFromSchemaType(schemaType)

	buffer.Printf("\nif(typeof(value) !== \"%s\")\n{", expectedType)
	buffer.AddIndentation(1)

	buffer.Printf("\nthrow new TypeError(\"Property \"+value+\" was not of the expected type '%s'\")", expectedType)

	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")

	// if this is an array or object, check the constructor
	shouldWriteCtorCheck = false

	switch schemaType {
	case SCHEMATYPE_ARRAY:
		shouldWriteCtorCheck = true
		expectedType = "Array"
	case SCHEMATYPE_OBJECT:
		shouldWriteCtorCheck = true
		expectedType = ToCamelCase(schema.GetTitle())
	}

	if shouldWriteCtorCheck {

		buffer.Printf("\nif(value.constructor !== %s)\n{", expectedType)
		buffer.AddIndentation(1)

		buffer.Printf("\nthrow new TypeError(\"Property '\"+value+\"'was not of the expected type '%s'\")", expectedType)

		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generateJSEnumCheck(schema interface{}, buffer *BufferedFormatString, enumValues []interface{}, prefix string, postfix string) {

	var length int

	length = len(enumValues)

	if length <= 0 {
		return
	}

	// write array of valid values
	buffer.Printf("\nvar validValues = [%s%v%s", prefix, enumValues[0], postfix)

	for _, enumValue := range enumValues[1:length] {
		buffer.Printf(",%s%v%s", prefix, enumValue, postfix)
	}
	buffer.Print("]\n")

	// compare
	buffer.Print("\nvar isValid = false")
	buffer.Print("\nfor(var i = 0; i < validValues.length; i++) \n{")
	buffer.AddIndentation(1)

	buffer.Print("\nif(validValues[i] === value)\n{\nisValid = true")
	buffer.Print("\nbreak;")
	buffer.AddIndentation(-1)
	buffer.Print("\n}")
	buffer.AddIndentation(-1)
	buffer.Print("\n}")

	buffer.Print("\nif(!isValid)\n{")
	buffer.AddIndentation(1)
	buffer.Print("\nthrow new Error(\"Given value '\"+value+\"' was not found in list of acceptable values\")")
	buffer.AddIndentation(-1)
	buffer.Print("\n}")
}

func getJSTypeFromSchemaType(schemaType SchemaType) string {

	switch schemaType {
	case SCHEMATYPE_NUMBER:
		fallthrough
	case SCHEMATYPE_INTEGER:
		return "number"
	case SCHEMATYPE_ARRAY:
		fallthrough
	case SCHEMATYPE_OBJECT:
		return "object"
	case SCHEMATYPE_STRING:
		return "string"
	case SCHEMATYPE_BOOLEAN:
		return "boolean"
	}

	return "object"
}
