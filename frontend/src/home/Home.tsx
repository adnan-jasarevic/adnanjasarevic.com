import type { ReactNode } from 'react'
import type { CV, Education, Experience, Project, Skill } from '../api'
import { useI18n, type Strings } from '../i18n'
import { site } from '../site'
import { ContactForm } from './ContactForm'
import { useApiStatus, useResource, type Resource } from './hooks'
import { SystemFigure } from './SystemFigure'

const YEAR = new Date().getFullYear()
const REPO_URL = 'https://github.com/adnan-jasarevic/adnanjasarevic.com'

const SECTIONS = ['features', 'system', 'projects', 'experience', 'education', 'links'] as const
type SectionId = (typeof SECTIONS)[number]

function Section({ id, children }: { id: SectionId; children: ReactNode }) {
  const { t } = useI18n()
  return (
    <section id={id} className="sheet-section" aria-labelledby={`${id}-heading`}>
      <h2 id={`${id}-heading`} className="section-heading">
        <span className="section-num">{SECTIONS.indexOf(id) + 1}</span>
        {sectionTitle(t.sections, id)}
      </h2>
      {children}
    </section>
  )
}

function sectionTitle(titles: Strings['sections'], id: SectionId) {
  return id === 'links' && site.cvPdf ? titles.linksCv : titles[id]
}

function Ghost({ rows = 3, note }: { rows?: number; note?: string }) {
  return (
    <div className="ghost" role="status">
      {note && <p className="ghost-note">{note}</p>}
      {Array.from({ length: rows }, (_, i) => (
        <span key={i} className="ghost-row" aria-hidden="true" />
      ))}
    </div>
  )
}

function Loadable<T>({ resource, rows, children }: { resource: Resource<T>; rows?: number; children: (data: T) => ReactNode }) {
  const { t } = useI18n()
  if (resource.status === 'loading') return <Ghost rows={rows} note={t.loading} />
  if (resource.status === 'error') return <p className="empty-row">{t.notLoaded}</p>
  return <>{children(resource.data)}</>
}

function Empty() {
  const { t } = useI18n()
  return <p className="empty-row">{t.noEntries}</p>
}

function formatMonth(iso: string, locale: string) {
  return new Intl.DateTimeFormat(locale, { month: 'short', year: 'numeric', timeZone: 'UTC' }).format(new Date(iso))
}

function Period({ start, end }: { start: string; end: string | null }) {
  const { t, locale } = useI18n()
  return (
    <span className="period">
      <time dateTime={start.slice(0, 7)}>{formatMonth(start, locale)}</time>
      {' – '}
      {end ? <time dateTime={end.slice(0, 7)}>{formatMonth(end, locale)}</time> : t.table.present}
    </span>
  )
}

function Features({ skills, t }: { skills: Skill[]; t: Strings }) {
  if (skills.length === 0) return <Empty />
  const groups = new Map<string, string[]>()
  for (const skill of [...skills].sort((a, b) => a.sortOrder - b.sortOrder)) {
    const key = skill.category || t.skillsOther
    groups.set(key, [...(groups.get(key) ?? []), skill.name])
  }
  return (
    <dl className="features">
      {[...groups].map(([category, names]) => (
        <div className="features-row" key={category}>
          <dt>{category}</dt>
          <dd>
            <ul>
              {names.map((name) => (
                <li key={name}>{name}</li>
              ))}
            </ul>
          </dd>
        </div>
      ))}
    </dl>
  )
}

