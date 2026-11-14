<script lang="ts">
	import { formatNumber } from '$lib/format';
	import { PAGE_SIZES, pageCount, pageItems } from '$lib/pagination';

	let {
		total,
		page,
		size,
		noun,
		label,
		onpage,
		onsize
	}: {
		total: number;
		page: number;
		size: number;
		/** Plural noun for the count, e.g. "modules". */
		noun: string;
		/** Accessible name of the navigation landmark. */
		label: string;
		onpage: (page: number) => void;
		onsize: (size: number) => void;
	} = $props();

	const totalPages = $derived(pageCount(total, size));
	const from = $derived(total === 0 ? 0 : (page - 1) * size + 1);
	const to = $derived(Math.min(page * size, total));
	const items = $derived(pageItems(page, totalPages));
	const barLeft = $derived(total ? ((from - 1) / total) * 100 : 0);
	const barWidth = $derived(total ? Math.max(4, ((to - from + 1) / total) * 100) : 0);
	const uid = $props.id();
	const sizeLabelId = `per-page-${uid}`;
</script>

<nav class="pagination" aria-label={label}>
	<div class="position">
		<span class="count">
			Showing <strong>{formatNumber(from)}–{formatNumber(to)}</strong> of {formatNumber(total)} {noun}
		</span>
		<span class="bar" aria-hidden="true"><span style:left="{barLeft}%" style:width="{barWidth}%"></span></span>
	</div>

	<div class="per-page">
		<span id={sizeLabelId}>Per page</span>
		<div class="segmented" role="group" aria-labelledby={sizeLabelId}>
			{#each PAGE_SIZES as option (option)}
				<button type="button" aria-pressed={option === size} class:on={option === size} onclick={() => onsize(option)}>
					{option}
				</button>
			{/each}
		</div>
	</div>

	<div class="pages">
		<button type="button" class="edge" disabled={page <= 1} onclick={() => onpage(page - 1)}>← Prev</button>
		{#each items as item, i (item === 'gap' ? `gap-${i}` : item)}
			{#if item === 'gap'}
				<span class="gap" aria-hidden="true">…</span>
			{:else}
				<button
					type="button"
					class="num"
					class:current={item === page}
					aria-label="Page {item}"
					aria-current={item === page ? 'page' : undefined}
					onclick={() => onpage(item)}
				>
					{item}
				</button>
			{/if}
		{/each}
		<button type="button" class="edge" disabled={page >= totalPages} onclick={() => onpage(page + 1)}>Next →</button>
	</div>
</nav>

<style>
	.pagination {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 16px 24px;
		padding: 14px 16px 14px 22px;
		background: rgba(255, 255, 255, 0.5);
		border: 1px solid var(--line);
		border-radius: 24px;
	}

	.position {
		display: flex;
		flex-direction: column;
		gap: 8px;
		min-width: 210px;
	}

	.count {
		font-size: 14px;
		color: var(--text-muted);
	}

	.count strong {
		color: var(--ink);
	}

	.bar {
		position: relative;
		display: block;
		width: 210px;
		height: 4px;
		border-radius: 2px;
		background: rgba(11, 11, 11, 0.1);
		overflow: hidden;
	}

	.bar span {
		position: absolute;
		top: 0;
		bottom: 0;
		border-radius: 2px;
		background: var(--ember);
		transition: left 0.2s ease, width 0.2s ease;
	}

	.per-page {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 14px;
		color: var(--text-subtle);
	}

	.segmented {
		display: flex;
		gap: 2px;
		padding: 4px;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.6);
		border: 1px solid var(--line);
	}

	.segmented button {
		height: 36px;
		min-width: 44px;
		padding: 0 12px;
		border: 0;
		border-radius: 999px;
		background: transparent;
		color: var(--text-subtle);
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
	}

	.segmented button.on {
		background: var(--ink);
		color: var(--white);
	}

	.pages {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 6px;
	}

	.num {
		min-width: 44px;
		height: 44px;
		padding: 0 10px;
		border: 0;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.7);
		font-size: 14px;
		cursor: pointer;
	}

	.num:hover:not(.current) {
		background: var(--white);
	}

	.num.current {
		background: var(--ember);
		color: var(--ink);
		font-weight: 700;
		cursor: default;
	}

	.edge {
		height: 44px;
		padding: 0 18px;
		border: 0;
		border-radius: 999px;
		background: var(--ink);
		color: var(--white);
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
	}

	.edge:hover:not(:disabled) {
		background: var(--ember);
		color: var(--ink);
	}

	.edge:disabled {
		border: 1px solid var(--line-strong);
		background: transparent;
		color: #6a6a6a;
		cursor: not-allowed;
	}

	.gap {
		padding: 0 6px;
		color: var(--text-faint);
	}

	@media (max-width: 640px) {
		.pagination {
			padding: 14px;
		}

		.per-page > span {
			display: none;
		}

		.num:not(.current) {
			display: none;
		}

		.gap {
			display: none;
		}

		.pages {
			width: 100%;
			justify-content: space-between;
		}
	}
</style>
