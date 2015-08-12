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

	ret.WriteString(generateJSConstructor(schema, module))
	ret.WriteString("\n")
	ret.WriteString(generateJSFunctions(schema, module))
	ret.WriteString("\n")

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
  toWrite = fmt.Sprintf("function %s.%s(", module, schema.Title)
  ret.WriteString(toWrite)

  ret.WriteString(strings.Join(parameterNames, ","))
  ret.WriteString(")\n{")

  // body
  for _, parameterName = range parameterNames {

    // TODO: use setters if the value is constrained
    toWrite = fmt.Sprintf("\n\tthis.%s = %s", parameterName, parameterName)
    ret.WriteString(toWrite)
  }

  ret.WriteString("\n}\n\n")

  return ret.String()
}

func generateJSFunctions(schema *ObjectSchema, module string) string {

  var ret bytes.Buffer

  

  return ret.String()
}
