package presilo

import (
	"fmt"
	"regexp"
	"strings"
)

/*
  Generates valid CSharp code for a given schema.

	C# code contains DataContractJsonSerializer annotations,
	which are required for proper serialization/deserialization of fields.
	Unfortunately this isn't available before .NET 4.5, so any generated code
	will need to be compiled with .NET 4.5+
*/
func GenerateCSharp(schema *ObjectSchema, module string, tabstyle string) string {

	var buffer *BufferedFormatString

	buffer = NewBufferedFormatString(tabstyle)

	generateCSharpImports(schema, buffer)
	buffer.Print("\n")
	generateCSharpNamespace(schema, buffer, module)
	buffer.Print("\n")
	generateCSharpTypeDeclaration(schema, buffer)
	buffer.Print("\n")
	generateCSharpConstructor(schema, buffer)
	buffer.Print("\n")
	generateCSharpFunctions(schema, buffer)
	buffer.AddIndentation(-1)
	buffer.Print("\n}")
	buffer.AddIndentation(-1)
	buffer.Print("\n}")

	return buffer.String()
}

func ValidateCSharpModule(module string) bool {

	pattern := "^[a-zA-Z_]+[0-9a-zA-Z_]*(\\.[a-zA-Z_]+[0-9a-zA-Z_]*)*$"
	matched, err := regexp.MatchString(pattern, module)
	return err == nil && matched
}

func generateCSharpImports(schema *ObjectSchema, buffer *BufferedFormatString) {

	buffer.Print("using System;")
	buffer.Print("\n using System.Runtime.Serialization;")

	// import regex if we need it
	if containsRegexpMatch(schema) {
		buffer.Print("\nusing System.Text.RegularExpressions;")
	}

	buffer.Print("\n")
}

func generateCSharpNamespace(schema *ObjectSchema, buffer *BufferedFormatString, module string) {

	buffer.Printf("namespace %s\n{", module)
	buffer.AddIndentation(1)
}

func generateCSharpTypeDeclaration(schema *ObjectSchema, buffer *BufferedFormatString) {

	var subschema TypeSchema
	var propertyName string

	buffer.Print("[DataContract]")
	buffer.Printf("\npublic class %s\n{", ToCamelCase(schema.Title))
	buffer.AddIndentation(1)

	for _, propertyName = range schema.GetOrderedPropertyNames() {

		subschema = schema.Properties[propertyName]

		buffer.Printf("\n[DataMember(Name = \"%s\")]", propertyName)
		buffer.Printf("\nprotected %s %s;", GenerateCSharpTypeForSchema(subschema), ToJavaCase(propertyName))
	}
}

func generateCSharpConstructor(schema *ObjectSchema, buffer *BufferedFormatString) {

	var subschema TypeSchema
	var declarations, setters []string
	var propertyName string
	var toWrite string

	buffer.Printf("\npublic %s(", ToCamelCase(schema.Title))

	for _, propertyName = range schema.RequiredProperties {

		subschema = schema.Properties[propertyName]
		propertyName = ToJavaCase(propertyName)

		toWrite = fmt.Sprintf("%s %s", GenerateCSharpTypeForSchema(subschema), propertyName)
		declarations = append(declarations, toWrite)

		toWrite = fmt.Sprintf("\nset%s(%s);", ToStrictCamelCase(propertyName), propertyName)
		setters = append(setters, toWrite)
	}

	buffer.Print(strings.Join(declarations, ","))
	buffer.Print(")\n{")
	buffer.AddIndentation(1)

	for _, setter := range setters {
		buffer.Print(setter)
	}

	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
}

func generateCSharpFunctions(schema *ObjectSchema, buffer *BufferedFormatString) {

	var subschema TypeSchema
	var propertyName, properName, camelName, typeName string

	for _, propertyName = range schema.GetOrderedPropertyNames() {

		subschema = schema.Properties[propertyName]

		properName = ToJavaCase(propertyName)
		camelName = ToStrictCamelCase(propertyName)
		typeName = GenerateCSharpTypeForSchema(subschema)

		// getter
		buffer.Printf("\npublic %s get%s()\n{", typeName, camelName)
		buffer.AddIndentation(1)

		buffer.Printf("\nreturn this.%s;", properName)

		buffer.AddIndentation(-1)
		buffer.Print("\n}")

		// setter
		buffer.Printf("\npublic void set%s(%s value)\n{", camelName, typeName)
		buffer.AddIndentation(1)

		switch subschema.GetSchemaType() {
		case SCHEMATYPE_STRING:
			generateCSharpStringSetter(subschema.(*StringSchema), buffer)
		case SCHEMATYPE_INTEGER:
			fallthrough
		case SCHEMATYPE_NUMBER:
			generateCSharpNumericSetter(subschema.(NumericSchemaType), buffer)
		case SCHEMATYPE_OBJECT:
			generateCSharpObjectSetter(subschema.(*ObjectSchema), buffer)
		case SCHEMATYPE_ARRAY:
			generateCSharpArraySetter(subschema.(*ArraySchema), buffer)
		}

		buffer.Printf("\nthis.%s = value;", properName)
		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}
}

