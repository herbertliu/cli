// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package contact

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/shortcuts/common"
)

func TestContactShortcuts(t *testing.T) {
	var commands []string
	for _, shortcut := range Shortcuts() {
		commands = append(commands, shortcut.Command)
	}

	want := []string{
		"+search-user",
		"+get-user",
		"+department-get",
		"+department-children",
		"+department-users-list",
	}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("Shortcuts() commands = %#v, want %#v", commands, want)
	}
}

func TestContactDepartmentGetDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "contact"}
	cmd.Flags().String("department-id", "", "")
	cmd.Flags().String("department-id-type", "open_department_id", "")
	cmd.Flags().String("user-id-type", "open_id", "")
	_ = cmd.Flags().Set("department-id", "od-123")

	runtime := common.TestNewRuntimeContext(cmd, contactTestConfig())
	dry := ContactDepartmentGet.DryRun(context.Background(), runtime)
	formatted := dry.Format()
	if !strings.Contains(formatted, "GET /open-apis/contact/v3/departments/od-123") {
		t.Fatalf("dry-run missing path: %s", formatted)
	}
	if !strings.Contains(formatted, "department_id_type=open_department_id") {
		t.Fatalf("dry-run missing department_id_type: %s", formatted)
	}
}

func TestContactDepartmentChildrenDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "contact"}
	cmd.Flags().String("department-id", "", "")
	cmd.Flags().String("department-id-type", "open_department_id", "")
	cmd.Flags().String("user-id-type", "open_id", "")
	cmd.Flags().String("page-size", "20", "")
	cmd.Flags().String("page-token", "", "")
	_ = cmd.Flags().Set("department-id", "od-123")
	_ = cmd.Flags().Set("page-size", "35")
	_ = cmd.Flags().Set("page-token", "pt_1")

	runtime := common.TestNewRuntimeContext(cmd, contactTestConfig())
	dry := ContactDepartmentChildren.DryRun(context.Background(), runtime)
	formatted := dry.Format()
	if !strings.Contains(formatted, "GET /open-apis/contact/v3/departments/od-123/children") {
		t.Fatalf("dry-run missing path: %s", formatted)
	}
	if !strings.Contains(formatted, "page_size=35") || !strings.Contains(formatted, "page_token=pt_1") {
		t.Fatalf("dry-run missing pagination params: %s", formatted)
	}
}

func TestContactDepartmentUsersListDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "contact"}
	cmd.Flags().String("department-id", "", "")
	cmd.Flags().String("department-id-type", "open_department_id", "")
	cmd.Flags().String("user-id-type", "open_id", "")
	cmd.Flags().String("page-size", "20", "")
	cmd.Flags().String("page-token", "", "")
	_ = cmd.Flags().Set("department-id", "od-123")
	_ = cmd.Flags().Set("page-size", "50")

	runtime := common.TestNewRuntimeContext(cmd, contactTestConfig())
	dry := ContactDepartmentUsersList.DryRun(context.Background(), runtime)
	formatted := dry.Format()
	if !strings.Contains(formatted, "GET /open-apis/contact/v3/users/find_by_department") {
		t.Fatalf("dry-run missing path: %s", formatted)
	}
	if !strings.Contains(formatted, "department_id=od-123") || !strings.Contains(formatted, "page_size=50") {
		t.Fatalf("dry-run missing params: %s", formatted)
	}
}

func TestContactUserStatus(t *testing.T) {
	if got := contactUserStatus(map[string]interface{}{"status": map[string]interface{}{"is_frozen": true}}); got != "frozen" {
		t.Fatalf("contactUserStatus() = %q, want frozen", got)
	}
	if got := contactUserStatus(map[string]interface{}{"status": map[string]interface{}{"is_activated": true}}); got != "active" {
		t.Fatalf("contactUserStatus() = %q, want active", got)
	}
}

func contactTestConfig() *core.CliConfig {
	return &core.CliConfig{
		Brand: core.BrandFeishu,
	}
}
