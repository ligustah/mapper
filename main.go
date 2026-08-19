package main

import (
	"flag"
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
	strictEnums := flag.Bool("strict-enums", false,
		"fail instead of warning when a generated conversion between two enums can produce a value the target does not declare")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <package_path>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	pkgPath := flag.Arg(0)
	outDir := filepath.Join(pkgPath, "gen")

	parser := mapper.NewParser()
	mappers, err := parser.Parse(pkgPath)
	if err != nil {
		log.Fatal(err)
	}

	generator := mapper.NewGenerator("gen", outDir)
	generator.SetSrcDir(pkgPath)
	generator.SetStrictEnums(*strictEnums)
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
