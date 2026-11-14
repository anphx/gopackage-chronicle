<script lang="ts">
	import { formatDate } from '$lib/format';
	import { KIND_LABEL, RELEASE_KINDS, type ReleaseKind } from '$lib/semver';
	import { placeLabels, type HistoryEntry } from '$lib/stats';

	let { history, now = Date.now() }: { history: HistoryEntry[]; now?: number } = $props();

	const DAY = 86_400_000;
	type Range = '3M' | '6M' | '1Y' | 'All';
	const RANGES: { id: Range; days: number | null; label: string }[] = [
		{ id: '3M', days: 91, label: 'Last 3 months' },
		{ id: '6M', days: 182, label: 'Last 6 months' },
		{ id: '1Y', days: 365, label: 'Last 12 months' },
		{ id: 'All', days: null, label: 'All time' }
	];
	const LABEL_PX = 92; // room one version label needs on its side of a lane
	const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

	const firstTime = $derived(history.length ? history[0].time : now);
	// Default to the last year for long histories, otherwise show everything.
	let range = $state<Range | null>(null);
	const activeRange = $derived<Range>(range ?? (now - firstTime > 400 * DAY ? '1Y' : 'All'));
	let plotWidth = $state(900);

	const domain = $derived.by(() => {
		const days = RANGES.find((r) => r.id === activeRange)?.days ?? null;
		const begin = days === null ? firstTime - (now - firstTime) * 0.02 : now - days * DAY;
		const span = Math.max(now - begin, DAY);
		// leave a little room after "today" so its marker isn't on the edge
		return { start: begin, end: now + span * 0.03 };
	});
	const pos = (t: number) => ((t - domain.start) / (domain.end - domain.start)) * 100;

	const visible = $derived(history.filter((h) => h.time >= domain.start && h.time <= domain.end));

	const lanes = $derived.by(() => {
		const minGap = (LABEL_PX / Math.max(plotWidth, 1)) * 100;
		return RELEASE_KINDS.map((kind) => {
			const entries = visible.filter((h) => h.kind === kind).map((h) => ({ ...h, left: pos(h.time) }));
			const sides = kind === 'patch' ? entries.map(() => null) : placeLabels(entries.map((e) => e.left), minGap);
			return {
				kind,
				dots: entries.map((e, i) => ({ ...e, side: sides[i], tip: `${e.version} · ${formatDate(e.time)}` }))
			};
		});
	});

	const ticks = $derived.by(() => {
		const spanDays = (domain.end - domain.start) / DAY;
		const out: { left: number; label: string }[] = [];
		const d = new Date(domain.start);
		if (spanDays > 760) {
			for (let y = d.getFullYear() + 1; y <= new Date(domain.end).getFullYear(); y++) {
				out.push({ left: pos(new Date(y, 0, 1).getTime()), label: String(y) });
			}
		} else {
			const step = spanDays > 400 ? 2 : 1;
			let m = new Date(d.getFullYear(), d.getMonth() + 1, 1);
			while (m.getTime() < domain.end) {
				const label = m.getMonth() === 0 ? `${MONTHS[0]} ${m.getFullYear()}` : MONTHS[m.getMonth()];
				out.push({ left: pos(m.getTime()), label });
				m = new Date(m.getFullYear(), m.getMonth() + step, 1);
			}
		}
		return out;
	});

	const todayLeft = $derived(pos(now));
	const counts = $derived(Object.fromEntries(RELEASE_KINDS.map((k) => [k, visible.filter((h) => h.kind === k).length])) as Record<ReleaseKind, number>);
	const summary = $derived(
		`${visible.length} releases, ${RANGES.find((r) => r.id === activeRange)?.label.toLowerCase()}: ` +
			RELEASE_KINDS.map((k) => `${counts[k]} ${KIND_LABEL[k].toLowerCase()}`).join(', ')
	);
</script>

