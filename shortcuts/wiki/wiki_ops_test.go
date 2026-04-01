// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package wiki

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/shortcuts/common"
)

func TestWikiShortcuts(t *testing.T) {
	var commands []string
	for _, shortcut := range Shortcuts() {
		commands = append(commands, shortcut.Command)
	}
	want := []string{"+export", "+member-list", "+member-add", "+member-remove"}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("Shortcuts() commands = %#v, want %#v", commands, want)
	}
}

func TestWikiExportDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "wiki"}
	cmd.Flags().String("wiki", "", "")
	_ = cmd.Flags().Set("wiki", "wiki_123")
	runtime := common.TestNewRuntimeContext(cmd, wikiTestConfig())
	got := WikiExport.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "fetch-doc") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func TestWikiMemberAddDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "wiki"}
	cmd.Flags().String("space-id", "", "")
	cmd.Flags().String("member-type", "", "")
	cmd.Flags().String("member-id", "", "")
	cmd.Flags().String("member-role", "member", "")
	cmd.Flags().Bool("notify", true, "")
	_ = cmd.Flags().Set("space-id", "123")
	_ = cmd.Flags().Set("member-type", "email")
	_ = cmd.Flags().Set("member-id", "user@example.com")
	runtime := common.TestNewRuntimeContext(cmd, wikiTestConfig())
	got := WikiMemberAdd.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "POST /open-apis/wiki/v2/spaces/123/members") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func TestWikiMemberRemoveDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "wiki"}
	cmd.Flags().String("space-id", "", "")
	cmd.Flags().String("member-type", "", "")
	cmd.Flags().String("member-id", "", "")
	cmd.Flags().String("member-role", "member", "")
	_ = cmd.Flags().Set("space-id", "123")
	_ = cmd.Flags().Set("member-type", "email")
	_ = cmd.Flags().Set("member-id", "user@example.com")
	runtime := common.TestNewRuntimeContext(cmd, wikiTestConfig())
	got := WikiMemberRemove.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "DELETE /open-apis/wiki/v2/spaces/123/members/user@example.com") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func wikiTestConfig() *core.CliConfig {
	return &core.CliConfig{Brand: core.BrandFeishu}
}
