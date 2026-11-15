<script lang="ts">
	import type { KindCounts, PackageActivity, RecentPackagesSort, ReleaseKindName } from '$lib/api/client';
	import { formatNumber, packageHref, splitModulePath, timeAgo } from '$lib/format';
	import { KIND_LABEL, RELEASE_KINDS, type ReleaseKind } from '$lib/semver';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import SemverBadge from '$lib/components/SemverBadge.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import Sparkline from '$lib/components/Sparkline.svelte';

	let {
		packages,
		total,
		kindCounts,
		weeks,
		loading,
		error,
		query,
		kind,
		sort,
		page,
		size,
		onquery,
		onkind,
		onsort,
		onpage,
		onsize,
		onretry
	}: {
		packages: PackageActivity[];
		total: number;
		kindCounts: KindCounts | null;
		weeks: number;
		loading: boolean;
		error: string | null;
		query: string;
		kind: ReleaseKindName | '';
		sort: RecentPackagesSort;
		page: number;
		size: number;
		onquery: (q: string) => void;
		onkind: (kind: ReleaseKindName | '') => void;
		onsort: (sort: RecentPackagesSort) => void;
		onpage: (page: number) => void;
		onsize: (size: number) => void;
		onretry: () => void;
	} = $props();

	const filters = $derived<{ value: ReleaseKind | ''; label: string; count: number | null }[]>([
		{ value: '', label: 'All', count: kindCounts?.all ?? null },
		...RELEASE_KINDS.map((k) => ({ value: k, label: KIND_LABEL[k], count: kindCounts ? kindCounts[k] : null }))
	]);

	const sorts: { value: RecentPackagesSort; label: string }[] = [
		{ value: 'recent', label: 'Newest' },
		{ value: 'active', label: 'Most active' },
		{ value: 'name', label: 'A–Z' }
	];

	const rows = $derived(packages.map((p) => ({ ...p, ...splitModulePath(p.path) })));
	const filtered = $derived(query.trim() !== '' || kind !== '');
</script>

