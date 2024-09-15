package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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
			// Check if kubectl executable is available
			if _, err := exec.LookPath("kubectl"); err != nil {
				return fmt.Errorf("kubectl executable not found in PATH")
			}

			// Call kubectl with provided args
			c := exec.Command("kubectl", os.Args[1:]...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("failed to run kubectl: %w", err)
			}

			return nil
		},
	}

	cobra.OnInitialize(initConfig)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if strings.HasPrefix(filepath.Base(os.Args[0]), "kubectl-") {
		cmd.SetUsageTemplate(kubectlUsageTemplate)
	}

	return cmd
}

func InitAndExecute() {
	rootCmd := RootCmd()

	// Get command
	kubernetesConfigFlags := genericclioptions.NewConfigFlags(false)
	kubernetesConfigFlags.AddFlags(getCmd.Flags())
	rootCmd.AddCommand(getCmd)

	// Set command
	rootCmd.AddCommand(SetCmd())

	// Template command
	rootCmd.AddCommand(TemplateCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}

var kubectlUsageTemplate = `Usage:{{if .Runnable}}
  kubectl {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  kubectl {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "kubectl {{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
