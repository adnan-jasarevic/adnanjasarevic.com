export const API_URL = (import.meta.env.VITE_API_URL ?? 'http://localhost:8080').replace(/\/$/, '')

export type Profile = {
  fullName: string
  headline: string
  summary: string
  location: string
  email: string | null
  githubUrl: string | null
  linkedinUrl: string | null
  updatedAt: string
}

export type Experience = {
  id: number
  company: string
  role: string
  location: string
  startDate: string
  endDate: string | null
  description: string
  highlights: string[]
}

export type Education = {
  id: number
  institution: string
  degree: string
  field: string
  startDate: string
  endDate: string | null
  description: string
}

export type Skill = {
  id: number
  name: string
  category: string
  sortOrder: number
}

export type CV = {
  profile: Profile | null
  experience: Experience[]
  education: Education[]
  skills: Skill[]
}

export type Project = {
  id: number
  title: string
  slug: string
  summary: string
  description: string
  techStack: string[]
  repoUrl: string | null
  liveUrl: string | null
}

// A cold start on the free tier takes about a minute.
const COLD_START_TIMEOUT = 90_000

export async function getJSON<T>(path: string, signal?: AbortSignal): Promise<T> {
  const timeout = AbortSignal.timeout(COLD_START_TIMEOUT)
  const res = await fetch(API_URL + path, {
    signal: signal ? AbortSignal.any([signal, timeout]) : timeout,
  })
  if (!res.ok) throw new Error(`GET ${path}: ${res.status}`)
  return res.json() as Promise<T>
}

export type ContactInput = {
  name: string
  email: string
  message: string
  website: string
  elapsedMs: number
}

export type ContactResult =
  | { kind: 'sent' }
  | { kind: 'invalid'; fields: Record<string, string> }
  | { kind: 'rateLimited'; retryAfterSeconds: number }
  | { kind: 'failed' }

export async function sendContact(input: ContactInput): Promise<ContactResult> {
  let res: Response
  try {
    res = await fetch(API_URL + '/contact', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
      signal: AbortSignal.timeout(COLD_START_TIMEOUT),
    })
  } catch {
    return { kind: 'failed' }
  }

  if (res.status === 201) return { kind: 'sent' }
  if (res.status === 429) {
    const seconds = Number(res.headers.get('Retry-After'))
    return { kind: 'rateLimited', retryAfterSeconds: Number.isFinite(seconds) && seconds > 0 ? seconds : 600 }
  }
  if (res.status === 400) {
    const body = (await res.json().catch(() => null)) as { fields?: Record<string, string> } | null
    if (body?.fields) return { kind: 'invalid', fields: body.fields }
  }
  return { kind: 'failed' }
}
