// TypeScript interfaces matching backend API responses

export interface Package {
	id: number;
	path: string;
	created_at: string;
}

export interface Release {
	id: number;
	package_id: number;
	version: string;
	released_at: string;
	indexed_at: string;
}

export interface ReleaseWithPackage extends Release {
	package_path: string;
}

export interface PackageDetail {
	package: Package;
	releases: Release[];
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
const getBaseURL = (): string => {
	// In SvelteKit, use PUBLIC_ prefix for client-side env vars
	if (typeof window !== 'undefined') {
		return import.meta.env.PUBLIC_API_BASE_URL || 'http://localhost:8080';
	}
	// Server-side can use different URL if needed
	return process.env.API_BASE_URL || 'http://localhost:8080';
};

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
 * Get package details with release history
 */
export async function getPackageDetail(
	packagePath: string
): Promise<PackageDetail> {
	// URL encode the package path to handle special characters
	const encodedPath = encodeURIComponent(packagePath);
	const response = await apiFetch<PackageDetailResponse>(`/api/packages/${encodedPath}`);
	return {
		package: response.package,
		releases: response.releases
	};
}

