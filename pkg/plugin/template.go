package plugin

import (
	"fmt"
	"log"
)

func RunTemplateCommand(resource string, showAll, showNamespaces bool, tmplName, namespace string) {
	fullResourceName, err := getFullResourceName(resource)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to get full resource name: %w", err))
	}

	// Load template config file if exists
	f, err := openTmplFile()
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	rtmap, err := unmarshalResourceTmplMap(f)
	if err != nil {
		log.Fatalln(err)
	}

	// Check if there is templates for the resource
	r := rtmap[fullResourceName]
	if r == nil {
		log.Fatalln(fmt.Sprintf("no templates found for %s", resource))
	}

	// If showAll is true, show all templates for the resource
	if showAll {
		for t := range r.Templates {
			fmt.Println(t)
		}
		return
	}

	// If showNamespaces is true, show all namespaces for which custom templates are defined for the resource
	if showNamespaces {
		for ns := range r.Namespaces {
			fmt.Println(ns)
		}
		return
	}

	// If tmplName is provided, show the template details
	if tmplName != "" {
		tmpl, ok := r.Templates[tmplName]
		if !ok {
			log.Fatalln(fmt.Sprintf("template %s not found", tmplName))
		}
		fmt.Println(tmpl)
		return
	}

	// If namespace is provided, show the default template for the resource in the namespace
	if namespace != "" {
		tmpl, ok := r.Namespaces[namespace]
		if !ok {
			log.Fatalln(fmt.Sprintf("default template for %s in namespace %s not found", resource, namespace))
		}
		fmt.Println(tmpl)
		return
	}

	// If none of the above, show the default template for the resource
	fmt.Println(r.Default)
}

func RunTemplateResourcesCommand() {
	// Load template config file if exists
	f, err := openTmplFile()
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	rtmap, err := unmarshalResourceTmplMap(f)
	if err != nil {
		log.Fatalln(err)
	}

	// Print all resources for which custom templates are defined
	for r := range rtmap {
		fmt.Println(r)
	}
}
