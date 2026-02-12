package cmds

import "github.com/spf13/cobra"

func newCompletionCommand() *cobra.Command {
	completion := &cobra.Command{Use: "completion", Short: "Generate shell completion"}

	completion.AddCommand(&cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		},
	})

	completion.AddCommand(&cobra.Command{
		Use:   "zsh",
		Short: "Generate zsh completion",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		},
	})

	return completion
}
