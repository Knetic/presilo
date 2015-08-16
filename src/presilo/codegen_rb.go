package presilo

import (
	"fmt"
	//"strings"
	"bytes"
)

func GenerateRuby(schema *ObjectSchema, module string) string {

	var ret bytes.Buffer
	var toWrite string

	toWrite = fmt.Sprintf("module %s\n\n", module)
	ret.WriteString(toWrite)

	ret.WriteString(generateRubySignature(schema))
	ret.WriteString("\n")
	ret.WriteString(generateRubyConstructor(schema))
	ret.WriteString("\n")
	ret.WriteString(generateRubyFunctions(schema))
	ret.WriteString("\n")

	return ret.String()
}

func generateRubySignature(schema *ObjectSchema) string {

	var ret bytes.Buffer
	return ret.String()
}

func generateRubyConstructor(schema *ObjectSchema) string {

	var ret bytes.Buffer
	return ret.String()
}

func generateRubyFunctions(schema *ObjectSchema) string {

	var ret bytes.Buffer
	return ret.String()
}
