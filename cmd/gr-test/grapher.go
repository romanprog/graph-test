package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func newGrapher() *Grapher {
	return &Grapher{
		DependencyMarkers: map[string]Dependency{},
		InsertYAMLMarkers: map[string]interface{}{},
		Modules:           map[string]*Module{},
	}
}

type Grapher struct {
	Modules           map[string]*Module
	DependencyMarkers map[string]Dependency
	InsertYAMLMarkers map[string]interface{}
}

func findModule(infra, name string, modsList map[string]*Module) *Module {
	mod, exists := modsList[fmt.Sprintf("%s.%s", infra, name)]
	// log.Printf("Check Mod: %s, exists: %v, list %v", name, exists, modsList)
	if !exists {
		return nil
	}
	return mod
}

func (g *Grapher) CheckDependencies() error {
	for _, mod := range g.Modules {
		for _, dep := range mod.Dependencies {
			if findModule(dep.Infra, dep.Module, g.Modules) == nil {
				return fmt.Errorf("module: '%s.%s': dependency not found: '%s.%s'", mod.Infra, mod.Name, dep.Infra, dep.Module)
			}
		}
	}
	return nil
}

func (g *Grapher) CheckGraph() error {

	modDone := map[string]*Module{}
	modWait := map[string]*Module{}

	for _, mod := range g.Modules {
		modWait[fmt.Sprintf("%s.%s", mod.Infra, mod.Name)] = mod
	}
	for c := 1; c < 20; c++ {
		doneLen := len(modDone)
		for _, mod := range modWait {
			modIndex := fmt.Sprintf("%s.%s", mod.Infra, mod.Name)
			if len(mod.Dependencies) == 0 {

				log.Printf("Mod '%s' done (%d)", modIndex, c)
				modDone[modIndex] = mod
				delete(modWait, modIndex)
				continue
			}
			var allDepsDone bool = true
			for _, dep := range mod.Dependencies {
				if findModule(dep.Infra, dep.Module, modDone) == nil {
					allDepsDone = false
					break
				}
			}
			if allDepsDone {
				log.Printf("Mod '%s' with deps %v done (%d)", modIndex, mod.Dependencies, c)
				modDone[modIndex] = mod
				delete(modWait, modIndex)
			}
		}
		if doneLen == len(modDone) {
			log.Fatalf("Unresolved dependency %v", modWait)
			return fmt.Errorf("Unresolved dependency %v", modWait)
		}
		if len(modWait) == 0 {
			return nil
		}
	}
	return nil
}

func (g *Grapher) AddDepMarker(path string) (string, error) {
	splittedPath := strings.Split(path, ".")
	if len(splittedPath) != 3 {
		return "", fmt.Errorf("bad dependency path")
	}
	dep := Dependency{
		Infra:  splittedPath[0],
		Module: splittedPath[1],
		Output: splittedPath[2],
	}
	marker := createMarker("DEP")
	g.DependencyMarkers[marker] = dep

	return fmt.Sprintf("%s", marker), nil
}