func generateCSharpStringSetter(schema *StringSchema, buffer *BufferedFormatString) {

	if !schema.Nullable {
		generateCSharpNullCheck(buffer)
	}

	if schema.MinLength != nil {
		generateCSharpRangeCheck(*schema.MinLength, "value.Length", "was shorter than allowable minimum", "%d", false, "<", "", buffer)
	}

	if schema.MaxLength != nil {
		generateCSharpRangeCheck(*schema.MaxLength, "value.Length", "was longer than allowable maximum", "%d", false, ">", "", buffer)
	}

	if schema.MinByteLength != nil {
		generateCSharpRangeCheck(*schema.MinByteLength, "value.Length * sizeof(Char)", "had fewer bytes than allowable minimum", "%d", false, "<", "", buffer)
	}

	if schema.MaxByteLength != nil {
		generateCSharpRangeCheck(*schema.MaxByteLength, "value.Length * sizeof(Char)", "had more bytes than allowable minimum", "%d", false, ">", "", buffer)
	}

	if schema.Pattern != nil {

		buffer.Printf("\nRegex regex = new Regex(\"%s\");", sanitizeQuotedString(*schema.Pattern))
		buffer.Printf("\nif(!regex.IsMatch(value))\n{")
		buffer.AddIndentation(1)

		buffer.Printf("\nthrow new Exception(\"Value '\"+value+\"' did not match pattern '%s'\");", *schema.Pattern)

		buffer.AddIndentation(-1)
		buffer.Print("\n}")
	}
}

func generateCSharpNumericSetter(schema NumericSchemaType, buffer *BufferedFormatString) {

	if schema.HasMinimum() {
		generateCSharpRangeCheck(schema.GetMinimum(), "value", "is under the allowable minimum", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<", buffer)
	}

	if schema.HasMaximum() {
		generateCSharpRangeCheck(schema.GetMaximum(), "value", "is over the allowable maximum", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">", buffer)
	}

	if schema.HasEnum() {
		generateCSharpEnumCheck(schema, buffer, schema.GetEnum(), "", "")
	}

	if schema.HasMultiple() {

		buffer.Printf("\nif(value %% %f != 0)\n{", schema.GetMultiple())
		buffer.AddIndentation(1)

		buffer.Printf("\nthrow new Exception(\"Property '\"+value+\"' was not a multiple of %s\");", schema.GetMultiple())

		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}
}

func generateCSharpObjectSetter(schema *ObjectSchema, buffer *BufferedFormatString) {

	if !schema.Nullable {
		generateCSharpNullCheck(buffer)
	}
}

func generateCSharpArraySetter(schema *ArraySchema, buffer *BufferedFormatString) {

	if !schema.Nullable {
		generateCSharpNullCheck(buffer)
	}

	if schema.MinItems != nil {
		generateCSharpRangeCheck(*schema.MinItems, "value.Length", "does not have enough items", "%d", false, "<", "", buffer)
	}

	if schema.MaxItems != nil {
		generateCSharpRangeCheck(*schema.MaxItems, "value.Length", "does not have enough items", "%d", false, ">", "", buffer)
	}
}

func generateCSharpNullCheck(buffer *BufferedFormatString) {

	buffer.Printf("\nif(value == null)\n{")
	buffer.AddIndentation(1)

	buffer.Print("\nthrow new NullReferenceException(\"Cannot set property to null value\");")

	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
}

func generateCSharpRangeCheck(value interface{}, reference, message, format string, exclusive bool, comparator, exclusiveComparator string, buffer *BufferedFormatString) {

	var compareString string

	if exclusive {
		compareString = exclusiveComparator
	} else {
		compareString = comparator
	}

	buffer.Printf("\nif(%s %s "+format+")\n{", reference, compareString, value)
	buffer.AddIndentation(1)

	buffer.Printf("\nthrow new Exception(\"Property '\"+value+\"' %s.\");", message)
	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generateCSharpEnumCheck(schema TypeSchema, buffer *BufferedFormatString, enumValues []interface{}, prefix string, postfix string) {

	var typeName string
	var length int

	length = len(enumValues)

	if length <= 0 {
		return
	}

	// write array of valid values
	typeName = GenerateCSharpTypeForSchema(schema)
	buffer.Printf("%s[] validValues = new %s[]{%s%v%s", typeName, typeName, prefix, enumValues[0], postfix)

	for _, enumValue := range enumValues[1:length] {
		buffer.Printf(",%s%v%s", prefix, enumValue, postfix)
	}

	buffer.Print("};\n")

	// compare
	buffer.Print("\nbool isValid = false;")
	buffer.Print("\nfor(int i = 0; i < validValues.Length; i++)\n{")
	buffer.AddIndentation(1)

	buffer.Print("\nif(validValues[i] == value)\n{")
	buffer.AddIndentation(1)
	buffer.Print("\nisValid = true;\nbreak;")

	buffer.AddIndentation(-1)
	buffer.Print("\n}")
	buffer.AddIndentation(-1)
	buffer.Print("\n}")

	buffer.Print("\nif(!isValid)\n{")
	buffer.AddIndentation(1)
	buffer.Print("\nthrow new Exception(\"Given value '\"+value+\"' was not found in list of acceptable values\");")

	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
}

func GenerateCSharpTypeForSchema(subschema TypeSchema) string {

	switch subschema.GetSchemaType() {
	case SCHEMATYPE_NUMBER:
		return "double"
	case SCHEMATYPE_INTEGER:
		return "int"
	case SCHEMATYPE_ARRAY:
		return ToCamelCase(subschema.(*ArraySchema).Items.GetTitle()) + "[]"
	case SCHEMATYPE_OBJECT:
		return ToCamelCase(subschema.GetTitle())
	case SCHEMATYPE_STRING:
		return "string"
	case SCHEMATYPE_BOOLEAN:
		return "bool"
	}

	return "Object"
}
