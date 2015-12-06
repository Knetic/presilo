package presilo

import (
	"bytes"
	"regexp"
	"strings"
)

/*
  Generates valid Go code for a given schema.
*/
func GenerateGo(schema *ObjectSchema, module string, tabstyle string) string {

	var buffer *BufferedFormatString

	buffer = NewBufferedFormatString(tabstyle)

	buffer.Printf("package %s", module)
	buffer.Print("\n")
	generateGoImports(schema, buffer)
	buffer.Print("\n")
	generateGoTypeDeclaration(schema, buffer)
	buffer.Print("\n")
	generateGoConstructor(schema, buffer)
	buffer.Print("\n")
	generateGoFunctions(schema, buffer)
	buffer.Print("\n")

	return buffer.String()
}

func ValidateGoModule(module string) bool {

	matched, err := regexp.MatchString("^[a-zA-Z_]+[0-9a-zA-Z_]*$", module)
	return err == nil && matched
}

func generateGoImports(schema *ObjectSchema, buffer *BufferedFormatString) {

	var imports []string

	// import errors if there are any constrained fields
	if len(schema.ConstrainedProperties) > 0 {
		imports = append(imports, "errors")
	}

	// if any string schema has a pattern match, import regex.
	if containsRegexpMatch(schema) {
		imports = append(imports, "regexp")
	}

	// if any number (but not integer!) has a multiple clause, import math
	if containsNumberMod(schema) {
		imports = append(imports, "math")
	}

	// write imports (if they exist)
	if len(imports) > 0 {

		buffer.Print("import (\n")
		for _, packageName := range imports {
			buffer.Printf("\"%s\"\n", packageName)
		}

		buffer.Print(")\n")
	}
}

/*
	Generates the type declaration for this schema,
	including all member fields (properly exported if they have no constraints),
	and struct tags.
	Also includes the doc comments.
*/
func generateGoTypeDeclaration(schema *ObjectSchema, buffer *BufferedFormatString) {

	var subschema TypeSchema
	var propertyName string

	// description first
	buffer.Printf("/*\n%s\n*/\n", schema.GetDescription())
	buffer.Printf("type %s struct {", schema.GetTitle())
	buffer.AddIndentation(1)

	// write all required fields as unexported fields.
	for _, propertyName = range schema.ConstrainedProperties {

		subschema = schema.Properties[propertyName]
		generateVariableDeclaration(subschema, buffer, propertyName, ToJavaCase)
	}

	// write all non-required fields as exported fields.
	for _, propertyName = range schema.UnconstrainedProperties {

		subschema = schema.Properties[propertyName]
		generateVariableDeclaration(subschema, buffer, propertyName, ToCamelCase)
	}

	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
}

/*
	Generates getters and setters for all fields in the given schema
	which have constraints.
	fields without constraints are assumed to be exported.
*/
func generateGoFunctions(schema *ObjectSchema, buffer *BufferedFormatString) {

	var subschema TypeSchema
	var propertyName, casedJavaName, casedCamelName string

	for _, propertyName = range schema.ConstrainedProperties {

		casedJavaName = ToJavaCase(propertyName)
		casedCamelName = ToCamelCase(propertyName)

		subschema = schema.Properties[propertyName]

		// getter
		buffer.Printf("\nfunc (this *%s) Get%s() (%s) {", schema.GetTitle(), casedCamelName, generateGoTypeForSchema(subschema))
		buffer.AddIndentation(1)
		buffer.Printf("\nreturn this.%s", casedJavaName)
		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")

		// setter
		buffer.Printf("func (this *%s) Set%s(value %s) (error) {", schema.GetTitle(), casedCamelName, generateGoTypeForSchema(subschema))
		buffer.AddIndentation(1)

		switch subschema.GetSchemaType() {
		case SCHEMATYPE_STRING:
			 generateGoStringSetter(subschema.(*StringSchema), buffer)
		case SCHEMATYPE_NUMBER:
			generateGoNumericSetter(subschema.(*NumberSchema), buffer)
		case SCHEMATYPE_INTEGER:
			generateGoNumericSetter(subschema.(*IntegerSchema), buffer)
		case SCHEMATYPE_ARRAY:
			generateGoArraySetter(subschema.(*ArraySchema), buffer)
		}


		buffer.Printf("\nthis.%s = value\nreturn nil", casedJavaName)
		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}
}

