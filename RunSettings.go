package main

import (
	"errors"
	"flag"
	"fmt"
)

type RunSettings struct {
	InputPath  string
	OutputPath string
	Language   string
	Module     string
	splitFiles bool
}

var SUPPORTED_LANGUAGES = []string{
	"go",
	"js",
	/*
	  "java",
	  "cs",
	  "py",
	  "sql",
	*/
}

func ParseRunSettings() (*RunSettings, error) {

	var ret *RunSettings

	ret = new(RunSettings)
	flag.StringVar(&ret.Language, "l", "go", "Language for the generated files")
	flag.StringVar(&ret.OutputPath, "o", "./", "Optional destination directory for generated files")
	flag.StringVar(&ret.Module, "m", "main", "Module to use for the generated files")

	flag.Parse()
	ret.InputPath = flag.Arg(0)

	if len(ret.InputPath) == 0 {
		return nil, errors.New("Input path not specified")
	}

	if !containsString(SUPPORTED_LANGUAGES, ret.Language) {
		errorMessage := fmt.Sprintf("Language '%s' not supported. Supported languages; %v\n", ret.Language, SUPPORTED_LANGUAGES)
		return nil, errors.New(errorMessage)
	}

	// determine if the given language is one that should split files.
	if(ret.Language == "js") {
		ret.splitFiles = false
	} else {
		ret.splitFiles = true
	}

	return ret, nil
}

func containsString(values []string, value string) bool {

	for _, supportedValue := range values {
		if supportedValue == value {
			return true
		}
	}

	return false
}
