// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package sheets

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/shortcuts/common"
)

func TestSheetShortcuts(t *testing.T) {
	var commands []string
	for _, shortcut := range Shortcuts() {
		commands = append(commands, shortcut.Command)
	}
	want := []string{
		"+info",
		"+read",
		"+write",
		"+read-rich",
		"+write-rich",
		"+merge",
		"+style",
		"+add-cols",
		"+append",
		"+find",
		"+create",
		"+export",
	}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("Shortcuts() commands = %#v, want %#v", commands, want)
	}
}

func TestSheetReadRichDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "sheets"}
	cmd.Flags().String("spreadsheet-token", "", "")
	cmd.Flags().String("sheet-id", "", "")
	cmd.Flags().String("ranges", "", "")
	cmd.Flags().String("datetime-render-option", "", "")
	cmd.Flags().String("value-render-option", "", "")
	cmd.Flags().String("user-id-type", "", "")
	_ = cmd.Flags().Set("spreadsheet-token", "sht_123")
	_ = cmd.Flags().Set("sheet-id", "0b12")
	_ = cmd.Flags().Set("ranges", `["0b12!A1:B2"]`)

	runtime := common.TestNewRuntimeContext(cmd, sheetsTestConfig())
	got := SheetReadRich.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "POST /open-apis/sheets/v3/spreadsheets/sht_123/sheets/0b12/values/batch_get") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func TestSheetWriteRichValidate(t *testing.T) {
	cmd := &cobra.Command{Use: "sheets"}
	cmd.Flags().String("value-ranges", "", "")
	_ = cmd.Flags().Set("value-ranges", `{}`)
	runtime := common.TestNewRuntimeContext(cmd, sheetsTestConfig())
	err := SheetWriteRich.Validate(context.Background(), runtime)
	if err == nil || !strings.Contains(err.Error(), "JSON array") {
		t.Fatalf("expected JSON array validation error, got %v", err)
	}
}

func TestSheetStyleDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "sheets"}
	cmd.Flags().String("spreadsheet-token", "", "")
	cmd.Flags().String("range", "", "")
	cmd.Flags().String("style-json", "", "")
	_ = cmd.Flags().Set("spreadsheet-token", "sht_123")
	_ = cmd.Flags().Set("range", "0b12!A1:B2")
	_ = cmd.Flags().Set("style-json", `{"hAlign":1}`)
	runtime := common.TestNewRuntimeContext(cmd, sheetsTestConfig())
	got := SheetStyle.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "PUT /open-apis/sheets/v2/spreadsheets/sht_123/style") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func TestSheetAddColsDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "sheets"}
	cmd.Flags().String("spreadsheet-token", "", "")
	cmd.Flags().String("sheet-id", "", "")
	cmd.Flags().String("count", "1", "")
	_ = cmd.Flags().Set("spreadsheet-token", "sht_123")
	_ = cmd.Flags().Set("sheet-id", "0b12")
	_ = cmd.Flags().Set("count", "3")
	runtime := common.TestNewRuntimeContext(cmd, sheetsTestConfig())
	got := SheetAddCols.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "\"majorDimension\":\"COLUMNS\"") || !strings.Contains(got, "\"length\":3") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func sheetsTestConfig() *core.CliConfig {
	return &core.CliConfig{Brand: core.BrandFeishu}
}
