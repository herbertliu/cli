// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

var SheetReadRich = common.Shortcut{
	Service:     "sheets",
	Command:     "+read-rich",
	Description: "Read rich-text cell values with the V3 Sheets API",
	Risk:        "read",
	Scopes:      []string{"sheets:spreadsheet:read"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "url", Desc: "spreadsheet URL"},
		{Name: "spreadsheet-token", Desc: "spreadsheet token"},
		{Name: "sheet-id", Desc: "sheet ID", Required: true},
		{Name: "ranges", Desc: "JSON array of ranges", Required: true},
		{Name: "datetime-render-option", Desc: "formatted_string | serial_number"},
		{Name: "value-render-option", Desc: "formatted_value | unformatted_value"},
		{Name: "user-id-type", Desc: "open_id | union_id | user_id"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		_, err := parseSheetRanges(runtime)
		return err
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		token := resolveSheetToken(runtime)
		ranges, _ := parseSheetRanges(runtime)
		return common.NewDryRunAPI().
			POST("/open-apis/sheets/v3/spreadsheets/:token/sheets/:sheet_id/values/batch_get").
			Params(sheetRichParams(runtime)).
			Body(map[string]interface{}{"ranges": ranges}).
			Set("token", token).
			Set("sheet_id", runtime.Str("sheet-id"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		token := resolveSheetToken(runtime)
		ranges, err := parseSheetRanges(runtime)
		if err != nil {
			return err
		}
		data, err := runtime.CallAPI("POST",
			fmt.Sprintf("/open-apis/sheets/v3/spreadsheets/%s/sheets/%s/values/batch_get", validate.EncodePathSegment(token), validate.EncodePathSegment(runtime.Str("sheet-id"))),
			sheetRichParams(runtime),
			map[string]interface{}{"ranges": ranges},
		)
		if err != nil {
			return err
		}
		valueRanges, _ := data["value_ranges"].([]interface{})
		runtime.OutFormat(data, nil, func(w io.Writer) {
			if len(valueRanges) == 0 {
				fmt.Fprintln(w, "No rich-text ranges found.")
				return
			}
			rows := make([]map[string]interface{}, 0, len(valueRanges))
			for _, item := range valueRanges {
				row, _ := item.(map[string]interface{})
				if row == nil {
					continue
				}
				values, _ := row["values"].([]interface{})
				rows = append(rows, map[string]interface{}{
					"range": row["range"],
					"rows":  len(values),
				})
			}
			output.PrintTable(w, rows)
		})
		return nil
	},
}

var SheetWriteRich = common.Shortcut{
	Service:     "sheets",
	Command:     "+write-rich",
	Description: "Write rich-text cell values with the V3 Sheets API",
	Risk:        "write",
	Scopes:      []string{"sheets:spreadsheet:write_only", "sheets:spreadsheet:read"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "url", Desc: "spreadsheet URL"},
		{Name: "spreadsheet-token", Desc: "spreadsheet token"},
		{Name: "sheet-id", Desc: "sheet ID", Required: true},
		{Name: "value-ranges", Desc: "JSON array of V3 value ranges", Required: true},
		{Name: "user-id-type", Desc: "open_id | union_id | user_id"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		_, err := parseJSONArrayFlag("--value-ranges", runtime.Str("value-ranges"))
		return err
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		token := resolveSheetToken(runtime)
		valueRanges, _ := parseJSONArrayFlag("--value-ranges", runtime.Str("value-ranges"))
		return common.NewDryRunAPI().
			POST("/open-apis/sheets/v3/spreadsheets/:token/sheets/:sheet_id/values/batch_update").
			Params(sheetUserIDParam(runtime)).
			Body(map[string]interface{}{"value_ranges": valueRanges}).
			Set("token", token).
			Set("sheet_id", runtime.Str("sheet-id"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		token := resolveSheetToken(runtime)
		valueRanges, err := parseJSONArrayFlag("--value-ranges", runtime.Str("value-ranges"))
		if err != nil {
			return err
		}
		data, err := runtime.CallAPI("POST",
			fmt.Sprintf("/open-apis/sheets/v3/spreadsheets/%s/sheets/%s/values/batch_update", validate.EncodePathSegment(token), validate.EncodePathSegment(runtime.Str("sheet-id"))),
			sheetUserIDParam(runtime),
			map[string]interface{}{"value_ranges": valueRanges},
		)
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

var SheetMerge = common.Shortcut{
	Service:     "sheets",
	Command:     "+merge",
	Description: "Merge cells in a spreadsheet range",
	Risk:        "write",
	Scopes:      []string{"sheets:spreadsheet:write_only", "sheets:spreadsheet:read"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "url", Desc: "spreadsheet URL"},
		{Name: "spreadsheet-token", Desc: "spreadsheet token"},
		{Name: "range", Desc: "range like <sheetId>!A1:C3", Required: true},
		{Name: "merge-type", Default: "MERGE_ALL", Desc: "MERGE_ALL | MERGE_ROWS | MERGE_COLUMNS"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		token := resolveSheetToken(runtime)
		return common.NewDryRunAPI().
			POST("/open-apis/sheets/v2/spreadsheets/:token/merge_cells").
			Body(map[string]interface{}{
				"range":     runtime.Str("range"),
				"mergeType": runtime.Str("merge-type"),
			}).
			Set("token", token)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		token := resolveSheetToken(runtime)
		data, err := runtime.CallAPI("POST",
			fmt.Sprintf("/open-apis/sheets/v2/spreadsheets/%s/merge_cells", validate.EncodePathSegment(token)),
			nil,
			map[string]interface{}{
				"range":     runtime.Str("range"),
				"mergeType": runtime.Str("merge-type"),
			},
		)
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

var SheetStyle = common.Shortcut{
	Service:     "sheets",
	Command:     "+style",
	Description: "Apply cell styles to a spreadsheet range",
	Risk:        "write",
	Scopes:      []string{"sheets:spreadsheet:write_only", "sheets:spreadsheet:read"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "url", Desc: "spreadsheet URL"},
		{Name: "spreadsheet-token", Desc: "spreadsheet token"},
		{Name: "range", Desc: "range like <sheetId>!A1:C3", Required: true},
		{Name: "style-json", Desc: "JSON object of style fields", Required: true},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		_, err := parseJSONObjectFlag("--style-json", runtime.Str("style-json"))
		return err
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		token := resolveSheetToken(runtime)
		style, _ := parseJSONObjectFlag("--style-json", runtime.Str("style-json"))
		return common.NewDryRunAPI().
			PUT("/open-apis/sheets/v2/spreadsheets/:token/style").
			Body(map[string]interface{}{
				"range":       runtime.Str("range"),
				"appendStyle": style,
			}).
			Set("token", token)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		token := resolveSheetToken(runtime)
		style, err := parseJSONObjectFlag("--style-json", runtime.Str("style-json"))
		if err != nil {
			return err
		}
		data, err := runtime.CallAPI("PUT",
			fmt.Sprintf("/open-apis/sheets/v2/spreadsheets/%s/style", validate.EncodePathSegment(token)),
			nil,
			map[string]interface{}{
				"range":       runtime.Str("range"),
				"appendStyle": style,
			},
		)
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

var SheetAddCols = common.Shortcut{
	Service:     "sheets",
	Command:     "+add-cols",
	Description: "Add columns to a sheet",
	Risk:        "write",
	Scopes:      []string{"sheets:spreadsheet:write_only", "sheets:spreadsheet:read"},
	AuthTypes:   []string{"user", "bot"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "url", Desc: "spreadsheet URL"},
		{Name: "spreadsheet-token", Desc: "spreadsheet token"},
		{Name: "sheet-id", Desc: "sheet ID", Required: true},
		{Name: "count", Default: "1", Desc: "number of columns to add"},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		token := resolveSheetToken(runtime)
		return common.NewDryRunAPI().
			POST("/open-apis/sheets/v2/spreadsheets/:token/dimension_range").
			Body(map[string]interface{}{
				"dimension": map[string]interface{}{
					"sheetId":        runtime.Str("sheet-id"),
					"majorDimension": "COLUMNS",
					"length":         sheetCount(runtime.Str("count")),
				},
			}).
			Set("token", token)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		token := resolveSheetToken(runtime)
		data, err := runtime.CallAPI("POST",
			fmt.Sprintf("/open-apis/sheets/v2/spreadsheets/%s/dimension_range", validate.EncodePathSegment(token)),
			nil,
			map[string]interface{}{
				"dimension": map[string]interface{}{
					"sheetId":        runtime.Str("sheet-id"),
					"majorDimension": "COLUMNS",
					"length":         sheetCount(runtime.Str("count")),
				},
			},
		)
		if err != nil {
			return err
		}
		runtime.Out(data, nil)
		return nil
	},
}

func resolveSheetToken(runtime *common.RuntimeContext) string {
	token := runtime.Str("spreadsheet-token")
	if runtime.Str("url") != "" {
		token = extractSpreadsheetToken(runtime.Str("url"))
	}
	return token
}

func parseSheetRanges(runtime *common.RuntimeContext) ([]interface{}, error) {
	if resolveSheetToken(runtime) == "" {
		return nil, common.FlagErrorf("specify --url or --spreadsheet-token")
	}
	return parseJSONArrayFlag("--ranges", runtime.Str("ranges"))
}

func sheetRichParams(runtime *common.RuntimeContext) map[string]interface{} {
	params := sheetUserIDParam(runtime)
	if v := runtime.Str("datetime-render-option"); v != "" {
		params["datetime_render_option"] = v
	}
	if v := runtime.Str("value-render-option"); v != "" {
		params["value_render_option"] = v
	}
	return params
}

func sheetUserIDParam(runtime *common.RuntimeContext) map[string]interface{} {
	params := map[string]interface{}{}
	if v := runtime.Str("user-id-type"); v != "" {
		params["user_id_type"] = v
	}
	return params
}

func parseJSONArrayFlag(flagName, raw string) ([]interface{}, error) {
	var items []interface{}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, output.ErrValidation("%s must be a valid JSON array: %v", flagName, err)
	}
	if len(items) == 0 {
		return nil, output.ErrValidation("%s must contain at least one item", flagName)
	}
	return items, nil
}

func parseJSONObjectFlag(flagName, raw string) (map[string]interface{}, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return nil, output.ErrValidation("%s must be a valid JSON object: %v", flagName, err)
	}
	if obj == nil {
		return nil, output.ErrValidation("%s must be a valid JSON object", flagName)
	}
	return obj, nil
}

func sheetCount(raw string) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n < 1 {
		return 1
	}
	if n > 1000 {
		return 1000
	}
	return n
}
