package cli

import (
	"fmt"
	"strings"

	"github.com/kunchenguid/no-mistakes/internal/agentcfg"
	"github.com/spf13/cobra"
)

const runAgentPushOptionPrefix = "no-mistakes.run-agent="

func runAgentFromFlags(cmd *cobra.Command) (agentcfg.RunAgent, error) {
	if !cmd.Flags().Changed("agent") {
		return "", nil
	}
	value, _ := cmd.Flags().GetString("agent")
	selection := agentcfg.RunAgent(value)
	if selection == "" {
		return "", fmt.Errorf("--agent requires codex")
	}
	if err := selection.Validate(); err != nil {
		return "", err
	}
	if cmd.Flags().Changed("model") || cmd.Flags().Changed("effort") {
		return "", fmt.Errorf("--agent cannot be combined with Pi --model or --effort")
	}
	return selection, nil
}

func formatRunAgentPushOptions(selection agentcfg.RunAgent) []string {
	if selection == "" {
		return nil
	}
	return []string{runAgentPushOptionPrefix + string(selection)}
}

func parseRunAgentPushOptions(options []string) (agentcfg.RunAgent, error) {
	var selection agentcfg.RunAgent
	seen := false
	for _, option := range options {
		value, ok := strings.CutPrefix(option, runAgentPushOptionPrefix)
		if !ok {
			continue
		}
		if seen {
			return "", fmt.Errorf("duplicate run agent push option")
		}
		seen = true
		selection = agentcfg.RunAgent(value)
		if selection == "" || selection.Validate() != nil {
			return "", fmt.Errorf("invalid run agent push option")
		}
	}
	return selection, nil
}
