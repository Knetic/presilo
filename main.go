package main

import (
  . "presilo"
  "os"
  "path/filepath"
  "fmt"
  "io/ioutil"
)

func main() {

  var schema TypeSchema
  var settings *RunSettings
  var err error

  settings, err = ParseRunSettings()
  if(err != nil) {
    exitWith("Unable to parse run settings: %s\n", err)
    return
  }

  settings.OutputPath, err = prepareOutputPath(settings.OutputPath)
  if(err != nil) {
    exitWith("Unable to create output directory: %s\n", err)
    return
  }

  schema, err = ParseSchemaFile(settings.InputPath)
  if(err != nil) {
    exitWith("Unable to parse schema file: %s\n", err)
    return
  }

  err = generateCode(schema, settings.Module, settings.OutputPath, settings.Language)
  if(err != nil) {
    exitWith("Unable to generate code: %s\n", err)
    return
  }
}

/*
  Given the output path, returns the absolute value of it,
  and ensures that the given path exists.
*/
func prepareOutputPath(targetPath string) (string, error) {

  var err error

  targetPath, err = filepath.Abs(targetPath)
  if(err != nil) {
    return "", err
  }

  err = os.MkdirAll(targetPath, os.ModePerm)
  if(err != nil) {
    return "", err
  }

  return targetPath, nil
}

func generateCode(schema TypeSchema, module string, targetPath string, language string) error {

  var schemas []TypeSchema
  var written string
  var schemaPath string
  var err error

  schemas = RecurseObjectSchemas(schema, schemas)

  for _, schema := range schemas {

    // i know, this does a switch on the language each iteration,
    // even though language doesn't change.
    // I'm ok with that redundancy.
    switch(language) {

    case "go": written = GenerateGo(schema.(*ObjectSchema), module)
    }

    schemaPath = fmt.Sprintf("%s%s%s.%s", targetPath, string(os.PathSeparator), schema.GetTitle(), language)
    err = ioutil.WriteFile(schemaPath, []byte(written), os.ModePerm)

    if(err != nil) {
      return err
    }
  }

  return nil
}

func exitWith(message string, err error) {

  errorMessage := fmt.Sprintf(message, err.Error())

  fmt.Fprintf(os.Stderr, errorMessage)
  os.Exit(1)
}