/*
	Generates a convenience "New*" method for the given schema,
	which accepts parameters that match the 'required' properties.
	Any properties which are both 'required' and have constraints
	will have their setters used, instead of setting the field directly.
*/
func generateGoConstructor(schema *ObjectSchema, buffer *BufferedFormatString) {

	var subschema TypeSchema
	var ret bytes.Buffer
	var parameters, parameterNames []string
	var title string

	for _, propertyName := range schema.RequiredProperties {

		subschema = schema.Properties[propertyName]
		parameterNames = append(parameterNames, propertyName)

		propertyName = getAppropriateGoCase(schema, propertyName)
		ret.WriteString(propertyName)
		ret.WriteString(" ")
		ret.WriteString(generateGoTypeForSchema(subschema))

		parameters = append(parameters, ret.String())
		ret.Reset()
	}

	// signature
	title = ToCamelCase(schema.Title)
	buffer.Printf("\nfunc New%s(%s)(*%s, error) {\n", title, strings.Join(parameters, ","), title)
	buffer.AddIndentation(1)

	buffer.Printf("\nvar err error = nil")

	// body
	buffer.Printf("\nret := new(%s)\n", title)

	for _, propertyName := range parameterNames {

		subschema = schema.Properties[propertyName]
		propertyName = getAppropriateGoCase(schema, propertyName)

		if(subschema.HasConstraints()) {

			buffer.Printf("\nerr = ret.Set%s(%s)", ToCamelCase(propertyName), propertyName)

			buffer.Printf("\nif(err != nil) {")
			buffer.AddIndentation(1)
			buffer.Printf("\nreturn nil, err")
			buffer.AddIndentation(-1)
			buffer.Printf("\n}")
		} else {
			buffer.Printf("\nret.%s = %s", propertyName, propertyName)
		}
	}

	buffer.Print("\nreturn ret, err")
	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
}

/*
	Generates a 'setter' method for the given numeric schema.
	Numeric schemas could either be NumberSchema or IntegerSchema,
	since they share the same constraint set, this works on either.
*/
func generateGoNumericSetter(schema NumericSchemaType, buffer *BufferedFormatString) {

	var minimum, maximum, multiple interface{}
	var formatString string
	var comparator string

	if !schema.HasConstraints() {
		return
	}

	formatString = schema.GetConstraintFormat()

	if schema.HasEnum() {
		generateGoEnumForSchema(schema, buffer, schema.GetEnum(), "", "")
	}

	if schema.HasMinimum() {

		if schema.IsExclusiveMinimum() {
			comparator = "<="
		} else {
			comparator = "<"
		}

		minimum = schema.GetMinimum()

		buffer.Printf("\nif(value %s " + formatString + ") {", comparator, minimum)
		buffer.AddIndentation(1)

		buffer.Printf("\nreturn errors.New(\"Minimum value of '" + formatString + "' not met\")", minimum)

		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}

	if schema.HasMaximum() {

		if schema.IsExclusiveMaximum() {
			comparator = ">="
		} else {
			comparator = ">"
		}

		maximum = schema.GetMaximum()

		buffer.Printf("\nif(value %s " + formatString + ") {", comparator, maximum)
		buffer.AddIndentation(1)

		buffer.Printf("\nreturn errors.New(\"Maximum value of '" + formatString + "' not met\")", maximum)

		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}

	if schema.HasMultiple() {

		multiple = schema.GetMultiple()

		if schema.GetSchemaType() == SCHEMATYPE_NUMBER {
			buffer.Printf("\nif(math.Mod(value, %f) != 0) {", multiple)
		} else {
			buffer.Printf("\nif(value %% %d != 0) {", multiple)
		}

		buffer.AddIndentation(1)

		buffer.Printf("\nreturn errors.New(\"Value is not a multiple of '" + formatString + "'\")", multiple)

		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}
}

/*
	Generates a 'setter' function for the given string schema.
	Generates code which validates all schema constraints before setting.
*/
func generateGoStringSetter(schema *StringSchema, buffer *BufferedFormatString) {

	var cutoff int

	if schema.Enum != nil {
		generateGoEnumForSchema(schema, buffer, schema.GetEnum(), "\"", "\"")
	}

	if schema.MinLength != nil {

		cutoff = *schema.MinLength

		buffer.Printf("\nif(len(value) < %d) {", cutoff)
		buffer.AddIndentation(1)
		buffer.Printf("\nreturn errors.New(\"Value is shorter than minimum length of %d\")", cutoff)
		buffer.AddIndentation(-1)
		buffer.Printf("\n}\n")
	}

	if schema.MaxLength != nil {

		cutoff = *schema.MaxLength

		buffer.Printf("\nif(len(value) > %d) {", cutoff)
		buffer.AddIndentation(1)
		buffer.Printf("\nreturn errors.New(\"Value is longer than minimum length of %d\")", cutoff)
		buffer.AddIndentation(-1)
		buffer.Printf("\n}\n")
	}

	if schema.Pattern != nil {

		buffer.Printf("\nmatched, err := regexp.Match(\"%s\", []byte(value))", sanitizeQuotedString(*schema.Pattern))
		buffer.Printf("\nif(err != nil){return err}")
		buffer.Printf("\nif(!matched) {")
		buffer.AddIndentation(1)

		buffer.Printf("\nreturn errors.New(\"Value did not match regex '%s'\")", *schema.Pattern)

		buffer.AddIndentation(-1)
		buffer.Print("\n}\n")
	}
}

