package cli

import (
	"log"

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
				name, tmpl, namespace string
				overwrite, setDefault bool
				err                   error
			)
			name, err = cmd.Flags().GetString("name")
			if err != nil {
				log.Fatalln(err)
			}

			namespace, err = cmd.Flags().GetString("namespace")
			if err != nil {
				log.Fatalln(err)
			}

			tmpl, err = cmd.Flags().GetString("custom-columns")
			if err != nil {
				log.Fatalln(err)
			}

			overwrite, err = cmd.Flags().GetBool("overwrite")
			if err != nil {
				log.Fatalln(err)
			}

			setDefault, err = cmd.Flags().GetBool("set-default")
			if err != nil {
				log.Fatalln(err)
			}

			plugin.RunSetCmd(args[0], name, tmpl, namespace, overwrite, setDefault)
		},
	}

	// Set required flags
	setCmd.PersistentFlags().String("name", "", "Name of the custom column template")
	if err := setCmd.MarkPersistentFlagRequired("name"); err != nil {
		log.Fatalln(err)
	}

	setCmd.Flags().StringP("custom-columns", "c", "", "Custom columns template, can be either a string or a file path")
	if err := setCmd.MarkFlagRequired("custom-columns"); err != nil {
		log.Fatalln(err)
	}

	// Set optional flags
	setCmd.PersistentFlags().StringP("namespace", "n", "", "Namespace to use, if not provided, template will be used for all namespaces.")
	setCmd.Flags().BoolP("overwrite", "o", false, "Overwrite existing template if exists.")
	setCmd.Flags().BoolP("set-default", "d", true, "Set template as default, true by default. If namespace is provided, template will be set as default for that namespace.")

	return setCmd
}
