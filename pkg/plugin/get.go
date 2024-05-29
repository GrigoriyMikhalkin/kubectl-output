package plugin

import (
	"fmt"
	"os"
	"os/exec"
)

func RunGetCmd(args []string) {
	fmt.Println("get called: ", args)
	getOpts := args[1:]

	// Check if kubectl executable is available
	if _, err := exec.LookPath("kubectl"); err != nil {
		// TODO: log that executable isn't found in PATH
	} else {
		fmt.Println("kubectl executable found")
	}

	// Call kubectl with provided args
	c := exec.Command("kubectl", getOpts...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		// TODO: log that kubectl failed to run
	}
}
