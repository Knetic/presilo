package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	. "presilo"
	"sync"
)

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

func writeGeneratedCode(schema TypeSchema, module string, targetPath string, language string, splitFiles bool) error {

	var wg sync.WaitGroup
	var err error

	err = generateCode(schema, module, targetPath, language, splitFiles, &wg)
	wg.Wait()

	return err
}

func generateCode(schema TypeSchema, module string, targetPath string, language string, splitFiles bool, wg *sync.WaitGroup) error {

	var schemas []*ObjectSchema
	var objectSchema *ObjectSchema
	var generator func(*ObjectSchema, string) string
	var writtenChannel chan string
	var fileNameChannel chan string
	var errorChannel chan error
	var written string
	var schemaPath string

	if schema.GetSchemaType() != SCHEMATYPE_OBJECT {
		errorMsg := fmt.Sprintf("Could not generate code for '%s', it was not an object.", schema.GetTitle())
		return errors.New(errorMsg)
	}

	schemas = RecurseObjectSchemas(schema, schemas)

	// figure out which code generator to use
	switch language {

	case "go":
		generator = GenerateGo
	case "js":
		generator = GenerateJS
	case "java":
		generator = GenerateJava
	case "cs":
		generator = GenerateCSharp
	case "rb":
		generator = GenerateRuby
	case "py":
		generator = GeneratePython
	default: return errors.New("No valid language specified")
	}

	writtenChannel = make(chan string)
	errorChannel = make(chan error)
	defer close(writtenChannel)
	defer close(errorChannel)

	// one for the file writer, one for error listener.
	wg.Add(2)

	// Start writer goroutines based on our split strategy
	if splitFiles {

		fileNameChannel = make(chan string)
		defer close(fileNameChannel)
		go writeSplitFiles(writtenChannel, fileNameChannel, errorChannel, wg)

	} else {
		schemaPath = fmt.Sprintf("%s%s%s.%s", targetPath, string(os.PathSeparator), schema.GetTitle(), language)
		go writeSingleFile(schemaPath, writtenChannel, errorChannel, wg)
	}

	// write errors to stderr, no matter where they come from.
	go writeErrors(errorChannel, wg)

	// generate schemas, pass to writers.
	for _, objectSchema = range schemas {

		if splitFiles {
			schemaPath = fmt.Sprintf("%s%s%s.%s", targetPath, string(os.PathSeparator), objectSchema.GetTitle(), language)
			fileNameChannel <- schemaPath
		}

		written = generator(objectSchema, module)
		writtenChannel <- written
	}

	return nil
}

/*
	Writes incoming strings to incoming file names.
*/
func writeSplitFiles(source chan string, fileNames chan string, resultError chan error, wg *sync.WaitGroup) {

	var schemaPath, contents string
	var err error
	var ok bool

	defer wg.Done()

	for {

		schemaPath, ok = <-fileNames
		if !ok {
			return
		}

		contents = <-source

		err = ioutil.WriteFile(schemaPath, []byte(contents), os.ModePerm)

		if err != nil {
			resultError <- err
		}
	}
}

/*
	Writes all incoming contents from [source] to a file at the given [schemaPath],
	returning all found errors to [resultError], and returning only once a value is
	received on [exit], or if the file was unable to be opened.
*/
func writeSingleFile(schemaPath string, source chan string, resultError chan error, wg *sync.WaitGroup) {

	var contents string
	var outFile *os.File
	var writer *bufio.Writer
	var err error
	var ok bool

	defer wg.Done()

	outFile, err = os.Create(schemaPath)
	if err != nil {
		resultError <- err
		return
	}

	writer = bufio.NewWriter(outFile)

	for {
		contents, ok = <-source
		if !ok {
			break
		}

		_, err = writer.Write([]byte(contents))
		if err != nil {
			resultError <- err
		}
	}

	writer.Flush()
}

/*
  Writes all incoming errors to stderr.
*/
func writeErrors(intake chan error, wg *sync.WaitGroup) {

	var err error
	var ok bool

	defer wg.Done()

	for {
		err, ok = <-intake
		if !ok {
			return
		}

		fmt.Fprintf(os.Stderr, err.Error())
	}
}
