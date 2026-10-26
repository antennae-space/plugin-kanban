package main

import (
	"strconv"

	"antennae.space/plugin-sdk/backend"

	"github.com/antennae-space/plugin-kanban/backend/internal/i18n"
	kb "github.com/antennae-space/plugin-kanban/backend/internal/kanban"
)

// NotifyFeedback sends the push notification of a test feedback the assistant or Claude Code
// added to a ticket (grant notify): it opens the ticket.
func (s *K) NotifyFeedback(loc kb.Location, id int64, text string) {
	ticket := ""
	if t, err := s.Kanban.Get(loc, id); err == nil {
		ticket = t.Title
	}
	name := s.root().Name()
	if name == "" {
		name = loc.Project
	}
	n := strconv.FormatInt(id, 10)
	texts := map[string]backend.NoteTexts{}
	for _, lang := range append([]string{"en"}, i18n.Languages()...) {
		texts[lang] = backend.NoteTexts{Title: i18n.T(lang, "Feedback on #%d %s", id, ticket), Body: name + " · " + text}
	}
	_ = backend.Notify(s.base, backend.Note{Kind: "feedback", Query: "ticket=" + n, Tag: "ticket-" + n, Urgency: "normal", Texts: texts})
}
