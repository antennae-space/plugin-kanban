// The types of the IDE the Kanban's views use, as the IDE has them (web/src/plugins of the IDE):
// a tab, a place in a file, a menu item, the services of the IDE it reaches through its
// bridges, and the tickets service it gives the IDE.

/** A tab of a view of the Kanban: its state (the ticket, milestone or objective shown). */
export interface TabState {
  id: string
  kind?: string
  state?: unknown
}

/** A place in a file: its path, its line from 0. */
export interface NavTarget {
  path: string
  line?: number
}

export interface MenuItem {
  label: string
  action?: () => void
  disabled?: boolean
  hint?: string
  title?: string
  danger?: boolean
  separator?: boolean
  checked?: boolean
}

/** An item of the status bar of a tab, its menu's actions run in the frame. */
export interface StatusItem {
  id: string
  text: string
  icon?: string
  title?: string
  testid?: string
  action?: () => void
  menu?: () => MenuItem[]
}

export type TicketRole = 'briefing' | 'plan' | 'dev' | 'correction' | 'resolve'

/** A ticket as the IDE sees it. */
export interface TicketSummary {
  id: number
  title: string
  status: 'new' | 'todo' | 'in_progress' | 'to_test' | 'done' | 'abandoned'
  branch?: string
  worktree?: string
}

export interface TicketBlocker {
  id: number
  kind: string
}

/** The assistant of the IDE (the bridge use:assistant). */
export interface AssistantService {
  chat: () => { id: string; ticket?: { id: number; role: TicketRole } }
  config: () => { server: string; model: string }
  loadConfig(): Promise<void>
  newChat(o: Record<string, unknown>): void
  send(text: string): Promise<void>
  openChat(id: string): Promise<void>
  /** The Assistant tool shown. */
  show(): void
}

/** The Agents tool of the IDE (the bridge use:agents). */
export interface AgentsService {
  catalog(project?: string): Promise<{ agents: { id: string; name: string; installed: boolean }[]; local: boolean }>
  /** A session shown in the Agents tool. */
  show(session: string): void
}

/** The consoles of the IDE (the bridge use:terminal): a console opened, its id. */
export interface TerminalService {
  open(o: { cwd?: string; command?: string[]; title?: string; kind?: 'terminal' | 'task' }): Promise<string | null>
}

/** A saved command of the project. */
export interface SavedCommand {
  id: string
  name: string
  description?: string
  command: string
  flag: 'serve' | 'build' | 'test' | 'lint' | 'setup' | 'other'
  cwd?: string
  port?: number
}

/** The saved commands of the project (the bridge use:commands). */
export interface CommandsService {
  list: () => SavedCommand[]
  load(): Promise<void>
  edit(o?: { select?: string; draft?: Partial<SavedCommand> }): void
}

export interface GitBranch {
  name: string
  current: boolean
  remote?: boolean
  upstream?: string
}
