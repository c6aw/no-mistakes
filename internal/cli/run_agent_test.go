package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kunchenguid/no-mistakes/internal/agentcfg"
	"github.com/kunchenguid/no-mistakes/internal/git"
	"github.com/kunchenguid/no-mistakes/internal/ipc"
	"github.com/kunchenguid/no-mistakes/internal/types"
)

func TestSelectedRunRefusesOldDaemonBeforeLaunch(t *testing.T) {
	launched := olderDaemonFixture(t, nil)
	writeGlobalConfig(t, "")
	cmd := newAxiRunCmd()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--agent", "codex", "--intent", "verify selection"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err == nil || !strings.Contains(out.String(), "cannot honor --agent") {
		t.Fatalf("old daemon accepted selection: %v\n%s", err, out.String())
	}
	if len(*launched) != 0 {
		t.Fatalf("old daemon took custody: %v", *launched)
	}
}

func TestSelectedRunRefusesMismatchedActiveRunBeforeLaunch(t *testing.T) {
	var head string
	launched := olderDaemonFixture(t, nil, func(srv *ipc.Server) {
		srv.Handle(ipc.MethodProbeRunAgent, func(context.Context, json.RawMessage) (interface{}, error) {
			return ipc.ProbeRunAgentResult{OK: true}, nil
		})
		srv.Handle(ipc.MethodGetActiveRun, func(context.Context, json.RawMessage) (interface{}, error) {
			return ipc.GetActiveRunResult{Run: &ipc.RunInfo{ID: "existing-run", Status: types.RunRunning, HeadSHA: head}}, nil
		})
	})
	var err error
	head, err = git.Run(context.Background(), ".", "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	cmd := newAxiRunCmd()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--agent", "codex", "--intent", "verify selection"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err == nil || !strings.Contains(out.String(), "different agent selection") {
		t.Fatalf("mismatched active run accepted: %v\n%s", err, out.String())
	}
	if len(*launched) != 0 {
		t.Fatalf("active run was replaced: %v", *launched)
	}
}

func TestRunAgentFlagAndPushOption(t *testing.T) {
	cmd := newAxiRunCmd()
	if err := cmd.ParseFlags([]string{"--agent", "codex"}); err != nil {
		t.Fatal(err)
	}
	if got, err := runAgentFromFlags(cmd); err != nil || got != agentcfg.RunAgentCodex {
		t.Fatalf("selection = %q, %v", got, err)
	}
	selection, err := parseRunAgentPushOptions(formatRunAgentPushOptions(agentcfg.RunAgentCodex))
	if err != nil || selection != agentcfg.RunAgentCodex {
		t.Fatalf("push option = %q, %v", selection, err)
	}
	if _, err := parseRunAgentPushOptions([]string{"no-mistakes.run-agent=codex", "no-mistakes.run-agent=codex"}); err == nil {
		t.Fatal("duplicate selection accepted")
	}
}

func TestRunAgentRejectsUnsupportedAndPiSelection(t *testing.T) {
	for _, flags := range [][]string{{"--agent", "claude"}, {"--agent", "codex", "--model", "provider/model"}, {"--agent", "codex", "--effort", "high"}} {
		cmd := newAxiRunCmd()
		if err := cmd.ParseFlags(flags); err != nil {
			t.Fatal(err)
		}
		if _, err := runAgentFromFlags(cmd); err == nil {
			t.Fatalf("accepted %v", flags)
		}
	}
}
