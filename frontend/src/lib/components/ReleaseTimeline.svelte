<script lang="ts">
	import type { Release, ReleaseWithPackage } from '$lib/api/client';

	// Accept either Release or ReleaseWithPackage
	export let releases: (Release | ReleaseWithPackage)[];
	export let showPackagePath: boolean = false;

	// Format the released_at timestamp
	const formatDate = (dateString: string): string => {
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	};

	// Type guard to check if release has package_path
	const hasPackagePath = (release: Release | ReleaseWithPackage): release is ReleaseWithPackage => {
		return 'package_path' in release;
	};
</script>

<div class="timeline">
	{#if releases.length === 0}
		<p class="no-releases">No releases found</p>
	{:else}
		<ul class="release-list">
			{#each releases as release}
				<li class="release-item">
					<div class="release-content">
						{#if showPackagePath && hasPackagePath(release)}
							<a href="/packages/{encodeURIComponent(release.package_path)}" class="package-link">
								{release.package_path}
							</a>
						{/if}
						<div class="release-info">
							<span class="version">{release.version}</span>
							<span class="date">{formatDate(release.released_at)}</span>
						</div>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.timeline {
		margin: 1rem 0;
	}

	.no-releases {
		color: #5f6368;
		font-style: italic;
		text-align: center;
		padding: 2rem;
	}

	.release-list {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.release-item {
		position: relative;
		padding-left: 2rem;
		padding-bottom: 1.5rem;
		border-left: 2px solid #e0e0e0;
	}

	.release-item:last-child {
		border-left: none;
	}

	.release-item::before {
		content: '';
		position: absolute;
		left: -6px;
		top: 0;
		width: 10px;
		height: 10px;
		border-radius: 50%;
		background-color: #1a73e8;
		border: 2px solid #fff;
	}

	.release-content {
		background: #f8f9fa;
		border-radius: 6px;
		padding: 0.75rem 1rem;
	}

	.package-link {
		display: block;
		font-weight: 600;
		color: #1a73e8;
		text-decoration: none;
		margin-bottom: 0.5rem;
	}

	.package-link:hover {
		text-decoration: underline;
	}

	.release-info {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
	}

	.version {
		font-family: 'Courier New', monospace;
		font-weight: 600;
		color: #202124;
	}

	.date {
		font-size: 0.875rem;
		color: #5f6368;
		white-space: nowrap;
	}
</style>
