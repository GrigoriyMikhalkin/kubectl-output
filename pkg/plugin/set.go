package plugin

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	tmplDirPath  = "%s/.kube-output" // used as template. %s is should be replaced with home dir.
	tmplFileName = "resource_tmpl_map.yaml"
)

type ResourceTmplMap map[string]*Resource

type Resource struct {
	Default    string
	Namespaces map[string]string
	Templates  map[string]string
}

func RunSetCmd(resource, tmplName, tmpl, namespace string, overwrite, setDefault bool) {
	// Update files in ~/.kube-output
	// File resource_tmpl_map.yaml has the following format:
	// resource:
	//   default: tmplName
	//   namespaces:
	//     namespace: tmplName
	//   templates:
	//     tmplName: tmpl
	var f *os.File

	fullResourceName, err := getFullResourceName(resource)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to get full resource name: %w", err))
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	dirpath := fmt.Sprintf(tmplDirPath, homedir)
	filepath := fmt.Sprintf("%s/%s", dirpath, tmplFileName)
	_, err = os.Stat(filepath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dirpath, 0755); err != nil {
			panic(err)
		}

		f, err = os.Create(filepath)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	} else {
		f, err = os.Open(filepath)
		if err != nil {
			panic(err)
		}
	}

	defer f.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, f); err != nil {
		panic(fmt.Errorf("failed to read %s: %w", filepath, err))
	}

	rtmap := ResourceTmplMap{}
	if err := yaml.Unmarshal(buf.Bytes(), rtmap); err != nil {
		panic(fmt.Errorf("failed to unmarshal ResourceTmplMap: %w", err))
	}

	r := rtmap[fullResourceName]
	if r == nil {
		r = &Resource{
			Default:    "",
			Namespaces: map[string]string{},
			Templates:  map[string]string{},
		}
		rtmap[fullResourceName] = r
	}

	if r.Templates[tmplName] != "" {
		if overwrite {
			r.Templates[tmplName] = tmpl
		} else {
			log.Fatalln(fmt.Sprintf("template %s already exists", tmplName))
		}
	} else {
		r.Templates[tmplName] = tmpl
		// TODO: copy tmp to ~/.kube-output/templates/tmpName if tmp is a file
	}

	if setDefault {
		if namespace == "" {
			r.Default = tmplName
		} else {
			r.Namespaces[namespace] = tmplName
		}
	} else if namespace != "" {
		log.Println("--namespace flag is ignored if --set-default flag is set to false")
	}

	// Write to resource_tmpl_map.yaml file
	rtmapBytes, err := yaml.Marshal(rtmap)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(rtmapBytes)
	if err != nil {
		panic(err)
	}
}
