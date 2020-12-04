package main

import (
	"fmt"
	"log"

	"github.com/rodaine/hclencoder"
)

func TestEncoder() {

	type ModuleVars map[string]string

	type HCLModule struct {
		Name       string `hcl:",key"`
		ModuleVars `hcl:",squash"`
	}
	type Config struct {
		Mod HCLModule `hcl:"module"`
	}

	mod := HCLModule{
		Name: "modName",
		ModuleVars: ModuleVars{
			"Var1": "asddasd",
			"Var2": "2312345",
		},
	}

	input := Config{
		Mod: mod,
	}

	hcl, err := hclencoder.Encode(input)
	if err != nil {
		log.Fatal("unable to encode: ", err)
	}

	fmt.Print(string(hcl))
}
