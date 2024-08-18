package plugin

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v3"
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
	fullResourceName, err := getFullResourceName(resource)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to get full resource name: %w", err))
	}

	f, err := openTmplFile()
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	rtmap, err := unmarshalResourceTmplMap(f)
	if err != nil {
		log.Fatalln(err)
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
		// TODO: copy tmpl to ~/.kube-output/templates/tmplName if tmpl is a filepath
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
		log.Fatalln(err)
	}

	_, err = f.Write(rtmapBytes)
	if err != nil {
		log.Fatalln(err)
	}
}
