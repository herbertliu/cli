// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/shortcuts/common"
)

func TestChatDeleteDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "im"}
	cmd.Flags().String("chat-id", "", "")
	_ = cmd.Flags().Set("chat-id", "oc_123")

	runtime := common.TestNewRuntimeContext(cmd, imTestConfig())
	dry := ImChatDelete.DryRun(context.Background(), runtime)
	if got := dry.Format(); !strings.Contains(got, "DELETE /open-apis/im/v1/chats/oc_123") {
		t.Fatalf("dry-run missing delete path: %s", got)
	}
}

func TestChatLinkDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "im"}
	cmd.Flags().String("chat-id", "", "")
	cmd.Flags().String("validity-period", "week", "")
	_ = cmd.Flags().Set("chat-id", "oc_123")
	_ = cmd.Flags().Set("validity-period", "year")

	runtime := common.TestNewRuntimeContext(cmd, imTestConfig())
	dry := ImChatLink.DryRun(context.Background(), runtime)
	got := dry.Format()
	if !strings.Contains(got, "POST /open-apis/im/v1/chats/oc_123/link") {
		t.Fatalf("dry-run missing link path: %s", got)
	}
	if !strings.Contains(got, "\"validity_period\":\"year\"") {
		t.Fatalf("dry-run missing validity period: %s", got)
	}
}

func TestChatLinkValidate(t *testing.T) {
	cmd := &cobra.Command{Use: "im"}
	cmd.Flags().String("chat-id", "", "")
	cmd.Flags().String("validity-period", "week", "")
	_ = cmd.Flags().Set("chat-id", "oc_123")
	_ = cmd.Flags().Set("validity-period", "bad")

	runtime := common.TestNewRuntimeContext(cmd, imTestConfig())
	err := ImChatLink.Validate(context.Background(), runtime)
	if err == nil || !strings.Contains(err.Error(), "validity-period") {
		t.Fatalf("expected validity-period validation error, got %v", err)
	}
}

func imTestConfig() *core.CliConfig {
	return &core.CliConfig{
		Brand: core.BrandFeishu,
	}
}
