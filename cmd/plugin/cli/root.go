package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "kubectl-output is a kubectl plugin that allows to set custom output format for specific resources/namespaces.",
		Long: `kubectl-output is a kubectl plugin that allows to set custom output format for specific resources/namespaces.
Custom output format is based on custom-columns: [https://kubernetes.io/docs/reference/kubectl/#custom-columns]. 

Example: kubectl output set pods -o custom-columns=NAME:.metadata.name,STATUS:.status.phase,NAMESPACE:.metadata.namespace`,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(os.Args[1:])

			// Check if kubectl executable is available
			if _, err := exec.LookPath("kubectl"); err != nil {
				return fmt.Errorf("kubectl executable not found in PATH")
			} else {
				fmt.Println("kubectl executable found")
			}

			// Call kubectl with provided args
			c := exec.Command("kubectl", os.Args[1:]...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("failed to run kubectl: %w", err)
			}

			//_ := get.CustomColumnsPrinter{}

			//log := logger.NewLogger()
			//log.Info("")
			//
			//s := spin.New()
			//finishedCh := make(chan bool, 1)
			//namespaceName := make(chan string, 1)
			//go func() {
			//	lastNamespaceName := ""
			//	for {
			//		select {
			//		case <-finishedCh:
			//			fmt.Printf("\r")
			//			return
			//		case n := <-namespaceName:
			//			lastNamespaceName = n
			//		case <-time.After(time.Millisecond * 100):
			//			if lastNamespaceName == "" {
			//				fmt.Printf("\r  \033[36mSearching for namespaces\033[m %s", s.Next())
			//			} else {
			//				fmt.Printf("\r  \033[36mSearching for namespaces\033[m %s (%s)", s.Next(), lastNamespaceName)
			//			}
			//		}
			//	}
			//}()
			//defer func() {
			//	finishedCh <- true
			//}()

			//if err := plugin.RunPlugin(KubernetesConfigFlags, namespaceName); err != nil {
			//	return errors.Unwrap(err)
			//}

			//log.Info("")

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}

func InitAndExecute() {
	rootCmd := RootCmd()

	// Get command
	rootCmd.AddCommand(getCmd)

	// Set command
	setCmd := SetCmd()
	rootCmd.AddCommand(setCmd)

	// List command
	rootCmd.AddCommand(listCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}
