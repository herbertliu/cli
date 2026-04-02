// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/util"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

var DocsImport = common.Shortcut{
	Service:     "docs",
	Command:     "+import",
	Description: "Create a Lark document from a local Markdown file",
	Risk:        "write",
	Scopes:      []string{"docx:document:create"},
	AuthTypes:   []string{"user", "bot"},
	Flags: []common.Flag{
		{Name: "file", Desc: "local Markdown file", Required: true},
		{Name: "title", Desc: "document title (default: file name)"},
		{Name: "folder-token", Desc: "parent folder token"},
		{Name: "wiki-node", Desc: "wiki node token"},
		{Name: "wiki-space", Desc: "wiki space ID"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		if _, err := validate.SafeInputPath(runtime.Str("file")); err != nil {
			return output.ErrValidation("unsafe file path: %s", err)
		}
		return validateDocCreateTarget(runtime)
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		args := docsCreateArgs(runtime)
		args["markdown"] = "<contents from " + runtime.Str("file") + ">"
		return common.NewDryRunAPI().
			POST(common.MCPEndpoint(runtime.Config.Brand)).
			Desc("MCP tool: create-doc from local Markdown file").
			Body(map[string]interface{}{"method": "tools/call", "params": map[string]interface{}{"name": "create-doc", "arguments": args}}).
			Set("mcp_tool", "create-doc").Set("args", args)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		filePath, err := validate.SafeInputPath(runtime.Str("file"))
		if err != nil {
			return output.ErrValidation("unsafe file path: %s", err)
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return output.ErrValidation("cannot read file: %s", err)
		}
		args := docsCreateArgs(runtime)
		if _, ok := args["title"]; !ok {
			args["title"] = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		}
		args["markdown"] = string(content)
		result, err := common.CallMCPTool(runtime, "create-doc", args)
		if err != nil {
			return err
		}
		result["source_file"] = runtime.Str("file")
		runtime.Out(result, nil)
		return nil
	},
}

var DocsExport = common.Shortcut{
	Service:     "docs",
	Command:     "+export",
	Description: "Export a Lark document to a local Markdown file",
	Risk:        "read",
	Scopes:      []string{"docx:document:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "doc", Desc: "document URL or token", Required: true},
		{Name: "output", Desc: "output Markdown path"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		args := map[string]interface{}{"doc_id": runtime.Str("doc")}
		return common.NewDryRunAPI().
			POST(common.MCPEndpoint(runtime.Config.Brand)).
			Desc("MCP tool: fetch-doc, then write markdown to local file").
			Body(map[string]interface{}{"method": "tools/call", "params": map[string]interface{}{"name": "fetch-doc", "arguments": args}}).
			Set("mcp_tool", "fetch-doc").Set("args", args).Set("output", runtime.Str("output"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		result, err := common.CallMCPTool(runtime, "fetch-doc", map[string]interface{}{"doc_id": runtime.Str("doc")})
		if err != nil {
			return err
		}
		title, _ := result["title"].(string)
		markdown, _ := result["markdown"].(string)
		markdown = fixExportedMarkdown(markdown)
		outputPath, err := resolveDocTextOutputPath(runtime.Str("output"), title, "document", ".md")
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
			"doc":    runtime.Str("doc"),
			"title":  title,
			"output": outputPath,
		}, nil)
		return nil
	},
}

var DocsExportFile = common.Shortcut{
	Service:     "docs",
	Command:     "+export-file",
	Description: "Export a document as PDF or DOCX via async export task",
	Risk:        "read",
	Scopes:      []string{"drive:file:download"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "doc", Desc: "document URL or token", Required: true},
		{Name: "file-extension", Default: "pdf", Desc: "pdf | docx", Enum: []string{"pdf", "docx"}},
		{Name: "output", Desc: "output file path"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		target, dry := dryRunResolvedDocumentTarget(ctx, runtime, runtime.Str("doc"))
		step := 1
		if target.InputKind == "wiki" {
			step = 2
		}
		return dry.
			POST("/open-apis/drive/v1/export_tasks").
			Desc(fmt.Sprintf("[%d] Create export task", step)).
			Body(map[string]interface{}{
				"export_task": map[string]interface{}{
					"token":          target.Token,
					"type":           target.Kind,
					"file_extension": runtime.Str("file-extension"),
				},
			}).
			GET("/open-apis/drive/v1/export_tasks/:ticket").
			Desc(fmt.Sprintf("[%d] Poll export task result", step+1)).
			Params(map[string]interface{}{"token": target.Token}).
			Set("ticket", "<ticket>").
			GET("/open-apis/drive/v1/export_tasks/:file_token/download").
			Desc(fmt.Sprintf("[%d] Download exported file", step+2)).
			Set("file_token", "<exported_file_token>")
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		target, err := resolveDocumentTarget(runtime, runtime.Str("doc"))
		if err != nil {
			return err
		}
		if target.Kind != "doc" && target.Kind != "docx" {
			return output.ErrValidation("docs +export-file supports doc/docx only, got %q", target.Kind)
		}
		ticketData, err := runtime.CallAPI("POST", "/open-apis/drive/v1/export_tasks", nil, map[string]interface{}{
			"export_task": map[string]interface{}{
				"token":          target.Token,
				"type":           target.Kind,
				"file_extension": runtime.Str("file-extension"),
			},
		})
		if err != nil {
			return err
		}
		ticket := common.GetString(ticketData, "ticket")
		if ticket == "" {
			return output.Errorf(output.ExitAPI, "api_error", "export task created without ticket")
		}
		fileToken, err := waitDocExportTask(runtime, ticket, target.Token)
		if err != nil {
			return err
		}
		outputPath, err := resolveDocTextOutputPath(runtime.Str("output"), target.Title, target.Token, "."+runtime.Str("file-extension"))
		if err != nil {
			return err
		}
		if err := downloadDocExportFile(runtime, fileToken, outputPath); err != nil {
			return err
		}
		runtime.Out(map[string]interface{}{
			"doc":        runtime.Str("doc"),
			"doc_token":  target.Token,
			"format":     runtime.Str("file-extension"),
			"file_token": fileToken,
			"output":     outputPath,
		}, nil)
		return nil
	},
}

var DocsImportFile = common.Shortcut{
	Service:     "docs",
	Command:     "+import-file",
	Description: "Import a local DOCX file as a new document via async import task",
	Risk:        "write",
	Scopes:      []string{"drive:file:upload"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "file", Desc: "local DOCX file", Required: true},
		{Name: "title", Desc: "new document title"},
		{Name: "folder-token", Desc: "target folder token"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		if _, err := validate.SafeInputPath(runtime.Str("file")); err != nil {
			return output.ErrValidation("unsafe file path: %s", err)
		}
		if strings.ToLower(filepath.Ext(runtime.Str("file"))) != ".docx" {
			return output.ErrValidation("docs +import-file currently supports .docx files only")
		}
		return nil
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		fileName := runtime.Str("title")
		if fileName == "" {
			fileName = strings.TrimSuffix(filepath.Base(runtime.Str("file")), filepath.Ext(runtime.Str("file")))
		}
		return common.NewDryRunAPI().
			POST("/open-apis/drive/v1/files/upload_all").
			Desc("[1] Upload local DOCX file").
			Body(map[string]interface{}{
				"file_name":   filepath.Base(runtime.Str("file")),
				"parent_type": "explorer",
				"parent_node": runtime.Str("folder-token"),
				"file":        "@" + runtime.Str("file"),
			}).
			POST("/open-apis/drive/v1/import_tasks").
			Desc("[2] Create import task").
			Body(map[string]interface{}{
				"import_task": map[string]interface{}{
					"file_extension": "docx",
					"file_token":     "<uploaded_file_token>",
					"type":           "docx",
					"file_name":      fileName,
					"point": map[string]interface{}{
						"mount_type": 1,
						"mount_key":  runtime.Str("folder-token"),
					},
				},
			}).
			GET("/open-apis/drive/v1/import_tasks/:ticket").
			Desc("[3] Poll import task result").
			Set("ticket", "<ticket>")
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		filePath, err := validate.SafeInputPath(runtime.Str("file"))
		if err != nil {
			return output.ErrValidation("unsafe file path: %s", err)
		}
		info, err := os.Stat(filePath)
		if err != nil {
			return output.ErrValidation("cannot read file: %s", err)
		}
		fileName := runtime.Str("title")
		if fileName == "" {
			fileName = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		}
		fileToken, err := uploadImportSourceFile(ctx, runtime, filePath, filepath.Base(filePath), runtime.Str("folder-token"), info.Size())
		if err != nil {
			return err
		}
		body := map[string]interface{}{
			"import_task": map[string]interface{}{
				"file_extension": "docx",
				"file_token":     fileToken,
				"type":           "docx",
				"file_name":      fileName,
			},
		}
		if folderToken := runtime.Str("folder-token"); folderToken != "" {
			body["import_task"].(map[string]interface{})["point"] = map[string]interface{}{
				"mount_type": 1,
				"mount_key":  folderToken,
			}
		}
		taskData, err := runtime.CallAPI("POST", "/open-apis/drive/v1/import_tasks", nil, body)
		if err != nil {
			return err
		}
		ticket := common.GetString(taskData, "ticket")
		if ticket == "" {
			return output.Errorf(output.ExitAPI, "api_error", "import task created without ticket")
		}
		docToken, url, err := waitDocImportTask(runtime, ticket)
		if err != nil {
			return err
		}
		runtime.Out(map[string]interface{}{
			"ticket":      ticket,
			"source_file": runtime.Str("file"),
			"file_token":  fileToken,
			"doc_token":   docToken,
			"url":         url,
		}, nil)
		return nil
	},
}

var DocsCallout = common.Shortcut{
	Service:     "docs",
	Command:     "+callout",
	Description: "Insert a callout block using Markdown alert syntax",
	Risk:        "write",
	Scopes:      []string{"docx:document:write_only", "docx:document:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "doc", Desc: "document URL or token", Required: true},
		{Name: "type", Default: "note", Desc: "note | tip | important | warning | caution", Enum: []string{"note", "tip", "important", "warning", "caution"}},
		{Name: "title", Desc: "callout title"},
		{Name: "text", Desc: "callout body text", Required: true},
		{Name: "mode", Default: "append", Desc: "append | overwrite | replace_range | replace_all | insert_before | insert_after", Enum: []string{"append", "overwrite", "replace_range", "replace_all", "insert_before", "insert_after"}},
		{Name: "selection-with-ellipsis", Desc: "content locator (e.g. start...end)"},
		{Name: "selection-by-title", Desc: "title locator (e.g. ## Section)"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		mode := runtime.Str("mode")
		if mode == "delete_range" || !validModes[mode] {
			return common.FlagErrorf("invalid --mode %q for callout", mode)
		}
		selEllipsis := runtime.Str("selection-with-ellipsis")
		selTitle := runtime.Str("selection-by-title")
		if selEllipsis != "" && selTitle != "" {
			return common.FlagErrorf("--selection-with-ellipsis and --selection-by-title are mutually exclusive")
		}
		if needsSelection[mode] && selEllipsis == "" && selTitle == "" {
			return common.FlagErrorf("--%s mode requires --selection-with-ellipsis or --selection-by-title", mode)
		}
		return nil
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		args := docsUpdateArgs(runtime, buildCalloutMarkdown(runtime))
		return common.NewDryRunAPI().
			POST(common.MCPEndpoint(runtime.Config.Brand)).
			Desc("MCP tool: update-doc with callout markdown").
			Body(map[string]interface{}{"method": "tools/call", "params": map[string]interface{}{"name": "update-doc", "arguments": args}}).
			Set("mcp_tool", "update-doc").Set("args", args)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		args := docsUpdateArgs(runtime, buildCalloutMarkdown(runtime))
		result, err := common.CallMCPTool(runtime, "update-doc", args)
		if err != nil {
			return err
		}
		runtime.Out(result, nil)
		return nil
	},
}

var DocsBlocks = common.Shortcut{
	Service:     "docs",
	Command:     "+blocks",
	Description: "List document blocks; supports recursive traversal for all blocks",
	Risk:        "read",
	Scopes:      []string{"docx:document:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "doc", Desc: "document URL or token", Required: true},
		{Name: "all", Type: "bool", Desc: "recursively collect all blocks"},
		{Name: "page-size", Default: "200", Desc: "page size for children listing"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		target, dry := dryRunResolvedDocumentTarget(ctx, runtime, runtime.Str("doc"))
		step := 1
		if target.InputKind == "wiki" {
			step = 2
		}
		return dry.GET("/open-apis/docx/v1/documents/:document_id/blocks/:block_id/children").
			Desc(fmt.Sprintf("[%d] List document blocks", step)).
			Params(map[string]interface{}{"page_size": docsPageSize(runtime.Str("page-size"))}).
			Set("document_id", target.Token).
			Set("block_id", target.Token)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		target, err := resolveDocumentTarget(runtime, runtime.Str("doc"))
		if err != nil {
			return err
		}
		if target.Kind != "docx" {
			return output.ErrValidation("docs +blocks supports docx only, got %q", target.Kind)
		}
		blocks, err := collectDocBlocks(runtime, target.Token, runtime.Bool("all"), docsPageSize(runtime.Str("page-size")))
		if err != nil {
			return err
		}
		runtime.OutFormat(map[string]interface{}{
			"document_id": target.Token,
			"items":       blocks,
			"count":       len(blocks),
		}, nil, func(w io.Writer) {
			rows := make([]map[string]interface{}, 0, len(blocks))
			for _, block := range blocks {
				children, _ := block["children"].([]interface{})
				rows = append(rows, map[string]interface{}{
					"block_id":    common.GetString(block, "block_id"),
					"block_type":  block["block_type"],
					"parent_id":   common.GetString(block, "parent_id"),
					"child_count": len(children),
				})
			}
			if len(rows) == 0 {
				fmt.Fprintln(w, "No blocks found.")
				return
			}
			output.PrintTable(w, rows)
		})
		return nil
	},
}

func docsCreateArgs(runtime *common.RuntimeContext) map[string]interface{} {
	args := map[string]interface{}{}
	if v := runtime.Str("title"); v != "" {
		args["title"] = v
	}
	if v := runtime.Str("folder-token"); v != "" {
		args["folder_token"] = v
	}
	if v := runtime.Str("wiki-node"); v != "" {
		args["wiki_node"] = v
	}
	if v := runtime.Str("wiki-space"); v != "" {
		args["wiki_space"] = v
	}
	return args
}

func docsUpdateArgs(runtime *common.RuntimeContext, markdown string) map[string]interface{} {
	args := map[string]interface{}{
		"doc_id":   runtime.Str("doc"),
		"mode":     runtime.Str("mode"),
		"markdown": markdown,
	}
	if v := runtime.Str("selection-with-ellipsis"); v != "" {
		args["selection_with_ellipsis"] = v
	}
	if v := runtime.Str("selection-by-title"); v != "" {
		args["selection_by_title"] = v
	}
	return args
}

func validateDocCreateTarget(runtime *common.RuntimeContext) error {
	count := 0
	for _, flag := range []string{"folder-token", "wiki-node", "wiki-space"} {
		if runtime.Str(flag) != "" {
			count++
		}
	}
	if count > 1 {
		return common.FlagErrorf("--folder-token, --wiki-node, and --wiki-space are mutually exclusive")
	}
	return nil
}

func buildCalloutMarkdown(runtime *common.RuntimeContext) string {
	alert := strings.ToUpper(runtime.Str("type"))
	lines := []string{"> [!" + alert + "]"}
	if title := strings.TrimSpace(runtime.Str("title")); title != "" {
		lines = append(lines, "> **"+title+"**")
	}
	for _, line := range strings.Split(runtime.Str("text"), "\n") {
		lines = append(lines, "> "+line)
	}
	return strings.Join(lines, "\n")
}

func resolveDocTextOutputPath(rawOutput, title, fallback, ext string) (string, error) {
	path := rawOutput
	if path == "" {
		base := sanitizeDocOutputBase(title)
		if base == "" {
			base = sanitizeDocOutputBase(fallback)
		}
		path = base + ext
	}
	safePath, err := validate.SafeOutputPath(path)
	if err != nil {
		return "", output.ErrValidation("unsafe output path: %s", err)
	}
	return safePath, nil
}

func sanitizeDocOutputBase(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-", "\n", "-", "\t", "-")
	s = replacer.Replace(s)
	s = strings.Trim(s, "-._")
	if s == "" {
		return "document"
	}
	return s
}

func waitDocExportTask(runtime *common.RuntimeContext, ticket, token string) (string, error) {
	for i := 0; i < 30; i++ {
		data, err := runtime.CallAPI("GET", fmt.Sprintf("/open-apis/drive/v1/export_tasks/%s", ticket), map[string]interface{}{"token": token}, nil)
		if err != nil {
			return "", err
		}
		result := common.GetMap(data, "result")
		jobStatus, _ := util.ToFloat64(result["job_status"])
		if int(jobStatus) == 0 {
			fileToken := common.GetString(result, "file_token")
			if fileToken == "" {
				return "", output.Errorf(output.ExitAPI, "api_error", "export task finished without file_token")
			}
			return fileToken, nil
		}
		if int(jobStatus) != 1 && int(jobStatus) != 2 {
			return "", output.Errorf(output.ExitAPI, "api_error", "export task failed: %s", common.GetString(result, "job_error_msg"))
		}
		time.Sleep(time.Second)
	}
	return "", output.Errorf(output.ExitAPI, "timeout", "export task timed out")
}

func downloadDocExportFile(runtime *common.RuntimeContext, fileToken, outputPath string) error {
	apiResp, err := runtime.DoAPI(&larkcore.ApiReq{
		HttpMethod: http.MethodGet,
		ApiPath:    fmt.Sprintf("/open-apis/drive/v1/export_tasks/%s/download", fileToken),
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return output.Errorf(output.ExitInternal, "internal_error", "cannot create output directory: %v", err)
	}
	if err := validate.AtomicWrite(outputPath, apiResp.RawBody, 0644); err != nil {
		return output.Errorf(output.ExitInternal, "internal_error", "cannot create file: %v", err)
	}
	return nil
}

func uploadImportSourceFile(ctx context.Context, runtime *common.RuntimeContext, filePath, fileName, folderToken string, fileSize int64) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	fd := larkcore.NewFormdata()
	fd.AddField("file_name", fileName)
	fd.AddField("parent_type", "explorer")
	fd.AddField("parent_node", folderToken)
	fd.AddField("size", fmt.Sprintf("%d", fileSize))
	fd.AddFile("file", f)
	apiResp, err := runtime.DoAPI(&larkcore.ApiReq{
		HttpMethod: http.MethodPost,
		ApiPath:    "/open-apis/drive/v1/files/upload_all",
		Body:       fd,
	}, larkcore.WithFileUpload())
	if err != nil {
		var exitErr *output.ExitError
		if errors.As(err, &exitErr) {
			return "", err
		}
		return "", output.ErrNetwork("upload failed: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(apiResp.RawBody, &result); err != nil {
		return "", output.Errorf(output.ExitAPI, "api_error", "upload failed: invalid response JSON: %v", err)
	}
	code, _ := util.ToFloat64(result["code"])
	if code != 0 {
		return "", output.Errorf(output.ExitAPI, "api_error", "upload failed: %s", common.GetString(result, "msg"))
	}
	data, _ := result["data"].(map[string]interface{})
	fileToken, _ := data["file_token"].(string)
	if fileToken == "" {
		return "", output.Errorf(output.ExitAPI, "api_error", "upload failed: file_token missing")
	}
	_ = ctx
	return fileToken, nil
}

func waitDocImportTask(runtime *common.RuntimeContext, ticket string) (string, string, error) {
	for i := 0; i < 30; i++ {
		data, err := runtime.CallAPI("GET", fmt.Sprintf("/open-apis/drive/v1/import_tasks/%s", ticket), nil, nil)
		if err != nil {
			return "", "", err
		}
		result := common.GetMap(data, "result")
		jobStatus, _ := util.ToFloat64(result["job_status"])
		if int(jobStatus) == 0 {
			docToken := common.GetString(result, "token")
			url := common.GetString(result, "url")
			if docToken == "" {
				return "", "", output.Errorf(output.ExitAPI, "api_error", "import task finished without token")
			}
			return docToken, url, nil
		}
		if int(jobStatus) != 1 && int(jobStatus) != 2 {
			return "", "", output.Errorf(output.ExitAPI, "api_error", "import task failed: %s", common.GetString(result, "job_error_msg"))
		}
		time.Sleep(time.Second)
	}
	return "", "", output.Errorf(output.ExitAPI, "timeout", "import task timed out")
}

func docsPageSize(raw string) int {
	if raw == "" {
		return 200
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 200
	}
	if n < 1 {
		return 1
	}
	if n > 500 {
		return 500
	}
	return n
}

func collectDocBlocks(runtime *common.RuntimeContext, documentID string, recursive bool, pageSize int) ([]map[string]interface{}, error) {
	queue := []string{documentID}
	visited := map[string]bool{}
	var blocks []map[string]interface{}
	for len(queue) > 0 {
		parentID := queue[0]
		queue = queue[1:]
		if visited[parentID] {
			continue
		}
		visited[parentID] = true
		pageToken := ""
		for {
			params := map[string]interface{}{"page_size": pageSize}
			if pageToken != "" {
				params["page_token"] = pageToken
			}
			data, err := runtime.CallAPI("GET", fmt.Sprintf("/open-apis/docx/v1/documents/%s/blocks/%s/children", documentID, parentID), params, nil)
			if err != nil {
				return nil, err
			}
			items, _ := data["items"].([]interface{})
			for _, item := range items {
				block, _ := item.(map[string]interface{})
				if block == nil {
					continue
				}
				blocks = append(blocks, block)
				if recursive {
					children, _ := block["children"].([]interface{})
					if len(children) > 0 {
						queue = append(queue, common.GetString(block, "block_id"))
					}
				}
			}
			hasMore, _ := data["has_more"].(bool)
			pageToken = common.GetString(data, "page_token")
			if !hasMore || pageToken == "" {
				break
			}
		}
		if !recursive {
			break
		}
	}
	return blocks, nil
}

// fixExportedMarkdown applies post-processing to Lark-exported Markdown to
// improve round-trip fidelity on re-import:
//
//  1. fixSetextAmbiguity: inserts a blank line before any "---" that immediately
//     follows a non-empty line, preventing it from being parsed as a Setext H2.
//
//  2. fixTopLevelSoftbreaks: inserts a blank line between adjacent non-empty
//     lines at the top level (outside tables, callouts, code blocks, etc.).
//     Lark exports each block element on its own line with only \n between them;
//     standard Markdown parsers collapse those into a single paragraph on
//     re-import, losing the original block structure entirely.
func fixExportedMarkdown(md string) string {
	md = fixBoldSpacing(md)
	md = fixSetextAmbiguity(md)
	md = fixTopLevelSoftbreaks(md)
	// Collapse runs of 3+ consecutive newlines into exactly 2 (one blank line).
	for strings.Contains(md, "\n\n\n") {
		md = strings.ReplaceAll(md, "\n\n\n", "\n\n")
	}
	md = strings.TrimRight(md, "\n") + "\n"
	return md
}

// fixBoldSpacing fixes two issues with bold markers exported by Lark:
//
//  1. Trailing whitespace before closing **: "**text **" → "**text**"
//     CommonMark requires no space before a closing delimiter; otherwise the
//     ** is rendered as literal text.
//
//  2. Redundant bold in ATX headings: "# **text**" → "# text"
//     Headings are already bold, so the inner ** is visually redundant and
//     some renderers display the markers literally.
var (
	boldTrailingSpaceRe   = regexp.MustCompile(`(\*\*\S[^*]*?)\s+(\*\*)`)
	italicTrailingSpaceRe = regexp.MustCompile(`(\*\S[^*]*?)\s+(\*)`)
	headingBoldRe         = regexp.MustCompile(`(?m)^(#{1,6})\s+\*\*(.+?)\*\*\s*$`)
)

func fixBoldSpacing(md string) string {
	// Process line-by-line to avoid cross-line mismatches where ** from
	// different bold spans on different lines confuse the regex engine.
	lines := strings.Split(md, "\n")
	for i, line := range lines {
		lines[i] = boldTrailingSpaceRe.ReplaceAllString(line, "$1$2")
		lines[i] = italicTrailingSpaceRe.ReplaceAllString(lines[i], "$1$2")
	}
	md = strings.Join(lines, "\n")
	md = headingBoldRe.ReplaceAllString(md, "$1 $2")
	return md
}

var setextRe = regexp.MustCompile(`(?m)^([^\n]+)\n(-{3,}\s*$)`)

func fixSetextAmbiguity(md string) string {
	return setextRe.ReplaceAllString(md, "$1\n\n$2")
}

// opaqueBlocks are block elements whose interior must never be modified.
var opaqueBlocks = [][2]string{
	{"<callout", "</callout>"},
	{"<quote-container>", "</quote-container>"},
	{"```", "```"},
}

// isTableStructuralTag returns true for lark-table tags that are structural
// (table/tr/td open/close) and should not themselves trigger blank-line insertion.
func isTableStructuralTag(s string) bool {
	return strings.HasPrefix(s, "<lark-t") ||
		strings.HasPrefix(s, "</lark-t")
}

// fixTopLevelSoftbreaks ensures that adjacent non-empty content lines are
// separated by a blank line in two contexts:
//  1. Top level (depth == 0): every Lark block becomes its own Markdown paragraph.
//  2. Inside <lark-td>: multi-line cell content is preserved as separate paragraphs.
//
// Structural table tags (<lark-table>, <lark-tr>, <lark-td> and their closing
// counterparts) never trigger blank-line insertion themselves. Opaque blocks
// (callout, quote-container, code fences) are left untouched.
func fixTopLevelSoftbreaks(md string) string {
	lines := strings.Split(md, "\n")
	out := make([]string, 0, len(lines)*2)

	// opaqueDepth tracks nesting inside opaque blocks (callout, quote, code).
	opaqueDepth := 0
	inCodeBlock := false
	// inTableCell is true when we are between <lark-td> and </lark-td>.
	inTableCell := false
	// tableDepth tracks <lark-table> nesting (for the outer structure).
	tableDepth := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// --- Track fenced code blocks (``` toggles). ---
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				inCodeBlock = false
				opaqueDepth--
			} else {
				inCodeBlock = true
				opaqueDepth++
			}
			out = append(out, line)
			continue
		}

		if !inCodeBlock {
			// --- Track opaque blocks (other than ```). ---
			for _, bd := range opaqueBlocks {
				if bd[0] == "```" {
					continue
				}
				if strings.HasPrefix(trimmed, bd[0]) {
					opaqueDepth++
				}
				if strings.Contains(trimmed, bd[1]) {
					opaqueDepth--
					if opaqueDepth < 0 {
						opaqueDepth = 0
					}
				}
			}

			// --- Track table structure. ---
			if strings.HasPrefix(trimmed, "<lark-table") {
				tableDepth++
			}
			if strings.Contains(trimmed, "</lark-table>") {
				tableDepth--
				if tableDepth < 0 {
					tableDepth = 0
				}
			}
			if strings.HasPrefix(trimmed, "<lark-td>") {
				inTableCell = true
			}
			if strings.Contains(trimmed, "</lark-td>") {
				inTableCell = false
			}
		}

		// --- Decide whether to insert a blank line before this line. ---
		// Skip if inside an opaque block.
		if opaqueDepth == 0 && trimmed != "" && i > 0 {
			// Skip structural table tags — they are not content lines.
			isStructural := isTableStructuralTag(trimmed)

			// Don't split consecutive blockquote lines ("> ...") — they form
			// one continuous blockquote in the original document.
			isBlockquote := strings.HasPrefix(trimmed, "> ")

			// Insert blank line if: (a) top level, or (b) inside a table cell,
			// AND this line is a content line, AND the previous output is non-empty.
			if !isStructural && !isBlockquote && (tableDepth == 0 || inTableCell) {
				prev := ""
				if len(out) > 0 {
					prev = strings.TrimSpace(out[len(out)-1])
				}
				// Don't insert blank line after a structural tag either.
				if prev != "" && !isTableStructuralTag(prev) {
					out = append(out, "")
				}
			}
		}

		out = append(out, line)
	}

	return strings.Join(out, "\n")
}
