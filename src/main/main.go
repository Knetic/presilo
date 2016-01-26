package main

import (
	"fmt"
	"os"
	"path/filepath"
	"presilo"
)

func main() {

	var parseContext *presilo.SchemaParseContext
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

	parseContext, err = parseSchemas(settings.InputPaths)
	if(err != nil) {
		exitWith("Unable to parse schemas: %s\n", err)
		return
	}

	err = presilo.WriteGeneratedCode(parseContext, settings.Module, settings.OutputPath, settings.Language, settings.TabStyle, settings.UnsafeModule, settings.splitFiles)
	if err != nil {
		exitWith("Unable to generate code: %s\n", err)
		return
	}
}

func parseSchemas(paths []string) (*presilo.SchemaParseContext, error) {

	var parseContext *presilo.SchemaParseContext
	var err error

	parseContext = presilo.NewSchemaParseContext()

	for _, path := range paths {
		
		_, err = presilo.ParseSchemaFileContinue(path, parseContext)
		if err != nil {
			return nil, err
		}
	}

	err = presilo.LinkSchemas(parseContext)
	return parseContext, err
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
