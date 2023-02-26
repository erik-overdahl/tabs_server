package main

import (
	// "encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/dave/jennifer/jen"
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
	cleaned:= generate.Clean(result)
	output, err := generate.Convert(cleaned)
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range output {
		ns, ok := item.(*generate.SchemaNamespace)
		if !ok || ns.Name == "manifest" {
			continue
		}
		pkg := generate.MakePkg(ns.Name)
		if err != nil {
			log.Fatal(err)
		}
		pkg.AddNamespaceProperties(ns.Properties)
		if len(ns.Functions) > 0 {
			pkg.AddClient()
		}
		for _, f := range ns.Functions {
			pkg.AddFunction(f)
		}
		for _, t := range ns.Types {
			switch t := t.(type) {
			case *generate.SchemaObjectProperty:
				pkg.AddStruct(t, "")
			case *generate.SchemaStringProperty:
				pkg.AddEnum(t)
			}
		}
		for _, f := range []*jen.File{pkg.TypeFile, pkg.ClientFile} {
			rendered, err := generate.RenderGo(f)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(rendered))
			fmt.Println("--------------")
		}
	}
	log.Println("SUCCESS")
}
