<div class="skeleton-card">
	<div class="skeleton-title"></div>
	<div class="skeleton-meta"></div>
</div>

<style>
	.skeleton-card {
		border: 1px solid #e0e0e0;
		border-radius: 8px;
		padding: 1rem;
		margin-bottom: 0.75rem;
	}

	.skeleton-title {
		width: 70%;
		height: 1.5rem;
		background: linear-gradient(90deg, #e0e0e0 25%, #f0f0f0 50%, #e0e0e0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.5s infinite;
		border-radius: 4px;
		margin-bottom: 0.5rem;
	}

	.skeleton-meta {
		width: 40%;
		height: 1rem;
		background: linear-gradient(90deg, #e0e0e0 25%, #f0f0f0 50%, #e0e0e0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.5s infinite;
		border-radius: 4px;
	}

	@keyframes shimmer {
		0% {
			background-position: -200% 0;
		}
		100% {
			background-position: 200% 0;
		}
	}
</style>
