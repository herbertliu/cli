// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

var ImChatDelete = common.Shortcut{
	Service:     "im",
	Command:     "+chat-delete",
	Description: "Delete a group chat; user/bot; dissolves a chat when caller has permission",
	Risk:        "high-risk-write",
	Scopes:      []string{"im:chat:update"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "chat-id", Desc: "chat ID (oc_xxx)", Required: true},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		_, err := common.ValidateChatID(runtime.Str("chat-id"))
		return err
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			DELETE(fmt.Sprintf("/open-apis/im/v1/chats/%s", validate.EncodePathSegment(runtime.Str("chat-id"))))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		chatID := runtime.Str("chat-id")
		if _, err := runtime.DoAPIJSON(http.MethodDelete,
			fmt.Sprintf("/open-apis/im/v1/chats/%s", validate.EncodePathSegment(chatID)),
			nil,
			nil,
		); err != nil {
			return err
		}

		runtime.OutFormat(map[string]interface{}{"chat_id": chatID, "deleted": true}, nil, func(w io.Writer) {
			fmt.Fprintf(w, "Group deleted successfully (chat_id: %s)\n", chatID)
		})
		return nil
	},
}
