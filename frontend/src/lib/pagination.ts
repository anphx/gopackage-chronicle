// Pure helpers behind the Pagination component.

export const PAGE_SIZES = [10, 25, 50, 100] as const;
export const DEFAULT_PAGE_SIZE = 25;

export type PageItem = number | 'gap';

export function pageCount(total: number, size: number): number {
	return Math.max(1, Math.ceil(total / Math.max(1, size)));
}

export function clampPage(page: number, total: number, size: number): number {
	if (!Number.isFinite(page)) return 1;
	return Math.min(Math.max(1, Math.floor(page)), pageCount(total, size));
}

/**
 * Page buttons to show: always the first and last page and the pages next to the current
 * one; runs of hidden pages collapse into a single 'gap'. A gap that would hide exactly one
 * page shows that page instead.
 */
export function pageItems(current: number, totalPages: number): PageItem[] {
	const items: PageItem[] = [];
	for (let p = 1; p <= totalPages; p++) {
		const near = Math.abs(p - current) <= 1;
		const edge = p === 1 || p === totalPages;
		const bridges = (p === 2 && current === 4) || (p === totalPages - 1 && current === totalPages - 3);
		if (near || edge || bridges) items.push(p);
		else if (items[items.length - 1] !== 'gap') items.push('gap');
	}
	return items;
}

export function parsePageSize(value: string | null): number {
	const n = Number(value);
	return (PAGE_SIZES as readonly number[]).includes(n) ? n : DEFAULT_PAGE_SIZE;
}

export function parsePage(value: string | null): number {
	const n = Number(value);
	return Number.isInteger(n) && n > 0 ? n : 1;
}
