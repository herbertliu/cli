// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/shortcuts/common"
)

func TestMessagesForwardDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "im"}
	cmd.Flags().String("message-id", "", "")
	cmd.Flags().String("chat-id", "", "")
	cmd.Flags().String("user-id", "", "")
	cmd.Flags().String("uuid", "", "")
	_ = cmd.Flags().Set("message-id", "om_123")
	_ = cmd.Flags().Set("chat-id", "oc_123")

	runtime := common.TestNewRuntimeContext(cmd, imTestConfig())
	got := ImMessagesForward.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "POST /open-apis/im/v1/messages/om_123/forward") || !strings.Contains(got, "receive_id_type=chat_id") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func TestMessagesMergeForwardDryRun(t *testing.T) {
	cmd := &cobra.Command{Use: "im"}
	cmd.Flags().String("message-ids", "", "")
	cmd.Flags().String("chat-id", "", "")
	cmd.Flags().String("user-id", "", "")
	cmd.Flags().String("uuid", "", "")
	_ = cmd.Flags().Set("message-ids", "om_1,om_2")
	_ = cmd.Flags().Set("chat-id", "oc_123")

	runtime := common.TestNewRuntimeContext(cmd, imTestConfig())
	got := ImMessagesMergeForward.DryRun(context.Background(), runtime).Format()
	if !strings.Contains(got, "POST /open-apis/im/v1/messages/merge_forward") || !strings.Contains(got, "\"message_id_list\":[\"om_1\",\"om_2\"]") {
		t.Fatalf("unexpected dry-run: %s", got)
	}
}

func TestMessagesForwardValidate(t *testing.T) {
	cmd := &cobra.Command{Use: "im"}
	cmd.Flags().String("message-id", "", "")
	cmd.Flags().String("chat-id", "", "")
	cmd.Flags().String("user-id", "", "")
	_ = cmd.Flags().Set("message-id", "om_123")

	runtime := common.TestNewRuntimeContext(cmd, imTestConfig())
	err := ImMessagesForward.Validate(context.Background(), runtime)
	if err == nil || !strings.Contains(err.Error(), "--chat-id or --user-id") {
		t.Fatalf("expected target validation error, got %v", err)
	}
}
