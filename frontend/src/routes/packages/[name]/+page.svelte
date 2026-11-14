<script lang="ts">
	import { replaceState } from '$app/navigation';
	import { page as appPage } from '$app/state';
	import { getPackageWithAllReleases, type PackageDetail } from '$lib/api/client';
	import { formatDate, formatDayMonth, formatNumber, moduleDisplayName, pluralize } from '$lib/format';
	import { DEFAULT_PAGE_SIZE, parsePage, parsePageSize } from '$lib/pagination';
	import { cadenceSentence, countByKind, median, toHistory } from '$lib/stats';
	import CadenceChart from '$lib/components/CadenceChart.svelte';
	import ErrorState from '$lib/components/ErrorState.svelte';
	import PackageHero from '$lib/components/PackageHero.svelte';
	import ReleaseTimeline from '$lib/components/ReleaseTimeline.svelte';
	import SemverMix from '$lib/components/SemverMix.svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import StatTile from '$lib/components/StatTile.svelte';
	import VersionsTable from '$lib/components/VersionsTable.svelte';

	let { data }: { data: { packagePath: string } } = $props();

	const DAY = 86_400_000;

	let detail = $state<PackageDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let notFound = $state(false);
	let requestId = 0;

	// Versions table paging lives in the URL (?page=, ?per_page=) so it survives reloads.
	let tablePage = $state(parsePage(appPage.url.searchParams.get('page')));
	let tableSize = $state(parsePageSize(appPage.url.searchParams.get('per_page')));

	const syncURL = () => {
		const url = new URL(appPage.url);
		if (tablePage === 1) url.searchParams.delete('page');
		else url.searchParams.set('page', String(tablePage));
		if (tableSize === DEFAULT_PAGE_SIZE) url.searchParams.delete('per_page');
		else url.searchParams.set('per_page', String(tableSize));
		replaceState(url, appPage.state);
	};

	const load = async (path: string) => {
		const id = ++requestId;
		loading = true;
		error = null;
		notFound = false;
		detail = null;
		try {
			const result = await getPackageWithAllReleases(path);
			if (id === requestId) detail = result;
		} catch (e) {
			if (id !== requestId) return;
			const message = e instanceof Error ? e.message : 'Failed to load the module';
			notFound = message.includes('404');
			error = message;
		} finally {
			if (id === requestId) loading = false;
		}
	};

	$effect(() => {
		load(data.packagePath);
	});

	const now = Date.now();
	const history = $derived(detail ? toHistory(detail.releases) : []);
	const latest = $derived(history.length ? history[history.length - 1] : null);
	const gaps = $derived(history.flatMap((h) => (h.gapDays === null ? [] : [{ days: h.gapDays, version: h.version }])));
	const typicalGap = $derived(median(gaps.map((g) => g.days)));
	const longest = $derived(gaps.reduce<(typeof gaps)[number] | null>((a, g) => (!a || g.days > a.days ? g : a), null));
	const shortest = $derived(gaps.reduce<(typeof gaps)[number] | null>((a, g) => (!a || g.days < a.days ? g : a), null));
	const recent90 = $derived(history.filter((h) => now - h.time <= 90 * DAY).length);
	const counts = $derived(countByKind(history));
	const display = $derived(moduleDisplayName(data.packagePath));
</script>

<svelte:head>
	<title>{display.name}{display.major ? ` ${display.major}` : ''} · go/chronicle</title>
	<meta name="description" content="Release history of {data.packagePath}" />
</svelte:head>

<nav class="breadcrumb" aria-label="Breadcrumb">
	<a href="/#modules">Modules</a>
	<span aria-hidden="true">/</span>
	<span aria-current="page" title={data.packagePath}>{display.name}{display.major ? ` ${display.major}` : ''}</span>
</nav>

