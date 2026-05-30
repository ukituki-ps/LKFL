import { useEffect } from 'react'
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom'
import { Title, Text, Container, Stack } from '@mantine/core'

export function Login() {
	const navigate = useNavigate()
	const location = useLocation()
	const [searchParams] = useSearchParams()

	useEffect(() => {
		// В test mode пропускаем redirect — E2E тесты проверяют UI напрямую
		if (import.meta.env.VITE_TEST_MODE === 'true') {
			return
		}

		// Очистка счётчика попыток при старте login
		sessionStorage.removeItem('lkfl_login_attempts')

		// Получаем attempted URL из router state (для redirect после логина)
		const from =
			(location.state as { from?: { pathname?: string } } | null)?.from?.pathname || '/'

		// Редирект на backend login endpoint
		// Backend генерирует state, сохраняет в Redis и делает 302 на Keycloak
		const retry = searchParams.get('retry')
		window.location.href = `/api/v1/auth/login?post_redirect=${encodeURIComponent(from)}&retry=${retry || '0'}`
	}, [navigate, location, searchParams])

	return (
		<Container size="sm" px={0}>
			<Stack align="center" justify="center" min-h="80vh" gap="xl">
				<Title order={2}>Вход в ЛКФЛ</Title>
				<Text c="dimmed">Перенаправление на страницу входа...</Text>
			</Stack>
		</Container>
	)
}