func (g *Grapher) insertYAMLMarker(data interface{}) (string, error) {
	marker := createMarker("YAML")
	g.InsertYAMLMarkers[marker] = data
	return fmt.Sprintf("%s", marker), nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func createMarker(t string) string {
	const markerLen = 10
	hash := randSeq(markerLen)
	return fmt.Sprintf("%s.%s.%s", hash, t, hash)
}

func Version() string {
	return "v0.1.1"
}

func (g *Grapher) checkMarkers(data reflect.Value, infra, module string) (reflect.Value, bool) {
	subVal := reflect.ValueOf(data.Interface())
	if subVal.Kind() == reflect.String {
		for hash := range g.InsertYAMLMarkers {
			if subVal.String() == hash {
				return reflect.ValueOf(g.InsertYAMLMarkers[hash]), true
			}
		}
		for key, marker := range g.DependencyMarkers {
			if subVal.String() == key {
				if marker.Infra == "this" {
					marker.Infra = infra
				}
				modKey := fmt.Sprintf("%s.%s", infra, module)
				mDeps := g.Modules[modKey]
				// log.Println(modKey, g.Modules)
				mDeps.Dependencies = append(mDeps.Dependencies, Dependency{
					Infra:  marker.Infra,
					Module: marker.Module,
				})
				remoteStateRef := fmt.Sprintf("${data.terraform_remote_state.%s-%s.%s}", marker.Infra, marker.Module, marker.Output)
				return reflect.ValueOf(remoteStateRef), true
			}
		}
	} else {
		g.ProcessingRecursive(data.Interface(), infra, module)
	}
	return reflect.ValueOf(nil), false
}

func (g *Grapher) ProcessingRecursive(data interface{}, infra, module string) error {
	out := reflect.ValueOf(data)
	if out.Kind() == reflect.Ptr && !out.IsNil() {
		out = out.Elem()
	}
	switch out.Kind() {
	case reflect.Slice:
		for i := 0; i < out.Len(); i++ {
			val, found := g.checkMarkers(out.Index(i), infra, module)
			if found {
				out.Index(i).Set(val)
			}
			g.checkMarkers(out.Index(i), infra, module)
		}
	case reflect.Map:
		for _, key := range out.MapKeys() {
			val, found := g.checkMarkers(out.MapIndex(key), infra, module)
			if found {
				out.SetMapIndex(key, val)
			}
		}
	}
	return nil
}

func (g *Grapher) appendModules(data interface{}, infra string) error {

	modulesSliceIf, ok := data.(map[string]interface{})["modules"]
	modulesSlice := modulesSliceIf.([]interface{})
	if !ok {
		return fmt.Errorf("Incompatible struct")
	}
	for _, moduleData := range modulesSlice {
		mName, ok := moduleData.(map[interface{}]interface{})["name"]
		if !ok {
			return fmt.Errorf("Incorrect module name")
		}
		mType, ok := moduleData.(map[interface{}]interface{})["type"]
		if !ok {
			return fmt.Errorf("Incorrect module type")
		}
		mSource, ok := moduleData.(map[interface{}]interface{})["source"]
		if !ok {
			return fmt.Errorf("Incorrect module source")
		}
		mInputs, ok := moduleData.(map[interface{}]interface{})["inputs"]
		if !ok {
			return fmt.Errorf("Incorrect module inputs")
		}
		mod := Module{
			Infra:        infra,
			Name:         mName.(string),
			Type:         mType.(string),
			Source:       mSource.(string),
			Inputs:       map[string]interface{}{},
			Dependencies: []Dependency{},
		}
		inputs := mInputs.(map[interface{}]interface{})
		for ky, vl := range inputs {
			mod.Inputs[ky.(string)] = vl
		}
		modKey := fmt.Sprintf("%s.%s", infra, mName)
		g.Modules[modKey] = &mod
		g.ProcessingRecursive(mod.Inputs, infra, mName.(string))
	}
	return nil
}

func (g *Grapher) GenCode(codeStructName string) error {
	baseOutDir := filepath.Join("./", ".outputs")
	if _, err := os.Stat(baseOutDir); os.IsNotExist(err) {
		err := os.Mkdir(baseOutDir, 0755)
		if err != nil {
			return err
		}
	}
	codeDir := filepath.Join(baseOutDir, codeStructName)
	if _, err := os.Stat(codeDir); os.IsNotExist(err) {
		err := os.Mkdir(codeDir, 0755)
		if err != nil {
			return err
		}
	}
	err := RemoveContents(codeDir)
	if err != nil {
		return err
	}
	for mName, module := range g.Modules {
		modDir := filepath.Join(codeDir, mName)
		err := os.Mkdir(modDir, 0755)
		if err != nil {
			return err
		}

		codeBlock, err := module.GenMainCodeBlockHCL()
		if err != nil {
			log.Fatal(err)
			return err
		}

		tfFile := filepath.Join(modDir, "main.tf")
		ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)

		codeBlock, err = module.GenBackendCodeBlockHCL(codeStructName)
		if err != nil {
			return err
		}

		tfFile = filepath.Join(modDir, "init.tf")
		ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
		codeBlock, err = module.GenRemoteStateCodeBlockHCL(codeStructName)
		if err != nil {
			return err
		}
		if len(codeBlock) > 1 {
			tfFile = filepath.Join(modDir, "remote_state.tf")
			ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
		}
	}
	return nil
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
