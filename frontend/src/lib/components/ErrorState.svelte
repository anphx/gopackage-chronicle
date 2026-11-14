<script lang="ts">
	let {
		title,
		message,
		onretry
	}: { title: string; message?: string; onretry?: () => void } = $props();
</script>

<div class="error" role="alert">
	<span class="title">{title}</span>
	{#if message}<span class="message">{message}</span>{/if}
	{#if onretry}
		<button type="button" onclick={onretry}>Try again</button>
	{/if}
</div>

<style>
	.error {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
		padding: 40px 16px;
		text-align: center;
	}

	.title {
		font-size: 18px;
		font-weight: 600;
	}

	.message {
		max-width: 460px;
		font-size: 14px;
		color: var(--text-subtle);
	}

	button {
		height: 44px;
		padding: 0 20px;
		border: 0;
		border-radius: 999px;
		background: var(--ink);
		color: var(--white);
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
	}

	button:hover {
		background: var(--ember);
		color: var(--ink);
	}
</style>
