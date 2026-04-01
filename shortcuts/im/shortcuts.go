// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package im

import "github.com/larksuite/cli/shortcuts/common"

// Shortcuts returns all im shortcuts.
func Shortcuts() []common.Shortcut {
	return []common.Shortcut{
		ImChatCreate,
		ImChatDelete,
		ImChatLink,
		ImChatMessageList,
		ImChatSearch,
		ImChatUpdate,
		ImMessagesForward,
		ImMessagesMergeForward,
		ImMessagesMGet,
		ImMessagesReply,
		ImMessagesResourcesDownload,
		ImMessagesSearch,
		ImMessagesSend,
		ImThreadsMessagesList,
	}
}
