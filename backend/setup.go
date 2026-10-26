package main

import (
	"context"
	"log"
	"strings"
	"sync"

	"antennae.space/plugin-sdk/backend"
)

// Command is a saved command of the project, as the IDE keeps it (the core calls
// commands.list and commands.save).
type Command struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Command     string `json:"command"`
	Flag        string `json:"flag"`
	Cwd         string `json:"cwd,omitempty"`
	Port        int    `json:"port,omitempty"`
}

// migrateMu keeps two loads from moving the setup command of the kanban twice.
var migrateMu sync.Mutex

// projectCommands loads the saved commands of the project. The setup command of the kanban
// settings (before the setup flag) becomes a setup command once, then is cleared: when a
// window opens the project, or before a worktree is set up.
func (s *K) projectCommands(ctx context.Context) ([]Command, error) {
	migrateMu.Lock()
	defer migrateMu.Unlock()
	var r struct {
		Commands []Command `json:"commands"`
	}
	if err := backend.Core(ctx, "commands.list", map[string]any{}, &r); err != nil {
		return nil, err
	}
	list := r.Commands
	root := s.root()
	loc := kanbanLoc(root)
	if !s.Kanban.Has(loc) {
		return list, nil
	}
	meta, err := s.Kanban.Meta(loc)
	setup := strings.TrimSpace(meta["setup"])
	if err != nil || setup == "" {
		return list, nil
	}
	var saved []Command
	if err := backend.Core(ctx, "commands.save", map[string]any{"commands": append(list, Command{Name: "Worktree setup", Command: setup, Flag: "setup",
		Description: "Moved from the kanban settings"})}, &saved); err != nil {
		log.Printf("setup command of the kanban of %s not moved: %v", root.ID, err)
		return list, nil
	}
	return saved, s.Kanban.SetMeta(loc, map[string]string{"setup": ""})
}
