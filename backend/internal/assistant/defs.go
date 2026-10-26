package assistant

import "fmt"

// The limits and the schemas of the kanban tools, as the IDE's assistant has them
// (pod/assistant/defs.go of the IDE): the MCP tools of the Kanban use them.

type obj = map[string]any

func str(description string) obj     { return obj{"type": "string", "description": description} }
func integer(description string) obj { return obj{"type": "integer", "description": description} }
func enum(description string, values ...string) obj {
	o := obj{"type": "string", "enum": values}
	if description != "" {
		o["description"] = description
	}
	return o
}

// MaxDescription and MaxNote are the limits of the kanban (characters).
const (
	MaxDescription = 1500
	MaxNote        = 1000
)

var Complexities = []string{"low", "medium", "high"}

// MaxStory is the longest user story of a milestone, in characters.
const MaxStory = 500

// MilestoneDescription and MilestoneProps: the kanban_milestone tool, for the assistant and the
// MCP endpoint.
const MilestoneDescription = `Milestones group the tickets of one feature (across lineages) with its spec, user stories and feedback; the user validates them. Actions: list; get (id); create (title, description, spec); update (id; title, description, spec, objective); story_add (id, text: "As a <user>, I <can do>…"), story_edit (id, story, text), story_delete (id, story); feedback_add (id, kind, text, complexity), feedback_done / feedback_reopen (id, feedback); note (id, text: a decision). Only the user validates a story or a milestone. Put a ticket in a milestone with the milestone field of kanban_create / kanban_update.`

var MilestoneProps = obj{
	"action":      enum("What to do", "list", "get", "create", "update", "story_add", "story_edit", "story_delete", "feedback_add", "feedback_done", "feedback_reopen", "note"),
	"id":          integer("Milestone number (all actions but list and create)"),
	"title":       str("Title (create, update)"),
	"description": str(fmt.Sprintf("Description in Markdown, %d characters max: the feature, its scope (create, update)", MaxDescription)),
	"spec":        str("The spec: a path of the repository or an address (create, update)"),
	"story":       integer("User story id (story_edit, story_delete)"),
	"text":        str(fmt.Sprintf("User story (%d characters max), feedback or note (%d characters max)", MaxStory, MaxNote)),
	"kind":        enum("Kind of feedback (feedback_add)", "bug", "feature", "info"),
	"complexity":  enum("How hard handling the feedback is (feedback_add)", Complexities...),
	"feedback":    integer("Feedback id (feedback_done, feedback_reopen)"),
	"objective":   integer("Default objective of the milestone; its open tickets go into it (update; 0: none); see kanban_objective"),
}

// ObjectiveDescription and ObjectiveProps: the kanban_objective tool, for the assistant and the
// MCP endpoint.
const ObjectiveDescription = `Objectives are named deadlines (MVP, a launch…) in order: a milestone says what (one feature), an objective says when, taking tickets of several milestones and lone ones. A ticket has one objective at most, apart from its milestone. Actions: list (in order); get (id: its tickets grouped by milestone); create (title, description, due); update (id; title, description, due); note (id, text: a decision). Only the user reaches, abandons or reorders an objective. Put a ticket in an objective with the objective field of kanban_create / kanban_update, a milestone's open tickets with the objective field of kanban_milestone update.`

var ObjectiveProps = obj{
	"action":      enum("What to do", "list", "get", "create", "update", "note"),
	"id":          integer("Objective number (get, update, note)"),
	"title":       str("Title (create, update)"),
	"description": str(fmt.Sprintf("Description in Markdown, %d characters max: what reaching it means (create, update)", MaxDescription)),
	"due":         str("Date aimed at, YYYY-MM-DD; empty: none (create, update)"),
	"text":        str(fmt.Sprintf("Note, %d characters max (note)", MaxNote)),
}