function Projects({ projects }: { projects: Project[] }) {
  const { t } = useI18n()
  if (projects.length === 0) return <Empty />
  return (
    <div className="projects">
      {projects.map((p, i) => (
        <article className="project" key={p.id} aria-labelledby={`project-${p.id}`}>
          <div className="project-text">
            <h3 id={`project-${p.id}`}>
              <span className="project-num">3.{i + 1}</span>
              {p.title}
            </h3>
            {p.summary && <p className="project-summary">{p.summary}</p>}
            {p.description && <p className="project-desc">{p.description}</p>}
          </div>
          <table className="spec">
            <tbody>
              <tr>
                <th scope="row">{t.projects.tech}</th>
                <td>{p.techStack.length > 0 ? p.techStack.join(', ') : t.projects.none}</td>
              </tr>
              <tr>
                <th scope="row">{t.projects.source}</th>
                <td>{p.repoUrl ? <a href={p.repoUrl}>{t.projects.viewSource}</a> : t.projects.none}</td>
              </tr>
              <tr>
                <th scope="row">{t.projects.live}</th>
                <td>{p.liveUrl ? <a href={p.liveUrl}>{t.projects.viewLive}</a> : t.projects.none}</td>
              </tr>
            </tbody>
          </table>
        </article>
      ))}
    </div>
  )
}

function ExperienceTable({ items }: { items: Experience[] }) {
  const { t } = useI18n()
  if (items.length === 0) return <Empty />
  return (
    <div className="rows" role="list">
      <div className="rows-head" aria-hidden="true">
        <span>{t.table.period}</span>
        <span>{t.table.role}</span>
        <span>{t.table.company}</span>
        <span>{t.table.location}</span>
      </div>
      {items.map((x) => (
        <article className="row" role="listitem" key={x.id}>
          <Period start={x.startDate} end={x.endDate} />
          <h3 className="row-title">{x.role}</h3>
          <span className="row-org">{x.company}</span>
          <span className="row-loc">{x.location}</span>
          {(x.description || x.highlights.length > 0) && (
            <div className="row-body">
              {x.description && <p>{x.description}</p>}
              {x.highlights.length > 0 && (
                <ul>
                  {x.highlights.map((h) => (
                    <li key={h}>{h}</li>
                  ))}
                </ul>
              )}
            </div>
          )}
        </article>
      ))}
    </div>
  )
}

function EducationTable({ items }: { items: Education[] }) {
  const { t } = useI18n()
  if (items.length === 0) return <Empty />
  return (
    <div className="rows" role="list">
      <div className="rows-head is-edu" aria-hidden="true">
        <span>{t.table.period}</span>
        <span>{t.table.degree}</span>
        <span>{t.table.institution}</span>
      </div>
      {items.map((x) => (
        <article className="row is-edu" role="listitem" key={x.id}>
          <Period start={x.startDate} end={x.endDate} />
          <h3 className="row-title">{[x.degree, x.field].filter(Boolean).join(', ')}</h3>
          <span className="row-org">{x.institution}</span>
          {x.description && (
            <div className="row-body">
              <p>{x.description}</p>
            </div>
          )}
        </article>
      ))}
    </div>
  )
}

