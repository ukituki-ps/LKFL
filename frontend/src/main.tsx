import React from 'react'
import ReactDOM from 'react-dom/client'
import { App } from './App'
import { AprilProviders } from './lib/providers'
import { setupAuthForTest } from './stores/authStore'
import '@mantine/core/styles.css'

// April tokens CSS (подключает CSS variables)
// import '@ukituki-ps/april-tokens/dist/index.css'

// Экспортируем setupAuthForTest для E2E тестов (Playwright page.evaluate)
if (typeof window !== 'undefined') {
	;(window as any).__LKFL_AUTH_STORE__ = { setupAuthForTest }
}

ReactDOM.createRoot(document.getElementById('root')!).render(
	<React.StrictMode>
		<AprilProviders>
			<App />
		</AprilProviders>
	</React.StrictMode>,
)
