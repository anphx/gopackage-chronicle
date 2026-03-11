<script lang="ts">
	import { getPackageDetail, type PackageDetail } from '$lib/api/client';
	import ReleaseGraph from '$lib/components/ReleaseGraph.svelte';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import SkeletonRelease from '$lib/components/SkeletonRelease.svelte';

	export let data: { packagePath: string };

	let packageDetail: PackageDetail | null = null;
	let loading = true;
	let error: string | null = null;

	// Format the created_at timestamp
	const formatDate = (dateString: string): string => {
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	};

	const loadPackageDetail = async (path: string) => {
		loading = true;
		error = null;
		packageDetail = null;

		try {
			const apiData = await getPackageDetail(path);
			packageDetail = apiData;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load package details';
			console.error('Error loading package details:', e);
		} finally {
			loading = false;
		}
	};

	// Reactive statement to load data when packagePath changes
	$: if (data.packagePath) {
		loadPackageDetail(data.packagePath);
	}
</script>

<svelte:head>
	<title>{data.packagePath || 'Package Details'} - Go Package Chronicles</title>
</svelte:head>

<div class="container">
	<SearchBar />

	<nav class="breadcrumb">
		<a href="/">Home</a>
		<span class="separator">/</span>
		<span class="current">Package Details</span>
	</nav>

	{#if loading}
		<!-- Skeleton loading state -->
		<div class="skeleton-container">
			<div class="skeleton-header">
				<div class="skeleton-title"></div>
				<div class="skeleton-meta"></div>
			</div>

			<div class="skeleton-stats">
				{#each Array(3) as _}
					<div class="skeleton-stat"></div>
				{/each}
			</div>

			<div class="skeleton-section">
				<div class="skeleton-section-title"></div>
				<div class="skeleton-releases">
					{#each Array(5) as _}
						<SkeletonRelease />
					{/each}
				</div>
			</div>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error">Error: {error}</p>
			<p class="error-hint">
				Make sure the package path is correct and the package has been indexed.
			</p>
			<button class="retry-button" on:click={() => loadPackageDetail(data.packagePath)}>
				Try Again
			</button>
		</div>
	{:else if packageDetail}
		<section class="package-section">
			<header class="package-header">
				<h1 class="package-name">{packageDetail.package.path}</h1>
				<p class="package-meta">
					Indexed on {formatDate(packageDetail.package.created_at)}
				</p>
			</header>

			<div class="stats">
				<div class="stat-item">
					<span class="stat-label">Total Releases</span>
					<span class="stat-value">{packageDetail.releases.length}</span>
				</div>
				{#if packageDetail.releases.length > 0}
					<div class="stat-item">
						<span class="stat-label">Latest Version</span>
						<span class="stat-value">{packageDetail.releases[0].version}</span>
					</div>
					<div class="stat-item">
						<span class="stat-label">Latest Release</span>
						<span class="stat-value">{formatDate(packageDetail.releases[0].released_at)}</span>
					</div>
				{/if}
			</div>

			{#if packageDetail.releases.length === 0}
				<p class="no-releases">No releases found for this package.</p>
			{:else}
				<ReleaseGraph releases={packageDetail.releases} packagePath={packageDetail.package.path} />
			{/if}
		</section>
	{/if}
</div>

<style>
	.container {
		max-width: 900px;
		margin: 0 auto;
		padding: 2rem 1rem;
	}

	.breadcrumb {
		margin-bottom: 2rem;
		font-size: 0.875rem;
		color: #5f6368;
	}

	.breadcrumb a {
		color: #1a73e8;
		text-decoration: none;
	}

	.breadcrumb a:hover {
		text-decoration: underline;
	}

	.separator {
		margin: 0 0.5rem;
	}

	.current {
		color: #5f6368;
	}

	/* Skeleton loading styles */
	.skeleton-container {
		margin-top: 2rem;
	}

	.skeleton-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 2px solid #e0e0e0;
	}

	.skeleton-title {
		width: 60%;
		height: 2.5rem;
		background: linear-gradient(90deg, #e0e0e0 25%, #f0f0f0 50%, #e0e0e0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.5s infinite;
		border-radius: 4px;
		margin-bottom: 0.75rem;
	}

	.skeleton-meta {
		width: 30%;
		height: 1rem;
		background: linear-gradient(90deg, #e0e0e0 25%, #f0f0f0 50%, #e0e0e0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.5s infinite;
		border-radius: 4px;
	}

	.skeleton-stats {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 3rem;
	}

	.skeleton-stat {
		height: 5rem;
		background: linear-gradient(90deg, #e0e0e0 25%, #f0f0f0 50%, #e0e0e0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.5s infinite;
		border-radius: 8px;
	}

	.skeleton-section {
		margin-top: 2rem;
	}

	.skeleton-section-title {
		width: 25%;
		height: 1.5rem;
		background: linear-gradient(90deg, #e0e0e0 25%, #f0f0f0 50%, #e0e0e0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.5s infinite;
		border-radius: 4px;
		margin-bottom: 1.5rem;
	}

	.skeleton-releases {
		margin-top: 1rem;
	}

	@keyframes shimmer {
		0% {
			background-position: -200% 0;
		}
		100% {
			background-position: 200% 0;
		}
	}

	.error-container {
		text-align: center;
		padding: 3rem 1rem;
	}

	.error {
		color: #d93025;
		font-size: 1.125rem;
		margin-bottom: 0.5rem;
	}

	.error-hint {
		color: #5f6368;
		font-size: 0.875rem;
		margin-bottom: 1.5rem;
	}

	.retry-button {
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

	.retry-button:hover {
		background-color: #1557b0;
	}

	.package-section {
		margin-top: 2rem;
	}

	.package-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 2px solid #e0e0e0;
	}

	.package-name {
		font-size: 2rem;
		margin: 0 0 0.5rem 0;
		color: #202124;
		word-break: break-all;
	}

	.package-meta {
		margin: 0;
		color: #5f6368;
		font-size: 0.875rem;
	}

	.stats {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 3rem;
	}

	.stat-item {
		background: #f8f9fa;
		border-radius: 8px;
		padding: 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.stat-label {
		font-size: 0.875rem;
		color: #5f6368;
		font-weight: 500;
	}

	.stat-value {
		font-size: 1.25rem;
		font-weight: 600;
		color: #202124;
	}

	.releases-section h2 {
		font-size: 1.5rem;
		margin: 0 0 1.5rem 0;
		color: #202124;
	}

	.no-releases {
		text-align: center;
		padding: 3rem 1rem;
		color: #5f6368;
		font-style: italic;
	}
</style>
