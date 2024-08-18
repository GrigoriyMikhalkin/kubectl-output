/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"log"

	"github.com/GrigoriyMikhalkin/kubectl-output/pkg/plugin"

	"github.com/spf13/cobra"
)

// TemplateCmd represents the template command
func TemplateCmd() *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   "template TYPE[.VERSION][.GROUP] --name=template_name [flags]",
		Short: "Show info about available templates",
		Long: `Show info about available templates. To show all resources for which custom templates are define:
k o template --resources

To show all templates for a specific resource:
k o template deployment --all

To show default template for a resource in specific namespace:
k o template deployment --namespace=default

To show a specific template:
k o template deployment --name=replicas
`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				showResources, showAll, showNamespaces bool
				name, namespace                        string
				err                                    error
			)

			showResources, err = cmd.Flags().GetBool("resources")
			if err != nil {
				log.Fatalln(err)
			}

			showAll, err = cmd.Flags().GetBool("all")
			if err != nil {
				log.Fatalln(err)
			}

			showNamespaces, err = cmd.Flags().GetBool("namespaces")
			if err != nil {
				log.Fatalln(err)
			}

			name, err = cmd.Flags().GetString("name")
			if err != nil {
				log.Fatalln(err)
			}

			namespace, err = cmd.Flags().GetString("namespace")
			if err != nil {
				log.Fatalln(err)
			}

			if len(args) == 0 && !showResources {
				log.Fatalln("Either resource type or --resources flag must be provided")
			}

			if showResources {
				plugin.RunTemplateResourcesCommand()
			} else {
				plugin.RunTemplateCommand(args[0], showAll, showNamespaces, name, namespace)
			}

		},
	}

	// Set optional flags
	templateCmd.Flags().BoolP("resources", "r", false, "Show all resources for which custom templates are defined")
	templateCmd.Flags().BoolP("all", "a", false, "Show all templates for a specific resource")
	templateCmd.Flags().Bool("namespaces", false, "Show all namespaces for which custom templates are defined for the resource")
	templateCmd.Flags().StringP("name", "n", "", "Name of the template")
	templateCmd.Flags().String("namespace", "", "Namespace of the resource. When specified, will show default template for the resource in the namespace")

	return templateCmd
}
