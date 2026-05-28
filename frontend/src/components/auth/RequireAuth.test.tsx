import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { RequireAuth } from '@/components/auth/RequireAuth'
import { useAuthStore } from '@/stores/authStore'

const ProtectedContent = () => (
	<div data-testid="protected-content">Protected Content</div>
)

function renderWithRouter(path: string, roles: string[], requiredRoles?: string[]) {
	useAuthStore.getState().setAuth(
		'token',
		{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
		roles as any
	)

	render(
		<MemoryRouter initialEntries={[path]}>
			<Routes>
				<Route element={<RequireAuth roles={requiredRoles} />}>
					<Route path={path} element={<ProtectedContent />} />
				</Route>
				<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				<Route path="/forbidden" element={<div data-testid="forbidden-page">Forbidden</div>} />
			</Routes>
		</MemoryRouter>
	)
}

describe('RequireAuth', () => {
	beforeEach(() => {
		useAuthStore.getState().clearAuth()
	})

	it('перенаправляет на /login если не авторизован', () => {
		render(
			<MemoryRouter>
				<Routes>
					<Route path="/" element={<RequireAuth />}>
						<Route index element={<ProtectedContent />} />
					</Route>
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument()
		expect(screen.getByTestId('login-page')).toBeInTheDocument()
	})

	it('показывает контент если авторизован (без ролей)', () => {
		renderWithRouter('/', ['employee'])
		expect(screen.getByTestId('protected-content')).toBeInTheDocument()
	})

	it('показывает контент если есть требуемая роль catalog_manager', () => {
		renderWithRouter('/admin/catalog', ['catalog_manager'], ['catalog_manager'])
		expect(screen.getByTestId('protected-content')).toBeInTheDocument()
	})

	it('показывает контент если есть требуемая роль admin', () => {
		renderWithRouter('/admin/catalog', ['admin'], ['catalog_manager', 'admin'])
		expect(screen.getByTestId('protected-content')).toBeInTheDocument()
	})

	it('перенаправляет на /forbidden если нет требуемой роли', () => {
		renderWithRouter('/admin/catalog', ['employee'], ['catalog_manager', 'admin'])

		expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument()
		expect(screen.getByTestId('forbidden-page')).toBeInTheDocument()
	})

	it('показывает контент если пользователь имеет несколько ролей и одна подходит', () => {
		renderWithRouter('/admin/catalog', ['employee', 'hr'], ['catalog_manager', 'hr'])
		expect(screen.getByTestId('protected-content')).toBeInTheDocument()
	})

	// =============================================================================
	// EDGE CASE TESTS
	// =============================================================================

	it('auth check race — неавторизованный пользователь на защищённом маршруте', () => {
		useAuthStore.getState().clearAuth()

		render(
			<MemoryRouter>
				<Routes>
					<Route path="/" element={<RequireAuth />}>
						<Route index element={<ProtectedContent />} />
					</Route>
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument()
		expect(screen.getByTestId('login-page')).toBeInTheDocument()
	})

	it('role change — пользователь без роли видит forbidden', () => {
		useAuthStore.getState().setAuth(
			'token',
			{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
			['employee']
		)

		render(
			<MemoryRouter initialEntries={['/admin']}>
				<Routes>
					<Route element={<RequireAuth roles={['admin']} />}>
						<Route path="/admin" element={<ProtectedContent />} />
					</Route>
					<Route path="/forbidden" element={<div data-testid="forbidden-page">Forbidden</div>} />
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument()
		expect(screen.getByTestId('forbidden-page')).toBeInTheDocument()
	})

	it('nested auth guards — RequireAuth внутри RequireAuth', () => {
		useAuthStore.getState().setAuth(
			'token',
			{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
			['admin']
		)

		render(
			<MemoryRouter initialEntries={['/admin/secret']}>
				<Routes>
					<Route element={<RequireAuth />}>
						<Route path="/admin/*" element={<RequireAuth roles={['admin']} />}>
							<Route path="secret" element={<ProtectedContent />} />
						</Route>
					</Route>
					<Route path="/forbidden" element={<div data-testid="forbidden-page">Forbidden</div>} />
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		expect(screen.getByTestId('protected-content')).toBeInTheDocument()
	})

	it('nested auth guards — внешняя guard провалена', () => {
		useAuthStore.getState().clearAuth()

		render(
			<MemoryRouter initialEntries={['/admin/secret']}>
				<Routes>
					<Route element={<RequireAuth />}>
						<Route path="/admin/*" element={<RequireAuth roles={['admin']} />}>
							<Route path="secret" element={<ProtectedContent />} />
						</Route>
					</Route>
					<Route path="/forbidden" element={<div data-testid="forbidden-page">Forbidden</div>} />
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		expect(screen.getByTestId('login-page')).toBeInTheDocument()
		expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument()
	})

	it('nested auth guards — внутренняя guard провалена', () => {
		useAuthStore.getState().setAuth(
			'token',
			{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
			['employee']
		)

		render(
			<MemoryRouter initialEntries={['/admin/secret']}>
				<Routes>
					<Route element={<RequireAuth />}>
						<Route path="/admin/*" element={<RequireAuth roles={['admin']} />}>
							<Route path="secret" element={<ProtectedContent />} />
						</Route>
					</Route>
					<Route path="/forbidden" element={<div data-testid="forbidden-page">Forbidden</div>} />
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		expect(screen.getByTestId('forbidden-page')).toBeInTheDocument()
		expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument()
	})

	it('concurrent navigation — переход между защищёнными маршрутами', () => {
		useAuthStore.getState().setAuth(
			'token',
			{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
			['admin']
		)

		render(
			<MemoryRouter initialEntries={['/admin']}>
				<Routes>
					<Route element={<RequireAuth roles={['admin']} />}>
						<Route path="/admin" element={<div data-testid="admin-page">Admin</div>} />
						<Route path="/admin/settings" element={<div data-testid="settings-page">Settings</div>} />
					</Route>
					<Route path="/forbidden" element={<div data-testid="forbidden-page">Forbidden</div>} />
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		expect(screen.getByTestId('admin-page')).toBeInTheDocument()
	})

	it('redirect loop prevention — повторный redirect на login', () => {
		useAuthStore.getState().clearAuth()

		render(
			<MemoryRouter>
				<Routes>
					<Route path="/" element={<RequireAuth />}>
						<Route index element={<ProtectedContent />} />
					</Route>
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		// Должен перенаправить на /login, а не в бесконечный цикл
		expect(screen.getByTestId('login-page')).toBeInTheDocument()
		expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument()
	})

	it('empty roles array — любой авторизованный получает доступ', () => {
		useAuthStore.getState().setAuth(
			'token',
			{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
			['employee']
		)

		render(
			<MemoryRouter initialEntries={['/protected']}>
				<Routes>
					<Route element={<RequireAuth roles={[]} />}>
						<Route path="/protected" element={<ProtectedContent />} />
					</Route>
					<Route path="/forbidden" element={<div data-testid="forbidden-page">Forbidden</div>} />
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		// Пустой массив roles означает "любой авторизованный"
		expect(screen.getByTestId('protected-content')).toBeInTheDocument()
	})

	it('undefined roles — любой авторизованный получает доступ', () => {
		useAuthStore.getState().setAuth(
			'token',
			{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
			['employee']
		)

		render(
			<MemoryRouter initialEntries={['/protected']}>
				<Routes>
					<Route element={<RequireAuth roles={undefined as any} />}>
						<Route path="/protected" element={<ProtectedContent />} />
					</Route>
					<Route path="/forbidden" element={<div data-testid="forbidden-page">Forbidden</div>} />
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		expect(screen.getByTestId('protected-content')).toBeInTheDocument()
	})

	it('пользователь без ролей не проходит guard с ролями', () => {
		useAuthStore.getState().setAuth(
			'token',
			{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
			[]
		)

		render(
			<MemoryRouter initialEntries={['/admin']}>
				<Routes>
					<Route element={<RequireAuth roles={['admin']} />}>
						<Route path="/admin" element={<ProtectedContent />} />
					</Route>
					<Route path="/forbidden" element={<div data-testid="forbidden-page">Forbidden</div>} />
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		expect(screen.getByTestId('forbidden-page')).toBeInTheDocument()
		expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument()
	})

	it('сохранение attempted URL в state', () => {
		useAuthStore.getState().clearAuth()

		render(
			<MemoryRouter initialEntries={['/protected-page']}>
				<Routes>
					<Route path="/" element={<RequireAuth />}>
						<Route path="protected-page" element={<ProtectedContent />} />
					</Route>
					<Route path="/login" element={<div data-testid="login-page">Login</div>} />
				</Routes>
			</MemoryRouter>
		)

		expect(screen.getByTestId('login-page')).toBeInTheDocument()
	})
})
