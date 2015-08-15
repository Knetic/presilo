package presilo

import (
	"bytes"
	//"fmt"
	//"strings"
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
  ret.WriteString(generateJavaVariables(schema))
	ret.WriteString("\n")
	ret.WriteString(generateJavaConstructor(schema))
	ret.WriteString("\n")
	ret.WriteString(generateJavaFunctions(schema))
	ret.WriteString("\n")

	return ret.String()
}


func generateJavaImports(schema *ObjectSchema) string {

  var ret bytes.Buffer
  return ret.String()
}

func generateJavaTypeDeclaration(schema *ObjectSchema) string {

  var ret bytes.Buffer
  return ret.String()
}

func generateJavaVariables(schema *ObjectSchema) string {

  var ret bytes.Buffer
  return ret.String()
}

func generateJavaConstructor(schema *ObjectSchema) string {

  var ret bytes.Buffer
  return ret.String()
}

func generateJavaFunctions(schema *ObjectSchema) string {

  var ret bytes.Buffer
  return ret.String()
}
