import { expect, test } from '@playwright/test';

const modulePath = 'github.com/acme/clock';

const releases = [
	{
		id: 1,
		package_id: 1,
		version: 'v1.0.0',
		released_at: '2026-10-01T00:00:00Z',
		indexed_at: '2026-10-01T01:00:00Z'
	},
	{
		id: 2,
		package_id: 1,
		version: 'v1.1.0',
		released_at: '2026-10-15T00:00:00Z',
		indexed_at: '2026-10-15T01:00:00Z'
	},
	{
		id: 3,
		package_id: 1,
		version: 'v1.2.0',
		released_at: '2026-11-10T00:00:00Z',
		indexed_at: '2026-11-10T01:00:00Z'
	}
];

const activity = {
	package_id: 1,
	path: modulePath,
	latest_version: 'v1.2.0',
	latest_kind: 'minor',
	last_released_at: '2026-11-10T00:00:00Z',
	window_releases: 3,
	release_count: 8,
	recent_versions: ['v1.2.0', 'v1.1.0', 'v1.0.0'],
	weekly: Array(12).fill(0)
};

test.beforeEach(async ({ page }) => {
	await page.route('**/api/**', async (route) => {
		const url = new URL(route.request().url());
		if (!url.pathname.startsWith('/api/')) {
			await route.continue();
			return;
		}

		if (url.pathname === '/api/stats') {
			await route.fulfill({
				json: {
					total_releases: 1234,
					total_packages: 567,
					last_synced_at: '2026-11-15T20:00:00Z',
					weeks: [
						{
							week_start: '2026-11-09',
							major: 1,
							minor: 2,
							patch: 3,
							prerelease: 4
						}
					]
				}
			});
			return;
		}

		if (url.pathname === '/api/stats/packages') {
			await route.fulfill({
				json: {
					packages: [activity],
					total: 1,
					kind_counts: { all: 1, major: 0, minor: 1, patch: 0, prerelease: 0 },
					limit: 50,
					offset: 0,
					weeks: 12
				}
			});
			return;
		}

		if (url.pathname === '/api/releases') {
			await route.fulfill({
				json: {
					releases: [
						{
							...releases[2],
							package_path: modulePath
						}
					],
					limit: 7,
					offset: 0
				}
			});
			return;
		}

		if (url.pathname.startsWith('/api/packages/')) {
			const path = decodeURIComponent(url.pathname.slice('/api/packages/'.length));
			if (path === modulePath) {
				await route.fulfill({
					json: {
						package: {
							id: 1,
							path: modulePath,
							created_at: '2026-10-01T00:00:00Z'
						},
						releases: [...releases].reverse(),
						limit: 200,
						offset: 0
					}
				});
				return;
			}
		}

		await route.fulfill({ status: 404, json: { error: 'Unexpected API request' } });
	});
});

test('home page presents overview metrics, live releases, and module data', async ({ page }) => {
	await page.goto('/');

	await expect(page.getByRole('heading', { name: /Every Go release, in motion/ })).toBeVisible();
	await expect(page.getByText('Releases indexed', { exact: true }).locator('..')).toContainText('1,234');
	await expect(page.getByText('Modules tracked', { exact: true }).locator('..')).toContainText('567');

	const moduleRow = page
		.getByRole('table', { name: 'Recently updated modules' })
		.getByRole('row')
		.filter({ hasText: modulePath });
	await expect(moduleRow).toContainText('v1.2.0');
	await expect(page.getByRole('heading', { name: 'Just released' })).toBeVisible();
	await expect(page.locator('.feed').getByRole('link', { name: /clock/ })).toBeVisible();
});

// Opening the URL directly is what a reload or a shared link does. Both forms must
// render: the plain path the app links to, and the older %2F-encoded form.
for (const [label, url] of [
	['plain path', `/packages/${modulePath}`],
	['encoded path', `/packages/${encodeURIComponent(modulePath)}`]
]) {
	test(`module page loads directly from the ${label}`, async ({ page }) => {
		await page.goto(url);
		await expect(page.getByRole('heading', { name: 'clock' })).toBeVisible();
		await expect(page.locator('code')).toHaveText(`go get ${modulePath}@v1.2.0`);
	});
}

test('home page links to the module page with the plain path', async ({ page }) => {
	await page.goto('/');
	await page.locator('.feed').getByRole('link', { name: /clock/ }).click();
	await expect(page).toHaveURL(`/packages/${modulePath}`);
	await expect(page.getByRole('heading', { name: 'clock' })).toBeVisible();
});

test('module page presents its latest release and version history', async ({ page }) => {
	await page.goto(`/packages/${modulePath}`);

	await expect(page.getByRole('heading', { name: 'clock' })).toBeVisible();
	await expect(page.locator('code')).toHaveText(`go get ${modulePath}@v1.2.0`);
	await expect(page.getByText('Latest', { exact: true }).locator('..')).toContainText('v1.2.0');
	await expect(page.locator('.tiles').getByText('Releases', { exact: true }).locator('..')).toContainText('3');

	const versions = page.getByRole('table').getByRole('row');
	await expect(versions.filter({ hasText: 'v1.2.0' })).toContainText('latest');
	await expect(versions.filter({ hasText: 'v1.1.0' })).toBeVisible();
	await expect(versions.filter({ hasText: 'v1.0.0' })).toBeVisible();
});

test('about page explains the index and presents live totals', async ({ page }) => {
	await page.goto('/about');

	await expect(page.getByRole('heading', { name: /A logbook for the Go ecosystem/ })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'How it works' })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'The indexer' })).toBeVisible();
	await expect(page.getByText('Releases indexed', { exact: true }).locator('..')).toContainText('1,234');
	await expect(page.getByText('Modules tracked', { exact: true }).locator('..')).toContainText('567');
});
