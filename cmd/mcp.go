package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start a Model Context Protocol server for real-time agent event analysis (coming soon)",
		Long: `Start a stdio-based MCP server that accepts agent event streams and exposes
Markov policy analysis as callable tools:

  - suggest   rank MFPT-reducing policy knobs from a live event stream
  - classes   detect epistemic thrashing loops in real time
  - drift     report Jensen-Shannon drift between fast and slow kernels

Compatible with LangGraph, LlamaIndex, Claude Agent SDK, and any MCP client.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("MCP server not yet implemented — coming soon")
		},
	}
}
