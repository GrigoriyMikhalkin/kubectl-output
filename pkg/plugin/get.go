package plugin

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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

	getOpts := cmdLine[1:]

	// TODO: Read template config file
	//var rtmap ResourceTmpMap
	//cmds := getCmdPerResource(getOpts)
	//for _, c := range cmds {
	//	runCmd(c, rtmap)
	//}

	// Check if kubectl executable is available
	if _, err = exec.LookPath("kubectl"); err != nil {
		log.Fatalln("kubectl executable not found in PATH")
	}

	// If --output/-o flag is not provided, set it to default value for resource
	// from ~/.kube-output/resource_tmpl_map.yaml
	var ofound bool
	for _, t := range getOpts {
		if t == "--output" || t == "-o" {
			ofound = true
			break
		}
	}

	if !ofound {
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
			if r.Default != "" {
				getOpts = append(getOpts, "--output", fmt.Sprintf("custom-columns=%s", r.Templates[r.Default]))
			}
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
