// Package plugins holds what the Kanban's code took from the IDE's plugin API when it was a
// built-in, rebuilt on what an external plugin gets: a project as initialize tells it, the
// scope of an MCP call, the MCP tools and their schemas, a coded error, a notification.
package plugins

import (
	"context"
	"encoding/json"
	"net/url"
)

// Project is a project of the IDE: the main project of the kanban, or the project of the
// worktree of a ticket.
type Project struct {
	ID, Path, Type string
	Title          string
}

// Name is the name of the project shown by the IDE.
func (p *Project) Name() string { return p.Title }

// MCPScope is what a call of the MCP works on, found by the IDE from the folder of the client.
type MCPScope struct {
	// Here is the project of the folder (a worktree has its own), Root its main project.
	Here, Root *Project
	// Ticket is the ticket of the worktree the client works in, 0 elsewhere.
	Ticket int64
	// URL is the address the user opens the IDE at.
	URL string
	// Lang is the language of the interface.
	Lang string
}

// ProjectURL is the page of a project in the IDE, at the address the user opens it.
func (sc MCPScope) ProjectURL(id string) string { return sc.URL + "/project/" + url.PathEscape(id) }

// MCPTool is a tool of the MCP of the plugin: its name, description and parameters (besides
// cwd, which the IDE adds), described in the manifest's toolSchemas too.
type MCPTool struct {
	Name        string
	Description string
	Props       map[string]any
	Required    []string
	Run         func(ctx context.Context, sc MCPScope, args json.RawMessage) (string, error)
}

// Str, StrList and Enum are the JSON schemas of a parameter.
func Str(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func StrList(description string) map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": description}
}

func Enum(description string, values ...string) map[string]any {
	return map[string]any{"type": "string", "enum": values, "description": description}
}

// CodedError is an error with a code the page tells apart (not_git…).
type CodedError struct {
	Code string
	Data map[string]string
	Err  error
}

func (e *CodedError) Error() string { return e.Err.Error() }
func (e *CodedError) Unwrap() error { return e.Err }