<section class="card timeline">
	<div class="head">
		<div class="titles">
			<h2 class="section-title">Release timeline</h2>
			<span class="section-sub">Every version on one axis, one lane per kind of change</span>
		</div>
		<div class="segmented" role="group" aria-label="Time range">
			{#each RANGES as r (r.id)}
				<button type="button" class:on={activeRange === r.id} aria-pressed={activeRange === r.id} aria-label={r.label} onclick={() => (range = r.id)}>
					{r.id}
				</button>
			{/each}
		</div>
	</div>

	<div class="scroll">
		<div class="chart" role="img" aria-label={summary}>
			{#each lanes as lane (lane.kind)}
				<div class="lane">
					<div class="lane-label"><span class="swatch {lane.kind}"></span>{KIND_LABEL[lane.kind]}</div>
					<div class="track" bind:clientWidth={plotWidth}>
						{#each ticks as t (t.left)}<span class="gridline" style:left="{t.left}%"></span>{/each}
						<span class="axis-line"></span>
						{#each lane.dots as d (d.version)}
							<span class="point" style:left="{d.left}%" title={d.tip}>
								<span class="dot {lane.kind}"></span>
								{#if d.side}<span class="label {d.side}">{d.version}</span>{/if}
							</span>
						{/each}
					</div>
				</div>
			{/each}

			<div class="ticks">
				<div class="lane-label"></div>
				<div class="tick-track">
					{#each ticks as t (t.left)}<span style:left="{t.left}%">{t.label}</span>{/each}
					<span class="today" style:left="{todayLeft}%">Today</span>
				</div>
			</div>

			<div class="today-line" style:left="calc(128px + (100% - 128px) * {todayLeft / 100})" aria-hidden="true"></div>
		</div>
	</div>

	{#if visible.length === 0}
		<p class="empty">No releases in this range. Try a longer one.</p>
	{/if}
</section>

<style>
	.timeline {
		display: flex;
		flex-direction: column;
		gap: 24px;
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

	.segmented {
		display: flex;
		gap: 2px;
		padding: 4px;
		border-radius: 999px;
		background: var(--surface-strong);
	}

	.segmented button {
		height: 40px;
		min-width: 52px;
		padding: 0 14px;
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

	.scroll {
		position: relative;
		overflow-x: auto;
	}

	.chart {
		position: relative;
		min-width: 760px;
		padding-top: 6px;
	}

	.lane {
		display: flex;
		height: 76px;
		border-bottom: 1px solid var(--line);
	}

	.lane-label {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 128px;
		flex-shrink: 0;
		font-size: 14px;
		font-weight: 600;
	}

	.swatch {
		width: 12px;
		height: 12px;
		border-radius: 4px;
	}

	.swatch.major {
		background: var(--kind-major);
	}

	.swatch.minor {
		background: var(--kind-minor);
	}

	.swatch.patch {
		background: var(--kind-patch);
	}

	.swatch.prerelease {
		background: var(--kind-prerelease);
	}

	.track {
		position: relative;
		flex: 1 1 auto;
	}

	.gridline {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 1px;
		background: var(--line);
	}

	.axis-line {
		position: absolute;
		left: 0;
		right: 0;
		top: 50%;
		height: 2px;
		background: var(--surface-strong);
	}

	.point {
		position: absolute;
		top: 50%;
		transform: translate(-50%, -50%);
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.point:hover {
		z-index: 2;
	}

	.point:hover .dot {
		outline: 2px solid var(--ink);
		outline-offset: 2px;
	}

	.dot {
		display: block;
		border-radius: 50%;
	}

	.dot.major {
		width: 26px;
		height: 26px;
		background: var(--kind-major);
		box-shadow: 0 0 0 5px rgba(255, 255, 255, 0.7);
	}

	.dot.minor {
		width: 18px;
		height: 18px;
		background: var(--kind-minor);
		box-shadow: 0 0 0 4px rgba(255, 255, 255, 0.7);
	}

	.dot.patch {
		width: 14px;
		height: 14px;
		background: var(--kind-patch);
		box-shadow: 0 0 0 4px rgba(255, 255, 255, 0.7);
	}

	.dot.prerelease {
		width: 12px;
		height: 12px;
		background: var(--white);
		border: 3px solid #6a6a6a;
		box-sizing: content-box;
	}

	.label {
		position: absolute;
		max-width: 88px;
		overflow: hidden;
		text-overflow: ellipsis;
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 600;
		white-space: nowrap;
	}

	.label.above {
		bottom: calc(100% + 6px);
	}

	.label.below {
		top: calc(100% + 6px);
	}

	.ticks {
		display: flex;
		padding-top: 10px;
		height: 56px;
	}

	.tick-track {
		position: relative;
		flex: 1 1 auto;
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-faint);
	}

	.tick-track span {
		position: absolute;
		white-space: nowrap;
	}

	.tick-track .today {
		top: 20px;
		transform: translateX(-50%);
		padding: 3px 8px;
		border-radius: 8px;
		background: var(--ember);
		color: var(--ink);
		font-weight: 600;
	}

	.today-line {
		position: absolute;
		top: 0;
		bottom: 56px;
		width: 0;
		border-left: 2px dashed var(--ember);
		pointer-events: none;
	}

	.empty {
		margin: 0;
		font-size: 14px;
		color: var(--text-subtle);
	}
</style>
