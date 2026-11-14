// Derived numbers for charts. Pure functions, unit tested in stats.test.ts.

import type { Release, WeeklyActivity } from './api/client';
import { releaseKind, type ReleaseKind } from './semver';

const DAY_MS = 86_400_000;

export interface HistoryEntry {
	version: string;
	kind: ReleaseKind;
	time: number;
	/** Days since the previous release; null for the first one. */
	gapDays: number | null;
}

/** Releases sorted oldest first, with their kind and the gap to the previous release. */
export function toHistory(releases: Pick<Release, 'version' | 'released_at'>[]): HistoryEntry[] {
	const sorted = releases
		.map((r) => ({ version: r.version, time: new Date(r.released_at).getTime() }))
		.sort((a, b) => a.time - b.time);
	return sorted.map((r, i) => ({
		version: r.version,
		kind: releaseKind(r.version),
		time: r.time,
		gapDays: i === 0 ? null : Math.round((r.time - sorted[i - 1].time) / DAY_MS)
	}));
}

export function median(values: number[]): number | null {
	if (values.length === 0) return null;
	const s = [...values].sort((a, b) => a - b);
	const mid = Math.floor(s.length / 2);
	return s.length % 2 ? s[mid] : Math.round((s[mid - 1] + s[mid]) / 2);
}

export function countByKind(entries: { kind: ReleaseKind }[]): Record<ReleaseKind, number> {
	const counts: Record<ReleaseKind, number> = { major: 0, minor: 0, patch: 0, prerelease: 0 };
	for (const e of entries) counts[e.kind]++;
	return counts;
}

/**
 * Chooses a side for each label on one timeline lane so neighbours don't collide.
 * `positions` are percentages along the axis, sorted ascending; a label needs `minGap`
 * percent clear of the previous label on the same side. When both sides are taken the
 * label is hidden (null).
 */
export function placeLabels(positions: number[], minGap: number): ('above' | 'below' | null)[] {
	let lastAbove = -Infinity;
	let lastBelow = -Infinity;
	return positions.map((p) => {
		if (p - lastAbove >= minGap) {
			lastAbove = p;
			return 'above';
		}
		if (p - lastBelow >= minGap) {
			lastBelow = p;
			return 'below';
		}
		return null;
	});
}

/** One-line description of how a module ships, e.g. "Ships a minor about every 6 weeks…". */
export function cadenceSentence(history: HistoryEntry[]): string {
	if (history.length < 3) return 'Not enough releases yet to see a rhythm.';
	const features = history.filter((h) => h.kind === 'minor' || h.kind === 'major');
	const featureGaps = features.slice(1).map((h, i) => (h.time - features[i].time) / DAY_MS);
	const typical = median(featureGaps);
	const patches = history.some((h) => h.kind === 'patch');
	if (typical !== null && typical > 0) {
		const every = typical >= 14 ? `${Math.round(typical / 7)} weeks` : `${Math.round(typical)} days`;
		return `Ships a minor or major about every ${every}${patches ? ', with patches in between' : ''}.`;
	}
	const gaps = history.flatMap((h) => (h.gapDays === null ? [] : [h.gapDays]));
	const m = median(gaps) ?? 0;
	return `Mostly ${patches ? 'patch' : 'pre-'}releases, typically ${m} day${m === 1 ? '' : 's'} apart.`;
}

export interface ActivitySummary {
	total: number;
	average: number;
	busiest: { weekStart: string; total: number } | null;
	patchShare: number;
}

export function weekTotal(w: WeeklyActivity): number {
	return w.major + w.minor + w.patch + w.prerelease;
}

export function summarizeActivity(weeks: WeeklyActivity[]): ActivitySummary {
	let total = 0;
	let patches = 0;
	let busiest: ActivitySummary['busiest'] = null;
	for (const w of weeks) {
		const t = weekTotal(w);
		total += t;
		patches += w.patch;
		if (!busiest || t > busiest.total) busiest = { weekStart: w.week_start, total: t };
	}
	return {
		total,
		average: weeks.length ? Math.round(total / weeks.length) : 0,
		busiest: busiest && busiest.total > 0 ? busiest : null,
		patchShare: total ? Math.round((patches / total) * 100) : 0
	};
}

/** A round axis maximum at least `value`: 1, 1.5, 2, 2.5, 3, 4, 5, 6 or 8 times a power of ten. */
export function niceMax(value: number): number {
	if (value <= 0) return 1;
	const pow = 10 ** Math.floor(Math.log10(value));
	for (const step of [1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10]) {
		if (step * pow >= value) return step * pow;
	}
	return 10 * pow;
}
