package main

import (
	"log"

	"github.com/ligustah/mapper/mapper"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	parser := mapper.NewParser()
	mappers, err := parser.Parse("./example")
	if err != nil {
		log.Fatal(err)
	}

	generator := mapper.NewGenerator("gen")
	for _, goInterface := range mappers {
		err = generator.GenerateInterface(goInterface)
		if err != nil {
			log.Fatal(err)
		}
	}

	generator.Render()
}
