import { describe, expect, test } from 'bun:test';
import { releaseKind, shortVersion } from './semver';
import { moduleDisplayName, splitModulePath, timeAgo, formatCompact, formatDate } from './format';
import { clampPage, pageCount, pageItems, parsePage, parsePageSize } from './pagination';
import { cadenceSentence, countByKind, median, niceMax, placeLabels, summarizeActivity, toHistory } from './stats';

describe('releaseKind', () => {
	const cases: [string, ReturnType<typeof releaseKind>][] = [
		['v1.0.0', 'major'],
		['v2.0.0+incompatible', 'major'],
		['v1.4.0', 'minor'],
		['v0.39.0', 'minor'],
		['v1.4.2', 'patch'],
		['v1.10.12', 'patch'],
		['v3.0.0-rc.1', 'prerelease'],
		['v0.0.0-20261017001300-d82422bffc48', 'prerelease'],
		['v1.2.3+build-7', 'patch']
	];
	for (const [version, kind] of cases) {
		test(`${version} is ${kind}`, () => expect(releaseKind(version)).toBe(kind));
	}
	test('shortVersion drops the v', () => expect(shortVersion('v3.4.1')).toBe('3.4.1'));
});

describe('module names', () => {
	const cases: [string, { prefix: string; name: string }, { name: string; major: string | null }][] = [
		['github.com/spf13/cobra', { prefix: 'github.com/spf13/', name: 'cobra' }, { name: 'cobra', major: null }],
		['github.com/gofiber/fiber/v3', { prefix: 'github.com/gofiber/', name: 'fiber/v3' }, { name: 'fiber', major: 'v3' }],
		['gopkg.in/yaml.v3', { prefix: 'gopkg.in/', name: 'yaml.v3' }, { name: 'yaml.v3', major: null }],
		['example.com/v2', { prefix: 'example.com/', name: 'v2' }, { name: 'example.com', major: 'v2' }],
		['golang.org/x/tools', { prefix: 'golang.org/x/', name: 'tools' }, { name: 'tools', major: null }]
	];
	for (const [path, split, display] of cases) {
		test(path, () => {
			expect(splitModulePath(path)).toEqual(split);
			expect(moduleDisplayName(path)).toEqual(display);
		});
	}
});

describe('timeAgo', () => {
	const now = Date.UTC(2026, 10, 14, 12, 0, 0);
	const cases: [number, string][] = [
		[30_000, 'just now'],
		[5 * 60_000, '5m ago'],
		[3 * 3_600_000, '3h ago'],
		[2 * 86_400_000, '2d ago'],
		[15 * 86_400_000, '2w ago'],
		[95 * 86_400_000, '3mo ago'],
		[800 * 86_400_000, '2y ago'],
		[-60_000, 'just now']
	];
	for (const [ago, want] of cases) {
		test(want, () => expect(timeAgo(now - ago, now)).toBe(want));
	}
	test('formatDate', () => expect(formatDate(new Date(2026, 10, 11))).toBe('11 Nov 2026'));
});

describe('formatCompact', () => {
	test('shortens large numbers', () => {
		expect(formatCompact(950)).toBe('950');
		expect(formatCompact(612_904)).toBe('612.9K');
		expect(formatCompact(4_812_330)).toBe('4.8M');
	});
});

describe('pagination', () => {
	const cases: [number, number, (number | 'gap')[]][] = [
		[1, 1, [1]],
		[1, 4, [1, 2, 3, 4]],
		[2, 4, [1, 2, 3, 4]],
		[4, 4, [1, 2, 3, 4]],
		[1, 5, [1, 2, 'gap', 5]],
		[4, 50, [1, 2, 3, 4, 5, 'gap', 50]],
		[7, 50, [1, 'gap', 6, 7, 8, 'gap', 50]],
		[47, 50, [1, 'gap', 46, 47, 48, 49, 50]],
		[1, 3, [1, 2, 3]]
	];
	for (const [current, total, want] of cases) {
		test(`page ${current} of ${total}`, () => expect(pageItems(current, total)).toEqual(want));
	}
	test('pageCount', () => {
		expect(pageCount(0, 25)).toBe(1);
		expect(pageCount(100, 25)).toBe(4);
		expect(pageCount(101, 25)).toBe(5);
	});
	test('clampPage', () => {
		expect(clampPage(9, 100, 25)).toBe(4);
		expect(clampPage(0, 100, 25)).toBe(1);
		expect(clampPage(Number.NaN, 100, 25)).toBe(1);
	});
	test('parse query values', () => {
		expect(parsePageSize('50')).toBe(50);
		expect(parsePageSize('7')).toBe(25);
		expect(parsePageSize(null)).toBe(25);
		expect(parsePage('3')).toBe(3);
		expect(parsePage('-1')).toBe(1);
		expect(parsePage('abc')).toBe(1);
	});
});

describe('stats', () => {
	const releases = [
		{ version: 'v1.1.0', released_at: '2026-03-01T00:00:00Z' },
		{ version: 'v1.0.0', released_at: '2026-01-01T00:00:00Z' },
		{ version: 'v1.0.1', released_at: '2026-01-15T00:00:00Z' },
		{ version: 'v1.2.0-rc.1', released_at: '2026-04-01T00:00:00Z' }
	];

	test('toHistory sorts oldest first and computes gaps', () => {
		const h = toHistory(releases);
		expect(h.map((e) => e.version)).toEqual(['v1.0.0', 'v1.0.1', 'v1.1.0', 'v1.2.0-rc.1']);
		expect(h.map((e) => e.gapDays)).toEqual([null, 14, 45, 31]);
		expect(countByKind(h)).toEqual({ major: 1, minor: 1, patch: 1, prerelease: 1 });
	});

	test('median', () => {
		expect(median([])).toBeNull();
		expect(median([5])).toBe(5);
		expect(median([9, 1, 5])).toBe(5);
		expect(median([1, 2, 3, 10])).toBe(3);
	});

	test('placeLabels alternates sides and hides labels that cannot fit', () => {
		expect(placeLabels([3, 9, 15.6, 21.4, 63], 11)).toEqual(['above', 'below', 'above', 'below', 'above']);
		expect(placeLabels([10, 11, 12], 5)).toEqual(['above', 'below', null]);
	});

	test('cadenceSentence', () => {
		expect(cadenceSentence(toHistory(releases))).toBe(
			'Ships a minor or major about every 8 weeks, with patches in between.'
		);
		expect(cadenceSentence(toHistory(releases.slice(0, 2)))).toBe('Not enough releases yet to see a rhythm.');
	});

	test('summarizeActivity', () => {
		const s = summarizeActivity([
			{ week_start: 'a', major: 1, minor: 2, patch: 7, prerelease: 0 },
			{ week_start: 'b', major: 0, minor: 1, patch: 1, prerelease: 0 }
		]);
		expect(s).toEqual({ total: 12, average: 6, busiest: { weekStart: 'a', total: 10 }, patchShare: 67 });
		expect(summarizeActivity([]).busiest).toBeNull();
	});

	test('niceMax', () => {
		expect(niceMax(0)).toBe(1);
		expect(niceMax(7)).toBe(8);
		expect(niceMax(130)).toBe(150);
		expect(niceMax(214)).toBe(250);
		expect(niceMax(480)).toBe(500);
		expect(niceMax(2000)).toBe(2000);
	});
});
