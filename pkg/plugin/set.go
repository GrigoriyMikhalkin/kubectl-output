package plugin

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	// Check if tmpl is a file path, by checking if it contains path separators.
	// If it is, read the file and set tmpl to the content of the file.
	var tmplStr string
	if strings.Contains(tmpl, string(filepath.Separator)) {
		// Check that the file exists
		if _, err := os.Stat(tmpl); os.IsNotExist(err) {
			log.Fatalln(fmt.Sprintf("file %s does not exist", tmpl))
		}

		tmplBytes, err := os.ReadFile(tmpl)
		if err != nil {
			log.Fatalln(fmt.Errorf("failed to read file %s: %w", tmpl, err))
		}

		tmplBytesStr := string(tmplBytes)
		tmplLines := strings.Split(tmplBytesStr, "\n")
		tmplColumns := strings.Fields(tmplLines[0])
		tmplFields := strings.Fields(tmplLines[1])
		if len(tmplColumns) != len(tmplFields) {
			log.Fatalln(fmt.Sprintf("number of defined columns and fields do not match in file %s", tmpl))
		}

		for i := 0; i < len(tmplColumns); i++ {
			if tmplStr != "" {
				tmplStr += ","
			}
			tmplStr += tmplColumns[i] + ":" + tmplFields[i]
		}
	} else {
		tmplStr = tmpl
	}

	if r.Templates[tmplName] != "" {
		if overwrite {
			r.Templates[tmplName] = tmplStr
		} else {
			log.Fatalln(fmt.Sprintf("template %s already exists", tmplName))
		}
	} else {
		r.Templates[tmplName] = tmplStr
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

	if err := f.Truncate(0); err != nil {
		log.Fatalln(fmt.Errorf("failed to truncate file %s: %w", f.Name(), err))
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		log.Fatalln(fmt.Errorf("failed to seek to start of file %s: %w", f.Name(), err))
	}
	_, err = f.Write(rtmapBytes)
	if err != nil {
		log.Fatalln(err)
	}
}
