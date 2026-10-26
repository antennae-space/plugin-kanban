// The Kanban's door to the IDE, in a frame of the sandbox (ide, the plugin SDK): the names the
// board, the tickets, the milestones and the objectives use, as they were when the Kanban was a
// built-in of the IDE. The assistant, the Agents tool, the consoles and the saved commands are
// the bridges of the IDE (use:<capability>); the git branches its core method (git.branches).
import { createSignal, type JSX } from 'solid-js'
import '@antennae-space/plugin-sdk'
import { EmptyState as KitEmptyState, fuzzy as kitFuzzy, HelpLink as KitHelpLink, Icon as KitIcon, Markdown as KitMarkdown, Modal as KitModal, openMenu, type IconName } from '@antennae-space/plugin-sdk/ui'
import type { SandboxAPI } from '@antennae-space/plugin-sdk'
import type { AgentsService, AssistantService, CommandsService, GitBranch, SavedCommand, TerminalService } from './ide'

export type { TabState, NavTarget, MenuItem } from './ide'

const ide = (): SandboxAPI => globalThis.ide

export const t = (key: string, params?: Record<string, unknown>) => ide().t(key, params)
export const tn = (n: number, one: string, other: string, params?: Record<string, unknown>) => ide().tn(n, one, other, params)

const locale = () => ide().lang || 'en'
export const fmtDate = (d: number | string | Date) => new Date(d).toLocaleString(locale())
export function fmtSize(n: number): string {
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024
    i++
  }
  return `${i === 0 ? n : n.toFixed(n < 10 ? 1 : 0)} ${units[i]}`
}

export const Icon = (p: { name: string; size?: number; title?: string; class?: string }): JSX.Element => KitIcon({ ...p, name: p.name as IconName })
export const EmptyState = (p: { icon: string; text: JSX.Element; action?: string; onAction?: () => void; testid?: string }): JSX.Element =>
  KitEmptyState({
    icon: p.icon as IconName,
    text: typeof p.text === 'string' ? p.text : '',
    get children() {
      return [
        typeof p.text === 'string' ? null : p.text,
        p.action ? (
          <button class="btn primary" onClick={() => p.onAction?.()}>
            {p.action}
          </button>
        ) : null,
      ]
    },
  })

/** A modal: the frame itself in a modal view of the IDE (its body and its buttons), else the kit's, in the frame. */
export const Modal = (p: { title: string; onClose: () => void; children: JSX.Element; class?: string; footer?: JSX.Element }): JSX.Element => {
  if (ide().modal) {
    return (
      <div class="kb-modal-view">
        {p.children}
        {p.footer ? <div class="form-actions">{p.footer}</div> : null}
      </div>
    )
  }
  return KitModal({
    title: p.title,
    onClose: p.onClose,
    footer: p.footer,
    get children() {
      return p.children
    },
  })
}
export const Markdown = (p: { text: string; final?: boolean }): JSX.Element => KitMarkdown({ text: p.text })
export const HelpLink = (p: { topic: string }): JSX.Element => KitHelpLink({ topic: p.topic, title: t('Help: open the documentation page') })

/** A menu at a click: the kit's, in the frame. */
export function contextMenu(e: MouseEvent, items: { label: string; action?: () => void; disabled?: boolean; danger?: boolean; separator?: boolean; checked?: boolean; hint?: string; title?: string }[]) {
  e.preventDefault()
  openMenu(
    items.map((it) => (it.separator ? 'separator' : { label: (it.checked ? '✓ ' : '') + it.label, run: () => it.action?.(), disabled: it.disabled, danger: it.danger, keys: it.hint })),
    e,
  )
}

/** A yes or no asked by the IDE (a frame has no confirm of the browser). */
export const ask = (text: string, danger = false) => ide().confirm({ title: t('Kanban'), text, danger })

/** A text asked by the IDE (null when cancelled). */
export const prompt = (o: { title: string; label?: string; value?: string; placeholder?: string; password?: boolean }) => ide().prompt(o)

type PickItem<T> = { label: string; detail?: string; value: T }

/**
 * An item chosen in a list by the IDE: its value (undefined when cancelled). A provider gives
 * the items once (the IDE's list filters them as the user types, 500 at most).
 */
export async function pick<T>(o: { placeholder?: string; items?: PickItem<T>[]; provider?: (q: string) => PickItem<T>[] | Promise<PickItem<T>[]>; pathDetail?: boolean; initial?: number }): Promise<T | undefined> {
  const items = (o.items ?? (o.provider ? await o.provider('') : [])).slice(0, 500)
  const i = await ide().pick({ title: o.placeholder ?? '', items: items.map((x) => ({ label: x.label, detail: x.detail })), initial: o.initial })
  return i === null ? undefined : items[i]?.value
}

