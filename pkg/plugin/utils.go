package plugin

import (
	"fmt"
	"os"
	"slices"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

	if fullName == "" {
		return "", fmt.Errorf("resource %s not found", resource)
	}
	return fullName, nil
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
