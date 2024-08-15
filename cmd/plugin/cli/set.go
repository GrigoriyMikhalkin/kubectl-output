package cli

import (
	"github.com/spf13/cobra"

	"github.com/GrigoriyMikhalkin/kubectl-output/pkg/plugin"
)

// SetCmd represents the set command
func SetCmd() *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set TYPE[.VERSION][.GROUP] --name=template_name -c=... [flags]",
		Short: "Set custom column template for a resource",
		Long: `Set custom column template for a resource. For example:
k o set deployment --name=replicas -c=NAME:.metadata.name,READY:.status.readyReplicas,REPLICAS:.spec.replicas
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var (
				name, tmp, namespace  string
				overwrite, setDefault bool
				err                   error
			)
			name, err = cmd.Flags().GetString("name")
			if err != nil {
				panic(err)
			}

			namespace, err = cmd.Flags().GetString("namespace")
			if err != nil {
				panic(err)
			}

			tmp, err = cmd.Flags().GetString("custom-columns")
			if err != nil {
				panic(err)
			}

			overwrite, err = cmd.Flags().GetBool("overwrite")
			if err != nil {
				panic(err)
			}

			setDefault, err = cmd.Flags().GetBool("set-default")
			if err != nil {
				panic(err)
			}

			plugin.RunSetCmd(args[0], name, tmp, namespace, overwrite, setDefault)
		},
	}

	// Set required flags
	setCmd.PersistentFlags().String("name", "", "Name of the custom column template")
	if err := setCmd.MarkPersistentFlagRequired("name"); err != nil {
		panic(err)
	}

	setCmd.Flags().StringP("custom-columns", "c", "", "Custom columns template, can be either a string or a file path")
	if err := setCmd.MarkFlagRequired("custom-columns"); err != nil {
		panic(err)
	}

	// Set optional flags
	setCmd.PersistentFlags().StringP("namespace", "n", "", "Namespace to use, if not provided, template will be used for all namespaces.")
	setCmd.Flags().BoolP("overwrite", "o", false, "Overwrite existing template if exists.")
	setCmd.Flags().BoolP("set-default", "d", true, "Set template as default, true by default. If namespace is provided, template will be set as default for that namespace.")

	return setCmd
}
