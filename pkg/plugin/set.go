package plugin

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type ResourceTmpMap map[string]*Resource

type Resource struct {
	Default    string
	Namespaces map[string]string
	Templates  map[string]string
}

func RunSetCmd(resource, tmpName, tmp, namespace string, overwrite, setDefault bool) {
	fmt.Println("set called")
	// Update files in ~/.kube-output
	// File resource_tmp_map.yaml has following format:
	// resource:
	//   default: tmpName
	//   namespaces:
	//     namespace: tmpName
	//   templates:
	//     tmpName: tmp

	var (
		err error
		f   *os.File
	)

	_, err = os.Stat("~/.kube-output/resource_tmp_map.yaml")
	if os.IsNotExist(err) {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		if err = os.MkdirAll(fmt.Sprintf("%s/.kube-output", home), 0755); err != nil {
			panic(err)
		}

		f, err = os.Create(fmt.Sprintf("%s/.kube-output/resource_tmp_map.yaml", home))
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	} else {
		// Open file
		f, err = os.Open("~/.kube-output/resource_tmp_map.yaml")
		if err != nil {
			panic(err)
		}
	}

	defer f.Close()

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, f)

	rtmap := ResourceTmpMap{}
	yaml.Unmarshal(buf.Bytes(), &rtmap)

	r := rtmap[resource]
	if r == nil {
		r = &Resource{
			Default:    "",
			Namespaces: map[string]string{},
			Templates:  map[string]string{},
		}
		rtmap[resource] = r
	}

	if r.Templates[tmpName] != "" {
		if overwrite {
			r.Templates[tmpName] = tmp
		} else {
			// TODO: log that template already exists
		}
	} else {
		r.Templates[tmpName] = tmp
		// TODO: copy tmp to ~/.kube-output/templates/tmpName if tmp is a file
	}

	if setDefault {
		if namespace == "" {
			r.Default = tmpName
		} else {
			r.Namespaces[namespace] = tmpName
		}
	} else if namespace != "" {
		// TODO: log that --namespace flag is ignored if set-default is set to false
	}

	// Write to resource_tmp_map.yaml file
	rtmapBytes, err := yaml.Marshal(rtmap)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(rtmapBytes)
	if err != nil {
		panic(err)
	}
}
