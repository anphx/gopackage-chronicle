<script lang="ts">
	import { onMount } from 'svelte';
	import { getStats, type StatsOverview } from '$lib/api/client';
	import { formatCompact, formatNumber, timeAgo } from '$lib/format';
	import type { ReleaseKind } from '$lib/semver';
	import AboutHero from '$lib/components/AboutHero.svelte';
	import SemverBadge from '$lib/components/SemverBadge.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import StatTile from '$lib/components/StatTile.svelte';

	const REPO_URL = 'https://github.com/anphx/gopackage-chronicle';

	let stats = $state<StatsOverview | null>(null);
	let statsFailed = $state(false);

	onMount(async () => {
		try {
			// Only the totals and last sync time are needed, so ask for the smallest window.
			stats = await getStats(1);
		} catch {
			statsFailed = true;
		}
	});

	type Tone = 'fog' | 'ink' | 'surface' | 'ember';

	const steps: { tone: Tone; title: string; body: string; tag: string; arrow: string }[] = [
		{
			tone: 'fog',
			title: 'The Go module index',
			body: 'When anyone fetches a new version through proxy.golang.org, it lands in the public feed at index.golang.org.',
			tag: 'source of truth',
			arrow: '↗'
		},
		{
			tone: 'ink',
			title: 'The indexer',
			body: 'A small Go program reads the feed every hour. It remembers where it stopped, so each run picks up only what is new, and running it twice never adds a duplicate.',
			tag: 'hourly · GitHub Actions',
			arrow: '↗'
		},
		{
			tone: 'surface',
			title: 'The archive',
			body: "Every module gets one row, every version another, with the time the mirror first saw it. That's all the history the charts need.",
			tag: 'PostgreSQL',
			arrow: '↗'
		},
		{
			tone: 'ember',
			title: 'This site',
			body: 'A read-only Go API serves the numbers; this page draws them. Nothing is cached in your browser beyond the page you’re on.',
			tag: 'Go API · SvelteKit',
			arrow: '↙'
		}
	];

	const kinds: { kind: ReleaseKind; example: string; meaning: string }[] = [
		{ kind: 'major', example: 'v2.0.0', meaning: 'A new major line. In Go this usually means a new import path, like /v2.' },
		{ kind: 'minor', example: 'v1.4.0', meaning: 'New features that shouldn’t break existing code.' },
		{ kind: 'patch', example: 'v1.4.2', meaning: 'Fixes. Anything that isn’t a .0 release.' },
		{
			kind: 'prerelease',
			example: 'v1.5.0-rc.1',
			meaning: 'Anything with a suffix: betas, release candidates and pseudo-versions. Shown hatched in charts.'
		}
	];

	const notes: { lead: string; rest: string }[] = [
		{
			lead: 'Times are when the mirror first saw a version,',
			rest: 'not when it was tagged. They’re usually minutes apart, but can be days for a version nobody fetched.'
		},
		{
			lead: 'Only public modules appear.',
			rest: 'Private modules and anything fetched with the proxy turned off never reach the index.'
		},
		{
			lead: 'It’s a hobby project.',
			rest: 'Numbers can lag behind for a while if a sync fails; the “Last sync” tile tells you how fresh they are.'
		}
	];

	const stack: { label: string; tone: Tone | 'sand' }[] = [
		{ label: 'Go', tone: 'ink' },
		{ label: 'PostgreSQL', tone: 'fog' },
		{ label: 'SvelteKit', tone: 'ember' },
		{ label: 'Bun', tone: 'sand' },
		{ label: 'Fly.io · Supabase · Cloudflare', tone: 'surface' }
	];

	const short = (value: number | undefined) => (value === undefined ? '—' : formatCompact(value));
	const exact = (value: number | undefined) => (value === undefined ? undefined : formatNumber(value));
</script>

<svelte:head>
	<title>About · go/chronicle</title>
	<meta
		name="description"
		content="What go/chronicle is, where its data comes from, and how to read the release colours."
	/>
