<script lang="ts">
	import { shortVersion } from '$lib/semver';
	import { median, type HistoryEntry } from '$lib/stats';

	let { history, limit = 24 }: { history: HistoryEntry[]; limit?: number } = $props();

	// The most recent releases, each with its gap to the previous one.
	const bars = $derived(history.filter((h) => h.gapDays !== null).slice(-limit));
	const allGaps = $derived(history.flatMap((h) => (h.gapDays === null ? [] : [h.gapDays])));
	const typical = $derived(median(allGaps));
	const max = $derived(Math.max(1, ...bars.map((b) => b.gapDays ?? 0)));
	const pct = (days: number) => Math.max(1.5, (days / max) * 100);
	// Label every bar when there is room (about 56px per label), otherwise every n-th one.
	let chartWidth = $state(800);
	const labelEvery = $derived(Math.max(1, Math.ceil((bars.length * 56) / Math.max(chartWidth, 1))));
	const description = $derived(
		`Days between releases for the last ${bars.length} releases. Median gap ${typical ?? 0} days.`
	);
</script>

<section class="card cadence">
	<div class="titles">
		<h2 class="section-title">Release cadence</h2>
		<span class="section-sub">
			Days since the previous version · {bars.length < allGaps.length ? `last ${bars.length} releases · ` : ''}dashed line = median ({typical ?? 0} days)
		</span>
	</div>

	{#if bars.length === 0}
		<p class="empty">Cadence appears once a module has two or more releases.</p>
	{:else}
		<div class="scroll">
			<div class="chart" role="img" aria-label={description} bind:clientWidth={chartWidth}>
				<div class="plot">
					{#if typical !== null}
						<span class="median-layer" aria-hidden="true">
							<span class="median" style:bottom="{Math.min(100, (typical / max) * 100)}%"></span>
						</span>
					{/if}
					{#each bars as b (b.version)}
						<div class="col" title="{b.version}: {b.gapDays} days after the previous release">
							<span class="value">{b.gapDays}</span>
							<span class="bar {b.kind}" style:height="{pct(b.gapDays ?? 0)}%"></span>
						</div>
					{/each}
				</div>
				<div class="labels" aria-hidden="true">
					{#each bars as b, i (b.version)}<span title={b.version}>{(bars.length - 1 - i) % labelEvery === 0 ? shortVersion(b.version) : ''}</span>{/each}
				</div>
			</div>
		</div>
	{/if}
</section>

<style>
	.cadence {
		display: flex;
		flex-direction: column;
		gap: 24px;
		min-width: 0;
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

	.chart {
		min-width: 560px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.plot {
		position: relative;
		height: 220px;
		display: flex;
		align-items: flex-end;
		gap: 8px;
		padding-top: 20px;
		border-bottom: 2px solid var(--ink);
	}

	.median-layer {
		position: absolute;
		left: 0;
		right: 0;
		top: 20px;
		bottom: 0;
		pointer-events: none;
	}

	.median {
		position: absolute;
		left: 0;
		right: 0;
		border-top: 2px dashed #8f8f8f;
		pointer-events: none;
	}

	.col {
		position: relative;
		flex: 1 1 0;
		min-width: 0;
		height: 100%;
		display: flex;
		flex-direction: column;
		justify-content: flex-end;
		align-items: center;
		gap: 6px;
	}

	.value {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-muted);
	}

	.bar {
		width: 100%;
		border-radius: 8px 8px 0 0;
	}

	.col:hover .bar {
		outline: 2px solid var(--ink);
		outline-offset: 1px;
	}

	.bar.major {
		background: var(--kind-major);
	}

	.bar.minor {
		background: var(--kind-minor);
	}

	.bar.patch {
		background: var(--kind-patch);
	}

	.bar.prerelease {
		background: repeating-linear-gradient(45deg, #c2c2c2 0 4px, #e9e9e9 4px 8px);
	}

	.labels {
		display: flex;
		gap: 8px;
		font-family: var(--font-mono);
		font-size: 10px;
		color: var(--text-subtle);
	}

	.labels span {
		flex: 1 1 0;
		min-width: 0;
		overflow: visible;
		display: flex;
		justify-content: center;
		white-space: nowrap;
	}

	.empty {
		margin: 0;
		font-size: 14px;
		color: var(--text-subtle);
	}
</style>
