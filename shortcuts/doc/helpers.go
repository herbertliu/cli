// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/shortcuts/common"
)

type documentRef struct {
	Kind  string
	Token string
}

type resolvedDocumentTarget struct {
	InputKind string
	Kind      string
	Token     string
	WikiToken string
	Title     string
}

func parseDocumentRef(input string) (documentRef, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return documentRef{}, output.ErrValidation("--doc cannot be empty")
	}

	if token, ok := extractDocumentToken(raw, "/wiki/"); ok {
		return documentRef{Kind: "wiki", Token: token}, nil
	}
	if token, ok := extractDocumentToken(raw, "/docx/"); ok {
		return documentRef{Kind: "docx", Token: token}, nil
	}
	if token, ok := extractDocumentToken(raw, "/doc/"); ok {
		return documentRef{Kind: "doc", Token: token}, nil
	}
	if strings.Contains(raw, "://") {
		return documentRef{}, output.ErrValidation("unsupported --doc input %q: use a docx URL/token or a wiki URL that resolves to docx", raw)
	}
	if strings.ContainsAny(raw, "/?#") {
		return documentRef{}, output.ErrValidation("unsupported --doc input %q: use a docx token or a wiki URL", raw)
	}

	return documentRef{Kind: "docx", Token: raw}, nil
}

func extractDocumentToken(raw, marker string) (string, bool) {
	idx := strings.Index(raw, marker)
	if idx < 0 {
		return "", false
	}
	token := raw[idx+len(marker):]
	if end := strings.IndexAny(token, "/?#"); end >= 0 {
		token = token[:end]
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}
	return token, true
}

func buildDriveRouteExtra(docID string) (string, error) {
	extra, err := json.Marshal(map[string]string{"drive_route_token": docID})
	if err != nil {
		return "", output.Errorf(output.ExitInternal, "internal_error", "failed to marshal upload extra data: %v", err)
	}
	return string(extra), nil
}

func resolveDocumentTarget(runtime *common.RuntimeContext, input string) (resolvedDocumentTarget, error) {
	docRef, err := parseDocumentRef(input)
	if err != nil {
		return resolvedDocumentTarget{}, err
	}
	target := resolvedDocumentTarget{
		InputKind: docRef.Kind,
		Kind:      docRef.Kind,
		Token:     docRef.Token,
	}
	if docRef.Kind != "wiki" {
		return target, nil
	}
	data, err := runtime.CallAPI("GET", "/open-apis/wiki/v2/spaces/get_node", map[string]interface{}{"token": docRef.Token}, nil)
	if err != nil {
		return resolvedDocumentTarget{}, err
	}
	node := common.GetMap(data, "node")
	objType := common.GetString(node, "obj_type")
	objToken := common.GetString(node, "obj_token")
	if objType == "" || objToken == "" {
		return resolvedDocumentTarget{}, output.Errorf(output.ExitAPI, "api_error", "wiki get_node returned incomplete node data")
	}
	target.Kind = objType
	target.Token = objToken
	target.WikiToken = docRef.Token
	target.Title = common.GetString(node, "title")
	return target, nil
}

func dryRunResolvedDocumentTarget(ctx context.Context, runtime *common.RuntimeContext, input string) (resolvedDocumentTarget, *common.DryRunAPI) {
	docRef, err := parseDocumentRef(input)
	if err != nil {
		return resolvedDocumentTarget{}, common.NewDryRunAPI().Set("error", err.Error())
	}
	target := resolvedDocumentTarget{
		InputKind: docRef.Kind,
		Kind:      docRef.Kind,
		Token:     docRef.Token,
	}
	d := common.NewDryRunAPI()
	if docRef.Kind == "wiki" {
		target.Kind = "<resolved_obj_type>"
		target.Token = "<resolved_obj_token>"
		target.WikiToken = docRef.Token
		d.GET("/open-apis/wiki/v2/spaces/get_node").
			Desc("[1] Resolve wiki node to target document").
			Params(map[string]interface{}{"token": docRef.Token})
	}
	return target, d
}