<section id="modules" class="card modules" aria-labelledby="modules-title">
	<div class="head">
		<div class="titles">
			<h2 id="modules-title" class="section-title">Recently updated modules</h2>
			<span class="section-sub">Modules with a release in the last {weeks} weeks · {formatNumber(total)} found</span>
		</div>
		<label class="filter">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
				<path d="M3 6h18M7 12h10M11 18h2" />
			</svg>
			<span class="visually-hidden">Filter modules by path</span>
			<input type="search" placeholder="Filter by module path" value={query} oninput={(e) => onquery(e.currentTarget.value)} />
		</label>
	</div>

	<div class="controls">
		<div class="chips" role="group" aria-label="Kind of latest release">
			{#each filters as f (f.value)}
				<button type="button" class="chip" class:on={kind === f.value} aria-pressed={kind === f.value} onclick={() => onkind(f.value)}>
					{f.label}
					{#if f.count !== null}<span class="n">{formatNumber(f.count)}</span>{/if}
				</button>
			{/each}
		</div>
		<div class="segmented" role="group" aria-label="Sort">
			{#each sorts as s (s.value)}
				<button type="button" class:on={sort === s.value} aria-pressed={sort === s.value} onclick={() => onsort(s.value)}>{s.label}</button>
			{/each}
		</div>
	</div>

	<div class="table-wrap">
		<div class="table" role="table" aria-label="Recently updated modules" aria-busy={loading}>
			<div class="row header" role="row">
				<span role="columnheader">Module</span>
				<span role="columnheader">Latest</span>
				<span role="columnheader">Last {weeks} weeks</span>
				<span role="columnheader" class="num">Releases</span>
				<span role="columnheader" class="num">Updated</span>
				<span role="columnheader"><span class="visually-hidden">Open</span></span>
			</div>

			{#if error}
				<ErrorState title="Couldn't load modules" message={error} {onretry} />
			{:else if loading && rows.length === 0}
				{#each Array(8) as _, i (i)}
					<div class="row skeleton-row" role="row"><Skeleton height="44px" radius="14px" /></div>
				{/each}
			{:else if rows.length === 0}
				<div class="empty">
					<span class="empty-title">{filtered ? 'No modules match' : 'No releases in this window yet'}</span>
					{#if filtered}
						<button type="button" class="clear" onclick={() => { onquery(''); onkind(''); }}>Clear filters</button>
					{/if}
				</div>
			{:else}
				<div class="rows" class:stale={loading}>
					{#each rows as p (p.package_id)}
						<div class="row" role="row">
							<div class="module" role="cell">
								<span class="initial" aria-hidden="true">{p.name.charAt(0).toUpperCase()}</span>
								<span class="names">
									<a class="path" href={packageHref(p.path)}>
										<span class="prefix">{p.prefix}</span><span class="name">{p.name}</span>
									</a>
									<span class="versions">
										{#each p.recent_versions.slice(1) as v (v)}<span title={v}>{v}</span>{/each}
									</span>
								</span>
							</div>
							<div class="latest" role="cell">
								<span class="version">{p.latest_version}</span>
								<SemverBadge kind={p.latest_kind} />
							</div>
							<div role="cell">
								<Sparkline values={p.weekly} label="Releases per week over the last {weeks} weeks: {p.weekly.join(', ')}" />
							</div>
							<span class="num count" role="cell">{formatNumber(p.release_count)}</span>
							<span class="num ago" role="cell">{timeAgo(p.last_released_at)}</span>
							<span role="cell">
								<a class="go" href={packageHref(p.path)} aria-label="Open {p.path}">↗</a>
							</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>

	{#if !error && total > 0}
		<Pagination {total} {page} {size} noun="modules" label="Modules pagination" {onpage} {onsize} />
	{/if}
</section>

<style>
	.modules {
		display: flex;
		flex-direction: column;
		gap: 20px;
		scroll-margin-top: 16px;
	}

	.head {
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		align-items: flex-end;
		gap: 16px;
	}

	.titles {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.filter {
		display: flex;
		align-items: center;
		gap: 8px;
		height: 44px;
		min-width: 260px;
		padding: 0 16px;
		border-radius: 999px;
		border: 1px solid var(--line-strong);
		background: rgba(255, 255, 255, 0.5);
		color: var(--text-faint);
	}

	.filter:focus-within {
		border-color: var(--ink);
	}

	.filter input {
		width: 100%;
		border: 0;
		outline: none;
		background: transparent;
		color: var(--ink);
		font-size: 14px;
	}

	.controls {
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		gap: 12px;
	}

	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.chip {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		height: 44px;
		padding: 0 16px;
		border-radius: 999px;
		border: 1px solid rgba(11, 11, 11, 0.25);
		background: transparent;
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
	}

	.chip:hover {
		border-color: var(--ink);
	}

	.chip.on {
		background: var(--ink);
		border-color: var(--ink);
		color: var(--white);
	}

	.chip .n {
		font-family: var(--font-mono);
		font-size: 12px;
		opacity: 0.75;
	}

	.segmented {
		display: flex;
		gap: 2px;
		padding: 4px;
		border-radius: 999px;
		background: var(--surface-strong);
	}

	.segmented button {
		height: 40px;
		padding: 0 16px;
		border: 0;
		border-radius: 999px;
		background: transparent;
		color: var(--text-subtle);
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
	}

	.segmented button.on {
		background: var(--ember);
		color: var(--ink);
	}

	.table-wrap {
		position: relative;
		overflow-x: auto;
		margin: 0 -8px;
	}

	.table {
		min-width: 900px;
		padding: 0 8px;
	}

	.row {
		display: grid;
		grid-template-columns: minmax(0, 2.6fr) minmax(0, 1.3fr) 150px 90px 110px 48px;
		gap: 16px;
		align-items: center;
		padding: 16px;
		border-bottom: 1px solid var(--line);
		border-radius: 16px;
	}

	.rows .row {
		transition: background 0.15s ease;
	}

	.rows .row:hover {
		background: var(--surface-hover);
	}

	.rows.stale {
		opacity: 0.55;
		transition: opacity 0.15s ease;
	}

	.header {
		padding-top: 0;
		padding-bottom: 12px;
		border-bottom: 1px solid rgba(11, 11, 11, 0.1);
		border-radius: 0;
		font-size: 12px;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--text-faint);
	}

	.skeleton-row {
		display: block;
	}

	.num {
		text-align: right;
	}

	.module {
		display: flex;
		align-items: center;
		gap: 14px;
		min-width: 0;
	}

	.initial {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 44px;
		height: 44px;
		flex-shrink: 0;
		border-radius: var(--radius-sm);
		background: var(--ink);
		color: var(--sand);
		font-size: 18px;
		font-weight: 700;
	}

	.names {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
	}

	.path {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		text-decoration: none;
	}

	.prefix {
		font-size: 14px;
		color: var(--text-faint);
	}

	.name {
		font-size: 17px;
		font-weight: 600;
		letter-spacing: -0.01em;
	}

	.path:hover .name {
		color: var(--ember);
	}

	.versions {
		display: flex;
		gap: 6px;
		min-width: 0;
		overflow: hidden;
	}

	.versions span {
		flex-shrink: 0;
		max-width: 140px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		padding: 3px 7px;
		border-radius: 6px;
		background: var(--surface-strong);
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-subtle);
	}

	.latest {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}

	.version {
		max-width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-family: var(--font-mono);
		font-size: 15px;
		font-weight: 600;
	}

	.count {
		font-size: 17px;
		font-weight: 600;
	}

	.ago {
		font-size: 14px;
		color: var(--text-subtle);
	}

	.go {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 44px;
		height: 44px;
		border-radius: 50%;
		background: var(--ink);
		color: var(--white);
		text-decoration: none;
		font-size: 18px;
	}

	.go:hover {
		background: var(--ember);
		color: var(--ink);
	}

	.empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		padding: 56px 16px;
	}

	.empty-title {
		font-size: 18px;
		font-weight: 600;
	}

	.clear {
		height: 44px;
		padding: 0 20px;
		border: 0;
		border-radius: 999px;
		background: var(--ink);
		color: var(--white);
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
	}

	@media (max-width: 640px) {
		.filter {
			min-width: 0;
			width: 100%;
		}
	}
</style>
