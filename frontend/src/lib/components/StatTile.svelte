<script lang="ts">
	type Tone = 'ember' | 'ink' | 'fog' | 'surface';

	let {
		label,
		value,
		sub,
		tone = 'surface',
		mono = false,
		size = 'lg'
	}: {
		label: string;
		value: string | number;
		sub?: string;
		tone?: Tone;
		mono?: boolean;
		size?: 'lg' | 'md';
	} = $props();
</script>

<div class="tile {tone}">
	<span class="label">{label}</span>
	<span class="body">
		<span class="value {size}" class:mono title={String(value)}>{value}</span>
		{#if sub}<span class="sub">{sub}</span>{/if}
	</span>
</div>

<style>
	.tile {
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		gap: 20px;
		min-height: 140px;
		padding: 24px;
		border-radius: var(--radius-lg);
		min-width: 0;
		container-type: inline-size;
	}

	.ember {
		background: var(--ember);
	}

	.ink {
		background: var(--ink);
		color: var(--white);
	}

	.fog {
		background: var(--fog);
	}

	.surface {
		background: var(--surface);
		border: 1px solid var(--line);
	}

	.label {
		font-size: 14px;
		font-weight: 600;
	}

	.ink .label {
		color: var(--fog);
	}

	.surface .label {
		color: var(--text-subtle);
	}

	.body {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
	}

	.value {
		font-weight: 600;
		line-height: 1;
		letter-spacing: -0.04em;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		font-variant-numeric: tabular-nums;
	}

	/* Shrink with the tile width so values stay on one line. */
	.value.lg {
		font-size: min(44px, 26cqi);
	}

	.value.md {
		font-size: min(30px, 18cqi);
		letter-spacing: -0.03em;
	}

	.value.mono {
		font-family: var(--font-mono);
		letter-spacing: -0.03em;
	}

	.ink .value {
		color: var(--sand);
	}

	.sub {
		font-size: 13px;
		color: var(--text-faint);
	}

	.ink .sub {
		color: var(--fog);
	}

	/* Grey text fails contrast on orange. */
	.ember .sub {
		color: var(--ink);
	}

	@media (max-width: 640px) {
		.value.lg {
			font-size: min(34px, 26cqi);
		}
	}
</style>
