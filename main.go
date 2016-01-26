package main

import (
	"fmt"
	"os"
	"path/filepath"
	"presilo"
)

func main() {

	var schema presilo.TypeSchema
	var settings *RunSettings
	var err error

	settings, err = ParseRunSettings()
	if err != nil {
		exitWith("Unable to parse run settings: %s\n", err)
		return
	}

	if settings.ListLanguages {
		printLanguages()
		return
	}

	settings.OutputPath, err = prepareOutputPath(settings.OutputPath)
	if err != nil {
		exitWith("Unable to create output directory: %s\n", err)
		return
	}

	schema, _, err = presilo.ParseSchemaFile(settings.InputPath)
	if err != nil {
		exitWith("Unable to parse schema file: %s\n", err)
		return
	}

	err = presilo.WriteGeneratedCode(schema, settings.Module, settings.OutputPath, settings.Language, settings.TabStyle, settings.UnsafeModule, settings.splitFiles)
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

func printLanguages() {

	for _, language := range SUPPORTED_LANGUAGES {
		fmt.Println(language)
	}
}

func exitWith(message string, err error) {

	errorMessage := fmt.Sprintf(message, err.Error())

	fmt.Fprintf(os.Stderr, errorMessage)
	os.Exit(1)
}
