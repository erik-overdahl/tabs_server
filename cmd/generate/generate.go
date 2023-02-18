package main

import (
	"io"
	"log"
	"os"

	"github.com/erik-overdahl/tabs_server/pkg/generate"
)

const helpText = `
generate SCHEMA_FILE OUTPUT_FILE

Convert a TypeScript schema file to a Go source file

Args:
  SCHEMA_FILE    The file to convert
  OUTPUT_FILE	 The name of the resulting file, or - for stdout
`

func main() {
	log.SetFlags(log.Lshortfile)

	if len(os.Args) != 3 {
		log.Print(helpText)
		os.Exit(0)
	}

	schemaFile := os.Args[1]
	// outputFile := os.Args[2]
	file, err := os.Open(schemaFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	tokens, err := generate.TokenizeJson(content)
	if err != nil {
		log.Fatal(err)
	}
	parser := generate.MakeTokenParser()
	result, err := parser.Parse(tokens)
	if err != nil {
		log.Fatal(err)
	}
	result = generate.Clean(result)
	log.Println("SUCCESS")
}
