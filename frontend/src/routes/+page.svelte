<script lang="ts">
	import { onMount } from 'svelte';
	import { getRecentReleases, type ReleaseWithPackage } from '$lib/api/client';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import SkeletonRelease from '$lib/components/SkeletonRelease.svelte';

	// Raw releases from API
	let allReleases: ReleaseWithPackage[] = [];
	let loading = true;
	let error: string | null = null;

	// Pagination settings for grouped packages
	let currentPage = 1;
	const packagesPerPage = 20;

	// Filtering/sorting
	let sortBy: 'recent' | 'package' = 'recent';
	let filterText = '';

	// Fetch a large batch of releases upfront
	const fetchBatchSize = 1000;

	// Format relative time
	const timeAgo = (dateString: string): string => {
		const date = new Date(dateString);
		const now = new Date();
		const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

		if (seconds < 60) return 'just now';
		const minutes = Math.floor(seconds / 60);
		if (minutes < 60) return `${minutes}m ago`;
		const hours = Math.floor(minutes / 60);
		if (hours < 24) return `${hours}h ago`;
		const days = Math.floor(hours / 24);
		if (days < 7) return `${days}d ago`;
		const weeks = Math.floor(days / 7);
		if (weeks < 4) return `${weeks}w ago`;
		const months = Math.floor(days / 30);
		if (months < 12) return `${months}mo ago`;
		const years = Math.floor(days / 365);
		return `${years}y ago`;
	};

	// Group releases by package
	interface PackageGroup {
		package_path: string;
		releases: ReleaseWithPackage[];
		latest_release: ReleaseWithPackage;
	}

	const groupReleasesByPackage = (releases: ReleaseWithPackage[]): PackageGroup[] => {
		const grouped = new Map<string, ReleaseWithPackage[]>();

		releases.forEach(release => {
			if (!grouped.has(release.package_path)) {
				grouped.set(release.package_path, []);
			}
			grouped.get(release.package_path)!.push(release);
		});

		// Convert to array and sort releases within each group
		return Array.from(grouped.entries())
			.map(([path, releases]) => {
				const sorted = releases.sort((a, b) =>
					new Date(b.released_at).getTime() - new Date(a.released_at).getTime()
				);
				return {
					package_path: path,
					releases: sorted,
					latest_release: sorted[0]
				};
			})
			.sort((a, b) =>
				new Date(b.latest_release.released_at).getTime() -
				new Date(a.latest_release.released_at).getTime()
			);
	};

	// Step 1: Group all releases
	$: allGroups = groupReleasesByPackage(allReleases);

	// Step 2: Apply filtering and sorting
	$: filteredGroups = (() => {
		let result = [...allGroups];

		// Apply filter
		if (filterText.trim()) {
			const filter = filterText.toLowerCase();
			result = result.filter(g =>
				g.package_path.toLowerCase().includes(filter) ||
				g.releases.some(r => r.version.toLowerCase().includes(filter))
			);
		}

		// Apply sort
		if (sortBy === 'package') {
			result.sort((a, b) => a.package_path.localeCompare(b.package_path));
		}
		// 'recent' is already sorted from grouping

		return result;
	})();

	// Step 3: Paginate the filtered/sorted groups
	$: totalPages = Math.ceil(filteredGroups.length / packagesPerPage);
	$: paginatedGroups = filteredGroups.slice(
		(currentPage - 1) * packagesPerPage,
		currentPage * packagesPerPage
	);

	// Calculate stats
	$: totalReleases = allReleases.length;
	$: totalPackages = allGroups.length;

	const loadReleases = async () => {
		loading = true;
		error = null;

		try {
			const data = await getRecentReleases(fetchBatchSize, 0);
			allReleases = data.releases;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load releases';
			console.error('Error loading releases:', e);
		} finally {
			loading = false;
		}
	};

	const goToPage = (page: number) => {
		if (page >= 1 && page <= totalPages) {
			currentPage = page;
			window.scrollTo({ top: 0, behavior: 'smooth' });
		}
	};

	// Reset to page 1 when filter/sort changes
	$: if (filterText || sortBy) {
		currentPage = 1;
	}

	// Generate page numbers to display
	$: pageNumbers = (() => {
		const pages: (number | string)[] = [];

		if (totalPages <= 7) {
			// Show all pages if 7 or fewer
			for (let i = 1; i <= totalPages; i++) {
				pages.push(i);
			}
		} else {
			// Always show first page
			pages.push(1);

			// Show pages around current page
			const rangeStart = Math.max(2, currentPage - 1);
			const rangeEnd = Math.min(totalPages - 1, currentPage + 1);

			// Add ellipsis if needed
			if (rangeStart > 2) {
				pages.push('...');
			}

			// Add middle range
			for (let i = rangeStart; i <= rangeEnd; i++) {
				if (i !== 1 && i !== totalPages) {
					pages.push(i);
				}
			}

			// Add ellipsis if needed
			if (rangeEnd < totalPages - 1) {
				pages.push('...');
			}

			// Always show last page
			if (totalPages > 1) {
				pages.push(totalPages);
			}
		}

		return pages;
	})();

	onMount(() => {
		loadReleases();
	});
