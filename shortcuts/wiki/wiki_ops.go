// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package wiki

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

var WikiExport = common.Shortcut{
	Service:     "wiki",
	Command:     "+export",
	Description: "Export a wiki document to a local Markdown file",
	Risk:        "read",
	Scopes:      []string{},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "wiki", Desc: "wiki URL or token", Required: true},
		{Name: "output", Desc: "output Markdown path"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		args := map[string]interface{}{"doc_id": runtime.Str("wiki")}
		return common.NewDryRunAPI().
			POST(common.MCPEndpoint(runtime.Config.Brand)).
			Desc("MCP tool: fetch-doc for wiki, then write markdown to local file").
			Body(map[string]interface{}{"method": "tools/call", "params": map[string]interface{}{"name": "fetch-doc", "arguments": args}}).
			Set("mcp_tool", "fetch-doc").Set("args", args)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		result, err := common.CallMCPTool(runtime, "fetch-doc", map[string]interface{}{"doc_id": runtime.Str("wiki")})
		if err != nil {
			return err
		}
		title, _ := result["title"].(string)
		markdown, _ := result["markdown"].(string)
		outputPath, err := resolveWikiOutputPath(runtime.Str("output"), title, runtime.Str("wiki"))
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return output.Errorf(output.ExitInternal, "internal_error", "cannot create output directory: %v", err)
		}
		if err := validate.AtomicWrite(outputPath, []byte(markdown), 0644); err != nil {
			return output.Errorf(output.ExitInternal, "internal_error", "cannot write markdown file: %v", err)
		}
		runtime.Out(map[string]interface{}{
			"wiki":   runtime.Str("wiki"),
			"title":  title,
			"output": outputPath,
		}, nil)
		return nil
	},
}

var WikiMemberList = common.Shortcut{
	Service:     "wiki",
	Command:     "+member-list",
	Description: "List wiki space members",
	Risk:        "read",
	Scopes:      []string{},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "space-id", Desc: "wiki space ID", Required: true},
		{Name: "page-size", Default: "20", Desc: "page size"},
		{Name: "page-token", Desc: "page token"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		params := map[string]interface{}{"page_size": wikiPageSize(runtime.Str("page-size"))}
		if v := runtime.Str("page-token"); v != "" {
			params["page_token"] = v
		}
		return common.NewDryRunAPI().
			GET("/open-apis/wiki/v2/spaces/:space_id/members").
			Params(params).
			Set("space_id", runtime.Str("space-id"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		params := map[string]interface{}{"page_size": wikiPageSize(runtime.Str("page-size"))}
		if v := runtime.Str("page-token"); v != "" {
			params["page_token"] = v
		}
		data, err := runtime.CallAPI("GET", fmt.Sprintf("/open-apis/wiki/v2/spaces/%s/members", validate.EncodePathSegment(runtime.Str("space-id"))), params, nil)
		if err != nil {
			return err
		}
		items, _ := data["members"].([]interface{})
		if items == nil {
			items, _ = data["items"].([]interface{})
		}
		runtime.OutFormat(map[string]interface{}{
			"items":      items,
			"has_more":   data["has_more"],
			"page_token": data["page_token"],
		}, nil, func(w io.Writer) {
			if len(items) == 0 {
				fmt.Fprintln(w, "No wiki members found.")
				return
			}
			rows := make([]map[string]interface{}, 0, len(items))
			for _, item := range items {
				member, _ := item.(map[string]interface{})
				if member == nil {
					continue
				}
				rows = append(rows, map[string]interface{}{
					"member_type": common.GetString(member, "member_type"),
					"member_id":   common.GetString(member, "member_id"),
					"member_role": common.GetString(member, "member_role"),
					"type":        common.GetString(member, "type"),
				})
			}
			output.PrintTable(w, rows)
		})
		return nil
	},
}

var WikiMemberAdd = common.Shortcut{
	Service:     "wiki",
	Command:     "+member-add",
	Description: "Add a member to a wiki space",
	Risk:        "write",
	Scopes:      []string{},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "space-id", Desc: "wiki space ID", Required: true},
		{Name: "member-type", Desc: "openid | userid | email | open_app_id", Required: true},
		{Name: "member-id", Desc: "member identifier", Required: true},
		{Name: "member-role", Default: "member", Desc: "member | admin"},
		{Name: "notify", Type: "bool", Default: "true", Desc: "notify the added member"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			POST("/open-apis/wiki/v2/spaces/:space_id/members").
			Set("space_id", runtime.Str("space-id")).
			Body(map[string]interface{}{
				"member": map[string]interface{}{
					"member_type": runtime.Str("member-type"),
					"member_id":   runtime.Str("member-id"),
					"member_role": runtime.Str("member-role"),
				},
				"need_notification": runtime.Bool("notify"),
			})
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		data, err := runtime.CallAPI("POST", fmt.Sprintf("/open-apis/wiki/v2/spaces/%s/members", validate.EncodePathSegment(runtime.Str("space-id"))), nil, map[string]interface{}{
			"member": map[string]interface{}{
				"member_type": runtime.Str("member-type"),
				"member_id":   runtime.Str("member-id"),
				"member_role": runtime.Str("member-role"),
			},
			"need_notification": runtime.Bool("notify"),
		})
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

var WikiMemberRemove = common.Shortcut{
	Service:     "wiki",
	Command:     "+member-remove",
	Description: "Remove a member from a wiki space",
	Risk:        "high-risk-write",
	Scopes:      []string{},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "space-id", Desc: "wiki space ID", Required: true},
		{Name: "member-type", Desc: "openid | userid | email | open_app_id", Required: true},
		{Name: "member-id", Desc: "member identifier", Required: true},
		{Name: "member-role", Default: "member", Desc: "member | admin"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			DELETE("/open-apis/wiki/v2/spaces/:space_id/members/:member_id").
			Set("space_id", runtime.Str("space-id")).
			Set("member_id", runtime.Str("member-id")).
			Body(map[string]interface{}{
				"member": map[string]interface{}{
					"member_type": runtime.Str("member-type"),
					"member_id":   runtime.Str("member-id"),
					"member_role": runtime.Str("member-role"),
				},
			})
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		data, err := runtime.CallAPI("DELETE", fmt.Sprintf("/open-apis/wiki/v2/spaces/%s/members/%s", validate.EncodePathSegment(runtime.Str("space-id")), validate.EncodePathSegment(runtime.Str("member-id"))), nil, map[string]interface{}{
			"member": map[string]interface{}{
				"member_type": runtime.Str("member-type"),
				"member_id":   runtime.Str("member-id"),
				"member_role": runtime.Str("member-role"),
			},
		})
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

func resolveWikiOutputPath(rawOutput, title, fallback string) (string, error) {
	path := rawOutput
	if path == "" {
		base := strings.Trim(strings.ToLower(title), " ._-")
		if base == "" {
			base = strings.Trim(strings.ToLower(fallback), " ._-")
		}
		if base == "" {
			base = "wiki"
		}
		base = strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-").Replace(base)
		path = base + ".md"
	}
	safePath, err := validate.SafeOutputPath(path)
	if err != nil {
		return "", output.ErrValidation("unsafe output path: %s", err)
	}
	return safePath, nil
}

func wikiPageSize(raw string) int {
	if raw == "" {
		return 20
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 20
	}
	if n < 1 {
		return 1
	}
	if n > 200 {
		return 200
	}
	return n
}
