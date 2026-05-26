// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package minutes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

type minuteTodoPayload struct {
	Content   string   `json:"content"`
	Assignees []string `json:"assignees,omitempty"`
	IsDone    bool     `json:"is_done"`
}

// MinutesTodo updates one or more todos items on a minute.
var MinutesTodo = common.Shortcut{
	Service:     "minutes",
	Command:     "+todo",
	Description: "Update one or more todo items on a minute",
	Risk:        "write",
	Scopes:      []string{"minutes:minutes:update"},
	AuthTypes:   []string{"user"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "minute-token", Desc: "minute token (required)", Required: true},
		{Name: "todo-list", Desc: "JSON array of todos: [{\"content\":\"...\",\"is_done\":true}] (required unless --todo/--is-done are used)", Input: []string{common.File, common.Stdin}},
		{Name: "todo", Desc: "single todo plain-text content; must be used with --is-done", Input: []string{common.File, common.Stdin}},
		{Name: "is-done", Type: "bool", Desc: "single todo completion flag; must be used with --todo"},
	},
	Tips: []string{
		"Prefer `--todo-list @todos.json` to update multiple todos in one request.",
		"For a single todo, `--todo` and `--is-done` must be provided together.",
		"`content` is plain text only; markdown formatting is not supported.",
		"Use `lark-cli vc +notes --minute-tokens <token>` to read current todos before writing.",
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		minuteToken := runtime.Str("minute-token")
		if minuteToken == "" {
			return output.ErrValidation("--minute-token is required")
		}
		if err := validate.ResourceName(minuteToken, "--minute-token"); err != nil {
			return output.ErrValidation("%s", err)
		}
		if _, err := resolveMinuteTodoList(runtime); err != nil {
			return err
		}
		return nil
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			PUT(fmt.Sprintf("/open-apis/minutes/v1/minutes/%s/todo", validate.EncodePathSegment(runtime.Str("minute-token")))).
			Body(map[string]interface{}{
				"todo_list": "<todo_list array>",
			})
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		minuteToken := runtime.Str("minute-token")
		todoList, err := resolveMinuteTodoList(runtime)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/open-apis/minutes/v1/minutes/%s/todo", validate.EncodePathSegment(minuteToken))
		body := map[string]interface{}{
			"todo_list": todoList,
		}
		if _, err := runtime.CallAPI(http.MethodPut, path, nil, body); err != nil {
			return err
		}

		runtime.OutFormat(map[string]interface{}{
			"minute_token": minuteToken,
			"todo_count":   len(todoList),
			"updated":      true,
		}, nil, nil)
		return nil
	},
}

func resolveMinuteTodoList(runtime *common.RuntimeContext) ([]minuteTodoPayload, error) {
	todoListRaw := strings.TrimSpace(runtime.Str("todo-list"))
	todo := strings.TrimSpace(runtime.Str("todo"))
	hasTodoList := todoListRaw != ""
	hasTodo := todo != ""
	hasIsDone := runtime.Changed("is-done")

	if hasTodoList {
		if hasTodo || hasIsDone {
			return nil, output.ErrValidation("use either --todo-list or --todo/--is-done, not both")
		}
		return parseMinuteTodoListJSON(todoListRaw)
	}
	if hasTodo != hasIsDone {
		return nil, output.ErrValidation("--todo and --is-done must be provided together")
	}
	if !hasTodo {
		return nil, output.ErrValidation("--todo-list is required (or provide --todo with --is-done)")
	}
	return []minuteTodoPayload{{
		Content: todo,
		IsDone:  runtime.Bool("is-done"),
	}}, nil
}

func parseMinuteTodoListJSON(raw string) ([]minuteTodoPayload, error) {
	var todoList []minuteTodoPayload
	if err := json.Unmarshal([]byte(raw), &todoList); err != nil {
		return nil, output.ErrValidation("invalid --todo-list JSON: %v", err)
	}
	if len(todoList) == 0 {
		return nil, output.ErrValidation("--todo-list must contain at least one todo")
	}
	for i, item := range todoList {
		if strings.TrimSpace(item.Content) == "" {
			return nil, output.ErrValidation("todo_list[%d].content is required", i)
		}
	}
	return todoList, nil
}
