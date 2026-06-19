package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ligustah/mapper/mapper"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <package_path>\n", os.Args[0])
		os.Exit(1)
	}

	pkgPath := os.Args[1]
	outDir := filepath.Join(pkgPath, "gen")

	parser := mapper.NewParser()
	mappers, err := parser.Parse(pkgPath)
	if err != nil {
		log.Fatal(err)
	}

	generator := mapper.NewGenerator("gen", outDir)
	generator.SetSrcDir(pkgPath)
	for _, goInterface := range mappers {
		err = generator.GenerateInterface(goInterface)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := generator.Render(); err != nil {
		log.Fatal(err)
	}
}
