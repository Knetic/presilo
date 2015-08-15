package presilo

import (
	"bytes"
	"fmt"
	"strings"
)

/*
  Generates valid CSharp code for a given schema.
*/
func GenerateCSharp(schema *ObjectSchema, module string) string {

	var ret bytes.Buffer

	ret.WriteString(generateCSharpImports(schema))
  ret.WriteString("\n")
  ret.WriteString("namespace " + module + "\n{")
	ret.WriteString("\n")
	ret.WriteString(generateCSharpTypeDeclaration(schema))
	ret.WriteString("\n")
	ret.WriteString(generateCSharpConstructor(schema))
	ret.WriteString("\n")
	ret.WriteString(generateCSharpFunctions(schema))
	ret.WriteString("\n}\n}\n")

	return ret.String()
}

func generateCSharpImports(schema *ObjectSchema) string {

  var ret bytes.Buffer
  var usings []string
  var toWrite string

  usings = []string{"System"}

  // import regex if we need it
  if(containsRegexpMatch(schema)) {
    usings = append(usings, "System.Text.RegularExpressions")
  }

  for _, using := range usings {

    toWrite = fmt.Sprintf("using %s;", using)
    ret.WriteString(toWrite)
  }

  return ret.String()
}

func generateCSharpTypeDeclaration(schema *ObjectSchema) string {

  var ret bytes.Buffer
  var subschema TypeSchema
  var propertyName string
  var toWrite string

  toWrite = fmt.Sprintf("public class %s\n{\n", ToCamelCase(schema.Title))
  ret.WriteString(toWrite)

  for propertyName, subschema = range schema.Properties {

    propertyName = ToJavaCase(propertyName)
    toWrite = fmt.Sprintf("\n\tprotected %s %s;", generateCSharpTypeForSchema(subschema), propertyName)
    ret.WriteString(toWrite)
  }

  return ret.String()
}

func generateCSharpConstructor(schema *ObjectSchema) string {

  var ret bytes.Buffer
  var subschema TypeSchema
  var declarations, setters []string
  var propertyName string
  var toWrite string

  toWrite = fmt.Sprintf("\n\tpublic %s(", ToCamelCase(schema.Title))
  ret.WriteString(toWrite)

  for _, propertyName = range schema.RequiredProperties {

    subschema = schema.Properties[propertyName]
    propertyName = ToJavaCase(propertyName)

    toWrite = fmt.Sprintf("%s %s", generateCSharpTypeForSchema(subschema), propertyName)
    declarations = append(declarations, toWrite)

    toWrite = fmt.Sprintf("\n\t\tset%s(%s);", ToCamelCase(propertyName), propertyName)
    setters = append(setters, toWrite)
  }

  toWrite = strings.Join(declarations, ",")
  ret.WriteString(toWrite)
  ret.WriteString(")")

  ret.WriteString("\n\t{")

  for _, setter := range setters {
    ret.WriteString(setter)
  }

  ret.WriteString("\n\t}\n")
  return ret.String()
}

func generateCSharpFunctions(schema *ObjectSchema) string {

  var ret bytes.Buffer
  var subschema TypeSchema
  var toWrite string
  var propertyName, properName, camelName, typeName string

  for propertyName, subschema = range schema.Properties {

    properName = ToJavaCase(propertyName)
    camelName = ToCamelCase(propertyName)
    typeName = generateCSharpTypeForSchema(subschema)

    // getter
    toWrite = fmt.Sprintf("\n\tpublic %s get%s()\n\t{", typeName, camelName)
    ret.WriteString(toWrite)

    toWrite = fmt.Sprintf("\n\t\treturn this.%s;\n\t}", properName)
    ret.WriteString(toWrite)

    // setter
    toWrite = fmt.Sprintf("\n\tpublic void set%s(%s value)", camelName, typeName)
    ret.WriteString(toWrite)

    ret.WriteString("\n\t{")

    switch subschema.GetSchemaType() {
    case SCHEMATYPE_BOOLEAN:
      toWrite = ""
    case SCHEMATYPE_STRING:
      toWrite = generateCSharpStringSetter(subschema.(*StringSchema))
    case SCHEMATYPE_INTEGER: fallthrough
    case SCHEMATYPE_NUMBER:
      toWrite = generateCSharpNumericSetter(subschema.(NumericSchemaType))
    case SCHEMATYPE_OBJECT:
      toWrite = generateCSharpObjectSetter(subschema.(*ObjectSchema))
    case SCHEMATYPE_ARRAY:
      toWrite = generateCSharpArraySetter(subschema.(*ArraySchema))
    }

    ret.WriteString(toWrite)

    toWrite = fmt.Sprintf("\n\t\t%s = value;", properName)
    ret.WriteString(toWrite)

    ret.WriteString("\n\t}\n")
  }

  return ret.String()
}

func generateCSharpStringSetter(schema *StringSchema) string {

  var ret bytes.Buffer
  var toWrite string

  ret.WriteString(generateCSharpNullCheck())

  if(schema.MinLength!= nil) {
    ret.WriteString(generateCSharpRangeCheck(*schema.MinLength, "value.Length", "was shorter than allowable minimum", "%d", false, "<", ""))
  }

  if(schema.MaxLength != nil) {
    ret.WriteString(generateCSharpRangeCheck(*schema.MaxLength, "value.Length", "was longer than allowable maximum", "%d", false, ">", ""))
  }

  if(schema.Pattern != nil) {

    toWrite = fmt.Sprintf("\n\t\tRegex regex = new Regex(\"%s\");", sanitizeQuotedString(*schema.Pattern))
    ret.WriteString(toWrite)

    ret.WriteString("\n\t\tif(!regex.IsMatch(value))\n\t\t{")

    toWrite = fmt.Sprintf("\n\t\t\tthrow new Exception(\"Value '\"+value+\"' did not match pattern '%s'\");", *schema.Pattern)
    ret.WriteString(toWrite)

    ret.WriteString("\n\t\t}")
  }
  return ret.String()
}

