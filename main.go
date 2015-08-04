package main

import (
  "presilo"
  "os"
  "fmt"
)

func main() {

  var schema *Schema
  var schemaContents map[string]interface{}
  var settings *RunSettings
  var err error

  settings, err = ParseRunSettings()
  if(err != nil) {
    exitWith("Unable to parse run settings: %s\n", err)
    return
  }

  schemaContents, err = parseSchemaContents(settings.InputPath)
  if(err != nil) {
    exitWith("Unable to parse schema file: %s\n", err)
    return
  }

  schema, err = parseSchema(schemaContents)
  if(err != nil) {
    exitWith("Unable to parse schema file as schema: %s\n", err)
    return
  }

  fmt.Printf("Hello, %v\n", schema)
}

func exitWith(message string, err error) {

  errorMessage := fmt.Sprintf(message, err.Error())

  fmt.Fprintf(os.Stderr, errorMessage)
  os.Exit(1)
}
