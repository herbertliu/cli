// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/shortcuts/common"
)

func TestDocShortcuts(t *testing.T) {
	var commands []string
	for _, shortcut := range Shortcuts() {
		commands = append(commands, shortcut.Command)
	}
	want := []string{
		"+search",
		"+create",
		"+fetch",
		"+update",
		"+import",
		"+export",
		"+export-file",
		"+import-file",
		"+callout",
		"+blocks",
		"+media-insert",
		"+media-download",
	}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("Shortcuts() commands = %#v, want %#v", commands, want)
	}
}

func TestDocsImportDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "docs"}
	cmd.Flags().String("file", "", "")
	cmd.Flags().String("title", "", "")
	_ = cmd.Flags().Set("file", "report.md")
	_ = cmd.Flags().Set("title", "Report")

	runtime := common.TestNewRuntimeContext(cmd, docTestConfig())
	got := DocsImport.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "create-doc") || !strings.Contains(got, "\\u003ccontents from report.md\\u003e") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func TestDocsExportFileDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "docs"}
	cmd.Flags().String("doc", "", "")
	cmd.Flags().String("file-extension", "pdf", "")
	_ = cmd.Flags().Set("doc", "doccn_123")

	runtime := common.TestNewRuntimeContext(cmd, docTestConfig())
	got := DocsExportFile.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "POST /open-apis/drive/v1/export_tasks") || !strings.Contains(got, "\"file_extension\":\"pdf\"") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func TestDocsImportFileValidate(t *testing.T) {
	cmd := &cobra.Command{Use: "docs"}
	cmd.Flags().String("file", "", "")
	_ = cmd.Flags().Set("file", "report.pdf")

	runtime := common.TestNewRuntimeContext(cmd, docTestConfig())
	err := DocsImportFile.Validate(context.Background(), runtime)
	if err == nil || !strings.Contains(err.Error(), ".docx") {
		t.Fatalf("expected docx validation error, got %v", err)
	}
}

func TestDocsCalloutDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "docs"}
	cmd.Flags().String("doc", "", "")
	cmd.Flags().String("type", "note", "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("text", "", "")
	cmd.Flags().String("mode", "append", "")
	cmd.Flags().String("selection-with-ellipsis", "", "")
	cmd.Flags().String("selection-by-title", "", "")
	_ = cmd.Flags().Set("doc", "doccn_123")
	_ = cmd.Flags().Set("type", "warning")
	_ = cmd.Flags().Set("title", "Heads up")
	_ = cmd.Flags().Set("text", "Watch this")

	runtime := common.TestNewRuntimeContext(cmd, docTestConfig())
	got := DocsCallout.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "update-doc") || !strings.Contains(got, "[!WARNING]") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func TestDocsBlocksDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "docs"}
	cmd.Flags().String("doc", "", "")
	cmd.Flags().String("page-size", "200", "")
	_ = cmd.Flags().Set("doc", "doccn_123")
	_ = cmd.Flags().Set("page-size", "120")

	runtime := common.TestNewRuntimeContext(cmd, docTestConfig())
	got := DocsBlocks.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "GET /open-apis/docx/v1/documents/doccn_123/blocks/doccn_123/children") || !strings.Contains(got, "page_size=120") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func docTestConfig() *core.CliConfig {
	return &core.CliConfig{Brand: core.BrandFeishu}
}
