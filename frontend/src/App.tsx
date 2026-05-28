import React, { Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { RequireAuth } from '@/components/auth/RequireAuth'
import { Login } from '@/pages/Login'
import { Callback } from '@/pages/Callback'

// Lazy loaded employee pages
const Dashboard = React.lazy(
	() => import('@/pages/Dashboard').then((m) => ({ default: m.Dashboard })),
)
const Catalog = React.lazy(
	() => import('@/pages/Catalog').then((m) => ({ default: m.Catalog })),
)
const Points = React.lazy(
	() => import('@/pages/Points').then((m) => ({ default: m.Points })),
)
const Documents = React.lazy(
	() => import('@/pages/Documents').then((m) => ({ default: m.Documents })),
)
const Support = React.lazy(
	() => import('@/pages/Support').then((m) => ({ default: m.Support })),
)

// Lazy loaded admin pages
const AdminHR = React.lazy(
	() => import('@/pages/AdminHR').then((m) => ({ default: m.AdminHR })),
)
const AdminCatalog = React.lazy(
	() => import('@/pages/AdminCatalog').then((m) => ({ default: m.AdminCatalog })),
)
const AdminContent = React.lazy(
	() => import('@/pages/AdminContent').then((m) => ({ default: m.AdminContent })),
)

// Lazy loaded Shell
const Shell = React.lazy(
	() => import('@/components/layout/Shell').then((m) => ({ default: m.Shell })),
)

function Forbidden() {
	return (
		<div
			style={{
				display: 'flex',
				alignItems: 'center',
				justifyContent: 'center',
				minHeight: '100vh',
			}}
		>
			<div>
				<h1>403 — Доступ запрещён</h1>
				<p>У вас нет прав для доступа к этой странице.</p>
			</div>
		</div>
	)
}

const LoadingFallback = (
	<div
		style={{
			display: 'flex',
			alignItems: 'center',
			justifyContent: 'center',
			minHeight: '100vh',
		}}
	>
		Загрузка...
	</div>
)

export function App() {
	return (
		<BrowserRouter>
			<Suspense fallback={LoadingFallback}>
				<Routes>
					{/* Auth routes — без защиты */}
					<Route path="/login" element={<Login />} />
					<Route path="/callback" element={<Callback />} />
					<Route path="/forbidden" element={<Forbidden />} />

					{/* Employee routes — через Shell */}
					<Route element={<RequireAuth roles={['employee', 'catalog_manager', 'admin', 'hr']} />}>
						<Route element={<Shell />}>
							<Route path="/" element={<Dashboard />} />
							<Route path="/catalog" element={<Catalog />} />
							<Route path="/points" element={<Points />} />
							<Route path="/documents" element={<Documents />} />
							<Route path="/support" element={<Support />} />
						</Route>
					</Route>

					{/* Admin routes — через Shell */}
					<Route element={<RequireAuth roles={['hr', 'catalog_manager', 'admin']} />}>
						<Route element={<Shell />}>
							<Route path="/admin/hr" element={<AdminHR />} />
							<Route path="/admin/catalog" element={<AdminCatalog />} />
							<Route path="/admin/content" element={<AdminContent />} />
						</Route>
					</Route>

					{/* Catch-all — redirect to home */}
					<Route path="*" element={<Navigate to="/" replace />} />
				</Routes>
			</Suspense>
		</BrowserRouter>
	)
}
