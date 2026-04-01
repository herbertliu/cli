// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"fmt"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

var ImMessagesForward = common.Shortcut{
	Service:     "im",
	Command:     "+messages-forward",
	Description: "Forward a single message to a chat or user with bot identity",
	Risk:        "write",
	Scopes:      []string{"im:message:send_as_bot"},
	AuthTypes:   []string{"bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "message-id", Desc: "message ID (om_xxx)", Required: true},
		{Name: "chat-id", Desc: "target chat ID (oc_xxx)"},
		{Name: "user-id", Desc: "target user open_id (ou_xxx)"},
		{Name: "uuid", Desc: "idempotency key"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		if _, err := validateMessageID(runtime.Str("message-id")); err != nil {
			return err
		}
		return validateMessageForwardTarget(runtime)
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		receiveIDType, receiveID := messageForwardTarget(runtime)
		dry := common.NewDryRunAPI().
			POST("/open-apis/im/v1/messages/:message_id/forward").
			Params(map[string]interface{}{"receive_id_type": receiveIDType}).
			Body(map[string]interface{}{"receive_id": receiveID}).
			Set("message_id", runtime.Str("message-id"))
		if uuid := runtime.Str("uuid"); uuid != "" {
			dry.Params(map[string]interface{}{"receive_id_type": receiveIDType, "uuid": uuid})
		}
		return dry
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		receiveIDType, receiveID := messageForwardTarget(runtime)
		params := map[string]interface{}{"receive_id_type": receiveIDType}
		if uuid := runtime.Str("uuid"); uuid != "" {
			params["uuid"] = uuid
		}
		data, err := runtime.CallAPI("POST",
			fmt.Sprintf("/open-apis/im/v1/messages/%s/forward", validate.EncodePathSegment(runtime.Str("message-id"))),
			params,
			map[string]interface{}{"receive_id": receiveID},
		)
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

var ImMessagesMergeForward = common.Shortcut{
	Service:     "im",
	Command:     "+messages-merge-forward",
	Description: "Merge-forward multiple messages to a chat or user with bot identity",
	Risk:        "write",
	Scopes:      []string{"im:message:send_as_bot"},
	AuthTypes:   []string{"bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "message-ids", Desc: "comma-separated message IDs (om_xxx)", Required: true},
		{Name: "chat-id", Desc: "target chat ID (oc_xxx)"},
		{Name: "user-id", Desc: "target user open_id (ou_xxx)"},
		{Name: "uuid", Desc: "idempotency key"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		for _, id := range common.SplitCSV(runtime.Str("message-ids")) {
			if _, err := validateMessageID(id); err != nil {
				return err
			}
		}
		if len(common.SplitCSV(runtime.Str("message-ids"))) == 0 {
			return output.ErrValidation("--message-ids must contain at least one message ID")
		}
		return validateMessageForwardTarget(runtime)
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		receiveIDType, receiveID := messageForwardTarget(runtime)
		params := map[string]interface{}{"receive_id_type": receiveIDType}
		if uuid := runtime.Str("uuid"); uuid != "" {
			params["uuid"] = uuid
		}
		return common.NewDryRunAPI().
			POST("/open-apis/im/v1/messages/merge_forward").
			Params(params).
			Body(map[string]interface{}{
				"receive_id":      receiveID,
				"message_id_list": common.SplitCSV(runtime.Str("message-ids")),
			})
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		receiveIDType, receiveID := messageForwardTarget(runtime)
		params := map[string]interface{}{"receive_id_type": receiveIDType}
		if uuid := runtime.Str("uuid"); uuid != "" {
			params["uuid"] = uuid
		}
		data, err := runtime.CallAPI("POST",
			"/open-apis/im/v1/messages/merge_forward",
			params,
			map[string]interface{}{
				"receive_id":      receiveID,
				"message_id_list": common.SplitCSV(runtime.Str("message-ids")),
			},
		)
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

func validateMessageForwardTarget(runtime *common.RuntimeContext) error {
	chatID := runtime.Str("chat-id")
	userID := runtime.Str("user-id")
	if (chatID == "" && userID == "") || (chatID != "" && userID != "") {
		return output.ErrValidation("specify exactly one of --chat-id or --user-id")
	}
	if chatID != "" {
		_, err := common.ValidateChatID(chatID)
		return err
	}
	_, err := common.ValidateUserID(userID)
	return err
}

func messageForwardTarget(runtime *common.RuntimeContext) (string, string) {
	if chatID := runtime.Str("chat-id"); chatID != "" {
		return "chat_id", chatID
	}
	return "open_id", runtime.Str("user-id")
}
