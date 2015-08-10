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
    var subschema TypeSchema
    var propertyName string
    var isRequired bool

    ret.WriteString("type ")
    ret.WriteString(schema.GetTitle())
    ret.WriteString(" struct {\n")

    // write all required fields as unexported fields.
    for _, propertyName = range schema.RequiredProperties {

      subschema = schema.Properties[propertyName]

      ret.WriteString(generateVariableDeclaration(subschema, propertyName, ToJavaCase))
    }

    // write all non-required fields as exported fields.
    for propertyName, subschema = range schema.Properties {

      isRequired = false

      for _, requiredName := range schema.RequiredProperties {
        if(requiredName == propertyName) {
          isRequired = true
          break
        }
      }

      if(isRequired) {
        continue
      }

      ret.WriteString(generateVariableDeclaration(subschema, propertyName, ToCamelCase))
    }

    ret.WriteString("}\n")
    return ret.String()
}

func generateGoFunctions(schema *ObjectSchema) string {

    var ret bytes.Buffer
    var subschema TypeSchema
    var signature, body string
    var propertyName, casedJavaName, casedCamelName string

    for _, propertyName = range schema.RequiredProperties {

      casedJavaName = ToJavaCase(propertyName)
      casedCamelName = ToCamelCase(propertyName)

      subschema = schema.Properties[propertyName]

      signature = fmt.Sprintf("func (this *%s)Get%s()(%s){\n", schema.GetTitle(), casedCamelName, generateGoTypeForSchema(subschema))
      body = fmt.Sprintf("\treturn this.%s\n}\n\n", casedJavaName)

      ret.WriteString(signature)
      ret.WriteString(body)
    }

    return ret.String()
}

func generateGoConstructor(schema *ObjectSchema) string {

  var subschema TypeSchema
  var ret bytes.Buffer
  var parameters, parameterNames []string
  var title, signature, parameterDefinition string

  for _, propertyName := range schema.RequiredProperties {

    subschema = schema.Properties[propertyName]
    propertyName = ToJavaCase(propertyName)

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

func generateVariableDeclaration(subschema TypeSchema, propertyName string, casing func(string)(string)) (string) {

  var structTag string
  var ret bytes.Buffer

  structTag = fmt.Sprintf(" `json:\"%s\";`", ToJavaCase(propertyName))

  // TODO: this means unexported fields will have json deserialization struct tags,
  // which won't work.
  ret.WriteString("\tvar " + casing(propertyName) + " ")
  ret.WriteString(generateGoTypeForSchema(subschema))
  ret.WriteString(structTag)
  ret.WriteString("\n")

  return ret.String()
}
