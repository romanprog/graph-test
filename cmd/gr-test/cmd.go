package main

import (
	"flag"
	"log"
)

func main() {
	confName := flag.String("c", "", "main config filename")
	outDir := flag.String("n", "", "name of iac")
	flag.Parse()
	if *confName == "" {
		log.Fatalln("ERR: option -c <config_file.yaml> required")
	}
	if *outDir == "" {
		log.Fatalln("ERR: option -n <name> required")
	}
	ExecTemplate(*confName, *outDir)
}
