import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		// Fully static single-page app. Cloudflare Pages serves index.html for any
		// path that has no matching file (because there is no top-level 404.html),
		// so client-side routes like /packages/[name] work on direct visits.
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: 'index.html'
		})
	}
};

export default config;
