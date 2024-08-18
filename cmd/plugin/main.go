package main

import (
	"log"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE

	"github.com/GrigoriyMikhalkin/kubectl-output/cmd/plugin/cli"
)

func main() {
	// Disable date and time in logs
	log.SetFlags(0)

	cli.InitAndExecute()
}
