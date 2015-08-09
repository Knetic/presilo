package presilo

import (
  "bytes"
)

/*
  Generates valid Go code for a given schema.
*/
func GenerateGo(schema TypeSchema, module string) string {

  var ret bytes.Buffer

  ret.WriteString("package " + module)
  return ret.String()
}