</script>

<div class="container">
	<header class="header">
		<div class="hero">
			<h1>
				<svg class="logo-icon" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 2L2 7l10 5 10-5-10-5z"/>
					<path d="M2 17l10 5 10-5"/>
					<path d="M2 12l10 5 10-5"/>
				</svg>
				Go Package Chronicles
			</h1>
			<p class="subtitle">Track and explore the complete release history of the Go ecosystem</p>
		</div>

		{#if !loading && allGroups.length > 0}
			<div class="stats-cards">
				<div class="stat-card">
					<div class="stat-icon">📦</div>
					<div class="stat-content">
						<div class="stat-value">{totalReleases}</div>
						<div class="stat-label">Releases</div>
					</div>
				</div>
				<div class="stat-card">
					<div class="stat-icon">🔖</div>
					<div class="stat-content">
						<div class="stat-value">{totalPackages}</div>
						<div class="stat-label">Packages</div>
					</div>
				</div>
				<div class="stat-card">
					<div class="stat-icon">⏱️</div>
					<div class="stat-content">
						<div class="stat-value">{timeAgo(allGroups[0].latest_release.released_at)}</div>
						<div class="stat-label">Latest</div>
					</div>
				</div>
			</div>
		{/if}
	</header>

	<SearchBar />

	<section class="releases-section">
		<div class="section-header">
			<h2>
				<svg class="section-icon" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
				</svg>
				Recent Releases
			</h2>

			{#if !loading && allGroups.length > 0}
				<div class="controls">
					<div class="filter-group">
						<label for="filter">
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<circle cx="11" cy="11" r="8"/>
								<path d="M21 21l-4.35-4.35"/>
							</svg>
							Filter
						</label>
						<input
							id="filter"
							type="text"
							placeholder="Package or version..."
							bind:value={filterText}
							class="filter-input"
						/>
					</div>

					<div class="sort-group">
						<label for="sort">
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M3 6h18M7 12h10M11 18h2"/>
							</svg>
							Sort
						</label>
						<select id="sort" bind:value={sortBy} class="sort-select">
							<option value="recent">Most Recent</option>
							<option value="package">Package Name</option>
						</select>
					</div>
				</div>
			{/if}
		</div>

		{#if loading}
			<div class="skeleton-list">
				{#each Array(8) as _}
					<SkeletonRelease />
				{/each}
			</div>
		{:else if error}
			<p class="error">Error: {error}</p>
		{:else}
			{#if filteredGroups.length === 0}
				<div class="no-results">
					<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="11" cy="11" r="8"/>
						<path d="M21 21l-4.35-4.35"/>
					</svg>
					<p>No releases match your filter</p>
					<button class="clear-filter" on:click={() => filterText = ''}>Clear filter</button>
				</div>
			{:else}
				<div class="package-groups">
					{#each paginatedGroups as group}
						<div class="package-group">
							<div class="package-header">
								<a href="/packages/{encodeURIComponent(group.package_path)}" class="package-link">
									<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
									</svg>
									{group.package_path}
								</a>
								<span class="release-count">{group.releases.length} release{group.releases.length !== 1 ? 's' : ''}</span>
							</div>
							<div class="releases-grid">
								{#each group.releases as release}
									<div class="release-item">
										<span class="version">{release.version}</span>
										<span class="date">{timeAgo(release.released_at)}</span>
									</div>
								{/each}
							</div>
						</div>
					{/each}
				</div>
			{/if}

			{#if totalPages > 1}
				<div class="pagination">
					<button
						class="pagination-button"
						on:click={() => goToPage(currentPage - 1)}
						disabled={currentPage === 1}
					>
						← Previous
					</button>

					<div class="page-numbers">
						{#each pageNumbers as pageNum}
							{#if pageNum === '...'}
								<span class="ellipsis">...</span>
							{:else}
								<button
									class="page-number"
									class:active={pageNum === currentPage}
									on:click={() => goToPage(Number(pageNum))}
								>
									{pageNum}
								</button>
							{/if}
						{/each}
					</div>

					<button
						class="pagination-button"
						on:click={() => goToPage(currentPage + 1)}
						disabled={currentPage >= totalPages}
					>
						Next →
					</button>
				</div>

				<div class="pagination-info-text">
					Showing {(currentPage - 1) * packagesPerPage + 1}-{Math.min(currentPage * packagesPerPage, filteredGroups.length)} of {filteredGroups.length} packages
				</div>
			{/if}
		{/if}
	</section>
</div>

<style>
	.container {
		max-width: 900px;
		margin: 0 auto;
		padding: 2rem 1rem;
	}

	.header {
		text-align: center;
		margin-bottom: 3rem;
	}

	.hero {
		margin-bottom: 2rem;
	}

	.header h1 {
		font-size: 2.5rem;
		margin: 0 0 0.5rem 0;
		color: #202124;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 1rem;
	}

	.logo-icon {
		color: #1a73e8;
		flex-shrink: 0;
	}

	.subtitle {
		font-size: 1.125rem;
		color: #5f6368;
		margin: 0;
		max-width: 600px;
		margin: 0 auto;
	}

	.stats-cards {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-top: 2rem;
		max-width: 700px;
		margin-left: auto;
		margin-right: auto;
	}

	.stat-card {
		background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
		border: 1px solid #e0e0e0;
		border-radius: 12px;
		padding: 1.5rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: all 0.3s ease;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
	}

	.stat-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(26, 115, 232, 0.15);
		border-color: #1a73e8;
	}

	.stat-icon {
		font-size: 2rem;
		flex-shrink: 0;
	}

	.stat-content {
		flex: 1;
		text-align: left;
	}

	.stat-value {
		font-size: 1.75rem;
		font-weight: 700;
		color: #202124;
		line-height: 1;
		margin-bottom: 0.25rem;
	}

	.stat-label {
		font-size: 0.875rem;
		color: #5f6368;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1.5rem;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.section-header h2 {
		font-size: 1.5rem;
		margin: 0;
		color: #202124;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.section-icon {
		color: #1a73e8;
		flex-shrink: 0;
	}

	.controls {
		display: flex;
		gap: 1rem;
		align-items: flex-end;
		flex-wrap: wrap;
	}

	.filter-group,
	.sort-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.filter-group label,
	.sort-group label {
		font-size: 0.75rem;
		font-weight: 600;
		color: #5f6368;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.filter-group label svg,
	.sort-group label svg {
		color: #5f6368;
	}

	.filter-input,
	.sort-select {
		padding: 0.625rem 1rem;
		font-size: 0.875rem;
		border: 1px solid #e0e0e0;
		border-radius: 6px;
		background: white;
		color: #202124;
		transition: all 0.2s ease;
		font-family: inherit;
	}

	.filter-input {
		min-width: 200px;
	}

	.filter-input:focus,
	.sort-select:focus {
		outline: none;
		border-color: #1a73e8;
		box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.1);
	}

	.sort-select {
		cursor: pointer;
		padding-right: 2rem;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%235f6368' d='M6 9L1 4h10z'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.75rem center;
		appearance: none;
	}

	.no-results {
		text-align: center;
		padding: 4rem 2rem;
		color: #5f6368;
	}

	.no-results svg {
		margin-bottom: 1rem;
		color: #9e9e9e;
	}

	.no-results p {
		font-size: 1.125rem;
		margin: 0 0 1rem 0;
	}

	.clear-filter {
		padding: 0.625rem 1.5rem;
		font-size: 0.875rem;
		font-weight: 600;
		color: #1a73e8;
		background: transparent;
		border: 1px solid #1a73e8;
		border-radius: 6px;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.clear-filter:hover {
		background: #e8f0fe;
	}

	.package-groups {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.package-group {
		background: white;
		border: 1px solid #e0e0e0;
		border-radius: 8px;
		padding: 1.5rem;
		transition: all 0.2s ease;
	}

	.package-group:hover {
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		border-color: #1a73e8;
	}

	.package-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid #e0e0e0;
	}

	.package-link {
		font-size: 1.125rem;
		font-weight: 600;
		color: #1a73e8;
		text-decoration: none;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: color 0.2s ease;
	}

	.package-link:hover {
		color: #1557b0;
		text-decoration: underline;
	}

	.package-link svg {
		flex-shrink: 0;
	}

	.release-count {
		font-size: 0.875rem;
		color: #5f6368;
		background: #f8f9fa;
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-weight: 500;
	}

	.releases-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 0.75rem;
	}

	.release-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem;
		background: #f8f9fa;
		border-radius: 6px;
		gap: 0.5rem;
	}

	.version {
		font-family: 'Courier New', monospace;
		font-weight: 600;
		color: #202124;
		font-size: 0.875rem;
	}

	.date {
		font-size: 0.75rem;
		color: #5f6368;
		white-space: nowrap;
	}

	.releases-section h2 {
		font-size: 1.5rem;
		margin: 0 0 1.5rem 0;
		color: #202124;
	}

	.skeleton-list {
		margin: 1rem 0;
	}

	.error {
		text-align: center;
		padding: 2rem;
		color: #d93025;
	}

	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		margin-top: 2rem;
		padding: 1rem 0 0.5rem 0;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.pagination-info-text {
		text-align: center;
		font-size: 0.875rem;
		color: #5f6368;
		padding-bottom: 1rem;
	}

	.pagination-button {
		padding: 0.75rem 1.5rem;
		font-size: 1rem;
		font-weight: 600;
		color: #fff;
		background-color: #1a73e8;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		transition: background-color 0.2s ease;
	}

	.pagination-button:hover:not(:disabled) {
		background-color: #1557b0;
	}

	.pagination-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
		background-color: #9e9e9e;
	}

	.page-numbers {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.page-number {
		min-width: 40px;
		height: 40px;
		padding: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #5f6368;
		background-color: #f8f9fa;
		border: 1px solid #e0e0e0;
		border-radius: 6px;
		cursor: pointer;
		transition: all 0.2s ease;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.page-number:hover:not(.active) {
		background-color: #e8f0fe;
		border-color: #1a73e8;
		color: #1a73e8;
	}

	.page-number.active {
		background-color: #1a73e8;
		color: white;
		border-color: #1a73e8;
		font-weight: 700;
		cursor: default;
	}

	.ellipsis {
		padding: 0 0.5rem;
		color: #5f6368;
		font-weight: 500;
	}

	@media (max-width: 768px) {
		.header h1 {
			font-size: 2rem;
			flex-direction: column;
			gap: 0.5rem;
		}

		.subtitle {
			font-size: 1rem;
		}

		.stats-cards {
			grid-template-columns: 1fr;
		}

		.stat-card {
			padding: 1.25rem;
		}

		.section-header {
			flex-direction: column;
			align-items: stretch;
		}

		.controls {
			flex-direction: column;
			width: 100%;
		}

		.filter-group,
		.sort-group {
			width: 100%;
		}

		.filter-input,
		.sort-select {
			width: 100%;
		}

		.package-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.75rem;
		}

		.releases-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 640px) {
		.pagination {
			gap: 0.75rem;
		}

		.pagination-button {
			padding: 0.5rem 1rem;
			font-size: 0.875rem;
		}

		.page-number {
			min-width: 36px;
			height: 36px;
			font-size: 0.8rem;
		}

		.page-numbers {
			gap: 0.25rem;
		}

		.header h1 {
			font-size: 1.75rem;
		}

		.logo-icon {
			width: 32px;
			height: 32px;
		}

		.stat-value {
			font-size: 1.5rem;
		}

		.stat-icon {
			font-size: 1.5rem;
		}
	}
</style>
