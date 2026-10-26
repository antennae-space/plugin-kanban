// Package assistant is the part of the tool kit of the IDE's assistant the Kanban needs to answer
// its kanban tools (kanban.assistantTool): the result of a tool, its summary translated by the
// page, the arguments of a call, the limits and the schemas of the kanban tools. A ToolResult
// is encoded as the IDE's (Go field names), so the IDE reads it as its own.
package assistant

import (
	"encoding/json"
	"fmt"
)

// ToolResult is the result of a tool call: Content for the model, Summary for the user (a
// Text or a plain string, JSON), Status ok or error.
type ToolResult struct {
	Content string
	Failure string
	Summary json.RawMessage
	Status  string
}

// Text is a text of the interface translated by the page: t(Key, Params), or tn(N, Key,
// Other, Params) when Other is set.
type Text struct {
	Key    string         `json:"key"`
	Other  string         `json:"other,omitempty"`
	N      int            `json:"n,omitempty"`
	Params map[string]any `json:"params,omitempty"`
	Prefix string         `json:"prefix,omitempty"`
	Suffix string         `json:"suffix,omitempty"`
}

// With puts a plain text before and after the translated one.
func (t Text) With(prefix, suffix string) Text {
	t.Prefix, t.Suffix = prefix, suffix
	return t
}

// T is a text to translate; Tn its plural form (Key for one, Other for several).
func T(key string, params map[string]any) Text { return Text{Key: key, Params: params} }
func Tn(n int, key, other string, params map[string]any) Text {
	return Text{Key: key, Other: other, N: n, Params: params}
}

// Raw is the JSON of a summary.
func (t Text) Raw() json.RawMessage {
	data, _ := json.Marshal(t)
	return data
}

// Plain is a summary that needs no translation.
func Plain(s string) json.RawMessage {
	data, _ := json.Marshal(s)
	return data
}

// OK is a result whose summary is translated by the page.
func OK(content string, summary Text) ToolResult {
	return ToolResult{Content: content, Summary: summary.Raw(), Status: "ok"}
}

// OKPlain is a result with a summary already in the language of the user.
func OKPlain(content, summary string) ToolResult {
	return ToolResult{Content: content, Summary: Plain(summary), Status: "ok"}
}

// UsageError: the call itself is wrong (a parameter missing or invalid).
type UsageError struct{ Msg string }

func (e UsageError) Error() string { return e.Msg }

func Usagef(format string, a ...any) error { return UsageError{fmt.Sprintf(format, a...)} }

// Args of a tool call, read field by field.
type Args map[string]json.RawMessage

func (a Args) Str(k string) string {
	var s string
	if json.Unmarshal(a[k], &s) != nil {
		var n json.Number
		if json.Unmarshal(a[k], &n) == nil {
			return n.String()
		}
	}
	return s
}

func (a Args) Num(k string) int {
	var f float64
	if json.Unmarshal(a[k], &f) == nil {
		return int(f)
	}
	var s string
	if json.Unmarshal(a[k], &s) == nil {
		var n int
		fmt.Sscan(s, &n)
		return n
	}
	return 0
}

func (a Args) Bool(k string, def bool) bool {
	var b bool
	if json.Unmarshal(a[k], &b) == nil {
		return b
	}
	return def
}

func (a Args) Has(k string) bool { return len(a[k]) > 0 && string(a[k]) != "null" }

func (a Args) Strs(k string) []string {
	var out []string
	_ = json.Unmarshal(a[k], &out)
	return out
}

func (a Args) OptStr(k string) *string {
	if !a.Has(k) {
		return nil
	}
	v := a.Str(k)
	if v == "" {
		return nil
	}
	return &v
}