{#if loading}
	<section class="intro">
		<div class="hero-slot"><Skeleton height="300px" radius="32px" /></div>
		<div class="tiles">{#each Array(4) as _, i (i)}<Skeleton height="140px" radius="28px" />{/each}</div>
	</section>
	<Skeleton height="120px" radius="32px" />
	<Skeleton height="400px" radius="32px" />
{:else if error}
	<section class="card">
		{#if notFound}
			<ErrorState
				title="Module not found"
				message="{data.packagePath} hasn't been indexed. Check the path, or come back after the next daily sync."
			/>
		{:else}
			<ErrorState title="Couldn't load this module" message={error} onretry={() => load(data.packagePath)} />
		{/if}
	</section>
{:else if detail}
	<section class="intro">
		<div class="hero-slot">
			<PackageHero
				path={detail.package.path}
				latest={latest?.version ?? null}
				since={history.length ? formatDate(history[0].time) : null}
			/>
		</div>
		<div class="tiles">
			<StatTile tone="ember" label="Latest" value={latest?.version ?? '—'} mono size="md" />
			<StatTile label="Releases" value={formatNumber(history.length)} />
			<StatTile tone="fog" label="Last release" value={latest ? formatDayMonth(latest.time) : '—'} sub={latest ? String(new Date(latest.time).getFullYear()) : undefined} size="md" />
			<StatTile label="Typical gap" value={typicalGap === null ? '—' : `${typicalGap}d`} sub="median days between releases" />
		</div>
	</section>

	{#if history.length === 0}
		<section class="card"><ErrorState title="No releases yet" message="This module is indexed but has no releases recorded." /></section>
	{:else}
		<SemverMix {counts} />

		<ReleaseTimeline {history} {now} />

		<section class="cadence-row">
			<div class="cadence-slot"><CadenceChart {history} /></div>
			<div class="glance card-dark">
				<h2 class="section-title">At a glance</h2>
				<dl>
					<div><dt>Longest wait</dt><dd>{longest ? `${pluralize(longest.days, 'day')} (${longest.version})` : '—'}</dd></div>
					<div><dt>Fastest follow-up</dt><dd>{shortest ? `${pluralize(shortest.days, 'day')} (${shortest.version})` : '—'}</dd></div>
					<div><dt>Releases in the last 90 days</dt><dd>{recent90}</dd></div>
					<div><dt>Pre-releases</dt><dd>{counts.prerelease}</dd></div>
				</dl>
				<p class="insight">↗ {cadenceSentence(history)}</p>
			</div>
		</section>

		<VersionsTable
			path={detail.package.path}
			{history}
			page={tablePage}
			size={tableSize}
			onpage={(p) => {
				tablePage = p;
				syncURL();
			}}
			onsize={(s) => {
				tableSize = s;
				tablePage = 1;
				syncURL();
			}}
		/>
	{/if}
{/if}

<style>
	.breadcrumb {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 0 8px;
		font-size: 14px;
		color: var(--text-subtle);
	}

	.breadcrumb a {
		color: var(--text-subtle);
	}

	.breadcrumb [aria-current] {
		color: var(--ink);
		font-weight: 600;
	}

	.intro,
	.cadence-row {
		display: flex;
		flex-wrap: wrap;
		gap: 16px;
	}

	.hero-slot {
		flex: 999 1 600px;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}

	.hero-slot > :global(*) {
		flex: 1 1 auto;
	}

	.tiles {
		flex: 1 1 360px;
		min-width: 0;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 16px;
	}

	.cadence-slot {
		flex: 999 1 640px;
		min-width: 0;
		display: flex;
	}

	.cadence-slot > :global(*) {
		flex: 1 1 auto;
	}

	.glance {
		flex: 1 1 300px;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	dl {
		display: flex;
		flex-direction: column;
		margin: 0;
	}

	dl div {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		padding: 14px 0;
		border-top: 1px solid var(--graphite-line);
	}

	dt {
		color: var(--fog);
	}

	dd {
		margin: 0;
		font-weight: 600;
		text-align: right;
	}

	.insight {
		margin: auto 0 0;
		padding: 18px;
		border-radius: 20px;
		background: var(--ember);
		color: var(--ink);
		font-size: 15px;
		font-weight: 500;
		line-height: 1.45;
	}
</style>
