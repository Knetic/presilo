package presilo

import (
  "fmt"
  "bytes"
  "strings"
)

/*
  Generates valid Go code for a given schema.
*/
func GenerateGo(schema *ObjectSchema, module string) string {

  var ret bytes.Buffer

  ret.WriteString("package " + module)
  ret.WriteString("\n")
  ret.WriteString(generateGoImports(schema))
  ret.WriteString("\n")
  ret.WriteString(generateGoTypeDeclaration(schema))
  ret.WriteString("\n")
  ret.WriteString(generateGoConstructor(schema))
  ret.WriteString("\n")
  ret.WriteString(generateGoFunctions(schema))
  ret.WriteString("\n")

  return ret.String()
}

func generateGoImports(schema *ObjectSchema) string {

    var ret bytes.Buffer
    return ret.String()
}

func generateGoTypeDeclaration(schema *ObjectSchema) string {

    var ret bytes.Buffer

    ret.WriteString("type ")
    ret.WriteString(schema.GetTitle())
    ret.WriteString(" struct {\n")

    for propertyName, subschema := range schema.Properties {

      ret.WriteString("\tvar " + ToCamelCase(propertyName) + " ")
      ret.WriteString(generateGoTypeForSchema(subschema))
      ret.WriteString("\n")
    }

    ret.WriteString("}\n")
    return ret.String()
}

func generateGoFunctions(schema *ObjectSchema) string {

    var ret bytes.Buffer
    return ret.String()
}

func generateGoConstructor(schema *ObjectSchema) string {

  var ret bytes.Buffer
  var parameters, parameterNames []string
  var title, signature, parameterDefinition string

  for propertyName, subschema := range schema.Properties {

    propertyName = ToCamelCase(propertyName)

    ret.WriteString(propertyName)
    ret.WriteString(" ")
    ret.WriteString(generateGoTypeForSchema(subschema))

    parameterNames = append(parameterNames, propertyName)
    parameters = append(parameters, ret.String())
    ret.Reset()

  }

  // signature
  title = ToCamelCase(schema.Title)
  signature = fmt.Sprintf("func New%s(%s)(*%s) {\n", title, strings.Join(parameters, ","), title)
  ret.WriteString(signature)

  // body
  parameterDefinition = fmt.Sprintf("\tret := new(%s)\n", title)
  ret.WriteString(parameterDefinition)

  for _, propertyName := range parameterNames {

    parameterDefinition = fmt.Sprintf("\tret.%s = %s\n", propertyName, propertyName)
    ret.WriteString(parameterDefinition)
  }

  ret.WriteString("\treturn ret\n}\n\n")
  return ret.String()
}

func generateGoIntegerFunctions(schema *IntegerSchema) string {

  var ret bytes.Buffer

  if(!schema.HasConstraints()) {
    return ""
  }

  if(schema.Minimum != nil) {

  }

  return ret.String()
}

func generateGoStringFunctions(schema *StringSchema) string {

  var ret bytes.Buffer

  return ret.String()
}

func generateGoTypeForSchema(schema TypeSchema) string {

  switch schema.GetSchemaType() {
  case SCHEMATYPE_STRING:
    return "string"
  case SCHEMATYPE_INTEGER:
    return "int"
  case SCHEMATYPE_OBJECT:
    return "*" + ToCamelCase(schema.GetTitle())
  }

  return "interface{}"
}
