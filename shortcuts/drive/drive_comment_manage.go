// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package drive

import (
	"context"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

var DriveCommentResolve = common.Shortcut{
	Service:     "drive",
	Command:     "+comment-resolve",
	Description: "Resolve or unresolve a document comment",
	Risk:        "write",
	Scopes:      []string{"docs:document.comment:update", "docx:document:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "doc", Desc: "document URL/token, or wiki URL", Required: true},
		{Name: "comment-id", Desc: "comment ID", Required: true},
		{Name: "unresolve", Type: "bool", Desc: "restore a resolved comment"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		_, err := parseCommentDocRef(runtime.Str("doc"))
		return err
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		docRef, _ := parseCommentDocRef(runtime.Str("doc"))
		token, fileType, resolvedBy := dryRunResolvedCommentTarget(docRef, commentModeFull)
		dry := common.NewDryRunAPI()
		if resolvedBy == "wiki" {
			dry.GET("/open-apis/wiki/v2/spaces/get_node").
				Desc("[1] Resolve wiki node to target document").
				Params(map[string]interface{}{"token": docRef.Token})
		}
		step := "[1]"
		if resolvedBy == "wiki" {
			step = "[2]"
		}
		return dry.PATCH("/open-apis/drive/v1/files/:file_token/comments/:comment_id").
			Desc(step + " Resolve or unresolve comment").
			Params(map[string]interface{}{"file_type": fileType}).
			Body(map[string]interface{}{"resolved": !runtime.Bool("unresolve")}).
			Set("file_token", token).
			Set("comment_id", runtime.Str("comment-id"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		target, err := resolveCommentTarget(ctx, runtime, runtime.Str("doc"), commentModeFull)
		if err != nil {
			return err
		}
		commentID := runtime.Str("comment-id")
		resolved := !runtime.Bool("unresolve")
		data, err := runtime.CallAPI("PATCH",
			fmt.Sprintf("/open-apis/drive/v1/files/%s/comments/%s", validate.EncodePathSegment(target.FileToken), validate.EncodePathSegment(commentID)),
			map[string]interface{}{"file_type": target.FileType},
			map[string]interface{}{"resolved": resolved},
		)
		if err != nil {
			return err
		}
		out := map[string]interface{}{
			"doc_id":      target.DocID,
			"file_token":  target.FileToken,
			"file_type":   target.FileType,
			"comment_id":  commentID,
			"resolved":    resolved,
			"resolved_by": target.ResolvedBy,
		}
		if target.WikiToken != "" {
			out["wiki_token"] = target.WikiToken
		}
		for _, key := range []string{"comment_id", "resolved", "update_time"} {
			if v, ok := data[key]; ok {
				out[key] = v
			}
		}
		runtime.Out(out, nil)
		return nil
	},
}

var DriveCommentRepliesList = common.Shortcut{
	Service:     "drive",
	Command:     "+comment-replies-list",
	Description: "List replies under a document comment",
	Risk:        "read",
	Scopes:      []string{"docs:document.comment:read", "docx:document:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "doc", Desc: "document URL/token, or wiki URL", Required: true},
		{Name: "comment-id", Desc: "comment ID", Required: true},
		{Name: "page-size", Default: "20", Desc: "page size"},
		{Name: "page-token", Desc: "page token"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		_, err := parseCommentDocRef(runtime.Str("doc"))
		return err
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		docRef, _ := parseCommentDocRef(runtime.Str("doc"))
		token, fileType, resolvedBy := dryRunResolvedCommentTarget(docRef, commentModeFull)
		params := map[string]interface{}{
			"file_type": fileType,
			"page_size": driveCommentPageSize(runtime.Str("page-size")),
		}
		if pageToken := runtime.Str("page-token"); pageToken != "" {
			params["page_token"] = pageToken
		}
		dry := common.NewDryRunAPI()
		if resolvedBy == "wiki" {
			dry.GET("/open-apis/wiki/v2/spaces/get_node").
				Desc("[1] Resolve wiki node to target document").
				Params(map[string]interface{}{"token": docRef.Token})
		}
		step := "[1]"
		if resolvedBy == "wiki" {
			step = "[2]"
		}
		return dry.GET("/open-apis/drive/v1/files/:file_token/comments/:comment_id/replies").
			Desc(step + " List comment replies").
			Params(params).
			Set("file_token", token).
			Set("comment_id", runtime.Str("comment-id"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		target, err := resolveCommentTarget(ctx, runtime, runtime.Str("doc"), commentModeFull)
		if err != nil {
			return err
		}
		params := map[string]interface{}{
			"file_type": target.FileType,
			"page_size": driveCommentPageSize(runtime.Str("page-size")),
		}
		if pageToken := runtime.Str("page-token"); pageToken != "" {
			params["page_token"] = pageToken
		}
		data, err := runtime.CallAPI("GET",
			fmt.Sprintf("/open-apis/drive/v1/files/%s/comments/%s/replies", validate.EncodePathSegment(target.FileToken), validate.EncodePathSegment(runtime.Str("comment-id"))),
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
				fmt.Fprintln(w, "No replies found.")
				return
			}
			rows := make([]map[string]interface{}, 0, len(items))
			for _, item := range items {
				reply, _ := item.(map[string]interface{})
				if reply == nil {
					continue
				}
				rows = append(rows, map[string]interface{}{
					"reply_id":     firstDriveString(reply, "reply_id"),
					"user_id":      firstDriveString(reply, "user_id"),
					"content":      replyPreview(reply),
					"create_time":  common.FormatTimeWithSeconds(reply["create_time"]),
					"update_time":  common.FormatTimeWithSeconds(reply["update_time"]),
					"is_solved":    reply["resolved"],
				})
			}
			output.PrintTable(w, rows)
		})
		return nil
	},
}

var DriveCommentReplyDelete = common.Shortcut{
	Service:     "drive",
	Command:     "+comment-reply-delete",
	Description: "Delete a reply under a document comment",
	Risk:        "high-risk-write",
	Scopes:      []string{"docs:document.comment:delete", "docx:document:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "doc", Desc: "document URL/token, or wiki URL", Required: true},
		{Name: "comment-id", Desc: "comment ID", Required: true},
		{Name: "reply-id", Desc: "reply ID", Required: true},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		_, err := parseCommentDocRef(runtime.Str("doc"))
		return err
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		docRef, _ := parseCommentDocRef(runtime.Str("doc"))
		token, fileType, resolvedBy := dryRunResolvedCommentTarget(docRef, commentModeFull)
		dry := common.NewDryRunAPI()
		if resolvedBy == "wiki" {
			dry.GET("/open-apis/wiki/v2/spaces/get_node").
				Desc("[1] Resolve wiki node to target document").
				Params(map[string]interface{}{"token": docRef.Token})
		}
		step := "[1]"
		if resolvedBy == "wiki" {
			step = "[2]"
		}
		return dry.DELETE("/open-apis/drive/v1/files/:file_token/comments/:comment_id/replies/:reply_id").
			Desc(step + " Delete comment reply").
			Params(map[string]interface{}{"file_type": fileType}).
			Set("file_token", token).
			Set("comment_id", runtime.Str("comment-id")).
			Set("reply_id", runtime.Str("reply-id"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		target, err := resolveCommentTarget(ctx, runtime, runtime.Str("doc"), commentModeFull)
		if err != nil {
			return err
		}
		if _, err := runtime.CallAPI("DELETE",
			fmt.Sprintf("/open-apis/drive/v1/files/%s/comments/%s/replies/%s",
				validate.EncodePathSegment(target.FileToken),
				validate.EncodePathSegment(runtime.Str("comment-id")),
				validate.EncodePathSegment(runtime.Str("reply-id")),
			),
			map[string]interface{}{"file_type": target.FileType},
			nil,
		); err != nil {
			return err
		}
		runtime.Out(map[string]interface{}{
			"doc_id":     target.DocID,
			"file_token": target.FileToken,
			"file_type":  target.FileType,
			"comment_id": runtime.Str("comment-id"),
			"reply_id":   runtime.Str("reply-id"),
			"deleted":    true,
		}, nil)
		return nil
	},
}

func replyPreview(reply map[string]interface{}) string {
	content, _ := reply["content"].(map[string]interface{})
	if content == nil {
		return ""
	}
	elements, _ := content["elements"].([]interface{})
	parts := make([]string, 0, len(elements))
	for _, element := range elements {
		item, _ := element.(map[string]interface{})
		if item == nil {
			continue
		}
		switch item["type"] {
		case "text_run":
			if textRun, _ := item["text_run"].(map[string]interface{}); textRun != nil {
				if text, _ := textRun["text"].(string); text != "" {
					parts = append(parts, text)
				}
			}
		case "person":
			if person, _ := item["person"].(map[string]interface{}); person != nil {
				if name, _ := person["name"].(string); name != "" {
					parts = append(parts, "@"+name)
				}
			}
		case "docs_link":
			parts = append(parts, "[doc]")
		}
	}
	return strings.Join(parts, "")
}

func driveCommentPageSize(raw string) int {
	if raw == "" {
		return 20
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 20
	}
	return int(math.Min(math.Max(float64(n), 1), 100))
}