function Links({ cv }: { cv: Resource<CV> }) {
  const { t } = useI18n()
  const profile = cv.status === 'ready' ? cv.data.profile : null
  const email = profile?.email ?? site.email
  const rows: { label: string; value: ReactNode }[] = [
    { label: t.links.email, value: <a href={`mailto:${email}`}>{email}</a> },
  ]
  if (profile?.githubUrl) rows.push({ label: t.links.github, value: <a href={profile.githubUrl}>{profile.githubUrl.replace(/^https?:\/\//, '')}</a> })
  if (profile?.linkedinUrl) rows.push({ label: t.links.linkedin, value: <a href={profile.linkedinUrl}>{profile.linkedinUrl.replace(/^https?:\/\/(www\.)?/, '')}</a> })
  if (profile?.location) rows.push({ label: t.links.location, value: profile.location })
  if (site.cvPdf) rows.push({ label: t.links.cv, value: <a href={site.cvPdf} download>{t.links.downloadCv}</a> })

  return (
    <div className={`links${site.photo ? ' has-photo' : ''}`}>
      {site.photo && (
        <img className="portrait" src={site.photo} alt={t.links.photoAlt(profile?.fullName ?? site.name)} width={240} height={300} />
      )}
      <table className="spec">
        <tbody>
          {rows.map((r) => (
            <tr key={r.label}>
              <th scope="row">{r.label}</th>
              <td>{r.value}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export function Home() {
  const { t, lang, setLang, locale } = useI18n()
  const apiStatus = useApiStatus()
  const cv = useResource<CV>('/cv')
  const projects = useResource<Project[]>('/projects')

  const profile = cv.status === 'ready' ? cv.data.profile : null
  const name = profile?.fullName || site.name
  const email = profile?.email ?? site.email
  const failed = [cv, projects].flatMap((r) => (r.status === 'error' ? [r] : []))
  const rev = profile ? new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeZone: 'UTC' }).format(new Date(profile.updatedAt)) : null

  return (
    <>
      <header className="band" id="top">
        <span className="band-mark" aria-hidden="true">
          AJ
        </span>
        <div className="band-id">
          <span className="band-title">{t.datasheet}</span>
          <span className="band-sub">
            adnanjasarevic.com
            {rev && (
              <span className="band-rev">
                {' · '}
                {t.rev} {rev}
              </span>
            )}
          </span>
        </div>
        <button
          type="button"
          className="band-lang"
          lang={lang === 'en' ? 'bs' : 'en'}
          aria-label={t.langSwitchLabel}
          onClick={() => setLang(lang === 'en' ? 'bs' : 'en')}
        >
          {t.langSwitch}
        </button>
      </header>
      <nav className="band-nav" aria-label={t.navLabel}>
        <ol>
          {SECTIONS.map((id, i) => (
            <li key={id}>
              <a href={`#${id}`}>
                <span className="nav-num">{i + 1}</span>
                {sectionTitle(t.nav, id)}
              </a>
            </li>
          ))}
          <li>
            <a href="#contact" className="nav-contact">
              {t.nav.contact}
            </a>
          </li>
        </ol>
      </nav>

      <main className="sheet">
        <div className="masthead">
          <div className="masthead-text">
            <h1>{name}</h1>
            {profile?.headline && <p className="headline">{profile.headline}</p>}
            {cv.status === 'loading' && <span className="ghost-row ghost-headline" aria-hidden="true" />}
            {failed.length > 0 && (
              <div className="load-error" role="alert">
                <p>{t.loadFailed}</p>
                <button type="button" className="btn btn-quiet" onClick={() => failed.forEach((r) => r.retry())}>
                  {t.retry}
                </button>
              </div>
            )}
          </div>

          <section id="contact" className="contact-panel" aria-labelledby="contact-heading">
            <h2 id="contact-heading" className="contact-panel-head">
              {t.sections.contact}
            </h2>
            <div className="contact-panel-body">
              <p className="contact-intro">{t.contact.intro}</p>
              <ContactForm apiStatus={apiStatus} email={email} />
            </div>
          </section>
          {profile?.summary && <p className="summary">{profile.summary}</p>}
        </div>

        <div className="lead-grid">
          <Section id="features">
            <Loadable resource={cv} rows={4}>
              {(data) => <Features skills={data.skills} t={t} />}
            </Loadable>
          </Section>

          <Section id="system">
            <SystemFigure status={apiStatus} />
          </Section>
        </div>

        <Section id="projects">
          <Loadable resource={projects} rows={3}>
            {(data) => <Projects projects={data} />}
          </Loadable>
        </Section>

        <Section id="experience">
          <Loadable resource={cv} rows={3}>
            {(data) => <ExperienceTable items={data.experience} />}
          </Loadable>
        </Section>

        <Section id="education">
          <Loadable resource={cv} rows={2}>
            {(data) => <EducationTable items={data.education} />}
          </Loadable>
        </Section>

        <Section id="links">
          <Links cv={cv} />
        </Section>
      </main>

      <footer className="sheet-foot">
        <span className="foot-id">
          <b>AJ</b> · {t.datasheet}
          {rev && ` · ${t.rev} ${rev}`} · © {YEAR} {name}
        </span>
        <a href={REPO_URL}>{t.footer.source}</a>
        <a href="#top">{t.footer.top}</a>
      </footer>
    </>
  )
}
