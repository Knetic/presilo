package presilo

import (
	"bytes"
	"fmt"
	"strings"
)

/*
  Generates valid Java code for a given schema.
*/
func GenerateJava(schema *ObjectSchema, module string) string {

	var ret bytes.Buffer

	ret.WriteString("package " + module + ";")
	ret.WriteString("\n")
	ret.WriteString(generateJavaImports(schema))
	ret.WriteString("\n")
	ret.WriteString(generateJavaTypeDeclaration(schema))
	ret.WriteString("\n")
	ret.WriteString(generateJavaConstructor(schema))
	ret.WriteString("\n")
	ret.WriteString(generateJavaFunctions(schema))
	ret.WriteString("\n}\n")

	return ret.String()
}

func generateJavaImports(schema *ObjectSchema) string {

  // import regex if we need it
  if(containsRegexpMatch(schema)) {
    return "import java.util.regex.*;\n\n"
  }
  return ""
}

func generateJavaTypeDeclaration(schema *ObjectSchema) string {

  var ret bytes.Buffer
  var subschema TypeSchema
  var propertyName string
  var toWrite string

  toWrite = fmt.Sprintf("public class %s\n{\n", ToCamelCase(schema.Title))
  ret.WriteString(toWrite)

  for propertyName, subschema = range schema.Properties {

    propertyName = ToJavaCase(propertyName)
    toWrite = fmt.Sprintf("\n\tprotected %s %s;", generateJavaTypeForSchema(subschema), propertyName)
    ret.WriteString(toWrite)
  }

  return ret.String()
}

func generateJavaConstructor(schema *ObjectSchema) string {

  var ret bytes.Buffer
  var subschema TypeSchema
  var declarations, setters []string
  var propertyName string
  var toWrite string
  var constrained bool

  toWrite = fmt.Sprintf("\n\tpublic %s(", ToCamelCase(schema.Title))
  ret.WriteString(toWrite)

  for _, propertyName = range schema.RequiredProperties {

    subschema = schema.Properties[propertyName]
    propertyName = ToJavaCase(propertyName)

    if(subschema.HasConstraints()) {
      constrained = true
    }

    toWrite = fmt.Sprintf("%s %s", generateJavaTypeForSchema(subschema), propertyName)
    declarations = append(declarations, toWrite)

    toWrite = fmt.Sprintf("\n\t\tset%s(%s);", ToCamelCase(propertyName), propertyName)
    setters = append(setters, toWrite)
  }

  toWrite = strings.Join(declarations, ",")
  ret.WriteString(toWrite)
  ret.WriteString(")")

  if(constrained) {
    ret.WriteString(" throws Exception")
  }

  ret.WriteString("\n\t{")

  for _, setter := range setters {
    ret.WriteString(setter)
  }

  ret.WriteString("\n\t}\n")
  return ret.String()
}

func generateJavaFunctions(schema *ObjectSchema) string {

  var ret bytes.Buffer
  var subschema TypeSchema
  var toWrite string
  var propertyName, properName, camelName, typeName string

  for propertyName, subschema = range schema.Properties {

    properName = ToJavaCase(propertyName)
    camelName = ToCamelCase(propertyName)
    typeName = generateJavaTypeForSchema(subschema)

    // getter
    toWrite = fmt.Sprintf("\n\tpublic %s get%s()\n\t{", typeName, camelName)
    ret.WriteString(toWrite)

    toWrite = fmt.Sprintf("\n\t\treturn this.%s;\n\t}", properName)
    ret.WriteString(toWrite)

    // setter
    toWrite = fmt.Sprintf("\n\tpublic void set%s(%s value)", camelName, typeName)
    ret.WriteString(toWrite)

    if(subschema.HasConstraints()) {
      ret.WriteString(" throws Exception")
    }

    ret.WriteString("\n\t{")

    switch subschema.GetSchemaType() {
    case SCHEMATYPE_BOOLEAN:
      toWrite = ""
    case SCHEMATYPE_STRING:
      toWrite = generateJavaStringSetter(subschema.(*StringSchema))
    case SCHEMATYPE_INTEGER: fallthrough
    case SCHEMATYPE_NUMBER:
      toWrite = generateJavaNumericSetter(subschema.(NumericSchemaType))
    case SCHEMATYPE_OBJECT:
      toWrite = generateJavaObjectSetter(subschema.(*ObjectSchema))
    case SCHEMATYPE_ARRAY:
      toWrite = generateJavaArraySetter(subschema.(*ArraySchema))
    }

    ret.WriteString(toWrite)

    toWrite = fmt.Sprintf("\n\t\t%s = value;", properName)
    ret.WriteString(toWrite)

    ret.WriteString("\n\t}\n")
  }

  return ret.String()
}

