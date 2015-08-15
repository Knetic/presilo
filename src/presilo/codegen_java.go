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

	ret.WriteString("package " + module)
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
    return "import java.util.regex.Pattern;\n\n"
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

  toWrite = fmt.Sprintf("\n\tpublic %s(", ToCamelCase(schema.Title))
  ret.WriteString(toWrite)

  for _, propertyName = range schema.RequiredProperties {

    subschema = schema.Properties[propertyName]
    propertyName = ToJavaCase(propertyName)

    toWrite = fmt.Sprintf("%s %s", generateJavaTypeForSchema(subschema), propertyName)
    declarations = append(declarations, toWrite)

    toWrite = fmt.Sprintf("\n\t\tset%s(%s);", ToCamelCase(propertyName), propertyName)
    setters = append(setters, toWrite)
  }

  toWrite = strings.Join(declarations, ",")
  ret.WriteString(toWrite)
  ret.WriteString(")\n\t{")

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
    toWrite = fmt.Sprintf("\n\tpublic void set%s(%s value)\n\t{", camelName, typeName)
    ret.WriteString(toWrite)

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

    toWrite = fmt.Sprintf("\n\t\t%s = value;", properName)
    ret.WriteString(toWrite)

    ret.WriteString("\n\t}\n")
  }

  return ret.String()
}

func generateJavaRangeCheck(value interface{}, reference string, format string, exclusive bool, comparator, exclusiveComparator string) string {

	var ret bytes.Buffer
	var toWrite, compareString string

	if(exclusive) {
		compareString = exclusiveComparator
	} else {
		compareString = comparator
	}

	toWrite = "\n\tif("+ reference +" " + compareString + " " +format+ ")\n\t{"
	toWrite = fmt.Sprintf(toWrite, value)
	ret.WriteString(toWrite)

	toWrite = fmt.Sprintf("\n\t\tthrow new Exception(\"Property '\"+value+\"' was out of range.\")\n\t}\n")
	ret.WriteString(toWrite)

	return ret.String()
}

func generateJavaStringSetter(schema *StringSchema) string {

  var ret bytes.Buffer

  // TODO: null check
  // TODO: length checks
  // TODO: pattern checks
  return ret.String()
}

func generateJavaNumericSetter(schema NumericSchemaType) string {

  var ret bytes.Buffer

  // TODO: null check
  // TODO: min/max checks
  // TODO: multiple checks
  return ret.String()
}

func generateJavaObjectSetter(schema *ObjectSchema) string {

  var ret bytes.Buffer

  // TODO: null check
  return ret.String()
}

func generateJavaArraySetter(schema *ArraySchema) string {

  var ret bytes.Buffer

  // TODO: null check
  // TODO: min/max length checks
  return ret.String()
}

func generateJavaTypeForSchema(subschema TypeSchema) string {

  switch(subschema.GetSchemaType()) {
  case SCHEMATYPE_NUMBER: return "double"
  case SCHEMATYPE_INTEGER: return "int"
  case SCHEMATYPE_ARRAY: return subschema.GetTitle() + "[]"
  case SCHEMATYPE_OBJECT: return subschema.GetTitle()
  case SCHEMATYPE_STRING: return "String"
  case SCHEMATYPE_BOOLEAN: return "boolean"
  }

  return "Object"
}
