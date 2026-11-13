package cmd

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "soka",
		Short: "Self Optimizing Knowledge Agents — Markov-based policy analysis for agent event streams",
	}
	root.AddCommand(newSuggestCmd())
	root.AddCommand(newCheckCmd())
	root.AddCommand(newMCPCmd())
	return root
}
