<script lang="ts">
	import { formatDate, formatNumber, pkgGoDevURL, pluralize } from '$lib/format';
	import { clampPage } from '$lib/pagination';
	import type { HistoryEntry } from '$lib/stats';
	import Pagination from '$lib/components/Pagination.svelte';
	import SemverBadge from '$lib/components/SemverBadge.svelte';

	let {
		path,
		history,
		page,
		size,
		onpage,
		onsize
	}: {
		path: string;
		/** Oldest first, as returned by toHistory. */
		history: HistoryEntry[];
		page: number;
		size: number;
		onpage: (page: number) => void;
		onsize: (size: number) => void;
	} = $props();

	let section: HTMLElement | undefined = $state();

	const newestFirst = $derived([...history].reverse());
	const maxGap = $derived(Math.max(1, ...history.map((h) => h.gapDays ?? 0)));
	const current = $derived(clampPage(page, newestFirst.length, size));
	const rows = $derived(newestFirst.slice((current - 1) * size, current * size));
	const latestVersion = $derived(newestFirst[0]?.version);
	const firstYear = $derived(history.length ? new Date(history[0].time).getFullYear() : null);

	const goTo = (p: number) => {
		onpage(p);
		section?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	};
</script>

<section class="card versions" bind:this={section} aria-labelledby="versions-title">
	<div class="titles">
		<h2 id="versions-title" class="section-title">All versions</h2>
		<span class="section-sub">
			{formatNumber(history.length)} releases{firstYear ? ` since ${firstYear}` : ''} · newest first
		</span>
	</div>

	<div class="scroll">
		<table>
			<thead>
				<tr>
					<th scope="col">Version</th>
					<th scope="col">Change</th>
					<th scope="col">Released</th>
					<th scope="col" class="gap-col">Since previous</th>
					<th scope="col"><span class="visually-hidden">Open</span></th>
				</tr>
			</thead>
			<tbody>
				{#each rows as r (r.version)}
					<tr>
						<td class="version">
							{r.version}
							{#if r.version === latestVersion}<span class="latest">latest</span>{/if}
						</td>
						<td><SemverBadge kind={r.kind} /></td>
						<td class="date">{formatDate(r.time)}</td>
						<td>
							<span class="gap">
								<span class="track"><span style:width="{r.gapDays === null ? 0 : (r.gapDays / maxGap) * 100}%"></span></span>
								<span class="gap-text">{r.gapDays === null ? 'first seen' : pluralize(r.gapDays, 'day')}</span>
							</span>
						</td>
						<td class="open">
							<a href={pkgGoDevURL(path, r.version)} target="_blank" rel="noopener noreferrer" aria-label="Open {r.version} on pkg.go.dev">↗</a>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	<Pagination total={history.length} page={current} {size} noun="versions" label="Versions pagination" onpage={goTo} {onsize} />
</section>

<style>
	.versions {
		display: flex;
		flex-direction: column;
		gap: 16px;
		scroll-margin-top: 16px;
	}

	.titles {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.scroll {
		position: relative;
		overflow-x: auto;
	}

	table {
		width: 100%;
		min-width: 720px;
		border-collapse: separate;
		border-spacing: 0;
		font-size: 15px;
	}

	th {
		padding: 0 16px 12px;
		border-bottom: 1px solid rgba(11, 11, 11, 0.1);
		text-align: left;
		font-size: 12px;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--text-faint);
	}

	.gap-col {
		width: 34%;
	}

	td {
		padding: 12px 16px;
		border-bottom: 1px solid var(--line);
	}

	tbody tr:hover {
		background: var(--surface-hover);
	}

	.version {
		font-family: var(--font-mono);
		font-weight: 600;
		white-space: nowrap;
	}

	.latest {
		margin-left: 10px;
		padding: 3px 9px;
		border-radius: 999px;
		background: var(--ink);
		color: var(--sand);
		font-family: var(--font-sans);
		font-size: 12px;
		font-weight: 600;
	}

	.date {
		color: var(--text-muted);
		white-space: nowrap;
	}

	.gap {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.track {
		flex: 1 1 auto;
		height: 8px;
		border-radius: 4px;
		background: var(--surface-strong);
		overflow: hidden;
	}

	.track span {
		display: block;
		height: 8px;
		border-radius: 4px;
		background: var(--sand);
	}

	.gap-text {
		width: 84px;
		text-align: right;
		font-size: 14px;
		color: var(--text-muted);
	}

	.open {
		padding-top: 6px;
		padding-bottom: 6px;
		text-align: right;
	}

	.open a {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background: var(--surface-strong);
		text-decoration: none;
	}

	.open a:hover {
		background: var(--ember);
		color: var(--ink);
	}
</style>
