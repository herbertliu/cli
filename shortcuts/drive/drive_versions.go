// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package drive

import (
	"context"
	"fmt"
	"io"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

var DrivePermissionPasswordCreate = common.Shortcut{
	Service:     "drive",
	Command:     "+permission-password-create",
	Description: "Create a public sharing password for a file",
	Risk:        "write",
	Scopes:      []string{},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags:       permissionTokenFlags(),
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			POST("/open-apis/drive/v1/permissions/:token/public/password").
			Params(map[string]interface{}{"type": runtime.Str("file-type")}).
			Set("token", runtime.Str("token"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		data, err := runtime.CallAPI("POST",
			fmt.Sprintf("/open-apis/drive/v1/permissions/%s/public/password", validate.EncodePathSegment(runtime.Str("token"))),
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

var DrivePermissionPasswordDelete = common.Shortcut{
	Service:     "drive",
	Command:     "+permission-password-delete",
	Description: "Delete a public sharing password for a file",
	Risk:        "high-risk-write",
	Scopes:      []string{},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags:       permissionTokenFlags(),
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			DELETE("/open-apis/drive/v1/permissions/:token/public/password").
			Params(map[string]interface{}{"type": runtime.Str("file-type")}).
			Set("token", runtime.Str("token"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		data, err := runtime.CallAPI("DELETE",
			fmt.Sprintf("/open-apis/drive/v1/permissions/%s/public/password", validate.EncodePathSegment(runtime.Str("token"))),
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

var DriveVersionList = common.Shortcut{
	Service:     "drive",
	Command:     "+version-list",
	Description: "List file versions",
	Risk:        "read",
	Scopes:      []string{},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "file-token", Desc: "file token", Required: true},
		{Name: "file-type", Default: "docx", Desc: "doc | docx | sheet | bitable"},
		{Name: "page-size", Default: "50", Desc: "page size"},
		{Name: "page-token", Desc: "page token"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		params := map[string]interface{}{
			"type":      runtime.Str("file-type"),
			"page_size": driveCommentPageSize(runtime.Str("page-size")),
		}
		if v := runtime.Str("page-token"); v != "" {
			params["page_token"] = v
		}
		return common.NewDryRunAPI().
			GET("/open-apis/drive/v1/files/:file_token/versions").
			Params(params).
			Set("file_token", runtime.Str("file-token"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		params := map[string]interface{}{
			"type":      runtime.Str("file-type"),
			"page_size": driveCommentPageSize(runtime.Str("page-size")),
		}
		if v := runtime.Str("page-token"); v != "" {
			params["page_token"] = v
		}
		data, err := runtime.CallAPI("GET",
			fmt.Sprintf("/open-apis/drive/v1/files/%s/versions", validate.EncodePathSegment(runtime.Str("file-token"))),
			params,
			nil,
		)
		if err != nil {
			return err
		}
		items, _ := data["items"].([]interface{})
		runtime.OutFormat(map[string]interface{}{
			"items":      items,
			"has_more":   data["has_more"],
			"page_token": data["page_token"],
		}, nil, func(w io.Writer) {
			if len(items) == 0 {
				fmt.Fprintln(w, "No versions found.")
				return
			}
			rows := make([]map[string]interface{}, 0, len(items))
			for _, item := range items {
				version, _ := item.(map[string]interface{})
				if version == nil {
					continue
				}
				rows = append(rows, map[string]interface{}{
					"name":        common.GetString(version, "name"),
					"version":     common.GetString(version, "version"),
					"status":      common.GetString(version, "status"),
					"create_time": common.GetString(version, "create_time"),
				})
			}
			output.PrintTable(w, rows)
		})
		return nil
	},
}

var DriveVersionGet = common.Shortcut{
	Service:     "drive",
	Command:     "+version-get",
	Description: "Get details for a file version",
	Risk:        "read",
	Scopes:      []string{},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "file-token", Desc: "file token", Required: true},
		{Name: "version-id", Desc: "version ID", Required: true},
		{Name: "file-type", Default: "docx", Desc: "doc | docx | sheet | bitable"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			GET("/open-apis/drive/v1/files/:file_token/versions/:version_id").
			Params(map[string]interface{}{"type": runtime.Str("file-type")}).
			Set("file_token", runtime.Str("file-token")).
			Set("version_id", runtime.Str("version-id"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		data, err := runtime.CallAPI("GET",
			fmt.Sprintf("/open-apis/drive/v1/files/%s/versions/%s", validate.EncodePathSegment(runtime.Str("file-token")), validate.EncodePathSegment(runtime.Str("version-id"))),
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

var DriveVersionCreate = common.Shortcut{
	Service:     "drive",
	Command:     "+version-create",
	Description: "Create a file version",
	Risk:        "write",
	Scopes:      []string{},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "file-token", Desc: "file token", Required: true},
		{Name: "file-type", Default: "docx", Desc: "doc | docx | sheet | bitable"},
		{Name: "name", Desc: "version name", Required: true},
		{Name: "user-id-type", Default: "open_id", Desc: "open_id | union_id | user_id"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			POST("/open-apis/drive/v1/files/:file_token/versions").
			Params(map[string]interface{}{
				"type":         runtime.Str("file-type"),
				"user_id_type": runtime.Str("user-id-type"),
			}).
			Body(map[string]interface{}{"name": runtime.Str("name")}).
			Set("file_token", runtime.Str("file-token"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		data, err := runtime.CallAPI("POST",
			fmt.Sprintf("/open-apis/drive/v1/files/%s/versions", validate.EncodePathSegment(runtime.Str("file-token"))),
			map[string]interface{}{
				"type":         runtime.Str("file-type"),
				"user_id_type": runtime.Str("user-id-type"),
			},
			map[string]interface{}{"name": runtime.Str("name")},
		)
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

var DriveVersionDelete = common.Shortcut{
	Service:     "drive",
	Command:     "+version-delete",
	Description: "Delete a file version",
	Risk:        "high-risk-write",
	Scopes:      []string{},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "file-token", Desc: "file token", Required: true},
		{Name: "version-id", Desc: "version ID", Required: true},
		{Name: "file-type", Default: "docx", Desc: "doc | docx | sheet | bitable"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			DELETE("/open-apis/drive/v1/files/:file_token/versions/:version_id").
			Params(map[string]interface{}{"type": runtime.Str("file-type")}).
			Set("file_token", runtime.Str("file-token")).
			Set("version_id", runtime.Str("version-id"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		data, err := runtime.CallAPI("DELETE",
			fmt.Sprintf("/open-apis/drive/v1/files/%s/versions/%s", validate.EncodePathSegment(runtime.Str("file-token")), validate.EncodePathSegment(runtime.Str("version-id"))),
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
