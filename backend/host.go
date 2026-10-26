package main

import (
	"context"
	"os"
	"strconv"

	"antennae.space/plugin-sdk/backend"

	kb "github.com/antennae-space/plugin-kanban/backend/internal/kanban"
	"github.com/antennae-space/plugin-kanban/backend/internal/plugins"
)

// What the IDE lends the Kanban, through its core calls: the main project, git and gh, the
// projects of the worktrees, the events to the windows, the agents.

// Project is the main project of the kanban, or the project of the worktree of a ticket.
type Project = plugins.Project

// childID is the project of the worktree of a ticket of a main project, as the IDE names it.
func childID(parent string, ticket int64) string {
	return parent + "-t" + strconv.FormatInt(ticket, 10)
}

// root is the main project of this backend's project.
func (s *K) root() *Project {
	typ := "ssh"
	if s.info.Project.Local {
		typ = "local"
	}
	return &Project{ID: s.info.Project.Main, Title: s.info.Project.MainName, Path: s.info.Project.MainRoot, Type: typ}
}

// kanbanProject is the project whose kanban (and conversations) a project uses: its main one.
func (s *K) kanbanProject(string) (*Project, error) { return s.root(), nil }

func kanbanLoc(p *Project) kb.Location { return kb.LocationOf(p.ID, p.Path, p.Type == "local") }

// runner runs git and gh through the IDE.
type runner struct{}

func (runner) Git(ctx context.Context, dir string, args ...string) (string, error) {
	return backend.Git(ctx, dir, args...)
}

func (runner) GitAny(ctx context.Context, dir string, args ...string) (string, error) {
	out, _, err := backend.GitAny(ctx, dir, args...)
	return out, err
}

func (runner) GH(ctx context.Context, args ...string) (string, error) {
	return backend.GH(ctx, args...)
}

func (runner) HasGH(ctx context.Context) bool {
	_, err := backend.GH(ctx, "--version")
	return err == nil
}

// emitKanban tells the windows of the main project and of its worktrees that a ticket changed.
func (s *K) emitKanban(root string, id int64, _ *client) {
	_ = backend.EmitFamily(s.base, "kanban.changed", map[string]any{"project": root, "id": id})
}

// statPath tells whether a folder of the project is there (a worktree): on this machine the
// backend sees it, over SSH git tells.
func statPath(p string) bool {
	if p == "" {
		return false
	}
	if srv.info.Project.Local {
		_, err := os.Stat(p)
		return err == nil
	}
	_, err := backend.Git(srv.base, p, "rev-parse", "--git-dir")
	return err == nil
}

// putChild registers the worktree of a ticket as a project of the IDE (again if it was).
func (s *K) putChild(owner int64, title, path string) (*Project, error) {
	id, err := backend.RegisterWorktree(s.base, owner, title, path)
	if err != nil {
		return nil, err
	}
	return &Project{ID: id, Title: title, Path: path, Type: s.root().Type}, nil
}

// Runtime stands for the runtime of a project the built-in Kanban had: the backend runs in its
// project, the IDE does the rest.
type Runtime struct{}

// The agents of the IDE, through the bridge of the core (use:agents): the development session
// of a ticket, the sessions of a removed worktree, the times of the conversations.
type (
	ChatTime struct {
		Parent      string
		Ms, ThinkMs int64
	}
	TicketBrief struct {
		ID, Parent, Feedback int64
		Role, Title, Branch  string
		Markdown, Text       string
	}
	AgentSession struct {
		ID, Name, Title, State string
	}
)

type agentsBridge struct{ ctx context.Context }

// agents is the agents capability of the IDE through its bridge.
func (s *K) agents(*Runtime) *agentsBridge { return &agentsBridge{s.base} }

func (a *agentsBridge) Develop(project, agent string, t TicketBrief) (AgentSession, bool, error) {
	var r struct {
		Session AgentSession `json:"session"`
		Created bool         `json:"created"`
	}
	err := backend.ServiceCall(a.ctx, "agents", "agents.develop", map[string]any{"project": project, "agent": agent, "ticket": t}, &r)
	return r.Session, r.Created, err
}

func (a *agentsBridge) TicketSession(project string, ticket int64) (AgentSession, bool) {
	var r struct {
		Session AgentSession `json:"session"`
		Found   bool         `json:"found"`
	}
	if backend.ServiceCall(a.ctx, "agents", "agents.ticketSession", map[string]any{"project": project, "ticket": ticket}, &r) != nil {
		return AgentSession{}, false
	}
	return r.Session, r.Found
}

func (a *agentsBridge) Purge(project string) {
	_ = backend.ServiceCall(a.ctx, "agents", "agents.purge", map[string]any{"project": project}, nil)
}

func (a *agentsBridge) ChatTimes(project string) (map[string]ChatTime, error) {
	var r map[string]ChatTime
	err := backend.ServiceCall(a.ctx, "agents", "agents.chatTimes", map[string]any{"project": project}, &r)
	return r, err
}
