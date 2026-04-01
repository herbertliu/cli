// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package drive

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/shortcuts/common"
)

func TestDriveShortcuts(t *testing.T) {
	var commands []string
	for _, shortcut := range Shortcuts() {
		commands = append(commands, shortcut.Command)
	}

	want := []string{
		"+upload",
		"+download",
		"+add-comment",
		"+mkdir",
		"+stats",
	}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("Shortcuts() commands = %#v, want %#v", commands, want)
	}
}

func TestDriveMkdirDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "drive"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("folder-token", "", "")
	_ = cmd.Flags().Set("name", "New Folder")
	_ = cmd.Flags().Set("folder-token", "fld_123")

	runtime := common.TestNewRuntimeContext(cmd, driveManageTestConfig())
	got := DriveMkdir.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "POST /open-apis/drive/v1/files/create_folder") {
		t.Fatalf("dry-run missing create_folder path: %s", got)
	}
	if !strings.Contains(got, "\"folder_token\":\"fld_123\"") {
		t.Fatalf("dry-run missing folder_token: %s", got)
	}
}

func TestDriveStatsDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "drive"}
	cmd.Flags().String("file-token", "", "")
	cmd.Flags().String("file-type", "docx", "")
	_ = cmd.Flags().Set("file-token", "doccn_123")

	runtime := common.TestNewRuntimeContext(cmd, driveManageTestConfig())
	got := DriveStats.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "GET /open-apis/drive/v1/files/doccn_123/statistics") {
		t.Fatalf("dry-run missing statistics path: %s", got)
	}
	if !strings.Contains(got, "file_type=docx") {
		t.Fatalf("dry-run missing file_type: %s", got)
	}
}

func driveManageTestConfig() *core.CliConfig {
	return &core.CliConfig{
		Brand: core.BrandFeishu,
	}
}
