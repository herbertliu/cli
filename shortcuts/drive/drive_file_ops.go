// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package drive

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

var DriveMkdir = common.Shortcut{
	Service:     "drive",
	Command:     "+mkdir",
	Description: "Create a folder in Drive",
	Risk:        "write",
	Scopes:      []string{"drive:drive.metadata:readonly", "drive:file:upload"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "name", Desc: "folder name", Required: true},
		{Name: "folder-token", Desc: "parent folder token (omit to create in root if supported)"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		body := map[string]interface{}{
			"name": runtime.Str("name"),
		}
		if folderToken := strings.TrimSpace(runtime.Str("folder-token")); folderToken != "" {
			body["folder_token"] = folderToken
		}
		return common.NewDryRunAPI().
			POST("/open-apis/drive/v1/files/create_folder").
			Body(body)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		body := map[string]interface{}{
			"name": runtime.Str("name"),
		}
		if folderToken := strings.TrimSpace(runtime.Str("folder-token")); folderToken != "" {
			body["folder_token"] = folderToken
		}
		data, err := runtime.CallAPI("POST", "/open-apis/drive/v1/files/create_folder", nil, body)
		if err != nil {
			return err
		}

		file, _ := data["file"].(map[string]interface{})
		if file == nil {
			file = data
		}
		runtime.OutFormat(map[string]interface{}{"file": file}, nil, func(w io.Writer) {
			output.PrintTable(w, []map[string]interface{}{{
				"name":         firstDriveString(file, "name"),
				"token":        firstDriveString(file, "token"),
				"url":          firstDriveString(file, "url"),
				"parent_token": firstDriveString(file, "parent_token"),
				"type":         firstDriveString(file, "type"),
			}})
		})
		return nil
	},
}

var DriveStats = common.Shortcut{
	Service:     "drive",
	Command:     "+stats",
	Description: "Get file statistics such as views and likes",
	Risk:        "read",
	Scopes:      []string{"drive:drive.metadata:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "file-token", Desc: "file token", Required: true},
		{Name: "file-type", Default: "docx", Desc: "file type: doc | docx | sheet | file | slides"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			GET("/open-apis/drive/v1/files/:file_token/statistics").
			Params(map[string]interface{}{"file_type": runtime.Str("file-type")}).
			Set("file_token", runtime.Str("file-token"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		fileToken := runtime.Str("file-token")
		data, err := runtime.CallAPI("GET",
			fmt.Sprintf("/open-apis/drive/v1/files/%s/statistics", validate.EncodePathSegment(fileToken)),
			map[string]interface{}{"file_type": runtime.Str("file-type")},
			nil,
		)
		if err != nil {
			return err
		}

		statistics, _ := data["statistics"].(map[string]interface{})
		if statistics == nil {
			statistics = data
		}

		runtime.OutFormat(map[string]interface{}{"statistics": statistics}, nil, func(w io.Writer) {
			output.PrintTable(w, []map[string]interface{}{{
				"file_token":     fileToken,
				"file_type":      runtime.Str("file-type"),
				"view_count":     firstDriveValue(statistics, "view_count", "pv", "views"),
				"viewer_count":   firstDriveValue(statistics, "viewer_count", "uv"),
				"like_count":     firstDriveValue(statistics, "like_count", "likes"),
				"comment_count":  firstDriveValue(statistics, "comment_count", "comments"),
				"download_count": firstDriveValue(statistics, "download_count", "downloads"),
			}})
		})
		return nil
	},
}

func firstDriveString(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := m[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func firstDriveValue(m map[string]interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			return value
		}
	}
	return nil
}
