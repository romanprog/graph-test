package main

import (
	"fmt"
	"log"

	json "github.com/json-iterator/go"
	"github.com/rodaine/hclencoder"
)

type Module struct {
	Infra        string
	Name         string
	Type         string
	Source       string
	Inputs       map[string]interface{}
	Dependencies []Dependency
}

type Dependency struct {
	Infra  string
	Module string
	Output string
}

type ModuleTfJSON struct {
	Module map[string]interface{} `json:"module"`
}

func (m *Module) GenMainCodeBlockHCL() ([]byte, error) {
	type ModuleVars map[string]interface{}

	type HCLModule struct {
		Name       string `hcl:",key"`
		ModuleVars `hcl:",squash"`
	}
	type Config struct {
		Mod HCLModule `hcl:"module"`
	}

	inp, err := json.Marshal(m.Inputs)
	if err != nil {
		log.Fatalln(err)
	}
	unmInputs := ModuleVars{}
	err = json.Unmarshal(inp, &unmInputs)
	if err != nil {
		log.Fatalln(err)
	}

	unmInputs["source"] = m.Source
	mod := HCLModule{
		Name:       m.Name,
		ModuleVars: unmInputs,
	}

	input := Config{
		Mod: mod,
	}
	return hclencoder.Encode(input)

}

func (m *Module) GenBackendCodeBlockHCL(name string) ([]byte, error) {
	type BackendSpec struct {
		Bucket string `hcl:"bucket"`
		Key    string `hcl:"key"`
		Region string `hcl:"region"`
	}

	type BackendConfig struct {
		BlockKey    string `hcl:",key"`
		BackendSpec `hcl:",squash"`
	}

	type Terraform struct {
		Backend BackendConfig `hcl:"backend"`
		ReqVer  string        `hcl:"required_version"`
	}

	type Config struct {
		TfBlock Terraform `hcl:"terraform"`
	}

	bSpeck := BackendSpec{
		Bucket: name,
		Key:    fmt.Sprintf("%s/%s", m.Infra, m.Name),
		Region: "us-east1",
	}

	tf := Terraform{
		Backend: BackendConfig{
			BlockKey:    "s3",
			BackendSpec: bSpeck,
		},
		ReqVer: "~> 0.13",
	}

	input := Config{
		TfBlock: tf,
	}
	return hclencoder.Encode(input)

}

func (m *Module) GenRemoteStateCodeBlockHCL(name string) ([]byte, error) {

	type BackendSpec struct {
		Bucket string `hcl:"bucket"`
		Key    string `hcl:"key"`
		Region string `hcl:"region"`
	}

	type Data struct {
		KeyRemState  string      `hcl:",key"`
		KeyStateName string      `hcl:",key"`
		Backend      string      `hcl:"backend"`
		Config       BackendSpec `hcl:"config"`
	}

	type Config struct {
		TfBlock []Data `hcl:"data"`
	}

	input := Config{}

	for _, dep := range m.Dependencies {
		tf := Data{
			KeyRemState:  "terraform_remote_state",
			KeyStateName: fmt.Sprintf("%s-%s", dep.Infra, dep.Module),
			Config: BackendSpec{
				Bucket: name,
				Key:    fmt.Sprintf("%s/%s", dep.Infra, dep.Module),
				Region: "us-east1",
			},
			Backend: "s3",
		}

		input.TfBlock = append(input.TfBlock, tf)
	}

	return hclencoder.Encode(input)

}
