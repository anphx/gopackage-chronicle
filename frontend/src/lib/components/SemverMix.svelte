<script lang="ts">
	import { formatNumber } from '$lib/format';
	import { KIND_LABEL, RELEASE_KINDS, type ReleaseKind } from '$lib/semver';

	let { counts }: { counts: Record<ReleaseKind, number> } = $props();

	const total = $derived(RELEASE_KINDS.reduce((sum, k) => sum + counts[k], 0));
	const parts = $derived(RELEASE_KINDS.filter((k) => counts[k] > 0).map((k) => ({ kind: k, n: counts[k] })));
	const description = $derived(parts.map((p) => `${formatNumber(p.n)} ${KIND_LABEL[p.kind].toLowerCase()}`).join(', '));
</script>

<section class="card mix">
	<div class="head">
		<h2 class="title">What kind of releases</h2>
		<span class="section-sub">Share of all {formatNumber(total)} versions by kind of change</span>
	</div>
	<div class="bar" role="img" aria-label="Releases by kind: {description}">
		{#each parts as p (p.kind)}
			<div class="part {p.kind}" style:flex-grow={p.n} title="{KIND_LABEL[p.kind]}: {formatNumber(p.n)}">
				<span>{KIND_LABEL[p.kind]}</span>
				<span class="n">{formatNumber(p.n)}</span>
			</div>
		{/each}
	</div>
</section>

<style>
	.mix {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 28px 32px;
	}

	.head {
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		align-items: baseline;
		gap: 12px;
	}

	.title {
		margin: 0;
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}

	.bar {
		display: flex;
		gap: 4px;
		height: 56px;
		border-radius: 16px;
		overflow: hidden;
	}

	.part {
		flex-basis: 0;
		min-width: 6px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 0 16px;
		overflow: hidden;
		white-space: nowrap;
		font-size: 14px;
		font-weight: 600;
	}

	.n {
		font-family: var(--font-mono);
		font-weight: 400;
	}

	.major {
		background: var(--kind-major);
		color: var(--white);
	}

	.minor {
		background: var(--kind-minor);
	}

	.patch {
		background: var(--kind-patch);
	}

	.prerelease {
		background: repeating-linear-gradient(45deg, #c2c2c2 0 4px, #e9e9e9 4px 8px);
	}

	@media (max-width: 640px) {
		.mix {
			padding: 22px;
		}

		.part {
			padding: 0 8px;
		}

		.part span:first-child {
			display: none;
		}
	}
</style>
