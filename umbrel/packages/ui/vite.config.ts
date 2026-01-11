import path from 'node:path'
import react from '@vitejs/plugin-react-swc'
import {defineConfig} from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
	server: {
		headers: {
			"Content-Security-Policy": [
				"default-src 'self'",
				"script-src 'self' 'unsafe-eval' 'unsafe-inline'",
				"style-src 'self' 'unsafe-inline'",
				"img-src 'self' data:",
				"font-src 'self'",
				"connect-src 'self' https://traefik.umbrel-dev.local http://10.21.0.2:8080 https://traefik.umbrel.simonhaas.eu",
			].join("; ")
    	}
	},
	plugins: [react()],
	resolve: {
		alias: {
			'@/': `${path.resolve(__dirname, 'src')}/`,
		},
	},
	build: {
		rollupOptions: {
			output: {
				minifyInternalExports: true,
				manualChunks: {
					// remeda: ['remeda'],
					// motion: ['framer-motion'],
					// bignumber: ['bignumber.js'],
					// other: ['react-helmet-async', 'react-error-boundary'],
					// toaster: ['sonner'],
					react: ['react', 'react-dom'],
					i18n: ['i18next', 'react-i18next', 'i18next-browser-languagedetector', 'i18next-http-backend'],
					fetch: ['@tanstack/react-query', '@trpc/react-query', '@trpc/client'],
					css: ['tailwind-merge', 'clsx'],
					reactRouter: ['react-router-dom'],
					dev: ['@tanstack/react-query-devtools', 'react-json-tree'],
					// sorter: ['match-sorter'],
					// icons: ['react-icons', 'lucide-react'],
					// qr: ['react-qr-code'],
					// pin: ['rci'],
					colorThief: ['colorthief'],
				},
			},
			external: ['@radix-ui/react-label', '@radix-ui/react-radio-group'],
		},
	},
})
