// The time and cost of the agents on a ticket (the sessions of its worktree), from the Agents
// tool of the IDE (its method agents.stats, the grant core:agents.stats), asked again when the
// ticket changes; hidden before any turn or without the Agents tool.
import { createResource, Show } from 'solid-js'
import { request, t } from './plugin'
import { ticketVersion } from './state'

interface Totals {
  turns: number
  work: number
  wait: number
  tokens: number
  cost: number
  unpriced?: number
  apiEquivalent?: boolean
}

function fmtWork(ms: number) {
  const s = Math.round(ms / 1000)
  if (s < 60) return `${s} s`
  if (s < 3600) return `${Math.round(s / 60)} min`
  return `${Math.floor(s / 3600)} h ${String(Math.round((s % 3600) / 60)).padStart(2, '0')}`
}

const fmtCost = (c: number) => '$' + (c > 0 && c < 0.01 ? c.toFixed(4) : c.toFixed(2))

function statsText(s: Totals) {
  return s.turns - (s.unpriced ?? 0) > 0 ? `${fmtWork(s.work)} · ${fmtCost(s.cost)}` : fmtWork(s.work)
}

function statsTitle(s: Totals) {
  const lines = [
    t('Work: {time}', { time: fmtWork(s.work) }),
    t('Waiting for you: {time}', { time: fmtWork(s.wait) }),
    t('{n} turns, {tokens} tokens', { n: s.turns, tokens: s.tokens.toLocaleString() }),
  ]
  if (s.turns - (s.unpriced ?? 0) > 0) lines.push(s.apiEquivalent ? t('Cost: {cost} (API equivalent)', { cost: fmtCost(s.cost) }) : t('Cost: {cost}', { cost: fmtCost(s.cost) }))
  if (s.unpriced) lines.push(t('{n} turns without a price (local or unknown model)', { n: s.unpriced }))
  return lines.join('\n')
}

export function AgentStats(props: { ticket: number }) {
  const [s] = createResource(
    () => ({ id: props.ticket, v: ticketVersion(props.ticket) }),
    ({ id }) => request<Totals>('agents.stats', { ticket: id, since: 0 }).catch(() => null),
  )
  return (
    <Show when={s()?.turns}>
      <span class="badge" title={statsTitle(s()!)} data-testid="ticket-agent-stats">
        {t('Agents: {stats}', { stats: statsText(s()!) })}
      </span>
    </Show>
  )
}
