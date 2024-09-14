package plugin

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// RunGetCmd accepts args and cmdLine arguments and runs kubectl command with provided arguments.
// args -- are arguments passed to the command, not including flags.
// cmdLine -- is command line itself split by spaces.
func RunGetCmd(args []string, cmdLine []string) {
	resource := args[0]
	resourceName, err := getFullResourceName(resource)
	if err != nil {
		log.Fatalln(err)
	}

	// Check if kubectl executable is available
	if _, err = exec.LookPath("kubectl"); err != nil {
		log.Fatalln("kubectl executable not found in PATH")
	}

	// If --output/-o flag is not provided, set it to default value for resource
	// from ~/.kube-output/resource_tmpl_map.yaml
	var output string
	var ns string
	getOpts := cmdLine[1:]
	for i, t := range getOpts {
		if strings.HasPrefix(t, "--output") || strings.HasPrefix(t, "-o") {
			if t != "--output" && t != "-o" {
				output = strings.TrimPrefix(t, "--output")
				output = strings.TrimPrefix(output, "-o")
				output = strings.TrimPrefix(output, "=")
			} else {
				// Check that next argument is not a flag
				nextOpt := getOpts[i+1]
				if !strings.HasPrefix(nextOpt, "-") {
					output = nextOpt
				}
			}
		}
		if strings.HasPrefix(t, "--namespace") || strings.HasPrefix(t, "-n") {
			if t != "--namespace" && t != "-n" {
				ns = strings.TrimPrefix(t, "--namespace")
				ns = strings.TrimPrefix(ns, "-n")
				ns = strings.TrimPrefix(ns, "=")
			} else {
				// Check that next argument is not a flag
				nextOpt := getOpts[i+1]
				if !strings.HasPrefix(nextOpt, "-") {
					ns = nextOpt
				}
			}
		}
	}

	// Read template config file.
	// If resource is found in the file, set --output/-o flag to the value from the file
	// If resource is not found in the file, set --output/-o flag to default value
	f, err := openTmplFile()
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	rtmap, err := unmarshalResourceTmplMap(f)
	if err != nil {
		log.Fatalln(err)
	}

	r := rtmap[resourceName]
	if r != nil {
		var cc string

		if output != "" {
			if tmpl, ok := r.Templates[output]; ok {
				cc = fmt.Sprintf("custom-columns=%s", tmpl)
			}
		}

		if cc == "" && ns != "" {
			if tmplName, _ := r.Namespaces[ns]; tmplName != "" {
				if tmpl, ok := r.Templates[tmplName]; ok {
					cc = fmt.Sprintf("custom-columns=%s", tmpl)
				}
			}
		}

		if cc == "" && r.Default != "" {
			cc = fmt.Sprintf("custom-columns=%s", r.Templates[r.Default])
		}

		if cc != "" {
			getOpts = append(getOpts, "--output", cc)

		}
	}

	// Call kubectl with provided args
	c := exec.Command("kubectl", getOpts...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		log.Fatalln(fmt.Errorf("failed to run kubectl get command: %w", err))
	}
}
