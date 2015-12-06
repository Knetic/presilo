package presilo

import (
	"fmt"
	"regexp"
	"strings"
)

/*
  Generates valid Java code for a given schema.
*/
func GenerateJava(schema *ObjectSchema, module string, tabstyle string) string {

	var buffer *BufferedFormatString

	buffer = NewBufferedFormatString(tabstyle)

	buffer.Printf("package %s;\n", module)

	generateJavaImports(schema, buffer)
	buffer.Print("\n")
	generateJavaTypeDeclaration(schema, buffer)
	buffer.Print("\n")
	generateJavaConstructor(schema, buffer)
	buffer.Print("\n")
	generateJavaFunctions(schema, buffer)

	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")

	return buffer.String()
}

func ValidateJavaModule(module string) bool {

	pattern := "^[a-zA-Z_]+[0-9a-zA-Z_]*(\\.[a-zA-Z_]+[0-9a-zA-Z_]*)*$"
	matched, err := regexp.MatchString(pattern, module)
	return err == nil && matched
}

func generateJavaImports(schema *ObjectSchema, buffer *BufferedFormatString) {

	// import regex if we need it
	if containsRegexpMatch(schema) {
		buffer.Print("import java.util.regex.*;\n\n")
	}
}

func generateJavaTypeDeclaration(schema *ObjectSchema, buffer *BufferedFormatString) {

	buffer.Printf("public class %s\n{", ToCamelCase(schema.Title))
	buffer.AddIndentation(1)

	for propertyName, subschema := range schema.Properties {

		buffer.Printf("\nprotected %s %s;", generateJavaTypeForSchema(subschema), ToJavaCase(propertyName))
	}
}

func generateJavaConstructor(schema *ObjectSchema, buffer *BufferedFormatString) {

	var subschema TypeSchema
	var declarations, setters []string
	var propertyName string
	var toWrite string
	var constrained bool

	buffer.Printf("\npublic %s(", ToCamelCase(schema.Title))

	for _, propertyName = range schema.RequiredProperties {

		subschema = schema.Properties[propertyName]
		propertyName = ToJavaCase(propertyName)

		if subschema.HasConstraints() {
			constrained = true
		}

		toWrite = fmt.Sprintf("%s %s", generateJavaTypeForSchema(subschema), propertyName)
		declarations = append(declarations, toWrite)

		toWrite = fmt.Sprintf("\nset%s(%s);", ToCamelCase(propertyName), propertyName)
		setters = append(setters, toWrite)
	}

	buffer.Print(strings.Join(declarations, ","))
	buffer.Print(")")

	if constrained {
		buffer.Print(" throws Exception")
	}

	buffer.Print("\n{")
	buffer.AddIndentation(1)

	for _, setter := range setters {
		buffer.Print(setter)
	}

	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
}

func generateJavaFunctions(schema *ObjectSchema, buffer *BufferedFormatString) {

	var subschema TypeSchema
	var propertyName, properName, camelName, typeName string

	for propertyName, subschema = range schema.Properties {

		properName = ToJavaCase(propertyName)
		camelName = ToCamelCase(propertyName)
		typeName = generateJavaTypeForSchema(subschema)

		// getter
		buffer.Printf("\npublic %s get%s()\n{", typeName, camelName)
		buffer.AddIndentation(1)

		buffer.Printf("\nreturn this.%s;", properName)

		buffer.AddIndentation(-1)
		buffer.Print("\n}")

		// setter
		buffer.Printf("\npublic void set%s(%s value)", camelName, typeName)

		if subschema.HasConstraints() {
			buffer.Print(" throws Exception")
		}

		buffer.Print("\n{")
		buffer.AddIndentation(1)

		switch subschema.GetSchemaType() {
		case SCHEMATYPE_STRING:
			generateJavaStringSetter(subschema.(*StringSchema), buffer)
		case SCHEMATYPE_INTEGER:
			fallthrough
		case SCHEMATYPE_NUMBER:
			generateJavaNumericSetter(subschema.(NumericSchemaType), buffer)
		case SCHEMATYPE_OBJECT:
			generateJavaObjectSetter(subschema.(*ObjectSchema), buffer)
		case SCHEMATYPE_ARRAY:
			generateJavaArraySetter(subschema.(*ArraySchema), buffer)
		}

		buffer.Printf("\n%s = value;", properName)

		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}
}

