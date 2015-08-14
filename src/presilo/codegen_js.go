package presilo

import (
	"bytes"
  "fmt"
  "strings"
)

/*
  Generates valid JS code for a given schema.
*/
func GenerateJS(schema *ObjectSchema, module string) string {

	var ret bytes.Buffer

  ret.WriteString(generateJSModuleCheck(module))
	ret.WriteString("\n")
  ret.WriteString(generateJSConstructor(schema, module))
	ret.WriteString("\n")
	ret.WriteString(generateJSFunctions(schema, module))
	ret.WriteString("\n")

	return ret.String()
}

func generateJSModuleCheck(module string) string {

  var ret bytes.Buffer
  var check string

  // check for undefined, first.
  check = fmt.Sprintf("if(typeof(%s) === \"undefined\")\n{", module)
  ret.WriteString(check)

  check = fmt.Sprintf("\n\t%s = {}", module)
  ret.WriteString(check)
  ret.WriteString("\n}\n")

  // then null check.
  ret.WriteString("else\n{\n")

  check = fmt.Sprintf("\t%s = %s || {}", module, module)
  ret.WriteString(check)
  ret.WriteString("\n}\n")
  return ret.String()
}

func generateJSConstructor(schema *ObjectSchema, module string) string {

  var ret bytes.Buffer
  var parameterNames []string
  var toWrite, propertyName, parameterName string

  // generate list of property names
  for _, propertyName = range schema.RequiredProperties {

		propertyName = ToJavaCase(propertyName)
		parameterNames = append(parameterNames, propertyName)
	}

  // write constructor signature
	toWrite = fmt.Sprintf("\n/*\n%s\n*/\n", schema.Description)
	ret.WriteString(toWrite)

  toWrite = fmt.Sprintf("%s.%s = function(", module, schema.Title)
  ret.WriteString(toWrite)

  ret.WriteString(strings.Join(parameterNames, ","))
  ret.WriteString(")\n{")

  // body
  for _, parameterName = range parameterNames {

    toWrite = fmt.Sprintf("\n\tthis.set%s(%s)", ToCamelCase(parameterName), parameterName)
    ret.WriteString(toWrite)
  }

  ret.WriteString("\n}\n\n")

  return ret.String()
}

func generateJSFunctions(schema *ObjectSchema, module string) string {

  var ret bytes.Buffer
	var subschema TypeSchema
	var propertyName, propertyNameCamel, propertyNameJava, schemaName, toWrite string

	schemaName = ToCamelCase(schema.Title)

	for propertyName, subschema = range schema.Properties {

		propertyNameCamel = ToCamelCase(propertyName)
		propertyNameJava = ToJavaCase(propertyName)

		toWrite = fmt.Sprintf("%s.%s.prototype.set%s = function(value)\n{\n", module, schemaName, propertyNameCamel)
		ret.WriteString(toWrite)

    // undefined check
    ret.WriteString("\tif(typeof(value) === 'undefined')\n\t{\n")

    toWrite = fmt.Sprintf("\t\tthrow new ReferenceError(\"Cannot set property '%s', no value given\")", propertyNameJava)
    ret.WriteString(toWrite)

    ret.WriteString("\n\t}\n")

    switch subschema.GetSchemaType() {
    case SCHEMATYPE_BOOLEAN:
      toWrite = ""
    case SCHEMATYPE_STRING:
      toWrite = generateJSStringSetter(subschema.(*StringSchema))
    case SCHEMATYPE_INTEGER: fallthrough
    case SCHEMATYPE_NUMBER:
      toWrite = generateJSNumericSetter(subschema.(NumericSchemaType))
    case SCHEMATYPE_OBJECT:
      toWrite = generateJSObjectSetter(subschema.(*ObjectSchema))
    case SCHEMATYPE_ARRAY:
      toWrite = generateJSArraySetter(subschema.(*ArraySchema))
    }

    ret.WriteString(toWrite)

		toWrite = fmt.Sprintf("\n\tthis.%s = value\n}\n", propertyNameJava)
		ret.WriteString(toWrite)
	}
  return ret.String()
}

/*
	Returns checks appropriate for verifying an object's type.
*/
func generateJSObjectSetter(schema *ObjectSchema) string {

	var ret bytes.Buffer

	ret.WriteString(generateJSTypeCheck(schema))
	return ret.String()
}

