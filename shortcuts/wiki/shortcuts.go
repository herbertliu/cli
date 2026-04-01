// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package wiki

import "github.com/larksuite/cli/shortcuts/common"

func Shortcuts() []common.Shortcut {
	return []common.Shortcut{
		WikiExport,
		WikiMemberList,
		WikiMemberAdd,
		WikiMemberRemove,
	}
}
