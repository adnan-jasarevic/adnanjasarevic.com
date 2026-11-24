import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@fontsource-variable/archivo/wdth.css'
import '@fontsource-variable/archivo/wdth-italic.css'
import './index.css'
import { I18nProvider } from './i18n'
import { Home } from './home/Home'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <I18nProvider>
      <Home />
    </I18nProvider>
  </StrictMode>,
)
