import { env } from '$env/dynamic/public';

// TypeScript interfaces matching backend API responses

export interface Package {
	id: number;
	path: string;
}

/** A tagged version of a module, identified by (package_id, version). */
export interface Release {
	package_id: number;
	version: string;
	released_at: string;
}

export interface ReleaseWithPackage extends Release {
	package_path: string;
}

export interface PackageDetail {
	package: Package;
	releases: Release[];
}

export type ReleaseKindName = 'major' | 'minor' | 'patch' | 'prerelease';

export interface WeeklyActivity {
	week_start: string;
	major: number;
	minor: number;
	patch: number;
	prerelease: number;
}

export interface StatsOverview {
	total_releases: number;
	total_packages: number;
	last_synced_at: string | null;
	weeks: WeeklyActivity[];
}

export interface PackageActivity {
	package_id: number;
	path: string;
	latest_version: string;
	latest_kind: ReleaseKindName;
	last_released_at: string;
	/** Releases inside the requested window of weeks. */
	window_releases: number;
	/** Releases of the module, all time. */
	release_count: number;
	/** Up to four newest versions, newest first. */
	recent_versions: string[];
	/** Releases per week of the window, oldest week first. */
	weekly: number[];
}

export interface KindCounts {
	all: number;
	major: number;
	minor: number;
	patch: number;
	prerelease: number;
}

export interface RecentPackagesResponse {
	packages: PackageActivity[];
	total: number;
	kind_counts: KindCounts;
	limit: number;
	offset: number;
	weeks: number;
}

export type RecentPackagesSort = 'recent' | 'active' | 'name';

export interface RecentPackagesParams {
	limit: number;
	offset: number;
	weeks?: number;
	q?: string;
	kind?: ReleaseKindName | '';
	sort?: RecentPackagesSort;
}

// API response wrappers
export interface ReleasesResponse {
	releases: ReleaseWithPackage[];
	limit: number;
	offset: number;
}

interface PackagesResponse {
	packages: Package[];
	limit: number;
	offset: number;
}

interface PackageDetailResponse {
	package: Package;
	releases: Release[];
	limit: number;
	offset: number;
}

// API configuration
// $env/dynamic/public works on both server and client, and on Cloudflare Workers,
// where `process.env` does not exist.
const getBaseURL = (): string => env.PUBLIC_API_BASE_URL || 'http://localhost:8080';

// Generic fetch wrapper with error handling
async function apiFetch<T>(endpoint: string): Promise<T> {
	const url = `${getBaseURL()}${endpoint}`;

	try {
		const response = await fetch(url);

		if (!response.ok) {
			throw new Error(`HTTP error! status: ${response.status}`);
		}

		return await response.json() as T;
	} catch (error) {
		console.error(`API fetch error for ${endpoint}:`, error);
		throw error;
	}
}

// API client functions

/**
 * Get recent releases with pagination
 */
export async function getRecentReleases(
	limit: number = 50,
	offset: number = 0
): Promise<ReleasesResponse> {
	const response = await apiFetch<ReleasesResponse>(
		`/api/releases?limit=${limit}&offset=${offset}`
	);
	return response;
}

/**
 * Get all packages with pagination
 */
export async function getPackages(
	limit: number = 50,
	offset: number = 0
): Promise<Package[]> {
	const response = await apiFetch<PackagesResponse>(
		`/api/packages?limit=${limit}&offset=${offset}`
	);
	return response.packages;
}

/**
 * Get package details with one page of its release history (newest first)
 */
export async function getPackageDetail(
	packagePath: string,
	limit: number = 50,
	offset: number = 0
): Promise<PackageDetail> {
	// URL encode the package path to handle special characters
	const encodedPath = encodeURIComponent(packagePath);
	const response = await apiFetch<PackageDetailResponse>(
		`/api/packages/${encodedPath}?limit=${limit}&offset=${offset}`
	);
	return {
		package: response.package,
		releases: response.releases ?? []
	};
}

/**
 * Get a package with its complete release history (newest first), fetched in pages of 200
 * (the API maximum). Stops after `maxReleases` to keep very long histories bounded.
 */
export async function getPackageWithAllReleases(
	packagePath: string,
	maxReleases: number = 5000
): Promise<PackageDetail> {
	const pageSize = 200;
	const first = await getPackageDetail(packagePath, pageSize, 0);
	const releases = [...first.releases];
	let last = first.releases.length;
	while (last === pageSize && releases.length < maxReleases) {
		const next = await getPackageDetail(packagePath, pageSize, releases.length);
		releases.push(...next.releases);
		last = next.releases.length;
	}
	return { package: first.package, releases };
}

/**
 * Get index-wide totals and weekly release activity for the last `weeks` weeks
 */
export async function getStats(weeks: number = 30): Promise<StatsOverview> {
	return apiFetch<StatsOverview>(`/api/stats?weeks=${weeks}`);
}

/**
 * Get one page of modules released recently, with filters and sorting
 */
export async function getRecentPackages(params: RecentPackagesParams): Promise<RecentPackagesResponse> {
	const query = new URLSearchParams({
		limit: String(params.limit),
		offset: String(params.offset)
	});
	if (params.weeks) query.set('weeks', String(params.weeks));
	if (params.q?.trim()) query.set('q', params.q.trim());
	if (params.kind) query.set('kind', params.kind);
	if (params.sort) query.set('sort', params.sort);
	return apiFetch<RecentPackagesResponse>(`/api/stats/packages?${query}`);
}

