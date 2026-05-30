import '@testing-library/jest-dom'

// Polyfill window.matchMedia (required by Mantine in jsdom)
Object.defineProperty(window, 'matchMedia', {
	writable: true,
	value: function (query: string) {
		return {
			matches: false,
			media: query,
			onchange: null,
			addEventListener: function () {},
			removeEventListener: function () {},
			dispatchEvent: function () {
				return false
			},
		}
	},
	enumerable: true,
})

// Polyfill window.scrollTo (used by some Mantine components)
window.scrollTo = function () {}