export const fuzzy = kitFuzzy
export const toast = (text: string, kind?: 'info' | 'warn' | 'error' | 'ok') => ide().toast(text, kind)
export const errorToast = (e: unknown) => ide().toast(e instanceof Error ? e.message : String(e), 'error')
export const copyText = (text: string) => Promise.resolve(ide().copyText(text))
export const request = <T = any,>(method: string, params?: any): Promise<T> => ide().rpc<T>(method, params)

/** An event of the Kanban's backend (kanban.*). */
export function on(event: string, f: (data: any) => void): () => void {
  let live = true
  ide().on(event, (d) => live && f(d))
  return () => (live = false)
}

/** A file of the project; line from 0, as the IDE's own views count. */
export const openFile = (target: { path: string; line?: number } | string) =>
  typeof target === 'string' ? ide().openFile(target) : ide().openFile(target.path, target.line === undefined ? undefined : target.line + 1)

// The project of the window: it changes when the window shows another worktree.
const [proj, setProj] = createSignal(ide()?.project ?? null)
setInterval(() => {
  const p = ide()?.project ?? null
  if (p?.id !== proj()?.id) setProj(p)
}, 500)

export const root = () => (proj() ?? ide()?.project)?.root ?? ''
export const relPath = (p: string) => (p === root() ? '.' : p.startsWith(root() + '/') ? p.slice(root().length + 1) : p)
export const basename = (p: string) => p.slice(p.replace(/\/+$/, '').lastIndexOf('/') + 1)

/** The project of the window: its id, its parent (a worktree's main project), its ticket. */
export const project = () => {
  const p = proj() ?? ide()?.project
  if (!p) return undefined
  const m = p.id !== p.main ? /-t(\d+)$/.exec(p.id) : null
  return { id: p.id, name: p.name, local: p.local, parent: p.id !== p.main ? p.main : '', ticket: m ? Number(m[1]) : 0 }
}
export const showWorktree = (id: string) => ide().showWorktree(id)

// The capabilities the Kanban uses, through the bridges of the IDE (or the core methods).
const use = (name: string) => ide().use(name)

const assistantService = (): AssistantService => {
  const a = use('assistant')
  return {
    chat: () => (a.value('chat') as ReturnType<AssistantService['chat']>) ?? { id: '' },
    config: () => (a.value('config') as ReturnType<AssistantService['config']>) ?? { server: '', model: '' },
    loadConfig: () => a.call('loadConfig').then(() => undefined),
    newChat: (o) => void a.call('newChat', o),
    send: (text) => a.call('send', text).then(() => undefined),
    openChat: (id) => a.call('openChat', id).then(() => undefined),
    show: () => void a.call('show'),
  }
}

const agentsService = (): AgentsService => {
  const g = use('agents')
  return {
    catalog: (project) => g.call('catalog', project),
    show: (session) => void g.call('show', session),
  }
}

const terminalService = (): TerminalService => {
  const term = use('terminal')
  return { open: (o) => term.call('open', o) as Promise<string | null> }
}

// The saved commands, followed (the editor and the list of the IDE).
const [savedCommands, setSavedCommands] = createSignal<SavedCommand[]>([])
let watchingCommands = false
const commandsService = (): CommandsService => {
  const c = use('commands')
  if (!watchingCommands) {
    watchingCommands = true
    c.onValue('list', (l) => setSavedCommands((l as SavedCommand[]) ?? []))
  }
  return {
    list: savedCommands,
    load: () => c.call('load').then(() => undefined),
    edit: (o) => void c.call('edit', o),
  }
}

const gitService = () => ({ branches: () => request<GitBranch[]>('git.branches') })

type Services = { assistant: AssistantService; agents: AgentsService; terminal: TerminalService; commands: CommandsService; git: ReturnType<typeof gitService> }

const made: Partial<Services> = {}
const makers: { [K in keyof Services]: () => Services[K] } = { assistant: assistantService, agents: agentsService, terminal: terminalService, commands: commandsService, git: gitService }

/** A service of the IDE the Kanban uses (always there: a call fails when the IDE lacks it). */
export function capability<K extends keyof Services>(name: K): Services[K] {
  return (made[name] ??= makers[name]()) as Services[K]
}

// What the views did through the api of a built-in.
export const api = {
  openView: (view: string, o: { title?: string; key?: string; state?: unknown } = {}) => ide().openView(view, o.state, o.title, o.key),
  panes: { closeTab: (_id?: string) => ide().closeView() },
  showTool: (id: string) => (id === 'kanban' ? ide().showTool(id) : id === 'assistant' ? capability('assistant').show() : undefined),
}