/*
	Generates 'setter' code to validate the given array schema's constraints,
	then set the owner object's value to the one passed in.
*/
func generateGoArraySetter(schema *ArraySchema, buffer *BufferedFormatString) {

	if !schema.HasConstraints() {
		return
	}

	buffer.Print("\nlength := len(value)\n")

	if schema.MinItems != nil {

		buffer.Printf("\nif(length < %d) {", *schema.MinItems)
		buffer.AddIndentation(1)

		buffer.Printf("\nreturn errors.New(\"Minimum number of elements '%d' not present\")", *schema.MinItems)

		buffer.AddIndentation(-1)
		buffer.Printf("\n}\n")
	}

	if schema.MaxItems != nil {

		buffer.Printf("\nif(length > %d) {", *schema.MaxItems)
		buffer.AddIndentation(1)

		buffer.Printf("return errors.New(\"Maximum number of elements '%d' not present\")", *schema.MaxItems)

		buffer.AddIndentation(-1)
		buffer.Printf("\n}\n")
	}
}

/*
	Convenience method to generate an enum constraint check for the given schema and
	its provided enum values.
	Generates an inline set of constants, each value of which is prefixed and postfixed accordingly,
	then generates code to check against those constants.
*/
func generateGoEnumForSchema(schema interface{}, buffer *BufferedFormatString, enumValues []interface{}, prefix string, postfix string) {

	var length int

	length = len(enumValues)

	if length <= 0 {
		return
	}

	// write array of valid values
	buffer.Printf("\nvalidValues := []%s{%s%v%s", generateGoTypeForSchema(schema), prefix, enumValues[0], postfix)

	for _, enumValue := range enumValues[1:length] {
		buffer.Printf(",%s%v%s", prefix, enumValue, postfix)
	}
	buffer.Printf("}\n")

	// compare
	buffer.Printf("\nisValid := false")
	buffer.Printf("\nfor _, validValue := range validValues {")
	buffer.AddIndentation(1)

	buffer.Printf("\nif(validValue == value){")
	buffer.AddIndentation(1)
	buffer.Printf("\nisValid = true\nbreak")
	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")

	buffer.Print("\nif(!isValid){")
	buffer.AddIndentation(1)
	buffer.Print("\nreturn errors.New(\"Given value was not found in list of acceptable values\")")
	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")

	buffer.AddIndentation(-1)
	buffer.Print("\n}\n")
}

/*
	Returns the Go equivalent of type for the given schema.
	If no type is found, this returns "interface{}"
*/
func generateGoTypeForSchema(schema interface{}) string {

	switch schema.(type) {
	case *BooleanSchema:
		return "bool"
	case *StringSchema:
		return "string"
	case *IntegerSchema:
		return "int"
	case *NumberSchema:
		return "float64"
	case *ObjectSchema:
		return "*" + ToCamelCase(schema.(TypeSchema).GetTitle())
	case *ArraySchema:
		return "[]" + ToCamelCase(schema.(*ArraySchema).Items.GetTitle())
	}

	return "interface{}"
}

/*
	Generates a type variable declaration for the given schema and propertyName,
	using the given casing function to modify the name of the property in the correct places.
*/
func generateVariableDeclaration(subschema TypeSchema, buffer *BufferedFormatString, propertyName string, casing func(string) string) {

	// TODO: this means unexported fields will have json deserialization struct tags,
	// which won't work.
	buffer.Printf("\n%s %s", casing(propertyName), generateGoTypeForSchema(subschema))
	buffer.Printf(" `json:\"%s\";`", ToJavaCase(propertyName))
}

/*
	Determines and returns the appropriate case for the property of schema provided.
	Only unconstrained fields are exported in Go generated code;
	if the referenced field is constrained in any way, this generates an unexported field name.
*/
func getAppropriateGoCase(schema *ObjectSchema, propertyName string) string {

	for _, constrainedName := range schema.ConstrainedProperties {
		if constrainedName == propertyName {
			return ToJavaCase(propertyName)
		}
	}
	return ToCamelCase(propertyName)
}
