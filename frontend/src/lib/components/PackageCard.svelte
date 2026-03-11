<script lang="ts">
	import type { Package } from '$lib/api/client';

	export let pkg: Package;

	// Format the created_at timestamp
	const formatDate = (dateString: string): string => {
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	};
</script>

<div class="package-card">
	<a href="/packages/{encodeURIComponent(pkg.path)}" class="package-link">
		<h3 class="package-path">{pkg.path}</h3>
		<p class="package-meta">
			Indexed on {formatDate(pkg.created_at)}
		</p>
	</a>
</div>

<style>
	.package-card {
		border: 1px solid #e0e0e0;
		border-radius: 8px;
		padding: 1rem;
		margin-bottom: 0.75rem;
		transition: box-shadow 0.2s ease;
	}

	.package-card:hover {
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
	}

	.package-link {
		text-decoration: none;
		color: inherit;
		display: block;
	}

	.package-path {
		margin: 0 0 0.5rem 0;
		font-size: 1.1rem;
		font-weight: 600;
		color: #1a73e8;
	}

	.package-meta {
		margin: 0;
		font-size: 0.875rem;
		color: #5f6368;
	}
</style>
