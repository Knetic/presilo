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

    // description first
    description := fmt.Sprintf("/*\n%s\n*/\n", schema.GetDescription())
    ret.WriteString(description)

    ret.WriteString("type ")
    ret.WriteString(schema.GetTitle())
    ret.WriteString(" struct {\n")

    // write all required fields as unexported fields.
    for _, propertyName = range schema.ConstrainedProperties {

      subschema = schema.Properties[propertyName]
      ret.WriteString(generateVariableDeclaration(subschema, propertyName, ToJavaCase))
    }

    // write all non-required fields as exported fields.
    for _, propertyName = range schema.UnconstrainedProperties {

      subschema = schema.Properties[propertyName]
      ret.WriteString(generateVariableDeclaration(subschema, propertyName, ToCamelCase))
    }

    ret.WriteString("}\n")
    return ret.String()
}

func generateGoFunctions(schema *ObjectSchema) string {

    var ret bytes.Buffer
    var subschema TypeSchema
    var signature, body, constraintChecks string
    var propertyName, casedJavaName, casedCamelName string

    for _, propertyName = range schema.ConstrainedProperties {

      casedJavaName = ToJavaCase(propertyName)
      casedCamelName = ToCamelCase(propertyName)

      subschema = schema.Properties[propertyName]

      // getter
      signature = fmt.Sprintf("func (this *%s) Get%s() (%s) {\n", schema.GetTitle(), casedCamelName, generateGoTypeForSchema(subschema))
      body = fmt.Sprintf("\treturn this.%s\n}\n\n", casedJavaName)

      ret.WriteString(signature)
      ret.WriteString(body)

      // setter
      signature = fmt.Sprintf("func (this *%s) Set%s(value %s) (error) {\n", schema.GetTitle(), casedCamelName, generateGoTypeForSchema(subschema))
      body = fmt.Sprintf("\n\tthis.%s = value\n\treturn nil\n}\n\n", casedJavaName)

      switch subschema.GetSchemaType() {
        case SCHEMATYPE_STRING: constraintChecks = generateGoStringSetter(subschema.(*StringSchema))
        case SCHEMATYPE_NUMBER: constraintChecks = generateGoNumericSetter(subschema.(*NumberSchema))
        case SCHEMATYPE_INTEGER: constraintChecks = generateGoNumericSetter(subschema.(*IntegerSchema))
        case SCHEMATYPE_ARRAY: constraintChecks = generateGoArraySetter(subschema.(*ArraySchema))
      }

      ret.WriteString(signature)
      ret.WriteString(constraintChecks)
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
    propertyName = getAppropriateGoCase(schema, propertyName)

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

func generateGoNumericSetter(schema NumericSchemaType) string {

  var ret bytes.Buffer
  var minimum, maximum, multiple interface{}
  var formatString, constraintTemplate string
  var constraint, comparator string

  if(!schema.HasConstraints()) {
    return ""
  }

  formatString = schema.GetConstraintFormat()

  if(schema.HasMinimum()) {

    if(schema.IsExclusiveMinimum()) {
      comparator = "<="
    } else {
      comparator = "<"
    }

    minimum = schema.GetMinimum()

    constraintTemplate = "\tif(value %s "+formatString+") {"
    constraint = fmt.Sprintf(constraintTemplate, comparator, minimum)

    constraintTemplate = "\n\t\treturn errors.New(\"Minimum value of '"+formatString+"' not met\")"
    constraint += fmt.Sprintf(constraintTemplate, minimum)

    constraint += fmt.Sprintf("\n\t}\n")
    ret.WriteString(constraint)
  }

  if(schema.HasMaximum()) {

    if(schema.IsExclusiveMaximum()) {
      comparator = ">="
    } else {
      comparator = ">"
    }

    maximum = schema.GetMaximum()

    constraintTemplate = "\tif(value %s "+formatString+") {"
    constraint = fmt.Sprintf(constraintTemplate, comparator, maximum)

    constraintTemplate = "\n\t\treturn errors.New(\"Maximum value of '"+formatString+"' not met\")"
    constraint += fmt.Sprintf(constraintTemplate, maximum)
    constraint += fmt.Sprintf("\n\t}\n")
    ret.WriteString(constraint)
  }

  if(schema.HasMultiple()) {

    multiple = schema.GetMultiple()

    constraintTemplate = "\tif(value %% "+formatString+" != 0) {"
    constraint = fmt.Sprintf(constraintTemplate, multiple)

    constraintTemplate = "\n\t\treturn errors.New(\"Value is not a multiple of '"+formatString+"'\")"
    constraint += fmt.Sprintf(constraintTemplate, multiple)

    constraint += fmt.Sprintf("\n\t}\n")
    ret.WriteString(constraint)
  }

  return ret.String()
}

func generateGoStringSetter(schema *StringSchema) string {

  var ret bytes.Buffer

  return ret.String()
}

func generateGoArraySetter(schema *ArraySchema) string {

  var ret bytes.Buffer
  var constraintTemplate string

  if(!schema.HasConstraints()) {
    return ""
  }

  ret.WriteString("\tlength := len(value)\n\n")

  if(schema.MinItems != nil) {

    constraintTemplate = fmt.Sprintf("\tif(length < %d) {\n", *schema.MinItems)
    ret.WriteString(constraintTemplate)

    constraintTemplate = fmt.Sprintf("\t\treturn errors.New(\"Minimum number of elements '%d' not present\")\n", *schema.MinItems)
    ret.WriteString(constraintTemplate)

    ret.WriteString("\t}\n\n")
  }

  if(schema.MaxItems != nil) {

    constraintTemplate = fmt.Sprintf("\tif(length > %d) {\n", *schema.MaxItems)
    ret.WriteString(constraintTemplate)

    constraintTemplate = fmt.Sprintf("\t\treturn errors.New(\"Maximum number of elements '%d' not present\")\n", *schema.MaxItems)
    ret.WriteString(constraintTemplate)

    ret.WriteString("\t}\n\n")
  }

  return ret.String()
}

func generateGoTypeForSchema(schema TypeSchema) string {

  switch schema.GetSchemaType() {
  case SCHEMATYPE_STRING:
    return "string"
  case SCHEMATYPE_INTEGER:
    return "int"
  case SCHEMATYPE_NUMBER:
    return "float64"
  case SCHEMATYPE_OBJECT:
    return "*" + ToCamelCase(schema.GetTitle())
  case SCHEMATYPE_ARRAY:
    return "[]" + ToCamelCase(schema.(*ArraySchema).Items.GetTitle())
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

func getAppropriateGoCase(schema *ObjectSchema, propertyName string) string {

  for _, constrainedName := range schema.ConstrainedProperties {
    if(constrainedName == propertyName) {
      return ToJavaCase(propertyName)
    }
  }
  return ToCamelCase(propertyName)
}
