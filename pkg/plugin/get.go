package plugin

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"gopkg.in/yaml.v3"
)

// RunGetCmd accepts args and cmdLine arguments and runs kubectl command with provided arguments.
// args -- are arguments passed to the command, not including flags.
// cmdLine -- is command line itself split by spaces.
func RunGetCmd(args []string, cmdLine []string) {
	// Expect TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME,
	// in which case first argument should always be a resource name.
	// If not the case, the error will be thrown later.
	resource := args[0]
	resourceName, err := getFullResourceName(resource)
	if err != nil {
		panic(err)
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
		// TODO: log that executable isn't found in PATH
	} else {
		fmt.Println("kubectl executable found")
	}

	// If --output/-o flag is not provided, set it to default value for resource
	// from ~/.kube-output/resource_tmp_map.yaml
	var ofound bool
	for _, t := range getOpts {
		if t == "--output" || t == "-o" {
			ofound = true
			break
		}
	}

	// Check if ~/.kube-output/resource_tmp_map.yaml exists
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	fPath := fmt.Sprintf("%s/.kube-output/resource_tmp_map.yaml", home)
	_, err = os.Stat(fPath)
	exists := err == nil
	if !exists && !os.IsNotExist(err) {
		panic(err)
	}

	if !ofound && exists {
		// Read ~/.kube-output/resource_tmp_map.yaml
		// If resource is found in the file, set --output/-o flag to the value from the file
		// If resource is not found in the file, set --output/-o flag to default value
		f, err := os.Open(fPath)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		buf := bytes.NewBuffer(nil)
		io.Copy(buf, f)

		rtmap := ResourceTmpMap{}
		yaml.Unmarshal(buf.Bytes(), &rtmap)

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
		// TODO: log that kubectl failed to run
	}
}
