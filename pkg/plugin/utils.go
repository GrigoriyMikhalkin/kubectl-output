package plugin

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	tmplDirPath  = "%s/.kube-output" // used as template. %s is should be replaced with home dir.
	tmplFileName = "resource_tmpl_map.yaml"
)

func getFullResourceName(resource string) (string, error) {
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
		return "", fmt.Errorf("invalid resource name: %s", resource)
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
							return "", fmt.Errorf("resource name %s is too ambiguous", t)
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
						return "", fmt.Errorf("resource name %s is too ambiguous", t)
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

	if fullName == "" {
		return "", fmt.Errorf("resource %s not found", resource)
	}
	return fullName, nil
}

func discoverAPIResources() ([]*v1.APIGroup, []*v1.APIResourceList) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	kubeconfig := fmt.Sprintf("%s/.kube/config", home)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalln(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	groups, resources, err := clientset.Discovery().ServerGroupsAndResources()
	if err != nil {
		log.Fatalln(err)
	}

	return groups, resources
}

// openTmplFile opens template file. If it doesn't exist, it creates one.
func openTmplFile() (*os.File, error) {
	var f *os.File

	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}

	dirpath := fmt.Sprintf(tmplDirPath, homedir)
	filepath := fmt.Sprintf("%s/%s", dirpath, tmplFileName)
	_, err = os.Stat(filepath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dirpath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dirpath, err)
		}

		f, err = os.Create(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %w", filepath, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to check if file %s exists: %w", filepath, err)
	} else {
		f, err = os.OpenFile(filepath, os.O_RDWR, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", filepath, err)
		}
	}

	return f, nil
}

func unmarshalResourceTmplMap(f *os.File) (ResourceTmplMap, error) {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, f); err != nil {
		return nil, fmt.Errorf("failed to read templates file: %w", err)
	}

	rtmap := ResourceTmplMap{}
	if err := yaml.Unmarshal(buf.Bytes(), rtmap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ResourceTmplMap: %w", err)
	}

	return rtmap, nil
}