/*
	Returns checks appropriate for verifying a numeric value and its constraints.
*/
func generateJSNumericSetter(schema NumericSchemaType) string {

	var ret bytes.Buffer

	ret.WriteString(generateJSTypeCheck(schema))

	if(schema.HasMinimum()) {
		ret.WriteString(generateJSRangeCheck(schema.GetMinimum(), "value", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<"))
	}

	if(schema.HasMaximum()) {
		ret.WriteString(generateJSRangeCheck(schema.GetMaximum(), "value", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">"))
	}

  if(schema.HasEnum()) {
	   ret.WriteString(generateJSEnumCheck(schema, schema.GetEnum(), "", ""))
   }
	// TODO: multiple check
	return ret.String()
}

/*
	Returns checks appropriate for verifying a string value and its constraints.
*/
func generateJSStringSetter(schema *StringSchema) string {

	var ret bytes.Buffer
	var toWrite string

	ret.WriteString(generateJSTypeCheck(schema))

	if(schema.MinLength != nil) {
		ret.WriteString(generateJSRangeCheck(*schema.MinLength, "value.length", "%d", false, "<", ""))
	}

	if(schema.MaxLength != nil) {
		ret.WriteString(generateJSRangeCheck(*schema.MaxLength, "value.length", "%d", false, ">", ""))
	}

	if(schema.Pattern != nil) {

		toWrite = fmt.Sprintf("\tvar regex = new RegExp(\"%s\")\n", *schema.Pattern)
		ret.WriteString(toWrite)

		toWrite = fmt.Sprintf("\tif(!regex.test(value))\n\t{")
		ret.WriteString(toWrite)

		toWrite = fmt.Sprintf("\n\t\tthrow new Error(\"Property '\"+value+\"' did not match pattern '%s'\")", *schema.Pattern)
		ret.WriteString(toWrite)

		ret.WriteString("\n\t}\n")
	}

  if(schema.Enum != nil) {
	   ret.WriteString(generateJSEnumCheck(schema, schema.GetEnum(), "\"", "\""))
   }
	return ret.String()
}

/*
	Returns checks appropriate for verifying an array value and its constraints.
*/
func generateJSArraySetter(schema *ArraySchema) string {

	var ret bytes.Buffer

	ret.WriteString(generateJSTypeCheck(schema))
	// TODO: value uniformity check

	if(schema.MinItems != nil) {
		ret.WriteString(generateJSRangeCheck(*schema.MinItems, "value.length", "%d", false, "<", ""))
	}

	if(schema.MaxItems != nil) {
		ret.WriteString(generateJSRangeCheck(*schema.MaxItems, "value.length", "%d", false, ">", ""))
	}
	return ret.String()
}

func generateJSRangeCheck(value interface{}, reference string, format string, exclusive bool, comparator, exclusiveComparator string) string {

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

	toWrite = fmt.Sprintf("\n\t\tthrow new RangeError(\"Property '\"+value+\"' was out of range.\")\n\t}\n")
	ret.WriteString(toWrite)

	return ret.String()
}

/*
	Generates code which throws an error if the given [parameter]'s type name is not equal to the given [typeName]
*/
func generateJSTypeCheck(schema TypeSchema) string {

  var ret bytes.Buffer
	var schemaType SchemaType
  var toWrite, expectedType string
	var shouldWriteCtorCheck bool

	schemaType = schema.GetSchemaType()
  expectedType = getJSTypeFromSchemaType(schemaType)

  toWrite = fmt.Sprintf("\tif(typeof(value) !== \"%s\")\n\t{", expectedType)
  ret.WriteString(toWrite)

  toWrite = fmt.Sprintf("\n\t\tthrow new TypeError(\"Property \"+value+\" was not of the expected type '%s'\")", expectedType)
  ret.WriteString(toWrite)
  ret.WriteString("\n\t}\n")

	// if this is an array or object, check the constructor
	shouldWriteCtorCheck = false
	switch schemaType {
	case SCHEMATYPE_ARRAY:
		shouldWriteCtorCheck = true
		expectedType = "Array"
	case SCHEMATYPE_OBJECT:
		shouldWriteCtorCheck = true
		expectedType = ToCamelCase(schema.GetTitle())
	}

	if(shouldWriteCtorCheck) {

		toWrite = fmt.Sprintf("\tif(value.constructor !== %s)\n\t{", expectedType)
		ret.WriteString(toWrite)

		toWrite = fmt.Sprintf("\n\t\tthrow new TypeError(\"Property '\"+value+\"'was not of the expected type '%s'\")", expectedType)
		ret.WriteString(toWrite)

		ret.WriteString("\n\t}\n")
	}

  return ret.String()
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generateJSEnumCheck(schema interface{}, enumValues []interface{}, prefix string, postfix string) string {

	var ret bytes.Buffer
	var constraint string
	var length int

	length = len(enumValues)

	if(length <= 0) {
		return ""
	}

	// write array of valid values
	constraint = fmt.Sprintf("\tvar validValues = [%s%v%s", prefix, enumValues[0], postfix)
	ret.WriteString(constraint)

	for _, enumValue := range enumValues[1:length] {

		constraint = fmt.Sprintf(",%s%v%s", prefix, enumValue, postfix)
		ret.WriteString(constraint)
	}
	ret.WriteString("]\n")

	// compare
	ret.WriteString("\tvar isValid = false\n")
	ret.WriteString("\tfor(var i = 0; i < validValues.length; i++) \n\t{\n")
	ret.WriteString("\t\tif(validValues[i] === value)\n\t\t{\n\t\t\tisValid = true")
	ret.WriteString("\n\t\t\tbreak\n\t\t}\n\t}")

	ret.WriteString("\n\tif(!isValid)\n\t{")
	ret.WriteString("\n\t\tthrow new Error(\"Given value '\"+value+\"' was not found in list of acceptable values\")\n")
	ret.WriteString("\t}\n")

	return ret.String()
}

func getJSTypeFromSchemaType(schemaType SchemaType) string {

  switch(schemaType) {
  case SCHEMATYPE_NUMBER: fallthrough
  case SCHEMATYPE_INTEGER: return "number"
  case SCHEMATYPE_ARRAY: fallthrough
  case SCHEMATYPE_OBJECT: return "object"
  case SCHEMATYPE_STRING: return "string"
  case SCHEMATYPE_BOOLEAN: return "boolean"
  }

  return "object"
}
