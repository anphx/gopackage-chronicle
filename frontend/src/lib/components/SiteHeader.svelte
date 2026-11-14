<script lang="ts">
	import { page } from '$app/state';
	import SearchBar from '$lib/components/SearchBar.svelte';

	const onHome = $derived(page.url.pathname === '/');
	const onPackage = $derived(page.url.pathname.startsWith('/packages/'));
</script>

<header class="header">
	<a href="/" class="brand" aria-label="go/chronicle home">
		<svg width="32" height="32" viewBox="0 0 32 32" aria-hidden="true">
			<circle cx="16" cy="16" r="16" fill="#D2622A" />
			<path d="M6 14c3 0 4 6 7 6s3-8 6-8 3 6 7 6" fill="none" stroke="#FFFFFF" stroke-width="3.2" stroke-linecap="round" />
		</svg>
		<span>go/chronicle</span>
	</a>

	<nav aria-label="Main">
		<a href="/" class="nav-link" class:active={onHome} aria-current={onHome ? 'page' : undefined}>Releases</a>
		<a href="/#modules" class="nav-link" class:active={onPackage}>Modules</a>
		<a href="#about" class="nav-link">About</a>
	</nav>

	<div class="spacer"></div>

	<SearchBar />
</header>

<style>
	.header {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 12px;
		min-height: 64px;
		padding: 8px 8px 8px 20px;
		border-radius: 999px;
		background: var(--ink);
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 10px;
		color: var(--white);
		text-decoration: none;
		font-size: 20px;
		font-weight: 700;
		letter-spacing: -0.02em;
	}

	.brand:hover {
		color: var(--white);
	}

	nav {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 4px;
		margin-left: 12px;
	}

	.nav-link {
		padding: 12px 18px;
		border-radius: 999px;
		color: var(--fog);
		text-decoration: none;
		font-size: 15px;
		font-weight: 500;
	}

	.nav-link:hover {
		color: var(--white);
	}

	.nav-link.active {
		background: var(--ember);
		color: var(--ink);
		font-weight: 600;
	}

	.spacer {
		flex: 1 1 0;
		min-width: 8px;
	}

	@media (max-width: 720px) {
		.header {
			border-radius: var(--radius-lg);
			padding: 12px;
		}

		nav {
			margin-left: 0;
		}

		.spacer {
			display: none;
		}
	}
</style>
