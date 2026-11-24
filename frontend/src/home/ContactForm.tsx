import { useEffect, useRef, useState, type FormEvent } from 'react'
import { sendContact } from '../api'
import { useI18n } from '../i18n'
import type { ApiStatus } from './hooks'

const LIMITS = { name: 100, email: 254, message: 5000 }

type Field = 'name' | 'email' | 'message'
type Values = Record<Field, string>
type Errors = Partial<Record<Field, string>>

type Phase =
  | { kind: 'idle' }
  | { kind: 'sending' }
  | { kind: 'sent'; email: string }
  | { kind: 'rateLimited'; minutes: number }
  | { kind: 'failed' }

const EMPTY: Values = { name: '', email: '', message: '' }

export function ContactForm({ apiStatus, email }: { apiStatus: ApiStatus; email: string }) {
  const { t } = useI18n()
  const c = t.contact
  const [values, setValues] = useState<Values>(EMPTY)
  const [errors, setErrors] = useState<Errors>({})
  const [phase, setPhase] = useState<Phase>({ kind: 'idle' })
  const openedAt = useRef(0)
  const doneRef = useRef<HTMLDivElement>(null)
  const refocusName = useRef(false)
  const honeypot = useRef<HTMLInputElement>(null)
  const fieldRefs = {
    name: useRef<HTMLInputElement>(null),
    email: useRef<HTMLInputElement>(null),
    message: useRef<HTMLTextAreaElement>(null),
  }

  useEffect(() => {
    openedAt.current = performance.now()
  }, [])

  useEffect(() => {
    if (phase.kind === 'sent') doneRef.current?.focus()
    else if (refocusName.current) {
      refocusName.current = false
      fieldRefs.name.current?.focus()
    }
  }, [phase.kind]) // eslint-disable-line react-hooks/exhaustive-deps

  function validate(v: Values): Errors {
    const e: Errors = {}
    for (const field of ['name', 'email', 'message'] as const) {
      const value = v[field].trim()
      if (!value) e[field] = c.required
      else if (value.length > LIMITS[field]) e[field] = c.tooLong(LIMITS[field])
    }
    if (!e.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v.email.trim())) e.email = c.badEmail
    return e
  }

  function focusFirst(e: Errors) {
    const first = (['name', 'email', 'message'] as const).find((f) => e[f])
    if (first) fieldRefs[first].current?.focus()
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (phase.kind === 'sending') return

    const clientErrors = validate(values)
    setErrors(clientErrors)
    if (Object.keys(clientErrors).length > 0) {
      focusFirst(clientErrors)
      return
    }

    setPhase({ kind: 'sending' })
    const result = await sendContact({
      name: values.name.trim(),
      email: values.email.trim(),
      message: values.message.trim(),
      website: honeypot.current?.value ?? '',
      elapsedMs: Math.round(performance.now() - openedAt.current),
    })

    switch (result.kind) {
      case 'sent':
        setPhase({ kind: 'sent', email: values.email.trim() })
        setValues(EMPTY)
        break
      case 'invalid': {
        const serverErrors: Errors = {}
        for (const field of ['name', 'email', 'message'] as const) {
          if (result.fields[field]) serverErrors[field] = field === 'email' ? c.badEmail : c.required
        }
        setErrors(serverErrors)
        setPhase({ kind: 'idle' })
        focusFirst(serverErrors)
        break
      }
      case 'rateLimited':
        setPhase({ kind: 'rateLimited', minutes: Math.max(1, Math.ceil(result.retryAfterSeconds / 60)) })
        break
      case 'failed':
        setPhase({ kind: 'failed' })
        break
    }
  }

  function update(field: Field, value: string) {
    setValues((v) => ({ ...v, [field]: value }))
    if (errors[field]) setErrors((e) => ({ ...e, [field]: undefined }))
    if (phase.kind === 'failed' || phase.kind === 'rateLimited') setPhase({ kind: 'idle' })
  }

  if (phase.kind === 'sent') {
    return (
      <div className="contact-done" role="status" tabIndex={-1} ref={doneRef}>
        <p className="contact-done-text">{c.sent(phase.email)}</p>
        <button
          type="button"
          className="btn btn-quiet"
          onClick={() => {
            openedAt.current = performance.now()
            refocusName.current = true
            setPhase({ kind: 'idle' })
          }}
        >
          {c.sendAnother}
        </button>
      </div>
    )
  }

  const sending = phase.kind === 'sending'
  const cold = apiStatus.state !== 'up'
  const messageCount = values.message.length

  const fieldProps = (field: Field) => ({
    id: `contact-${field}`,
    name: field,
    value: values[field],
    'aria-invalid': errors[field] ? true : undefined,
    'aria-describedby': errors[field] ? `contact-${field}-error` : undefined,
    disabled: sending,
  })

  return (
    <form className="contact-form" onSubmit={onSubmit} noValidate>
      <div className="field">
        <label htmlFor="contact-name">{c.name}</label>
        <input
          {...fieldProps('name')}
          ref={fieldRefs.name}
          autoComplete="name"
          maxLength={LIMITS.name}
          onChange={(e) => update('name', e.target.value)}
        />
        {errors.name && <p className="field-error" id="contact-name-error">{errors.name}</p>}
      </div>

      <div className="field">
        <label htmlFor="contact-email">{c.email}</label>
        <input
          {...fieldProps('email')}
          ref={fieldRefs.email}
          type="email"
          inputMode="email"
          autoComplete="email"
          spellCheck={false}
          maxLength={LIMITS.email}
          onChange={(e) => update('email', e.target.value)}
        />
        {errors.email && <p className="field-error" id="contact-email-error">{errors.email}</p>}
      </div>

      <div className="field field-message">
        <div className="field-head">
          <label htmlFor="contact-message">{c.message}</label>
          {messageCount > LIMITS.message * 0.8 && (
            <span className="field-count">{c.counter(messageCount, LIMITS.message)}</span>
          )}
        </div>
        <textarea
          {...fieldProps('message')}
          ref={fieldRefs.message}
          rows={3}
          maxLength={LIMITS.message}
          onChange={(e) => update('message', e.target.value)}
        />
        {errors.message && <p className="field-error" id="contact-message-error">{errors.message}</p>}
      </div>

      <div className="hp" aria-hidden="true">
        <label htmlFor="contact-website">Website</label>
        <input id="contact-website" name="website" ref={honeypot} tabIndex={-1} autoComplete="off" />
      </div>

      <div className="contact-actions">
        <button type="submit" className="btn" disabled={sending} aria-busy={sending || undefined}>
          {sending ? c.sending : c.send}
        </button>
        <p className="contact-note" aria-live="polite">
          {sending && cold && c.sendingCold}
          {!sending && phase.kind === 'idle' && (
            <>
              {c.orEmail} <a href={`mailto:${email}`}>{email}</a>
            </>
          )}
          {phase.kind === 'rateLimited' && (
            <span className="is-error">
              {c.rateLimited(phase.minutes)} <a href={`mailto:${email}`}>{email}</a>
            </span>
          )}
          {phase.kind === 'failed' && (
            <span className="is-error">
              {c.failed} <a href={`mailto:${email}`}>{email}</a>
            </span>
          )}
        </p>
      </div>
    </form>
  )
}
