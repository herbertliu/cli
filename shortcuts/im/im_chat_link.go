// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

var ImChatLink = common.Shortcut{
	Service:     "im",
	Command:     "+chat-link",
	Description: "Get a group chat share link; user/bot; supports week/year/permanently validity",
	Risk:        "write",
	Scopes:      []string{"im:chat:read"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "chat-id", Desc: "chat ID (oc_xxx)", Required: true},
		{Name: "validity-period", Default: "week", Desc: "week | year | permanently", Enum: []string{"week", "year", "permanently"}},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		if _, err := common.ValidateChatID(runtime.Str("chat-id")); err != nil {
			return err
		}
		switch runtime.Str("validity-period") {
		case "week", "year", "permanently":
			return nil
		default:
			return output.ErrValidation("--validity-period must be one of: week, year, permanently")
		}
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			POST(fmt.Sprintf("/open-apis/im/v1/chats/%s/link", validate.EncodePathSegment(runtime.Str("chat-id")))).
			Body(map[string]interface{}{
				"validity_period": runtime.Str("validity-period"),
			})
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		chatID := runtime.Str("chat-id")
		data, err := runtime.DoAPIJSON(http.MethodPost,
			fmt.Sprintf("/open-apis/im/v1/chats/%s/link", validate.EncodePathSegment(chatID)),
			nil,
			map[string]interface{}{
				"validity_period": runtime.Str("validity-period"),
			},
		)
		if err != nil {
			return err
		}

		runtime.OutFormat(data, nil, func(w io.Writer) {
			output.PrintTable(w, []map[string]interface{}{{
				"chat_id":         chatID,
				"share_link":      data["share_link"],
				"expire_time":     common.FormatTimeWithSeconds(data["expire_time"]),
				"is_permanent":    data["is_permanent"],
				"validity_period": runtime.Str("validity-period"),
			}})
		})
		return nil
	},
}
