// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package drive

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

var DrivePermissionPublicGet = common.Shortcut{
	Service:     "drive",
	Command:     "+permission-public-get",
	Description: "Get public permission settings for a file",
	Risk:        "read",
	Scopes:      []string{"docs:permission.member:auth"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags:       permissionTokenFlags(),
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			GET("/open-apis/drive/v1/permissions/:token/public").
			Params(map[string]interface{}{"type": runtime.Str("file-type")}).
			Set("token", runtime.Str("token"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		data, err := runtime.CallAPI("GET",
			fmt.Sprintf("/open-apis/drive/v1/permissions/%s/public", validate.EncodePathSegment(runtime.Str("token"))),
			map[string]interface{}{"type": runtime.Str("file-type")},
			nil,
		)
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

var DrivePermissionPublicUpdate = common.Shortcut{
	Service:     "drive",
	Command:     "+permission-public-update",
	Description: "Update public permission settings for a file",
	Risk:        "write",
	Scopes:      []string{"docs:permission.member:auth"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: append(permissionTokenFlags(), common.Flag{
		Name: "json", Desc: "public permission JSON object", Required: true,
	}),
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		_, err := parseJSONObject(runtime.Str("json"))
		return err
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		body, _ := parseJSONObject(runtime.Str("json"))
		return common.NewDryRunAPI().
			PATCH("/open-apis/drive/v1/permissions/:token/public").
			Params(map[string]interface{}{"type": runtime.Str("file-type")}).
			Body(body).
			Set("token", runtime.Str("token"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		body, err := parseJSONObject(runtime.Str("json"))
		if err != nil {
			return err
		}
		data, err := runtime.CallAPI("PATCH",
			fmt.Sprintf("/open-apis/drive/v1/permissions/%s/public", validate.EncodePathSegment(runtime.Str("token"))),
			map[string]interface{}{"type": runtime.Str("file-type")},
			body,
		)
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

var DrivePermissionBatchAdd = common.Shortcut{
	Service:     "drive",
	Command:     "+permission-batch-add",
	Description: "Batch add permission members to a file",
	Risk:        "write",
	Scopes:      []string{"docs:permission.member:create"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: append(permissionTokenFlags(),
		common.Flag{Name: "members-json", Desc: "JSON array of {member_type, member_id, perm}", Required: true},
		common.Flag{Name: "notify", Type: "bool", Desc: "notify added members"},
	),
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		_, err := parseJSONArray(runtime.Str("members-json"))
		return err
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		members, _ := parseJSONArray(runtime.Str("members-json"))
		return common.NewDryRunAPI().
			POST("/open-apis/drive/v1/permissions/:token/members/batch_create").
			Params(map[string]interface{}{
				"type":              runtime.Str("file-type"),
				"need_notification": runtime.Bool("notify"),
			}).
			Body(map[string]interface{}{"members": members}).
			Set("token", runtime.Str("token"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		members, err := parseJSONArray(runtime.Str("members-json"))
		if err != nil {
			return err
		}
		data, err := runtime.CallAPI("POST",
			fmt.Sprintf("/open-apis/drive/v1/permissions/%s/members/batch_create", validate.EncodePathSegment(runtime.Str("token"))),
			map[string]interface{}{
				"type":              runtime.Str("file-type"),
				"need_notification": runtime.Bool("notify"),
			},
			map[string]interface{}{"members": members},
		)
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

var DrivePermissionTransferOwner = common.Shortcut{
	Service:     "drive",
	Command:     "+permission-transfer-owner",
	Description: "Transfer file ownership to another member",
	Risk:        "high-risk-write",
	Scopes:      []string{"docs:permission.member:transfer"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: append(permissionTokenFlags(),
		common.Flag{Name: "member-type", Desc: "email | openid | userid", Required: true},
		common.Flag{Name: "member-id", Desc: "target owner member ID", Required: true},
		common.Flag{Name: "notify", Type: "bool", Default: "true", Desc: "notify new owner"},
		common.Flag{Name: "remove-old-owner", Type: "bool", Desc: "remove original owner permission after transfer"},
		common.Flag{Name: "stay-put", Type: "bool", Desc: "keep file at current location when supported"},
		common.Flag{Name: "old-owner-perm", Default: "full_access", Desc: "permission kept for original owner when not removed"},
	),
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			POST("/open-apis/drive/v1/permissions/:token/members/transfer_owner").
			Params(map[string]interface{}{
				"type":              runtime.Str("file-type"),
				"need_notification": runtime.Bool("notify"),
				"remove_old_owner":  runtime.Bool("remove-old-owner"),
				"stay_put":          runtime.Bool("stay-put"),
				"old_owner_perm":    runtime.Str("old-owner-perm"),
			}).
			Body(map[string]interface{}{
				"member_type": runtime.Str("member-type"),
				"member_id":   runtime.Str("member-id"),
			}).
			Set("token", runtime.Str("token"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		data, err := runtime.CallAPI("POST",
			fmt.Sprintf("/open-apis/drive/v1/permissions/%s/members/transfer_owner", validate.EncodePathSegment(runtime.Str("token"))),
			map[string]interface{}{
				"type":              runtime.Str("file-type"),
				"need_notification": runtime.Bool("notify"),
				"remove_old_owner":  runtime.Bool("remove-old-owner"),
				"stay_put":          runtime.Bool("stay-put"),
				"old_owner_perm":    runtime.Str("old-owner-perm"),
			},
			map[string]interface{}{
				"member_type": runtime.Str("member-type"),
				"member_id":   runtime.Str("member-id"),
			},
		)
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

func permissionTokenFlags() []common.Flag {
	return []common.Flag{
		{Name: "token", Desc: "file token", Required: true},
		{Name: "file-type", Default: "docx", Desc: "file type: doc | docx | sheet | file | wiki | bitable | slides | folder"},
	}
}

func parseJSONObject(raw string) (map[string]interface{}, error) {
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		return nil, output.ErrValidation("--json must be a valid JSON object: %v", err)
	}
	if body == nil {
		return nil, output.ErrValidation("--json must be a valid JSON object")
	}
	return body, nil
}

func parseJSONArray(raw string) ([]interface{}, error) {
	var items []interface{}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, output.ErrValidation("--members-json must be a valid JSON array: %v", err)
	}
	if len(items) == 0 {
		return nil, output.ErrValidation("--members-json must contain at least one member")
	}
	return items, nil
}
