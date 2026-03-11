<script lang="ts">
	import type { Release } from '$lib/api/client';

	export let releases: Release[];
	export let packagePath: string;

	// Format date for display
	const formatDate = (dateString: string): string => {
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	};

	// Sort releases by date (oldest to newest)
	const sortedReleases = [...releases].sort((a, b) =>
		new Date(a.released_at).getTime() - new Date(b.released_at).getTime()
	);

	// Calculate SVG viewBox dimensions
	const chartWidth = 800;
	const chartHeight = 300;
	const padding = { top: 20, right: 20, bottom: 60, left: 60 };
	const plotWidth = chartWidth - padding.left - padding.right;
	const plotHeight = chartHeight - padding.top - padding.bottom;

	// Get date range
	const minDate = sortedReleases.length > 0 ? new Date(sortedReleases[0].released_at).getTime() : 0;
	const maxDate = sortedReleases.length > 0 ? new Date(sortedReleases[sortedReleases.length - 1].released_at).getTime() : 1;
	const dateRange = maxDate - minDate || 1;

	// Generate points for the line chart
	const points = sortedReleases.map((release, index) => {
		const timestamp = new Date(release.released_at).getTime();
		const x = padding.left + ((timestamp - minDate) / dateRange) * plotWidth;
		const y = padding.top + plotHeight - (index / Math.max(sortedReleases.length - 1, 1)) * plotHeight;
		return { x, y, release };
	});

	// Generate path string for the line
	const linePath = points.map((p, i) =>
		`${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`
	).join(' ');

	// Generate Y-axis labels (version numbers)
	const numYLabels = Math.min(6, sortedReleases.length);
	const yLabels = Array.from({ length: numYLabels }, (_, i) => {
		const index = Math.floor(i * (sortedReleases.length - 1) / Math.max(numYLabels - 1, 1));
		const release = sortedReleases[index];
		const y = padding.top + plotHeight - (index / Math.max(sortedReleases.length - 1, 1)) * plotHeight;
		return { y, label: `v${index + 1}`, version: release.version };
	});

	// Generate X-axis labels (dates)
	const numXLabels = Math.min(6, sortedReleases.length);
	const xLabels = Array.from({ length: numXLabels }, (_, i) => {
		const index = Math.floor(i * (sortedReleases.length - 1) / Math.max(numXLabels - 1, 1));
		const release = sortedReleases[index];
		const timestamp = new Date(release.released_at).getTime();
		const x = padding.left + ((timestamp - minDate) / dateRange) * plotWidth;
		const date = new Date(release.released_at);
		const label = date.toLocaleDateString('en-US', { month: 'short', year: 'numeric' });
		return { x, label };
	});

	// Detect GitHub/GitLab repository from package path
	const getRepoURL = (path: string): string | null => {
		if (path.startsWith('github.com/')) {
			const parts = path.replace('github.com/', '').split('/');
			if (parts.length >= 2) {
				return `https://github.com/${parts[0]}/${parts[1]}`;
			}
		} else if (path.startsWith('gitlab.com/')) {
			const parts = path.replace('gitlab.com/', '').split('/');
			if (parts.length >= 2) {
				return `https://gitlab.com/${parts[0]}/${parts[1]}`;
			}
		}
		return null;
	};

	const repoURL = getRepoURL(packagePath);
</script>

