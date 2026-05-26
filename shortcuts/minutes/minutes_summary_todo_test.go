// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package minutes

import (
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/cmdutil"
)

func TestMinutesSummary_DryRun(t *testing.T) {
	f, stdout, _, _ := cmdutil.TestFactory(t, defaultConfig())
	warmTokenCache(t)

	err := mountAndRun(t, MinutesSummary, []string{
		"+summary",
		"--minute-token", "obcn123456789",
		"--summary", "**Weekly sync**\n- follow up",
		"--dry-run",
		"--as", "user",
	}, f, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "PUT") || !strings.Contains(out, "/open-apis/minutes/v1/minutes/obcn123456789/summary") {
		t.Fatalf("dry-run output = %q", out)
	}
}

func TestMinutesTodo_DryRun(t *testing.T) {
	f, stdout, _, _ := cmdutil.TestFactory(t, defaultConfig())
	warmTokenCache(t)

	err := mountAndRun(t, MinutesTodo, []string{
		"+todo",
		"--minute-token", "obcn123456789",
		"--todo", "- finish deck",
		"--is-done",
		"--dry-run",
		"--as", "user",
	}, f, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "PUT") || !strings.Contains(out, "/open-apis/minutes/v1/minutes/obcn123456789/todo") {
		t.Fatalf("dry-run output = %q", out)
	}
}

func TestMinutesTodo_RequiresIsDone(t *testing.T) {
	f, _, stderr, _ := cmdutil.TestFactory(t, defaultConfig())
	warmTokenCache(t)

	err := mountAndRun(t, MinutesTodo, []string{
		"+todo",
		"--minute-token", "obcn123456789",
		"--todo", "finish deck",
		"--as", "user",
	}, f, stderr)
	if err == nil {
		t.Fatal("expected validation error for missing --is-done")
	}
	if !strings.Contains(err.Error(), "is-done") && !strings.Contains(err.Error(), "todo-list") {
		t.Fatalf("error = %q, want message mentioning is-done or todo-list", err.Error())
	}
}

func TestMinutesSummaryAndTodo_HelpMetadata(t *testing.T) {
	for _, tip := range MinutesSummary.Tips {
		if strings.Contains(tip, "raw text") {
			return
		}
	}
	t.Fatal("MinutesSummary tips should mention unsupported markdown display behavior")
}
