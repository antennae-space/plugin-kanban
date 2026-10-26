// The Kanban, an external plugin of the IDE (docs/design/kanban.md of the IDE): each view a frame
// (its tool, the board, ticket, milestone and objective views, the New ticket modal), its
// background frame serving the tickets service the IDE uses (the project bar, the links, the
// assistant and its Orchestrator) through the IDE's adapter of tickets.
import { createEffect, createRoot } from 'solid-js'
import { view } from '@antennae-space/plugin-sdk/solid'
import type { SandboxAPI } from '@antennae-space/plugin-sdk'
import './base.css'
import './kanban.css'
import { t } from './plugin'
import { KanbanPanel } from './Panel'
import { Board, NewTicket } from './Board'
import { TicketView } from './TicketView'
import { MilestoneView } from './MilestoneView'
import { ObjectiveView } from './ObjectiveView'
import { viewStatus } from './status'
import { startTicketChat, startWorkSession } from './sessions'
import { blockerNames, board, ensureBoard, followBoardView, getTicket, openBoard, openMilestone, openObjective, openTicket, roleLabels, statusLabels, summary, watchBoard, worktreeProject, type Ticket } from './state'
import type { TabState, TicketRole } from './ide'

const short = (s: string, n: number) => (s.length > n ? s.slice(0, n - 1) + '…' : s)

/** The title of a tab: the ticket number and title, the name of the milestone or objective. */
function viewTitle(st: { ticket?: number; milestone?: number; objective?: number }) {
  if (st.ticket) {
    const s = summary(st.ticket)
    return s ? `#${s.id} ${short(s.title, 28)}` : `#${st.ticket}`
  }
  if (st.milestone) {
    const m = board.milestones.find((x) => x.id === st.milestone)
    return m ? short(m.title, 32) : t('Milestone')
  }
  if (st.objective) {
    const o = board.objectives.find((x) => x.id === st.objective)
    return o ? short(o.title, 32) : t('Objective')
  }
  return 'Kanban'
}

/** A tab view: the board loaded and followed, its title and its status items kept up to date. */
function tabView(C: (p: { tab: TabState; paneId: string }) => ReturnType<typeof Board>) {
  return view(() => {
    const ide = globalThis.ide
    const tab: TabState = { id: 'self', kind: ide.view?.id, state: ide.view?.state ?? {} }
    watchBoard()
    ensureBoard()
    followBoardView()
    createEffect(() => ide.setTitle(viewTitle((tab.state ?? {}) as Record<string, number>)))
    createEffect(() => {
      const items = viewStatus(tab)
      ide.setStatus(items.map((it) => ({ id: it.id, text: it.text, title: it.title, menu: it.menu?.().map((m, i) => ({ id: String(i), label: m.label, checked: m.checked, disabled: m.disabled, title: m.title })) })))
    })
    ide.onStatus((e) => {
      const it = viewStatus(tab).find((x) => x.id === e.item)
      it?.menu?.()[Number(e.choice)]?.action?.()
    })
    return C({ tab, paneId: 'self' })
  })
}

/** The tickets service, served by the background frame to the IDE (its adapter of tickets). */
function serveTickets(ide: SandboxAPI) {
  ide.publish('tickets', 'boardView', 'board')
  watchBoard()
  ensureBoard()
  createRoot(() => {
    createEffect(() => ide.publish('tickets', 'tickets', board.tickets.map((x) => ({ id: x.id, title: x.title, status: x.status, branch: x.branch, worktree: x.worktree }))))
    createEffect(() => ide.publish('tickets', 'labels', { status: statusLabels, role: roleLabels, blocker: Object.fromEntries(Object.entries(blockerNames).map(([k, v]) => [k, t(v)])) }))
  })
}

const ticketOf = async (id: number) => (await getTicket(id)) as Ticket

globalThis.ide.register({
  views: {
    // The background frame (the view ""): it serves the tickets.
    '': (_el, ide) => serveTickets(ide),
    kanban: view(() => (watchBoard(), followBoardView(), KanbanPanel())),
    board: tabView(Board),
    ticket: tabView(TicketView),
    milestone: tabView(MilestoneView),
    objective: tabView(ObjectiveView),
    'new-ticket': view(() => NewTicket({ onClose: () => globalThis.ide.closeModal() })),
  },
  commands: { open: () => openBoard() },
  services: {
    tickets: {
      // The view of the board its frames share.
      setBoardView: (v: string) => void globalThis.ide.publish('tickets', 'boardView', v),
      get: (id: number) => getTicket(id),
      open: (id: number) => openTicket(id),
      openBoard: () => openBoard(),
      openMilestone: (id: number) => openMilestone(id),
      openObjective: (id: number) => openObjective(id),
      startChat: async (id: number, role: TicketRole) => startTicketChat(await ticketOf(id), role),
      startWork: async (id: number, role: TicketRole, adoptedBy?: string) => startWorkSession(await ticketOf(id), role, undefined, false, adoptedBy),
      worktreeProject: (id: number) => worktreeProject(id),
    },
  },
})