func generateCSharpNumericSetter(schema NumericSchemaType) string {

  var ret bytes.Buffer
  var toWrite string

  if(schema.HasMinimum()) {
		ret.WriteString(generateCSharpRangeCheck(schema.GetMinimum(), "value", "is under the allowable minimum", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<"))
	}

	if(schema.HasMaximum()) {
		ret.WriteString(generateCSharpRangeCheck(schema.GetMaximum(), "value", "is over the allowable maximum", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">"))
	}

  if(schema.HasEnum()) {
		ret.WriteString(generateCSharpEnumCheck(schema, schema.GetEnum(), "", ""))
	}

  if(schema.HasMultiple()) {

    toWrite = fmt.Sprintf("\n\tif(value %% %f != 0)\n\t{", schema.GetMultiple())
    ret.WriteString(toWrite)

    toWrite = fmt.Sprintf("\n\t\tthrow new Exception(\"Property '\"+value+\"' was not a multiple of %s\");", schema.GetMultiple())
    ret.WriteString(toWrite)

    ret.WriteString("\n\t}\n")
  }
  return ret.String()
}

func generateCSharpObjectSetter(schema *ObjectSchema) string {

  var ret bytes.Buffer

  ret.WriteString(generateCSharpNullCheck())
  return ret.String()
}

func generateCSharpArraySetter(schema *ArraySchema) string {

  var ret bytes.Buffer

  ret.WriteString(generateCSharpNullCheck())

  if(schema.MinItems != nil) {
    ret.WriteString(generateCSharpRangeCheck(*schema.MinItems, "value.Length", "does not have enough items", "%d", false, "<", ""))
  }

  if(schema.MaxItems != nil) {
    ret.WriteString(generateCSharpRangeCheck(*schema.MaxItems, "value.Length", "does not have enough items", "%d", false, ">", ""))
  }

  return ret.String()
}

func generateCSharpNullCheck() string {

  var ret bytes.Buffer

  ret.WriteString("\n\t\tif(value == null)\n\t\t{")
  ret.WriteString("\n\t\t\tthrow new NullReferenceException(\"Cannot set property to null value\");")
  ret.WriteString("\n\t\t}\n")

  return ret.String()
}

func generateCSharpRangeCheck(value interface{}, reference, message, format string, exclusive bool, comparator, exclusiveComparator string) string {

	var ret bytes.Buffer
	var toWrite, compareString string

	if(exclusive) {
		compareString = exclusiveComparator
	} else {
		compareString = comparator
	}

	toWrite = "\n\t\tif("+ reference +" " + compareString + " " +format+ ")\n\t\t{"
	toWrite = fmt.Sprintf(toWrite, value)
	ret.WriteString(toWrite)

	toWrite = fmt.Sprintf("\n\t\t\tthrow new Exception(\"Property '\"+value+\"' %s.\");\n\t\t}\n", message)
	ret.WriteString(toWrite)

	return ret.String()
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generateCSharpEnumCheck(schema TypeSchema, enumValues []interface{}, prefix string, postfix string) string {

	var ret bytes.Buffer
	var constraint, typeName string
	var length int

	length = len(enumValues)

	if(length <= 0) {
		return ""
	}

	// write array of valid values
  typeName = generateCSharpTypeForSchema(schema)
	constraint = fmt.Sprintf("\t%s[] validValues = new %s[]{%s%v%s", typeName, typeName, prefix, enumValues[0], postfix)
	ret.WriteString(constraint)

	for _, enumValue := range enumValues[1:length] {

		constraint = fmt.Sprintf(",%s%v%s", prefix, enumValue, postfix)
		ret.WriteString(constraint)
	}
	ret.WriteString("};\n")

	// compare
	ret.WriteString("\tbool isValid = false;\n")
	ret.WriteString("\tfor(int i = 0; i < validValues.Length; i++) \n\t{\n")
	ret.WriteString("\t\tif(validValues[i] == value)\n\t\t{\n\t\t\tisValid = true;")
	ret.WriteString("\n\t\t\tbreak;\n\t\t}\n\t}")

	ret.WriteString("\n\tif(!isValid)\n\t{")
	ret.WriteString("\n\t\tthrow new Exception(\"Given value '\"+value+\"' was not found in list of acceptable values\");\n")
	ret.WriteString("\t}\n")

	return ret.String()
}

func generateCSharpTypeForSchema(subschema TypeSchema) string {

  switch(subschema.GetSchemaType()) {
  case SCHEMATYPE_NUMBER: return "double"
  case SCHEMATYPE_INTEGER: return "int"
  case SCHEMATYPE_ARRAY: return ToCamelCase(subschema.(*ArraySchema).Items.GetTitle()) + "[]"
  case SCHEMATYPE_OBJECT: return ToCamelCase(subschema.GetTitle())
  case SCHEMATYPE_STRING: return "string"
  case SCHEMATYPE_BOOLEAN: return "bool"
  }

  return "Object"
}
