package main

import (
	"fmt"
	"os"
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

	if(settings.ListLanguages) {
		printLanguages()
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

	err = writeGeneratedCode(schema, settings.Module, settings.OutputPath, settings.Language, settings.splitFiles)
	if err != nil {
		exitWith("Unable to generate code: %s\n", err)
		return
	}
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
