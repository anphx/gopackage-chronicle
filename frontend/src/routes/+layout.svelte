<script lang="ts">
	import '../app.css';
	import type { Snippet } from 'svelte';
	import SiteHeader from '$lib/components/SiteHeader.svelte';

	let { children }: { children: Snippet } = $props();
</script>

<div class="shell">
	<SiteHeader />

	<main>
		{@render children()}
	</main>

	<footer class="footer">
		<span>go/chronicle · release history of Go modules, indexed daily from index.golang.org</span>
		<span>Built with Go, PostgreSQL and SvelteKit · <a href="/about">About</a></span>
	</footer>
</div>

<style>
	.shell {
		max-width: 1280px;
		margin: 0 auto;
		padding: 24px 24px 48px;
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	main {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.footer {
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		gap: 12px;
		padding: 8px 8px 0;
		font-size: 13px;
		color: var(--text-subtle);
	}

	@media (max-width: 640px) {
		.shell {
			padding: 12px 12px 32px;
		}
	}
</style>
