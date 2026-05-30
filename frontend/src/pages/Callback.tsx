import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuthStore, type UserProfile, type UserRole } from '@/stores/authStore'
import { Title, Text, Container, Stack } from '@mantine/core'

interface CallbackResponse {
	user: UserProfile
	roles: UserRole[]
	/** Session token — возвращается в body, но хранится в httpOnly cookie,
	 * не в localStorage. Поле сохраняется для обратной совместимости API. */
	token?: string
}

const ATTEMPTS_KEY = 'lkfl_login_attempts'
const MAX_ATTEMPTS = 2

/** Получает realm tenant из env; fallback на 'lkfl' если не задано. */
function getRealm(): string {
	return import.meta.env.VITE_KEYCLOAK_REALM ?? 'lkfl'
}

export function Callback() {
	const navigate = useNavigate()
	const [searchParams] = useSearchParams()
	const setAuth = useAuthStore((state) => state.setAuth)
	const [error, setError] = useState<string | null>(null)

	useEffect(() => {
		const code = searchParams.get('code')
		const state = searchParams.get('state')

		if (!code) {
			setError('Ошибка авторизации: отсутствует код')
			setTimeout(() => navigate('/login', { replace: true }), 3000)
			return
		}

		// Счётчик попыток — защита от бесконечного цикла login → 401 → login
		const attemptCount = parseInt(sessionStorage.getItem(ATTEMPTS_KEY) || '0', 10) + 1
		sessionStorage.setItem(ATTEMPTS_KEY, String(attemptCount))

		// Handle logout — использует realm из env (не hard-coded).
		const doLogout = () => {
			const currentOrigin = window.location.origin
			const realm = getRealm()
			window.location.href =
				`${currentOrigin}/realms/${realm}/protocol/openid-connect/logout` +
				`?post_logout_redirect_uri=${encodeURIComponent(currentOrigin + '/')}`
		}

		// Exchange authorization code for user+roles via backend callback endpoint.
		// Backend validates state, verifies token, creates/updates user, creates session.
		// Session token устанавливается в httpOnly cookie (Set-Cookie header).
		fetch(`/api/v1/auth/callback?code=${code}&state=${state || ''}`, {
			headers: { 'Accept': 'application/json' },
			credentials: 'include', // ← D2: отправляем/получаем cookies
		})
			.then(async (res) => {
				if (!res.ok) {
					// 410 Gone — session expired, backend уже обработал gracefully
					if (res.status === 410) {
						if (attemptCount >= MAX_ATTEMPTS) {
							// Разрыв цикла — logout из Keycloak и очистка
							sessionStorage.removeItem(ATTEMPTS_KEY)
							doLogout()
							return
						}
						// Повторная попытка — форсируем перелогин в Keycloak
						setTimeout(() => {
							navigate('/login?retry=1', { replace: true })
						}, 1000)
						setError('Сессия устарела, повторная попытка входа...')
						return
					}
					throw new Error(`Ошибка авторизации: ${res.status}`)
				}
				// Успех — сброс счётчика
				sessionStorage.removeItem(ATTEMPTS_KEY)
				const data: CallbackResponse = await res.json()

				// D2: token берётся из response body (для store), но реальная
				// сессия хранится в httpOnly cookie, установленной backend.
				// Фронтенд хранит только user + roles в state (zustand).
				// Token в store нужен только для авторизации API-запросов
				// в переходный период; конечная цель — cookies только.
				const token = data.token

				if (!token) {
					// D11: без токена — ошибка, не используем fallback-строку
					setError('Ошибка авторизации: токен не получен от сервера')
					setTimeout(() => navigate('/login', { replace: true }), 3000)
					return
				}

				setAuth(token, data.user, data.roles ?? [])
				navigate('/', { replace: true })
			})
			.catch((err) => {
				setError(err.message)
				if (attemptCount >= MAX_ATTEMPTS) {
					sessionStorage.removeItem(ATTEMPTS_KEY)
					doLogout()
					return
				}
				setTimeout(() => navigate('/login', { replace: true }), 3000)
			})
	}, [searchParams, navigate, setAuth])

	if (error) {
		return (
			<Container size="sm" px={0}>
				<Stack align="center" justify="center" min-h="80vh" gap="xl">
					<Title order={2} c="red">Ошибка входа</Title>
					<Text c="dimmed">{error}</Text>
					<Text c="dimmed">Перенаправление на страницу входа...</Text>
				</Stack>
			</Container>
		)
	}

	return (
		<Container size="sm" px={0}>
			<Stack align="center" justify="center" min-h="80vh" gap="xl">
				<Title order={2}>Вход выполнен</Title>
				<Text c="dimmed">Перенаправление...</Text>
			</Stack>
		</Container>
	)
}
