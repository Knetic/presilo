package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	. "presilo"
)

func main() {

	var schema TypeSchema
	var settings *RunSettings
	var err error

	settings, err = ParseRunSettings()
	if err != nil {
		exitWith("Unable to parse run settings: %s\n", err)
		return
	}

	settings.OutputPath, err = prepareOutputPath(settings.OutputPath)
	if err != nil {
		exitWith("Unable to create output directory: %s\n", err)
		return
	}

	schema, err = ParseSchemaFile(settings.InputPath)
	if err != nil {
		exitWith("Unable to parse schema file: %s\n", err)
		return
	}

	err = generateCode(schema, settings.Module, settings.OutputPath, settings.Language)
	if err != nil {
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
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(targetPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	return targetPath, nil
}

func generateCode(schema TypeSchema, module string, targetPath string, language string) error {

	var schemas []*ObjectSchema
	var objectSchema *ObjectSchema
	var generator func(*ObjectSchema, string) string
	var written string
	var schemaPath string
	var err error

	if schema.GetSchemaType() != SCHEMATYPE_OBJECT {
		errorMsg := fmt.Sprintf("Could not generate code for '%s', it was not an object.", schema.GetTitle())
		return errors.New(errorMsg)
	}

	objectSchema = schema.(*ObjectSchema)
	schemas = RecurseObjectSchemas(schema, schemas)

	// figure out which code generator to use
	switch language {

	case "go":
		generator = GenerateGo
	}

	// write schemas
	for _, objectSchema = range schemas {

		written = generator(objectSchema, module)

		schemaPath = fmt.Sprintf("%s%s%s.%s", targetPath, string(os.PathSeparator), objectSchema.GetTitle(), language)
		err = ioutil.WriteFile(schemaPath, []byte(written), os.ModePerm)

		if err != nil {
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