func generateJavaStringSetter(schema *StringSchema) string {

  var ret bytes.Buffer
  var toWrite string

  ret.WriteString(generateJavaNullCheck())

  if(schema.MinLength!= nil) {
    ret.WriteString(generateJavaRangeCheck(*schema.MinLength, "value.length()", "was shorter than allowable minimum", "%d", false, "<", ""))
  }

  if(schema.MaxLength != nil) {
    ret.WriteString(generateJavaRangeCheck(*schema.MaxLength, "value.length()", "was longer than allowable maximum", "%d", false, ">", ""))
  }

  if(schema.Pattern != nil) {

    toWrite = fmt.Sprintf("\n\t\tPattern regex = Pattern.compile(\"%s\");", sanitizeQuotedString(*schema.Pattern))
    ret.WriteString(toWrite)

    ret.WriteString("\n\t\tif(!regex.matcher(value).matches())\n\t\t{")

    toWrite = fmt.Sprintf("\n\t\t\tthrow new Exception(\"Value '\"+value+\"' did not match pattern '%s'\");", *schema.Pattern)
    ret.WriteString(toWrite)

    ret.WriteString("\n\t\t}")
  }
  return ret.String()
}

func generateJavaNumericSetter(schema NumericSchemaType) string {

  var ret bytes.Buffer
  var toWrite string

  if(schema.HasMinimum()) {
		ret.WriteString(generateJavaRangeCheck(schema.GetMinimum(), "value", "is under the allowable minimum", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<"))
	}

	if(schema.HasMaximum()) {
		ret.WriteString(generateJavaRangeCheck(schema.GetMaximum(), "value", "is over the allowable maximum", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">"))
	}

  if(schema.HasEnum()) {
		ret.WriteString(generateJavaEnumCheck(schema, schema.GetEnum(), "", ""))
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

func generateJavaObjectSetter(schema *ObjectSchema) string {

  var ret bytes.Buffer

  ret.WriteString(generateJavaNullCheck())
  return ret.String()
}

func generateJavaArraySetter(schema *ArraySchema) string {

  var ret bytes.Buffer

  ret.WriteString(generateJavaNullCheck())

  if(schema.MinItems != nil) {
    ret.WriteString(generateJavaRangeCheck(*schema.MinItems, "value.length", "does not have enough items", "%d", false, "<", ""))
  }

  if(schema.MaxItems != nil) {
    ret.WriteString(generateJavaRangeCheck(*schema.MaxItems, "value.length", "does not have enough items", "%d", false, ">", ""))
  }

  return ret.String()
}

func generateJavaNullCheck() string {

  var ret bytes.Buffer

  ret.WriteString("\n\t\tif(value == null)\n\t\t{")
  ret.WriteString("\n\t\t\tthrow new NullPointerException(\"Cannot set property to null value\");")
  ret.WriteString("\n\t\t}\n")

  return ret.String()
}

func generateJavaRangeCheck(value interface{}, reference, message, format string, exclusive bool, comparator, exclusiveComparator string) string {

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
func generateJavaEnumCheck(schema TypeSchema, enumValues []interface{}, prefix string, postfix string) string {

	var ret bytes.Buffer
	var constraint, typeName string
	var length int

	length = len(enumValues)

	if(length <= 0) {
		return ""
	}

	// write array of valid values
  typeName = generateJavaTypeForSchema(schema)
	constraint = fmt.Sprintf("\t%s[] validValues = new %s[]{%s%v%s", typeName, typeName, prefix, enumValues[0], postfix)
	ret.WriteString(constraint)

	for _, enumValue := range enumValues[1:length] {

		constraint = fmt.Sprintf(",%s%v%s", prefix, enumValue, postfix)
		ret.WriteString(constraint)
	}
	ret.WriteString("};\n")

	// compare
	ret.WriteString("\tboolean isValid = false;\n")
	ret.WriteString("\tfor(int i = 0; i < validValues.length; i++) \n\t{\n")
	ret.WriteString("\t\tif(validValues[i] == value)\n\t\t{\n\t\t\tisValid = true;")
	ret.WriteString("\n\t\t\tbreak;\n\t\t}\n\t}")

	ret.WriteString("\n\tif(!isValid)\n\t{")
	ret.WriteString("\n\t\tthrow new Exception(\"Given value '\"+value+\"' was not found in list of acceptable values\");\n")
	ret.WriteString("\t}\n")

	return ret.String()
}

func generateJavaTypeForSchema(subschema TypeSchema) string {

  switch(subschema.GetSchemaType()) {
  case SCHEMATYPE_NUMBER: return "double"
  case SCHEMATYPE_INTEGER: return "int"
  case SCHEMATYPE_ARRAY: return ToCamelCase(subschema.(*ArraySchema).Items.GetTitle()) + "[]"
  case SCHEMATYPE_OBJECT: return ToCamelCase(subschema.GetTitle())
  case SCHEMATYPE_STRING: return "String"
  case SCHEMATYPE_BOOLEAN: return "boolean"
  }

  return "Object"
}