</svelte:head>

<section class="intro">
	<div class="hero-slot"><AboutHero /></div>
	<div class="tiles">
		{#if stats || statsFailed}
			<StatTile
				tone="ember"
				label="Releases indexed"
				value={short(stats?.total_releases)}
				sub={exact(stats?.total_releases)}
			/>
			<StatTile
				tone="ink"
				label="Modules tracked"
				value={short(stats?.total_packages)}
				sub={exact(stats?.total_packages)}
			/>
		{:else}
			{#each Array(2) as _, i (i)}<Skeleton height="140px" radius="28px" />{/each}
		{/if}
		<StatTile tone="fog" label="Checks for new releases" value="Every hour" sub="via GitHub Actions" size="md" />
		{#if stats || statsFailed}
			<StatTile
				label="Last sync"
				value={stats?.last_synced_at ? timeAgo(stats.last_synced_at) : '—'}
				sub={statsFailed ? 'Couldn’t reach the API' : 'Live from the API'}
				size="md"
			/>
		{:else}
			<Skeleton height="140px" radius="28px" />
		{/if}
	</div>
</section>

<section class="card how" aria-labelledby="how-title">
	<div class="head">
		<h2 id="how-title" class="section-title">How it works</h2>
		<span class="section-sub">From a tag on a repository to a dot on a chart, in four steps.</span>
	</div>
	<ol class="steps">
		{#each steps as step, i (step.title)}
			<li class="step {step.tone}">
				<div class="step-top">
					<span class="num">{i + 1}</span>
					<span class="arrow" aria-hidden="true">{step.arrow}</span>
				</div>
				<h3>{step.title}</h3>
				<p>{step.body}</p>
				<span class="tag mono">{step.tag}</span>
			</li>
		{/each}
	</ol>
</section>

<section class="row">
	<div class="card colours" aria-labelledby="colours-title">
		<div class="head">
			<h2 id="colours-title" class="section-title">Reading the colours</h2>
			<span class="section-sub">
				Every chart, badge and legend uses the same four. The kind comes from the shape of the version number.
			</span>
		</div>
		<dl class="kinds">
			{#each kinds as k (k.kind)}
				<div class="kind-row">
					<dt><SemverBadge kind={k.kind} /><span class="example mono">{k.example}</span></dt>
					<dd>{k.meaning}</dd>
				</div>
			{/each}
		</dl>
		<span class="section-sub">A <code>+incompatible</code> suffix is ignored when deciding the kind.</span>
	</div>

	<div class="card-dark notes" aria-labelledby="notes-title">
		<h2 id="notes-title" class="section-title">Good to know</h2>
		<ul>
			{#each notes as note (note.lead)}
				<li>
					<span class="bullet" aria-hidden="true">↗</span>
					<span><strong>{note.lead}</strong> {note.rest}</span>
				</li>
			{/each}
		</ul>
	</div>
</section>

<section class="card built" aria-labelledby="built-title">
	<div class="built-text">
		<h2 id="built-title" class="section-title">Built with</h2>
		<ul class="stack">
			{#each stack as item (item.label)}
				<li class={item.tone}>{item.label}</li>
			{/each}
		</ul>
	</div>
	<a class="repo" href={REPO_URL} target="_blank" rel="noopener noreferrer">Source on GitHub <span aria-hidden="true">↗</span></a>
</section>

<style>
	.intro,
	.row {
		display: flex;
		flex-wrap: wrap;
		gap: 16px;
	}

	.hero-slot {
		flex: 999 1 560px;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}

	.hero-slot > :global(*) {
		flex: 1 1 auto;
	}

	.tiles {
		flex: 1 1 380px;
		min-width: 0;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 16px;
	}

	.head {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	/* How it works */
	.how {
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.steps {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		gap: 16px;
	}

	.step {
		display: flex;
		flex-direction: column;
		gap: 14px;
		min-height: 240px;
		padding: 24px;
		border-radius: var(--radius-lg);
	}

	.step-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.num {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: var(--ink);
		color: var(--white);
		font-size: 14px;
		font-weight: 600;
	}

	.arrow {
		font-size: 22px;
	}

	.step h3 {
		margin: 0;
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}

	.step p {
		margin: 0;
		font-size: 15px;
		line-height: 1.5;
	}

	.tag {
		margin-top: auto;
		align-self: flex-start;
		padding: 6px 10px;
		border-radius: 999px;
		font-size: 12px;
	}

	.step.fog {
		background: var(--fog);
	}

	.step.fog p {
		color: var(--graphite);
	}

	.step.fog .tag {
		background: rgba(255, 255, 255, 0.55);
	}

	.step.ink {
		background: var(--ink);
		color: var(--white);
	}

	.step.ink .num {
		background: var(--ember);
		color: var(--ink);
	}

	.step.ink .arrow,
	.step.ink .tag {
		color: var(--sand);
	}

	.step.ink p {
		color: var(--fog);
	}

	.step.ink .tag {
		background: var(--graphite);
	}

	.step.surface {
		background: var(--surface-strong);
		border: 1px solid var(--line);
	}

	.step.surface p {
		color: var(--text-muted);
	}

	.step.surface .tag {
		background: rgba(11, 11, 11, 0.06);
	}

	.step.ember {
		background: var(--ember);
	}

	.step.ember .tag {
		background: rgba(11, 11, 11, 0.12);
	}

	/* Reading the colours */
	.colours {
		flex: 999 1 620px;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.kinds {
		margin: 0;
		display: flex;
		flex-direction: column;
		border-bottom: 1px solid var(--line);
	}

	.kind-row {
		display: grid;
		grid-template-columns: 270px 1fr;
		align-items: center;
		gap: 16px;
		padding: 16px 0;
		border-top: 1px solid var(--line);
	}

	dt {
		display: grid;
		grid-template-columns: 120px 1fr;
		align-items: center;
		gap: 16px;
	}

	dt :global(.badge) {
		justify-self: start;
	}

	.example {
		font-size: 15px;
		font-weight: 600;
	}

	dd {
		margin: 0;
		font-size: 15px;
		line-height: 1.45;
		color: var(--text-muted);
	}

	/* Good to know */
	.notes {
		flex: 1 1 340px;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 18px;
	}

	.notes ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	.notes li {
		display: flex;
		gap: 12px;
		padding: 16px;
		border-radius: var(--radius-md);
		background: var(--graphite);
		font-size: 15px;
		line-height: 1.5;
		color: #e9e9e9;
	}

	.notes strong {
		color: var(--white);
	}

	.bullet {
		color: var(--sand);
		font-weight: 600;
	}

	/* Built with */
	.built {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 20px;
		padding: 28px 32px;
	}

	.built-text {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.stack {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.stack li {
		padding: 8px 14px;
		border-radius: 999px;
		font-size: 14px;
		font-weight: 600;
	}

	.stack .ink {
		background: var(--ink);
		color: var(--white);
	}

	.stack .fog {
		background: var(--fog);
	}

	.stack .ember {
		background: var(--ember);
	}

	.stack .surface {
		background: rgba(255, 255, 255, 0.7);
		border: 1px solid var(--line);
	}

	.stack .sand {
		background: var(--sand);
	}

	.repo {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		min-height: 48px;
		padding: 0 22px;
		border-radius: 999px;
		background: var(--ink);
		color: var(--white);
		text-decoration: none;
		font-size: 15px;
		font-weight: 600;
	}

	.repo:hover {
		background: var(--ember);
		color: var(--ink);
	}

	@media (max-width: 760px) {
		.kind-row {
			grid-template-columns: 1fr;
			gap: 8px;
		}
	}

	@media (max-width: 640px) {
		.tiles {
			gap: 10px;
		}

		.built {
			padding: 22px;
		}
	}
</style>