func generateJavaStringSetter(schema *StringSchema, buffer *BufferedFormatString) {

	generateJavaNullCheck(buffer)

	if schema.MinLength != nil {
		generateJavaRangeCheck(*schema.MinLength, "value.length()", "was shorter than allowable minimum", "%d", false, "<", "", buffer)
	}

	if schema.MaxLength != nil {
		generateJavaRangeCheck(*schema.MaxLength, "value.length()", "was longer than allowable maximum", "%d", false, ">", "", buffer)
	}

	if schema.Pattern != nil {

		buffer.Printf("\nPattern regex = Pattern.compile(\"%s\");", sanitizeQuotedString(*schema.Pattern))
		buffer.Printf("\nif(!regex.matcher(value).matches())\n{")
		buffer.AddIndentation(1)

		buffer.Printf("\nthrow new Exception(\"Value '\"+value+\"' did not match pattern '%s'\");", *schema.Pattern)

		buffer.AddIndentation(-1)
		buffer.Print("\n}")
	}

	if schema.HasEnum() {
		generateJavaEnumCheck(schema, schema.GetEnum(), "\"", "\"", buffer)
	}
}

func generateJavaNumericSetter(schema NumericSchemaType, buffer *BufferedFormatString) {

	if schema.HasMinimum() {
		generateJavaRangeCheck(schema.GetMinimum(), "value", "is under the allowable minimum", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<", buffer)
	}

	if schema.HasMaximum() {
		generateJavaRangeCheck(schema.GetMaximum(), "value", "is over the allowable maximum", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">", buffer)
	}

	if schema.HasEnum() {
		generateJavaEnumCheck(schema, schema.GetEnum(), "", "", buffer)
	}

	if schema.HasMultiple() {

		buffer.Printf("\nif(value %% %f != 0)\n{", schema.GetMultiple())
		buffer.AddIndentation(1)

		buffer.Printf("\nthrow new Exception(\"Property '\"+value+\"' was not a multiple of %s\");", schema.GetMultiple())

		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}
}

func generateJavaObjectSetter(schema *ObjectSchema, buffer *BufferedFormatString) {

	generateJavaNullCheck(buffer)
}

func generateJavaArraySetter(schema *ArraySchema, buffer *BufferedFormatString) {

	generateJavaNullCheck(buffer)

	if schema.MinItems != nil {
		generateJavaRangeCheck(*schema.MinItems, "value.length", "does not have enough items", "%d", false, "<", "", buffer)
	}

	if schema.MaxItems != nil {
		generateJavaRangeCheck(*schema.MaxItems, "value.length", "does not have enough items", "%d", false, ">", "", buffer)
	}
}

func generateJavaNullCheck(buffer *BufferedFormatString) {

	buffer.Printf("\nif(value == null)\n{")
	buffer.AddIndentation(1)

	buffer.Printf("\nthrow new NullPointerException(\"Cannot set property to null value\");")

	buffer.AddIndentation(-1)
	buffer.Printf("\n}\n")
}

func generateJavaRangeCheck(value interface{}, reference, message, format string, exclusive bool, comparator, exclusiveComparator string, buffer *BufferedFormatString) {

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
	buffer.Printf("\n}\n")
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generateJavaEnumCheck(schema TypeSchema, enumValues []interface{}, prefix string, postfix string, buffer *BufferedFormatString) {

	var typeName string
	var length int

	length = len(enumValues)

	if length <= 0 {
		return
	}

	// write array of valid values
	typeName = generateJavaTypeForSchema(schema)
	buffer.Printf("\n%s[] validValues = new %s[]{%s%v%s", typeName, typeName, prefix, enumValues[0], postfix)

	for _, enumValue := range enumValues[1:length] {
		buffer.Printf(",%s%v%s", prefix, enumValue, postfix)
	}
	buffer.Print("};\n")

	// compare

	buffer.Print("\nboolean isValid = false;")
	buffer.Print("\nfor(int i = 0; i < validValues.length; i++)\n{")
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

func generateJavaTypeForSchema(subschema TypeSchema) string {

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
		return "String"
	case SCHEMATYPE_BOOLEAN:
		return "boolean"
	}

	return "Object"
}
