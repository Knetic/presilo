package presilo

import (
  "bytes"
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

      switch subschema.GetSchemaType() {
      case SCHEMATYPE_STRING:
        ret.WriteString("string\n")
      case SCHEMATYPE_INTEGER:
        ret.WriteString("int\n")
      }
    }

    ret.WriteString("}\n")
    return ret.String()
}

func generateGoFunctions(schema *ObjectSchema) string {

    var ret bytes.Buffer
    return ret.String()
}
