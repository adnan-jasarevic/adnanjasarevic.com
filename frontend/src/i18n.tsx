import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'

export type Lang = 'en' | 'bs'

const en = {
  datasheet: 'Datasheet',
  rev: 'Rev',
  langSwitch: 'Bosanski',
  langSwitchLabel: 'Bosanski: prikaži stranicu na bosanskom',
  navLabel: 'Sections',
  nav: {
    features: 'Features',
    system: 'System',
    contact: 'Contact',
    projects: 'Projects',
    experience: 'Experience',
    education: 'Education',
    links: 'Links',
    linksCv: 'Links & CV',
  },
  sections: {
    features: 'Features',
    system: 'Typical application',
    contact: 'Contact',
    projects: 'Projects',
    experience: 'Experience',
    education: 'Education',
    links: 'Links',
    linksCv: 'Links & CV',
  },
  loading: 'Loading from the API. After a quiet spell the server needs up to a minute to wake.',
  loadFailed: "Couldn't load the projects and CV from the API.",
  notLoaded: 'Not loaded.',
  retry: 'Try again',
  noEntries: 'No entries yet.',
  skillsOther: 'Other',
  figure: {
    browser: 'Browser',
    browserSub: 'React · Cloudflare Pages',
    api: 'Go API',
    apiSub: 'Render · Frankfurt',
    db: 'PostgreSQL',
    dbSub: 'Neon · Frankfurt',
    mail: 'Resend',
    mailSub: 'Contact notifications',
    caption: 'Figure 1. The system serving this page. The API status is measured live from your browser.',
    checking: 'Checking…',
    waking: 'Waking up…',
    up: (ms: number) => `Up · ${ms.toLocaleString('en')} ms`,
    woke: (s: number) => `Up · woke in ${s} s`,
    down: 'Unreachable',
  },
  contact: {
    intro: 'Hiring, a project, or a question about the code: write here and the message lands in my inbox.',
    name: 'Name',
    email: 'Email',
    message: 'Message',
    send: 'Send message',
    sending: 'Sending…',
    sendingCold: 'Sending. The server is waking up, which can take up to a minute.',
    sent: (email: string) => `Message sent. I'll reply to ${email}.`,
    sendAnother: 'Send another message',
    required: 'Required.',
    tooLong: (max: number) => `Keep it under ${max.toLocaleString('en')} characters.`,
    badEmail: 'Enter an email address like name@example.com.',
    rateLimited: (minutes: number) =>
      `Too many messages from your connection. Try again in ${minutes} ${minutes === 1 ? 'minute' : 'minutes'}, or email me directly.`,
    failed: "The message couldn't be sent. Email me directly instead:",
    orEmail: 'Or email',
    counter: (n: number, max: number) => `${n.toLocaleString('en')} / ${max.toLocaleString('en')}`,
  },
  projects: {
    tech: 'Technology',
    source: 'Source',
    live: 'Live',
    viewSource: 'View on GitHub',
    viewLive: 'Open site',
    none: '—',
  },
  table: {
    period: 'Period',
    role: 'Role',
    company: 'Company',
    location: 'Location',
    degree: 'Degree',
    institution: 'Institution',
    present: 'Present',
  },
  links: {
    email: 'Email',
    github: 'GitHub',
    linkedin: 'LinkedIn',
    location: 'Based in',
    cv: 'CV (PDF)',
    downloadCv: 'Download CV',
    photoAlt: (name: string) => `Portrait of ${name}`,
  },
  footer: {
    source: 'Source code',
    top: 'Back to top',
  },
}

export type Strings = typeof en

