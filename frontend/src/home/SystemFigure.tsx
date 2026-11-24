import type { ReactNode } from 'react'
import { useI18n } from '../i18n'
import type { ApiStatus } from './hooks'

function Node({ area, title, sub, children }: { area: string; title: string; sub: string; children?: ReactNode }) {
  return (
    <div className={`sys-node sys-${area}`}>
      <span className="sys-title">{title}</span>
      <span className="sys-sub">{sub}</span>
      {children}
    </div>
  )
}

function Wire({ area, active = false }: { area: string; active?: boolean }) {
  return <span className={`sys-wire sys-${area}${active ? ' is-active' : ''}`} aria-hidden="true" />
}

export function SystemFigure({ status }: { status: ApiStatus }) {
  const { t } = useI18n()
  const f = t.figure

  const label =
    status.state === 'up' ? (status.wokeSeconds && !status.ms ? f.woke(status.wokeSeconds) : f.up(status.ms)) : status.state === 'waking' ? f.waking : status.state === 'down' ? f.down : f.checking

  return (
    <figure className="sys-figure">
      <div className="sys" data-api={status.state}>
        <Node area="web" title={f.browser} sub={f.browserSub} />
        <Wire area="w1" active={status.state === 'checking' || status.state === 'waking'} />
        <Node area="api" title={f.api} sub={f.apiSub}>
          <span className="sys-status" role="status">
            <span className="sys-dot" aria-hidden="true" />
            {label}
          </span>
        </Node>
        <Wire area="w2" />
        <Node area="db" title={f.db} sub={f.dbSub} />
        <Wire area="w3" />
        <Node area="mail" title={f.mail} sub={f.mailSub} />
      </div>
      <figcaption>{f.caption}</figcaption>
    </figure>
  )
}