<div class="release-graph">
	{#if repoURL}
		<div class="repo-link-container">
			<a href={repoURL} target="_blank" rel="noopener noreferrer" class="repo-link">
				<svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
					<path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
				</svg>
				View on {packagePath.startsWith('github.com/') ? 'GitHub' : 'GitLab'}
			</a>
		</div>
	{/if}

	<div class="graph-container">
		<div class="graph-header">
			<h3>Release Timeline</h3>
			<p class="graph-subtitle">{releases.length} releases total</p>
		</div>

		<div class="chart-wrapper">
			<svg viewBox="0 0 {chartWidth} {chartHeight}" class="line-chart">
				<!-- Grid lines -->
				<g class="grid">
					{#each yLabels as { y }}
						<line
							x1={padding.left}
							y1={y}
							x2={chartWidth - padding.right}
							y2={y}
							stroke="#e0e0e0"
							stroke-width="1"
						/>
					{/each}
				</g>

				<!-- X-axis -->
				<line
					x1={padding.left}
					y1={chartHeight - padding.bottom}
					x2={chartWidth - padding.right}
					y2={chartHeight - padding.bottom}
					stroke="#5f6368"
					stroke-width="2"
				/>

				<!-- Y-axis -->
				<line
					x1={padding.left}
					y1={padding.top}
					x2={padding.left}
					y2={chartHeight - padding.bottom}
					stroke="#5f6368"
					stroke-width="2"
				/>

				<!-- Line path -->
				{#if points.length > 0}
					<path
						d={linePath}
						fill="none"
						stroke="#1a73e8"
						stroke-width="2"
						class="release-line"
					/>
				{/if}

				<!-- Data points -->
				{#each points as point}
					<circle
						cx={point.x}
						cy={point.y}
						r="4"
						fill="#1a73e8"
						stroke="white"
						stroke-width="2"
						class="data-point"
					>
						<title>{point.release.version} - {formatDate(point.release.released_at)}</title>
					</circle>
				{/each}

				<!-- Y-axis labels -->
				{#each yLabels as { y, label }}
					<text
						x={padding.left - 10}
						y={y}
						text-anchor="end"
						dominant-baseline="middle"
						class="axis-label"
					>
						{label}
					</text>
				{/each}

				<!-- X-axis labels -->
				{#each xLabels as { x, label }}
					<text
						x={x}
						y={chartHeight - padding.bottom + 20}
						text-anchor="middle"
						class="axis-label"
					>
						{label}
					</text>
				{/each}

				<!-- Axis titles -->
				<text
					x={chartWidth / 2}
					y={chartHeight - 10}
					text-anchor="middle"
					class="axis-title"
				>
					Release Date
				</text>
				<text
					x={-chartHeight / 2}
					y={15}
					text-anchor="middle"
					transform="rotate(-90)"
					class="axis-title"
				>
					Version Count
				</text>
			</svg>
		</div>
	</div>

	<div class="releases-list">
		<h3>All Releases</h3>
		<div class="releases-grid">
			{#each releases as release}
				<div class="release-card">
					<div class="release-header">
						<span class="version-badge">{release.version}</span>
						<span class="release-date">{formatDate(release.released_at)}</span>
					</div>
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.release-graph {
		margin: 2rem 0;
	}

	.repo-link-container {
		margin-bottom: 2rem;
		padding: 1rem;
		background: #f8f9fa;
		border-radius: 8px;
		text-align: center;
	}

	.repo-link {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		background: #1a73e8;
		color: white;
		text-decoration: none;
		border-radius: 6px;
		font-weight: 600;
		transition: background-color 0.2s ease;
	}

	.repo-link:hover {
		background: #1557b0;
	}

	.graph-container {
		background: #fff;
		border: 1px solid #e0e0e0;
		border-radius: 8px;
		padding: 1.5rem;
		margin-bottom: 2rem;
	}

	.graph-header {
		margin-bottom: 2rem;
	}

	.graph-header h3 {
		margin: 0 0 0.25rem 0;
		font-size: 1.25rem;
		color: #202124;
	}

	.graph-subtitle {
		margin: 0;
		font-size: 0.875rem;
		color: #5f6368;
	}

	.chart-wrapper {
		width: 100%;
		overflow-x: auto;
		padding: 1rem 0;
	}

	.line-chart {
		width: 100%;
		height: auto;
		min-width: 600px;
	}

	.release-line {
		filter: drop-shadow(0 2px 4px rgba(26, 115, 232, 0.2));
	}

	.data-point {
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.data-point:hover {
		r: 6;
		fill: #1557b0;
	}

	.axis-label {
		font-size: 12px;
		fill: #5f6368;
		font-family: system-ui, -apple-system, sans-serif;
	}

	.axis-title {
		font-size: 14px;
		fill: #202124;
		font-weight: 500;
		font-family: system-ui, -apple-system, sans-serif;
	}

	.grid line {
		opacity: 0.5;
	}

	.releases-list h3 {
		margin: 0 0 1rem 0;
		font-size: 1.25rem;
		color: #202124;
	}

	.releases-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1rem;
	}

	.release-card {
		background: #f8f9fa;
		border: 1px solid #e0e0e0;
		border-radius: 8px;
		padding: 1rem;
		transition: all 0.2s ease;
	}

	.release-card:hover {
		border-color: #1a73e8;
		box-shadow: 0 2px 8px rgba(26, 115, 232, 0.1);
	}

	.release-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.75rem;
	}

	.version-badge {
		font-family: 'Courier New', monospace;
		font-weight: 600;
		color: #202124;
		background: white;
		padding: 0.25rem 0.75rem;
		border-radius: 4px;
		font-size: 0.875rem;
	}

	.release-date {
		font-size: 0.75rem;
		color: #5f6368;
	}

	/* Responsive adjustments */
	@media (max-width: 768px) {
		.releases-grid {
			grid-template-columns: 1fr;
		}

		.axis-label {
			font-size: 10px;
		}

		.axis-title {
			font-size: 12px;
		}
	}
</style>
