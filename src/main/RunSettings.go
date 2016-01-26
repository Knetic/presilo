package main

import (
	"errors"
	"flag"
	"fmt"
)

type RunSettings struct {
	InputPaths    []string
	OutputPath    string
	Language      string
	Module        string
	ListLanguages bool
	TabStyle      string
	UnsafeModule  bool
	splitFiles    bool
}

var SUPPORTED_LANGUAGES = []string{
	"go",
	"js",
	"java",
	"cs",
	"rb",
	"py",
	"mysql",
	/*
		"lua",
		"mssql",
		"c",
		"cpp",
	*/
}

func ParseRunSettings() (*RunSettings, error) {

	var ret *RunSettings

	ret = new(RunSettings)
	flag.StringVar(&ret.Language, "l", "go", "Language for the generated files")
	flag.StringVar(&ret.OutputPath, "o", "./", "Optional destination directory for generated files")
	flag.StringVar(&ret.Module, "m", "main", "Module to use for the generated files")
	flag.StringVar(&ret.TabStyle, "t", "\t", "The indent delimiter ('tab style') to use")
	flag.BoolVar(&ret.UnsafeModule, "unsafemodule", false, "Ignore errors regarding module name validation")
	flag.BoolVar(&ret.ListLanguages, "a", false, "Lists all supported languages")

	flag.Parse()
	ret.InputPaths = flag.Args()

	if ret.ListLanguages {
		return ret, nil
	}

	if len(ret.InputPaths) == 0 {
		return nil, errors.New("No input files specified")
	}

	if !containsString(SUPPORTED_LANGUAGES, ret.Language) {
		errorMessage := fmt.Sprintf("Language '%s' not supported. Supported languages; %v\n", ret.Language, SUPPORTED_LANGUAGES)
		return nil, errors.New(errorMessage)
	}

	// determine if the given language is one that should split files.
	if ret.Language == "js" || ret.Language == "mysql" {
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
