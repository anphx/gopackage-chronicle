<script lang="ts">
	import { onMount } from 'svelte';
	import { replaceState } from '$app/navigation';
	import { page as appPage } from '$app/state';
	import {
		getRecentPackages,
		getRecentReleases,
		getStats,
		type RecentPackagesResponse,
		type RecentPackagesSort,
		type ReleaseKindName,
		type ReleaseWithPackage,
		type StatsOverview
	} from '$lib/api/client';
	import { formatNumber, timeAgo } from '$lib/format';
	import { clampPage, DEFAULT_PAGE_SIZE, parsePage, parsePageSize } from '$lib/pagination';
	import { RELEASE_KINDS } from '$lib/semver';
	import { weekTotal } from '$lib/stats';
	import ActivityChart from '$lib/components/ActivityChart.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import HomeHero from '$lib/components/HomeHero.svelte';
	import LiveFeed from '$lib/components/LiveFeed.svelte';
	import ModuleList from '$lib/components/ModuleList.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import StatTile from '$lib/components/StatTile.svelte';

	const ACTIVITY_WEEKS = 30;
	const MODULE_WEEKS = 12;
	const SORTS: RecentPackagesSort[] = ['recent', 'active', 'name'];
	const errorMessage = (e: unknown, fallback: string) => (e instanceof Error ? e.message : fallback);

	// --- overview: tiles + activity chart ---
	let stats = $state<StatsOverview | null>(null);
	let statsError = $state<string | null>(null);

	const loadStats = async () => {
		statsError = null;
		try {
			stats = await getStats(ACTIVITY_WEEKS);
		} catch (e) {
			statsError = errorMessage(e, 'Failed to load statistics');
		}
	};

	// --- live feed ---
	let feed = $state<ReleaseWithPackage[]>([]);
	let feedLoading = $state(true);

	const loadFeed = async () => {
		feedLoading = true;
		try {
			feed = (await getRecentReleases(7, 0)).releases;
		} catch {
			feed = [];
		} finally {
			feedLoading = false;
		}
	};

	// --- recently updated modules (server-side filtering, sorting and pagination) ---
	let query = $state('');
	let kind = $state<ReleaseKindName | ''>('');
	let sort = $state<RecentPackagesSort>('recent');
	let pageNum = $state(1);
	let pageSize = $state(DEFAULT_PAGE_SIZE);

	let modules = $state<RecentPackagesResponse | null>(null);
	let modulesLoading = $state(true);
	let modulesError = $state<string | null>(null);
	let requestId = 0;
	let debounce: ReturnType<typeof setTimeout> | undefined;

	const syncURL = () => {
		const url = new URL(appPage.url);
		const set = (key: string, value: string, fallback: string) =>
			value === fallback ? url.searchParams.delete(key) : url.searchParams.set(key, value);
		set('q', query.trim(), '');
		set('kind', kind, '');
		set('sort', sort, 'recent');
		set('page', String(pageNum), '1');
		set('per_page', String(pageSize), String(DEFAULT_PAGE_SIZE));
		replaceState(url, appPage.state);
	};

	const loadModules = async () => {
		const id = ++requestId;
		modulesLoading = true;
		modulesError = null;
		try {
			const res = await getRecentPackages({
				limit: pageSize,
				offset: (pageNum - 1) * pageSize,
				weeks: MODULE_WEEKS,
				q: query,
				kind,
				sort
			});
			if (id !== requestId) return;
			// A page past the end (e.g. after the list shrank) snaps back to the last page.
			const last = clampPage(pageNum, res.total, pageSize);
			if (res.total > 0 && last !== pageNum) {
				pageNum = last;
				syncURL();
				return loadModules();
			}
			modules = res;
		} catch (e) {
			if (id === requestId) modulesError = errorMessage(e, 'Failed to load modules');
		} finally {
			if (id === requestId) modulesLoading = false;
		}
	};

	const update = (changes: () => void, { resetPage = true, delay = 0 } = {}) => {
		changes();
		if (resetPage) pageNum = 1;
		syncURL();
		clearTimeout(debounce);
		if (delay) debounce = setTimeout(loadModules, delay);
		else loadModules();
	};

	const scrollToModules = () => document.getElementById('modules')?.scrollIntoView({ behavior: 'smooth', block: 'start' });

	onMount(() => {
		const params = appPage.url.searchParams;
		query = params.get('q') ?? '';
		const k = params.get('kind') ?? '';
		kind = (RELEASE_KINDS as string[]).includes(k) ? (k as ReleaseKindName) : '';
		const s = params.get('sort') ?? '';
		sort = (SORTS as string[]).includes(s) ? (s as RecentPackagesSort) : 'recent';
		pageNum = parsePage(params.get('page'));
		pageSize = parsePageSize(params.get('per_page'));

		loadStats();
		loadFeed();
		loadModules();
		return () => clearTimeout(debounce);
	});

	const thisWeek = $derived(stats && stats.weeks.length ? weekTotal(stats.weeks[stats.weeks.length - 1]) : null);
</script>

<svelte:head>
	<title>go/chronicle · Every Go release, in motion</title>
	<meta name="description" content="Track the release history of every Go module, indexed daily from index.golang.org." />
</svelte:head>

<section class="intro">
	<div class="hero-slot"><HomeHero /></div>
	<div class="tiles">
		{#if stats}
			<StatTile tone="ember" label="Releases indexed" value={formatNumber(stats.total_releases)} />
			<StatTile tone="ink" label="Modules tracked" value={formatNumber(stats.total_packages)} />
			<StatTile tone="fog" label="This week" value={formatNumber(thisWeek ?? 0)} />
			<StatTile
				label="Last sync"
				value={stats.last_synced_at ? timeAgo(stats.last_synced_at) : '—'}
				sub="Daily at 20:00 UTC"
				size="md"
			/>
		{:else}
			{#each Array(4) as _, i (i)}<Skeleton height="140px" radius="28px" />{/each}
		{/if}
	</div>
</section>

<section class="activity-row">
	<div class="card activity">
		{#if statsError}
			<ErrorState title="Couldn't load release activity" message={statsError} onretry={loadStats} />
		{:else if stats}
			<ActivityChart weeks={stats.weeks} />
		{:else}
			<Skeleton height="32px" width="40%" />
			<Skeleton height="260px" radius="16px" />
			<Skeleton height="76px" radius="18px" />
		{/if}
	</div>
	<div class="feed-slot">
		<LiveFeed releases={feed} loading={feedLoading} />
	</div>
</section>

<ModuleList
	packages={modules?.packages ?? []}
	total={modules?.total ?? 0}
	kindCounts={modules?.kind_counts ?? null}
	weeks={MODULE_WEEKS}
	loading={modulesLoading}
	error={modulesError}
	{query}
	{kind}
	{sort}
	page={pageNum}
	size={pageSize}
	onquery={(q) => update(() => (query = q), { delay: 250 })}
	onkind={(k) => update(() => (kind = k))}
	onsort={(s) => update(() => (sort = s))}
	onpage={(p) => {
		update(() => (pageNum = p), { resetPage: false });
		scrollToModules();
	}}
	onsize={(n) => update(() => (pageSize = n))}
	onretry={loadModules}
/>

<style>
	.intro,
	.activity-row {
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

	.activity {
		flex: 999 1 620px;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.feed-slot {
		flex: 1 1 340px;
		min-width: 0;
		display: flex;
	}

	.feed-slot > :global(*) {
		flex: 1 1 auto;
	}

	@media (max-width: 640px) {
		.tiles {
			gap: 10px;
		}
	}
</style>
