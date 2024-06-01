package plugin

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"slices"
	"strings"
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

	fullResourceName := getFullResourceName(resource)

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

	r := rtmap[fullResourceName]
	if r == nil {
		r = &Resource{
			Default:    "",
			Namespaces: map[string]string{},
			Templates:  map[string]string{},
		}
		rtmap[fullResourceName] = r
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

func getFullResourceName(resource string) string {
	// Should split resource by '.' to extract type, version and group if available
	var t, v, g string
	resourceParts := strings.Split(resource, ".")
	switch len(resourceParts) {
	case 1:
		t = resourceParts[0]
	case 2:
		t, g = resourceParts[0], resourceParts[1]
	case 3:
		t, v, g = resourceParts[0], resourceParts[1], resourceParts[2]
	default:
		panic(fmt.Errorf("invalid resource name: %s", resource))
	}

	groups, resources := discoverAPIResources()

	groupVersions := make(map[string]struct{})
	if g != "" {
		for _, group := range groups {
			if group.Name == g {
				if v != "" {
					for _, version := range group.Versions {
						if version.Version == v {
							groupVersions[version.GroupVersion] = struct{}{}
							break
						}
					}
				} else {
					for _, version := range group.Versions {
						groupVersions[version.GroupVersion] = struct{}{}
					}
				}
				break
			}
		}
	}

	var fullName string
	for _, resourceList := range resources {
		if len(groupVersions) > 0 {
			if _, ok := groupVersions[resourceList.GroupVersion]; ok {
				for _, r := range resourceList.APIResources {
					matches := r.Name == t || r.SingularName == t || slices.Contains(r.ShortNames, t)
					if matches {
						if fullName != "" {
							panic(fmt.Errorf("resource name %s is too ambiguous", t))
						}

						gv := strings.Split(resourceList.GroupVersion, "/")
						if len(gv) == 1 {
							// Means that group is not specified
							fullName = r.Name
						} else {
							fullName = fmt.Sprintf("%s.%s.%s", r.Name, gv[1], gv[0])
						}
						break
					}
				}
			}
		} else {
			for _, r := range resourceList.APIResources {
				matches := r.Name == t || r.SingularName == t || slices.Contains(r.ShortNames, t)
				if matches {
					if fullName != "" {
						panic(fmt.Errorf("resource name %s is too ambiguous", t))
					}
					gv := strings.Split(resourceList.GroupVersion, "/")
					if len(gv) == 1 {
						// Means that group is not specified
						fullName = r.Name
					} else {
						fullName = fmt.Sprintf("%s.%s.%s", r.Name, gv[1], gv[0])
					}
					break
				}
			}
		}

	}

	return fullName
}

func discoverAPIResources() ([]*v1.APIGroup, []*v1.APIResourceList) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	kubeconfig := fmt.Sprintf("%s/.kube/config", home)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	groups, resources, err := clientset.Discovery().ServerGroupsAndResources()
	if err != nil {
		panic(err)
	}

	return groups, resources
}
