<div class="skeleton-release">
	<div class="skeleton-content">
		<div class="skeleton-package"></div>
		<div class="skeleton-info">
			<div class="skeleton-version"></div>
			<div class="skeleton-date"></div>
		</div>
	</div>
</div>

<style>
	.skeleton-release {
		position: relative;
		padding-left: 2rem;
		padding-bottom: 1.5rem;
		border-left: 2px solid #e0e0e0;
	}

	.skeleton-release::before {
		content: '';
		position: absolute;
		left: -6px;
		top: 0;
		width: 10px;
		height: 10px;
		border-radius: 50%;
		background-color: #e0e0e0;
		border: 2px solid #fff;
	}

	.skeleton-content {
		background: #f8f9fa;
		border-radius: 6px;
		padding: 0.75rem 1rem;
	}

	.skeleton-package {
		width: 60%;
		height: 1.2rem;
		background: linear-gradient(90deg, #e0e0e0 25%, #f0f0f0 50%, #e0e0e0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.5s infinite;
		border-radius: 4px;
		margin-bottom: 0.5rem;
	}

	.skeleton-info {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
	}

	.skeleton-version {
		width: 25%;
		height: 1rem;
		background: linear-gradient(90deg, #e0e0e0 25%, #f0f0f0 50%, #e0e0e0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.5s infinite;
		border-radius: 4px;
	}

	.skeleton-date {
		width: 35%;
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
