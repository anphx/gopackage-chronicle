<script lang="ts">
	let {
		width = '100%',
		height = '16px',
		radius = '10px'
	}: { width?: string; height?: string; radius?: string } = $props();
</script>

<span class="skeleton" style:width style:height style:border-radius={radius} aria-hidden="true"></span>

<style>
	.skeleton {
		display: block;
		background: linear-gradient(90deg, rgba(11, 11, 11, 0.06) 25%, rgba(255, 255, 255, 0.5) 50%, rgba(11, 11, 11, 0.06) 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite linear;
	}

	@keyframes shimmer {
		from {
			background-position: 200% 0;
		}
		to {
			background-position: -200% 0;
		}
	}
</style>
