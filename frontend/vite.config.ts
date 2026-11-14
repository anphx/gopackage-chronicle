import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ command, mode }) => {
	// The API URL is baked into the static build (build/_app/env.js). Without it
	// the client falls back to http://localhost:8080, which silently breaks the
	// deployed site, so refuse to build instead.
	if (command === 'build') {
		const env = loadEnv(mode, process.cwd(), 'PUBLIC_');
		if (!env.PUBLIC_API_BASE_URL) {
			throw new Error(
				'PUBLIC_API_BASE_URL is not set. Set it as a build variable (Cloudflare Workers Builds: ' +
					'Settings → Build → Variables and secrets), or in frontend/.env for local builds.'
			);
		}
	}

	return {
		plugins: [sveltekit()]
	};
});
