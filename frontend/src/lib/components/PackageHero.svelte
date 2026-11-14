<script lang="ts">
	import { moduleDisplayName, pkgGoDevURL } from '$lib/format';

	let { path, latest, since }: { path: string; latest: string | null; since: string | null } = $props();

	const display = $derived(moduleDisplayName(path));
	const command = $derived(`go get ${path}${latest ? `@${latest}` : ''}`);
	let copied = $state(false);
	let resetTimer: ReturnType<typeof setTimeout> | undefined;

	const copy = async () => {
		try {
			await navigator.clipboard.writeText(command);
			copied = true;
			clearTimeout(resetTimer);
			resetTimer = setTimeout(() => (copied = false), 1800);
		} catch {
			copied = false;
		}
	};
</script>

<section class="hero card-dark">
	<svg class="curve" width="360" height="280" viewBox="0 0 360 280" aria-hidden="true">
		<path d="M20 230 C 80 230, 90 90, 160 90 S 240 220, 300 220 S 340 60, 340 0" fill="none" stroke="#D2622A" stroke-width="56" stroke-linecap="round" />
	</svg>

	{#if since}<span class="pill">Go module · first release indexed {since}</span>{/if}

	<h1 title={path}>
		{display.name}
		{#if display.major}<span class="major">{display.major}</span>{/if}
	</h1>

	<div class="actions">
		<code class="command">{command}</code>
		<button type="button" class="copy" onclick={copy} aria-label={copied ? 'Copied' : 'Copy install command'}>
			{#if copied}
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" aria-hidden="true"><path d="M5 12l5 5 9-10" /></svg>
			{:else}
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="12" height="12" rx="3" /><path d="M5 15V5a2 2 0 0 1 2-2h8" /></svg>
			{/if}
		</button>
		<span class="visually-hidden" aria-live="polite">{copied ? 'Install command copied' : ''}</span>
		<a class="docs" href={pkgGoDevURL(path, latest ?? undefined)} target="_blank" rel="noopener noreferrer">pkg.go.dev ↗</a>
	</div>
</section>

<style>
	.hero {
		position: relative;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		gap: 22px;
		min-width: 0;
		padding: 40px;
	}

	.curve {
		position: absolute;
		right: -30px;
		top: -40px;
	}

	.pill,
	h1,
	.actions {
		position: relative;
	}

	.pill {
		align-self: flex-start;
		padding: 8px 14px;
		border-radius: var(--radius-sm);
		background: var(--graphite);
		color: var(--sand);
		font-size: 14px;
		font-weight: 600;
	}

	h1 {
		display: flex;
		align-items: flex-start;
		gap: 14px;
		margin: 0;
		font-size: 96px;
		font-weight: 600;
		line-height: 0.9;
		letter-spacing: -0.05em;
		overflow-wrap: anywhere;
	}

	.major {
		margin-top: 10px;
		padding: 6px 10px;
		border-radius: 10px;
		background: var(--ember);
		color: var(--ink);
		font-family: var(--font-mono);
		font-size: 18px;
		font-weight: 600;
		letter-spacing: 0;
		line-height: 1.2;
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 10px;
	}

	.command {
		max-width: 100%;
		overflow-x: auto;
		padding: 14px 18px;
		border-radius: var(--radius-sm);
		background: var(--graphite);
		font-size: 14px;
		white-space: nowrap;
	}

	.copy {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 48px;
		height: 48px;
		border: 0;
		border-radius: var(--radius-sm);
		background: var(--ember);
		color: var(--ink);
		cursor: pointer;
	}

	.copy:hover {
		background: var(--sand);
	}

	.docs {
		padding: 12px 8px;
		color: var(--sand);
		font-size: 15px;
		font-weight: 600;
		text-decoration: none;
	}

	.docs:hover {
		color: var(--white);
	}

	@media (max-width: 640px) {
		.hero {
			padding: 28px 22px;
		}

		.curve {
			opacity: 0.35;
		}

		h1 {
			font-size: 64px;
		}
	}
</style>
