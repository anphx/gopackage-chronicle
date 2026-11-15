// Formatting helpers shared by pages and components.

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
const DAY_MS = 86_400_000;

const toDate = (value: string | number | Date): Date => (value instanceof Date ? value : new Date(value));

/** "just now", "5m ago", "3h ago", "2d ago", "3w ago", "4mo ago", "2y ago". */
export function timeAgo(value: string | number | Date, now: number = Date.now()): string {
	const seconds = Math.max(0, Math.floor((now - toDate(value).getTime()) / 1000));
	if (seconds < 60) return 'just now';
	const minutes = Math.floor(seconds / 60);
	if (minutes < 60) return `${minutes}m ago`;
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours}h ago`;
	const days = Math.floor(hours / 24);
	if (days < 7) return `${days}d ago`;
	if (days < 30) return `${Math.floor(days / 7)}w ago`;
	if (days < 365) return `${Math.floor(days / 30)}mo ago`;
	return `${Math.floor(days / 365)}y ago`;
}

/** "11 Nov 2026". */
export function formatDate(value: string | number | Date): string {
	const d = toDate(value);
	return `${d.getDate()} ${MONTHS[d.getMonth()]} ${d.getFullYear()}`;
}

/** "11 Nov". */
export function formatDayMonth(value: string | number | Date): string {
	const d = toDate(value);
	return `${d.getDate()} ${MONTHS[d.getMonth()]}`;
}

export function monthLabel(value: string | number | Date): string {
	return MONTHS[toDate(value).getMonth()];
}

export function formatNumber(n: number): string {
	return n.toLocaleString('en-US');
}

/** Short form for tight spaces: 4.8M, 613K, 950. */
export function formatCompact(n: number): string {
	return new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 1 }).format(n);
}

/** Whole days between two instants, rounded. */
export function daysBetween(a: string | number | Date, b: string | number | Date): number {
	return Math.round(Math.abs(toDate(b).getTime() - toDate(a).getTime()) / DAY_MS);
}

export function pluralize(n: number, one: string, many = `${one}s`): string {
	return `${formatNumber(n)} ${n === 1 ? one : many}`;
}

const MAJOR_SUFFIX = /^v\d+$/;

/**
 * Splits a module path into a dimmed prefix and the part people recognise:
 * "github.com/spf13/cobra" → { prefix: "github.com/spf13/", name: "cobra" },
 * "github.com/gofiber/fiber/v3" → { prefix: "github.com/gofiber/", name: "fiber/v3" }.
 */
export function splitModulePath(path: string): { prefix: string; name: string } {
	const parts = path.split('/');
	const take = parts.length > 2 && MAJOR_SUFFIX.test(parts[parts.length - 1]) ? 2 : 1;
	const name = parts.slice(-take).join('/');
	const prefix = parts.length > take ? parts.slice(0, -take).join('/') + '/' : '';
	return { prefix, name };
}

/**
 * The short, official-looking name of a module and its major version suffix, if any:
 * "github.com/gofiber/fiber/v3" → { name: "fiber", major: "v3" }.
 */
export function moduleDisplayName(path: string): { name: string; major: string | null } {
	const parts = path.split('/').filter(Boolean);
	const last = parts[parts.length - 1] ?? path;
	if (parts.length > 1 && MAJOR_SUFFIX.test(last)) {
		return { name: parts[parts.length - 2], major: last };
	}
	return { name: last, major: null };
}

export function pkgGoDevURL(path: string, version?: string): string {
	return `https://pkg.go.dev/${path}${version ? `@${version}` : ''}`;
}

/**
 * Link to a module's page. Each path segment is encoded on its own so the slashes stay
 * literal (/packages/github.com/foo/bar). An encoded slash (%2F) would make Cloudflare
 * redirect to the decoded URL.
 */
export function packageHref(path: string): string {
	return `/packages/${path.split('/').map(encodeURIComponent).join('/')}`;
}
