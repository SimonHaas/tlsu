#!/usr/bin/env node

// npx ts-node test.ts
// npx vite-node src/utils/test.ts 

import { appToUrl } from './misc.ts'

const TARGET_PORT = 3700

function runTest() {
	const app = { port: TARGET_PORT, name: 'test-app', hiddenService: 'hidden.onion', path: '/' }
	try {
		const url = appToUrl(app as any)
		console.log('appToUrl =>', url)
	} catch (e) {
		console.error('calling appToUrl failed:', e)
	}
}

runTest()
