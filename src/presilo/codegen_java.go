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
  return ret.String()
}

func generateJavaTypeForSchema(subschema TypeSchema) string {
  return "String"
}
