package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"text/template"

	"gopkg.in/yaml.v2"
)

// ExecTemplate apply template data 'data' to template file 'filename'.
// Write result to 'result'.
func ExecTemplate(configName string, cName string) error {

	confData, err := ioutil.ReadFile(configName)

	if err != nil {
		log.Fatal(err)
	}

	graph := newGrapher()

	// Parse template.
	fMap := template.FuncMap{
		"remoteState":          graph.AddDepMarker,
		"insertYAML":           graph.insertYAMLMarker,
		"ReconcilerVersionTag": Version,
	}

	tmpl, err := template.New("main").Funcs(fMap).Option("missingkey=error").Parse(string(confData))

	if err != nil {
		log.Fatal(err)
		return err
	}

	templatedConf := bytes.Buffer{}
	err = tmpl.Execute(&templatedConf, nil)
	if err != nil {
		log.Fatalln(err)
	}

	//fmt.Println(templatedConf.String())
	var infrastructuresList []map[string]interface{}
	dec := yaml.NewDecoder(bytes.NewReader(templatedConf.Bytes()))
	for {
		var parsedConf = make(map[string]interface{})
		err = dec.Decode(&parsedConf)
		if err != nil {
			break
		}
		infrastructuresList = append(infrastructuresList, parsedConf)
		//log.Println(infrastructuresList)
	}

	//fmt.Println(infrastructuresList)

	for _, infra := range infrastructuresList {
		fileName, ok := infra["template"].(string)
		if !ok {
			log.Fatal("infra must contain template field")
		}

		infraName, ok := infra["name"].(string)
		if !ok {
			log.Fatal("infra must contain name field")
		}

		infraTemplate, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Fatal(err)
		}
		//log.Printf("Templating: %s", fileName)
		// fmt.Printf("%+v\n", infra)
		t, err := template.New("main").Funcs(fMap).Option("missingkey=error").Parse(string(infraTemplate))

		if err != nil {
			log.Fatal(err)
		}

		infraScenario := bytes.Buffer{}
		err = t.Execute(&infraScenario, infra)
		if err != nil {
			log.Fatal(err)
		}

		//log.Println(infraScenario.String())
		scenario := make(map[string]interface{})
		err = yaml.Unmarshal(infraScenario.Bytes(), &scenario)
		if err != nil {
			log.Fatal(err)
		}
		//log.Println(infraScenario.String())
		//graph.ProcessingRecursive(scenario)
		graph.appendModules(scenario, infraName)
	}
	for _, m := range graph.Modules {
		var depList string
		for _, dep := range m.Dependencies {
			depList += dep.Infra + "." + dep.Module + "\n"
		}
		//log.Printf("Processing module: %s\nSource: %s\nInputs:\n%s\nDeps:\n  %v\n\n", key, m.Source, JSONinputs, depList)
	}
	graph.GenCode(cName)
	graph.CheckGraph()
	return nil
}