const bs: Strings = {
  datasheet: 'Tehnički list',
  rev: 'Rev',
  langSwitch: 'English',
  langSwitchLabel: 'English: show the page in English',
  navLabel: 'Sekcije',
  nav: {
    features: 'Vještine',
    system: 'Sistem',
    contact: 'Kontakt',
    projects: 'Projekti',
    experience: 'Iskustvo',
    education: 'Obrazovanje',
    links: 'Linkovi',
    linksCv: 'Linkovi i CV',
  },
  sections: {
    features: 'Vještine',
    system: 'Tipična primjena',
    contact: 'Kontakt',
    projects: 'Projekti',
    experience: 'Radno iskustvo',
    education: 'Obrazovanje',
    links: 'Linkovi',
    linksCv: 'Linkovi i CV',
  },
  loading: 'Učitavanje s API-ja. Nakon neaktivnosti serveru treba do jedne minute da se pokrene.',
  loadFailed: 'Projekti i CV se nisu mogli učitati s API-ja.',
  notLoaded: 'Nije učitano.',
  retry: 'Pokušaj ponovo',
  noEntries: 'Još nema unosa.',
  skillsOther: 'Ostalo',
  figure: {
    browser: 'Preglednik',
    browserSub: 'React · Cloudflare Pages',
    api: 'Go API',
    apiSub: 'Render · Frankfurt',
    db: 'PostgreSQL',
    dbSub: 'Neon · Frankfurt',
    mail: 'Resend',
    mailSub: 'Obavještenja o porukama',
    caption: 'Slika 1. Sistem koji služi ovu stranicu. Status API-ja se mjeri uživo iz vašeg preglednika.',
    checking: 'Provjera…',
    waking: 'Pokretanje…',
    up: (ms: number) => `Radi · ${ms.toLocaleString('bs')} ms`,
    woke: (s: number) => `Radi · pokrenut za ${s} s`,
    down: 'Nedostupan',
  },
  contact: {
    intro: 'Posao, projekat ili pitanje o kodu: napišite ovdje i poruka stiže direktno u moj inbox.',
    name: 'Ime',
    email: 'Email',
    message: 'Poruka',
    send: 'Pošalji poruku',
    sending: 'Slanje…',
    sendingCold: 'Slanje. Server se pokreće, što može potrajati do jedne minute.',
    sent: (email: string) => `Poruka je poslana. Odgovorit ću na ${email}.`,
    sendAnother: 'Pošalji novu poruku',
    required: 'Obavezno polje.',
    tooLong: (max: number) => `Najviše ${max.toLocaleString('bs')} znakova.`,
    badEmail: 'Unesite email adresu, npr. ime@primjer.com.',
    rateLimited: (minutes: number) =>
      `Previše poruka s vaše konekcije. Pokušajte ponovo za ${minutes} min ili mi pišite direktno.`,
    failed: 'Poruka nije poslana. Pišite mi direktno na:',
    orEmail: 'Ili pišite na',
    counter: (n: number, max: number) => `${n.toLocaleString('bs')} / ${max.toLocaleString('bs')}`,
  },
  projects: {
    tech: 'Tehnologije',
    source: 'Kod',
    live: 'Uživo',
    viewSource: 'Pogledaj na GitHubu',
    viewLive: 'Otvori stranicu',
    none: '—',
  },
  table: {
    period: 'Period',
    role: 'Pozicija',
    company: 'Firma',
    location: 'Lokacija',
    degree: 'Zvanje',
    institution: 'Ustanova',
    present: 'Danas',
  },
  links: {
    email: 'Email',
    github: 'GitHub',
    linkedin: 'LinkedIn',
    location: 'Lokacija',
    cv: 'CV (PDF)',
    downloadCv: 'Preuzmi CV',
    photoAlt: (name: string) => `Portret: ${name}`,
  },
  footer: {
    source: 'Izvorni kod',
    top: 'Na vrh',
  },
}

const dictionaries: Record<Lang, Strings> = { en, bs }

function initialLang(): Lang {
  try {
    const saved = localStorage.getItem('lang')
    if (saved === 'en' || saved === 'bs') return saved
  } catch {
    // storage blocked: fall through to the browser language
  }
  return /^(bs|hr|sr)\b/i.test(navigator.language) ? 'bs' : 'en'
}

type I18n = { lang: Lang; t: Strings; locale: string; setLang: (lang: Lang) => void }

const I18nContext = createContext<I18n | null>(null)

export function I18nProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(initialLang)

  useEffect(() => {
    document.documentElement.lang = lang
  }, [lang])

  const setLang = (next: Lang) => {
    setLangState(next)
    try {
      localStorage.setItem('lang', next)
    } catch {
      // not persisted; the choice still applies to this visit
    }
  }

  const value = { lang, t: dictionaries[lang], locale: lang === 'bs' ? 'bs-BA' : 'en-GB', setLang }
  return <I18nContext value={value}>{children}</I18nContext>
}

export function useI18n() {
  const ctx = useContext(I18nContext)
  if (!ctx) throw new Error('useI18n outside I18nProvider')
  return ctx
}
