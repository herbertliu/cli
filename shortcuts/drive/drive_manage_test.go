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
		"+comment-resolve",
		"+comment-replies-list",
		"+comment-reply-delete",
		"+permission-public-get",
		"+permission-public-update",
		"+permission-batch-add",
		"+permission-password-create",
		"+permission-password-delete",
		"+permission-transfer-owner",
		"+version-list",
		"+version-get",
		"+version-create",
		"+version-delete",
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

func TestDriveCommentResolveDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "drive"}
	cmd.Flags().String("doc", "", "")
	cmd.Flags().String("comment-id", "", "")
	cmd.Flags().Bool("unresolve", false, "")
	_ = cmd.Flags().Set("doc", "https://example.larksuite.com/docx/doccn_123")
	_ = cmd.Flags().Set("comment-id", "cmt_123")

	runtime := common.TestNewRuntimeContext(cmd, driveManageTestConfig())
	got := DriveCommentResolve.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "PATCH /open-apis/drive/v1/files/doccn_123/comments/cmt_123") {
		t.Fatalf("dry-run missing comment patch path: %s", got)
	}
	if !strings.Contains(got, "\"resolved\":true") {
		t.Fatalf("dry-run missing resolved body: %s", got)
	}
}

func TestDriveCommentRepliesListDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "drive"}
	cmd.Flags().String("doc", "", "")
	cmd.Flags().String("comment-id", "", "")
	cmd.Flags().String("page-size", "20", "")
	cmd.Flags().String("page-token", "", "")
	_ = cmd.Flags().Set("doc", "doccn_123")
	_ = cmd.Flags().Set("comment-id", "cmt_123")
	_ = cmd.Flags().Set("page-size", "30")

	runtime := common.TestNewRuntimeContext(cmd, driveManageTestConfig())
	got := DriveCommentRepliesList.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "GET /open-apis/drive/v1/files/doccn_123/comments/cmt_123/replies") {
		t.Fatalf("dry-run missing replies path: %s", got)
	}
	if !strings.Contains(got, "page_size=30") {
		t.Fatalf("dry-run missing page size: %s", got)
	}
}

func TestDriveCommentReplyDeleteDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "drive"}
	cmd.Flags().String("doc", "", "")
	cmd.Flags().String("comment-id", "", "")
	cmd.Flags().String("reply-id", "", "")
	_ = cmd.Flags().Set("doc", "doccn_123")
	_ = cmd.Flags().Set("comment-id", "cmt_123")
	_ = cmd.Flags().Set("reply-id", "rpl_123")

	runtime := common.TestNewRuntimeContext(cmd, driveManageTestConfig())
	got := DriveCommentReplyDelete.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "DELETE /open-apis/drive/v1/files/doccn_123/comments/cmt_123/replies/rpl_123") {
		t.Fatalf("dry-run missing delete path: %s", got)
	}
}

func TestDrivePermissionPublicUpdateDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "drive"}
	cmd.Flags().String("token", "", "")
	cmd.Flags().String("file-type", "docx", "")
	cmd.Flags().String("json", "", "")
	_ = cmd.Flags().Set("token", "doccn_123")
	_ = cmd.Flags().Set("json", `{"external_access":true}`)

	runtime := common.TestNewRuntimeContext(cmd, driveManageTestConfig())
	got := DrivePermissionPublicUpdate.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "PATCH /open-apis/drive/v1/permissions/doccn_123/public") {
		t.Fatalf("dry-run missing permission public path: %s", got)
	}
	if !strings.Contains(got, "\"external_access\":true") {
		t.Fatalf("dry-run missing body: %s", got)
	}
}

func TestDrivePermissionBatchAddDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "drive"}
	cmd.Flags().String("token", "", "")
	cmd.Flags().String("file-type", "docx", "")
	cmd.Flags().String("members-json", "", "")
	cmd.Flags().Bool("notify", false, "")
	_ = cmd.Flags().Set("token", "doccn_123")
	_ = cmd.Flags().Set("members-json", `[{"member_type":"email","member_id":"user@example.com","perm":"view"}]`)
	_ = cmd.Flags().Set("notify", "true")

	runtime := common.TestNewRuntimeContext(cmd, driveManageTestConfig())
	got := DrivePermissionBatchAdd.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "POST /open-apis/drive/v1/permissions/doccn_123/members/batch_create") {
		t.Fatalf("dry-run missing batch_create path: %s", got)
	}
	if !strings.Contains(got, "\"member_type\":\"email\"") {
		t.Fatalf("dry-run missing members body: %s", got)
	}
}

func TestDrivePermissionTransferOwnerDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "drive"}
	cmd.Flags().String("token", "", "")
	cmd.Flags().String("file-type", "docx", "")
	cmd.Flags().String("member-type", "", "")
	cmd.Flags().String("member-id", "", "")
	cmd.Flags().Bool("notify", false, "")
	cmd.Flags().Bool("remove-old-owner", false, "")
	cmd.Flags().Bool("stay-put", false, "")
	cmd.Flags().String("old-owner-perm", "full_access", "")
	_ = cmd.Flags().Set("token", "doccn_123")
	_ = cmd.Flags().Set("member-type", "email")
	_ = cmd.Flags().Set("member-id", "user@example.com")
	_ = cmd.Flags().Set("notify", "true")

	runtime := common.TestNewRuntimeContext(cmd, driveManageTestConfig())
	got := DrivePermissionTransferOwner.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "POST /open-apis/drive/v1/permissions/doccn_123/members/transfer_owner") {
		t.Fatalf("dry-run missing transfer_owner path: %s", got)
	}
	if !strings.Contains(got, "\"member_id\":\"user@example.com\"") {
		t.Fatalf("dry-run missing owner body: %s", got)
	}
}

func TestDrivePermissionPasswordCreateDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "drive"}
	cmd.Flags().String("token", "", "")
	cmd.Flags().String("file-type", "docx", "")
	_ = cmd.Flags().Set("token", "doccn_123")
	runtime := common.TestNewRuntimeContext(cmd, driveManageTestConfig())
	got := DrivePermissionPasswordCreate.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "POST /open-apis/drive/v1/permissions/doccn_123/public/password") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func TestDriveVersionCreateDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "drive"}
	cmd.Flags().String("file-token", "", "")
	cmd.Flags().String("file-type", "docx", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("user-id-type", "open_id", "")
	_ = cmd.Flags().Set("file-token", "doccn_123")
	_ = cmd.Flags().Set("name", "v1")
	runtime := common.TestNewRuntimeContext(cmd, driveManageTestConfig())
	got := DriveVersionCreate.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "POST /open-apis/drive/v1/files/doccn_123/versions") || !strings.Contains(got, "\"name\":\"v1\"") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func driveManageTestConfig() *core.CliConfig {
	return &core.CliConfig{
		Brand: core.BrandFeishu,
	}
}
