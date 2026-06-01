import { create } from 'zustand'

const LS_THEME = 'lkfl_theme'
export type ThemeMode = 'light' | 'dark'

interface ThemeState {
	colorScheme: ThemeMode
	toggle: () => void
	set: (mode: ThemeMode) => void
}

// Восстановление из localStorage или системные предпочтения
function getInitial(): ThemeMode {
	const stored = localStorage.getItem(LS_THEME) as ThemeMode | null
	if (stored === 'light' || stored === 'dark') return stored
	if (typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches) {
		return 'dark'
	}
	return 'light'
}

export const useThemeStore = create<ThemeState>((set) => ({
	colorScheme: getInitial(),
	toggle: () =>
		set((s) => {
			const next = s.colorScheme === 'light' ? 'dark' : 'light'
			localStorage.setItem(LS_THEME, next)
			return { colorScheme: next }
		}),
	set: (mode) => {
		localStorage.setItem(LS_THEME, mode)
		set({ colorScheme: mode })
	},
}))
