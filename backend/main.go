// The backend of the Kanban: the tickets of a project, their lineages, milestones and objectives,
// the git workflow of a ticket (worktree, diff, merge) and the development sessions of the
// agents (docs/design/kanban.md of the IDE). The IDE starts it confined, one process per
// project (a worktree has its own, on the same base), and speaks JSON-RPC to it
// (antennae.space/plugin-sdk/backend). The base is <project>/.ide/kanban.db for a project of
// this machine, the data folder of the plugin for one over SSH. What it cannot do confined, the
// IDE does on its grants: git and gh with the user's keys, the projects of the worktrees and
// their setup, the push notifications, the agents (the bridge use:agents).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"antennae.space/plugin-sdk/backend"

	"github.com/antennae-space/plugin-kanban/backend/internal/i18n"
	kb "github.com/antennae-space/plugin-kanban/backend/internal/kanban"
	"github.com/antennae-space/plugin-kanban/backend/internal/plugins"
)

// K is the backend: the manager of the bases and the methods, tools and MCP tools it serves.
type K struct {
	Kanban *kb.Manager
	info   backend.Info
	// base is the context of initialize: it reaches the IDE for the life of the process (the
	// setup of a worktree outlives the call that started it).
	base    context.Context
	methods map[string]backend.Method
	mcp     map[string]backend.MCPTool
}

// srv is the backend, for what needs it beyond a method (statPath).
var srv *K

func main() {
	s := &K{methods: map[string]backend.Method{}, mcp: map[string]backend.MCPTool{}}
	srv = s
	s.registerKanban()
	s.registerKanbanGit()
	s.registerKanbanAgent()
	s.registerAssistantTools()
	s.registerMilestones()
	s.registerObjectives()
	s.registerMCP()
	// The schemas of the MCP tools, for the toolSchemas of plugin.json (scripts/schemas.sh).
	if len(os.Args) > 1 && os.Args[1] == "schemas" {
		out := map[string]any{}
		for _, t := range s.mcpTools() {
			out[t.Name] = map[string]any{"description": t.Description, "properties": t.Props, "required": t.Required}
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		os.Stdout.Write(append(data, '\n'))
		return
	}
	backend.Serve(backend.Plugin{
		Init:    s.start,
		Methods: s.methods,
		MCP:     s.mcp,
		Notify: map[string]func(context.Context, json.RawMessage){
			// A window opened the project: the old setup setting of the kanban moves to the
			// saved commands.
			"project.open": func(context.Context, json.RawMessage) { _, _ = s.projectCommands(s.base) },
		},
		Shutdown: func() {
			if s.Kanban != nil {
				s.Kanban.Close()
			}
		},
	})
}

func (s *K) start(ctx context.Context, info backend.Info) error {
	if exe, err := os.Executable(); err == nil {
		i18n.Load(filepath.Join(filepath.Dir(exe), "..", "i18n"))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	s.info, s.base = info, ctx
	if s.info.Project.Main == "" {
		s.info.Project.Main = info.Project.ID
	}
	if s.info.Project.MainRoot == "" {
		s.info.Project.MainRoot, s.info.Project.MainName = info.Project.Root, info.Project.Name
	}
	s.Kanban = kb.NewManager(home)
	return nil
}

// client is the call of a window: its project (this backend's) and its language.
type client struct {
	project string
	lang    string
}

func (c *client) language() string { return c.lang }

// runtime is the runtime of the call's project: the backend's.
func (c *client) runtime() (*Runtime, error) { return &Runtime{}, nil }

type handler func(ctx context.Context, c *client, p json.RawMessage) (any, error)

// handle serves a method of the kanban, run in the project of this backend; a coded error (not
// a git repository…) keeps its code for the views, and every error is in the caller's language.
func (s *K) handle(name string, f handler) {
	s.methods[name] = func(ctx context.Context, p json.RawMessage) (any, error) {
		res, err := f(ctx, &client{project: s.info.Project.ID, lang: backend.LangOf(ctx)}, p)
		var ce *plugins.CodedError
		if errors.As(err, &ce) {
			return nil, backend.Coded(ce.Code, i18n.Translate(backend.LangOf(ctx), ce.Err), ce.Data)
		}
		if err != nil {
			return nil, errors.New(i18n.Translate(backend.LangOf(ctx), err))
		}
		return res, nil
	}
}

func bind[T any](raw json.RawMessage) (T, error) {
	var v T
	if len(raw) == 0 || string(raw) == "null" {
		return v, nil
	}
	err := json.Unmarshal(raw, &v)
	return v, err
}
