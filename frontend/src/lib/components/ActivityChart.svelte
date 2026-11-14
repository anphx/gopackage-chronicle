<script lang="ts">
	import type { WeeklyActivity } from '$lib/api/client';
	import { formatDayMonth, formatNumber, monthLabel } from '$lib/format';
	import { KIND_LABEL, RELEASE_KINDS } from '$lib/semver';
	import { niceMax, summarizeActivity, weekTotal } from '$lib/stats';

	let { weeks }: { weeks: WeeklyActivity[] } = $props();

	const summary = $derived(summarizeActivity(weeks));
	const scale = $derived(niceMax(Math.max(0, ...weeks.map(weekTotal))));
	const pct = (n: number) => (n / scale) * 100;

	const bars = $derived(
		weeks.map((w, i) => {
			const month = monthLabel(w.week_start);
			const showMonth = i === 0 || monthLabel(weeks[i - 1].week_start) !== month;
			return {
				key: w.week_start,
				label: showMonth ? month : '',
				height: pct(weekTotal(w)),
				segments: [
					{ kind: 'prerelease', n: w.prerelease },
					{ kind: 'major', n: w.major },
					{ kind: 'minor', n: w.minor },
					{ kind: 'patch', n: w.patch }
				],
				tip: `Week of ${formatDayMonth(w.week_start)}: ${formatNumber(weekTotal(w))} releases — ${w.major} major, ${w.minor} minor, ${w.patch} patch, ${w.prerelease} pre-release`
			};
		})
	);

	const description = $derived(
		`Releases per week for the last ${weeks.length} weeks. Average ${formatNumber(summary.average)} a week` +
			(summary.busiest ? `, busiest week starting ${formatDayMonth(summary.busiest.weekStart)} with ${formatNumber(summary.busiest.total)}.` : '.')
	);
</script>

<div class="head">
	<div class="titles">
		<h2 class="section-title">Release activity</h2>
		<span class="section-sub">Releases per week, split by kind of change · last {weeks.length} weeks</span>
	</div>
	<ul class="legend" aria-label="Legend">
		{#each RELEASE_KINDS as kind (kind)}
			<li><span class="swatch {kind}"></span>{KIND_LABEL[kind]}</li>
		{/each}
	</ul>
</div>

<div class="chart" role="img" aria-label={description}>
	<div class="axis" aria-hidden="true">
		<span>{formatNumber(scale)}</span>
		<span>{formatNumber(scale / 2)}</span>
		<span>0</span>
	</div>
	<div class="plot-wrap">
		<div class="plot">
			{#each bars as bar (bar.key)}
				<div class="col" title={bar.tip}>
					<div class="stack" style:height="{bar.height}%">
						{#each bar.segments as seg (seg.kind)}
							<div class="seg {seg.kind}" style:flex-grow={seg.n}></div>
						{/each}
					</div>
				</div>
			{/each}
		</div>
		<div class="labels" aria-hidden="true">
			{#each bars as bar (bar.key)}<span>{bar.label}</span>{/each}
		</div>
	</div>
</div>

<div class="summary">
	<div>
		<span class="k">Busiest week</span>
		<span class="v">{summary.busiest ? `${formatDayMonth(summary.busiest.weekStart)} · ${formatNumber(summary.busiest.total)}` : '—'}</span>
	</div>
	<div>
		<span class="k">Weekly average</span>
		<span class="v">{formatNumber(summary.average)} releases</span>
	</div>
	<div>
		<span class="k">Patch share</span>
		<span class="v">{summary.patchShare}%</span>
	</div>
</div>

<style>
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

	.legend {
		display: flex;
		flex-wrap: wrap;
		gap: 14px;
		margin: 0;
		padding: 0;
		list-style: none;
		font-size: 13px;
		font-weight: 500;
	}

	.legend li {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.swatch {
		width: 12px;
		height: 12px;
		border-radius: 4px;
	}

	.chart {
		flex: 1 1 auto;
		min-height: 240px;
		display: flex;
		gap: 12px;
	}

	.axis {
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		width: 44px;
		flex-shrink: 0;
		padding-bottom: 27px;
		margin-top: -6px;
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-faint);
		text-align: right;
	}

	.plot-wrap {
		flex: 1 1 auto;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.plot {
		flex: 1 1 auto;
		display: flex;
		align-items: flex-end;
		gap: 5px;
		background:
			linear-gradient(rgba(11, 11, 11, 0.08), rgba(11, 11, 11, 0.08)) top / 100% 1px no-repeat,
			linear-gradient(rgba(11, 11, 11, 0.08), rgba(11, 11, 11, 0.08)) center / 100% 1px no-repeat;
		border-bottom: 1px solid rgba(11, 11, 11, 0.2);
	}

	.col {
		flex: 1 1 0;
		min-width: 0;
		height: 100%;
		display: flex;
		flex-direction: column;
		justify-content: flex-end;
	}

	.col:hover .stack {
		outline: 2px solid var(--ink);
		outline-offset: 1px;
	}

	.stack {
		display: flex;
		flex-direction: column;
		min-height: 0;
		border-radius: 7px 7px 2px 2px;
		overflow: hidden;
	}

	.labels {
		display: flex;
		gap: 5px;
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-faint);
	}

	.labels span {
		flex: 1 1 0;
		min-width: 0;
		white-space: nowrap;
	}

	.seg {
		flex-basis: 0;
		min-height: 0;
	}

	.major {
		background: var(--kind-major);
	}

	.minor {
		background: var(--kind-minor);
	}

	.patch {
		background: var(--kind-patch);
	}

	.prerelease {
		background: var(--kind-prerelease);
	}

	.summary {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 12px;
	}

	.summary > div {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 16px;
		border-radius: var(--radius-md);
		background: var(--surface-strong);
	}

	.k {
		font-size: 13px;
		color: var(--text-faint);
	}

	.v {
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}

	@media (max-width: 640px) {
		.plot,
		.labels {
			gap: 2px;
		}

		.summary {
			grid-template-columns: 1fr;
		}
	}
</style>
