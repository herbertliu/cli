// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package contact

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/url"
	"strconv"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

var ContactDepartmentGet = common.Shortcut{
	Service:     "contact",
	Command:     "+department-get",
	Description: "Get department info by department ID",
	Risk:        "read",
	Scopes:      []string{"contact:department.base:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "department-id", Desc: "department ID (open_department_id by default)", Required: true},
		{Name: "department-id-type", Default: "open_department_id", Desc: "department ID type: open_department_id | department_id"},
		{Name: "user-id-type", Default: "open_id", Desc: "user ID type: open_id | union_id | user_id"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		departmentID := runtime.Str("department-id")
		return common.NewDryRunAPI().
			GET("/open-apis/contact/v3/departments/:department_id").
			Params(map[string]interface{}{
				"department_id_type": runtime.Str("department-id-type"),
				"user_id_type":       runtime.Str("user-id-type"),
			}).
			Set("department_id", departmentID)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		departmentID := runtime.Str("department-id")
		data, err := runtime.CallAPI("GET",
			"/open-apis/contact/v3/departments/"+validate.EncodePathSegment(departmentID),
			map[string]interface{}{
				"department_id_type": runtime.Str("department-id-type"),
				"user_id_type":       runtime.Str("user-id-type"),
			},
			nil,
		)
		if err != nil {
			return err
		}

		department, _ := data["department"].(map[string]interface{})
		if department == nil {
			department = data
		}

		runtime.OutFormat(map[string]interface{}{"department": department}, nil, func(w io.Writer) {
			output.PrintTable(w, []map[string]interface{}{departmentRow(department)})
		})
		return nil
	},
}

var ContactDepartmentChildren = common.Shortcut{
	Service:     "contact",
	Command:     "+department-children",
	Description: "List child departments under a department",
	Risk:        "read",
	Scopes:      []string{"contact:department.base:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "department-id", Desc: "department ID (open_department_id by default)", Required: true},
		{Name: "department-id-type", Default: "open_department_id", Desc: "department ID type: open_department_id | department_id"},
		{Name: "user-id-type", Default: "open_id", Desc: "user ID type: open_id | union_id | user_id"},
		{Name: "page-size", Default: "20", Desc: "page size"},
		{Name: "page-token", Desc: "page token"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		departmentID := runtime.Str("department-id")
		return common.NewDryRunAPI().
			GET("/open-apis/contact/v3/departments/:department_id/children").
			Params(contactListParams(runtime)).
			Set("department_id", departmentID)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		departmentID := runtime.Str("department-id")
		data, err := runtime.CallAPI("GET",
			"/open-apis/contact/v3/departments/"+url.PathEscape(departmentID)+"/children",
			contactListParams(runtime),
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
				fmt.Fprintln(w, "No child departments found.")
				return
			}
			rows := make([]map[string]interface{}, 0, len(items))
			for _, item := range items {
				if dept, _ := item.(map[string]interface{}); dept != nil {
					rows = append(rows, departmentRow(dept))
				}
			}
			output.PrintTable(w, rows)
		})
		return nil
	},
}

var ContactDepartmentUsersList = common.Shortcut{
	Service:     "contact",
	Command:     "+department-users-list",
	Description: "List direct users in a department",
	Risk:        "read",
	UserScopes:  []string{"contact:department.base:readonly", "contact:user.basic_profile:readonly"},
	BotScopes:   []string{"contact:department.base:readonly", "contact:user.base:readonly"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "department-id", Desc: "department ID (open_department_id by default)", Required: true},
		{Name: "department-id-type", Default: "open_department_id", Desc: "department ID type: open_department_id | department_id"},
		{Name: "user-id-type", Default: "open_id", Desc: "user ID type: open_id | union_id | user_id"},
		{Name: "page-size", Default: "20", Desc: "page size"},
		{Name: "page-token", Desc: "page token"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		params := contactListParams(runtime)
		params["department_id"] = runtime.Str("department-id")
		return common.NewDryRunAPI().
			GET("/open-apis/contact/v3/users/find_by_department").
			Params(params)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		params := contactListParams(runtime)
		params["department_id"] = runtime.Str("department-id")
		data, err := runtime.CallAPI("GET", "/open-apis/contact/v3/users/find_by_department", params, nil)
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
				fmt.Fprintln(w, "No users found in the department.")
				return
			}
			rows := make([]map[string]interface{}, 0, len(items))
			for _, item := range items {
				if user, _ := item.(map[string]interface{}); user != nil {
					rows = append(rows, map[string]interface{}{
						"name":       pickUserName(user),
						"open_id":    firstNonEmpty(user, "open_id"),
						"user_id":    firstNonEmpty(user, "user_id"),
						"email":      firstNonEmpty(user, "email", "enterprise_email"),
						"mobile":     firstNonEmpty(user, "mobile"),
						"department": firstNonEmpty(user, "department_name"),
						"status":     contactUserStatus(user),
					})
				}
			}
			output.PrintTable(w, rows)
		})
		return nil
	},
}

func contactListParams(runtime *common.RuntimeContext) map[string]interface{} {
	params := map[string]interface{}{
		"department_id_type": runtime.Str("department-id-type"),
		"user_id_type":       runtime.Str("user-id-type"),
		"page_size":          contactPageSize(runtime.Str("page-size")),
	}
	if pageToken := runtime.Str("page-token"); pageToken != "" {
		params["page_token"] = pageToken
	}
	return params
}

func contactPageSize(raw string) int {
	if raw == "" {
		return 20
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 20
	}
	return int(math.Min(math.Max(float64(n), 1), 100))
}

func departmentRow(department map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":                 firstNonEmpty(department, "name"),
		"department_id":        firstNonEmpty(department, "department_id"),
		"open_department_id":   firstNonEmpty(department, "open_department_id"),
		"parent_department_id": firstNonEmpty(department, "parent_department_id"),
		"leader_user_id":       firstNonEmpty(department, "leader_user_id"),
		"member_count":         department["member_count"],
		"chat_id":              firstNonEmpty(department, "chat_id"),
	}
}

func contactUserStatus(user map[string]interface{}) string {
	status, _ := user["status"].(map[string]interface{})
	if status == nil {
		return ""
	}
	isFrozen, _ := status["is_frozen"].(bool)
	if isFrozen {
		return "frozen"
	}
	if isActivated, ok := status["is_activated"].(bool); ok && isActivated {
		return "active"
	}
	return ""
}
