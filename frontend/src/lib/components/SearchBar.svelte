<script lang="ts">
	import { goto } from '$app/navigation';

	let query = $state('');
	let input: HTMLInputElement | undefined = $state();

	const handleSubmit = (e: SubmitEvent) => {
		e.preventDefault();
		const path = query.trim();
		if (!path) return;
		goto(`/packages/${encodeURIComponent(path)}`);
		query = '';
		input?.blur();
	};

	// "/" focuses the search from anywhere, unless the user is already typing somewhere.
	const handleKeydown = (e: KeyboardEvent) => {
		if (e.key !== '/' || e.metaKey || e.ctrlKey || e.altKey) return;
		const target = e.target as HTMLElement | null;
		if (target && (target.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName))) return;
		e.preventDefault();
		input?.focus();
	};
</script>

<svelte:window onkeydown={handleKeydown} />

<form class="search" role="search" onsubmit={handleSubmit}>
	<label class="field">
		<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
			<circle cx="11" cy="11" r="7" />
			<path d="M20 20l-4-4" />
		</svg>
		<span class="visually-hidden">Open a module by its path</span>
		<input
			bind:this={input}
			bind:value={query}
			type="search"
			placeholder="Open a module, e.g. github.com/spf13/cobra"
			autocomplete="off"
			spellcheck="false"
		/>
		<kbd aria-hidden="true">/</kbd>
	</label>
</form>

<style>
	.search {
		flex: 0 1 380px;
		min-width: 220px;
	}

	.field {
		display: flex;
		align-items: center;
		gap: 10px;
		height: 48px;
		padding: 0 8px 0 18px;
		border-radius: 999px;
		background: var(--graphite);
		color: var(--fog);
	}

	.field:focus-within {
		box-shadow: 0 0 0 2px var(--ember);
	}

	input {
		flex: 1 1 auto;
		min-width: 0;
		border: 0;
		outline: none;
		background: transparent;
		color: var(--white);
		font-size: 15px;
	}

	input::placeholder {
		color: var(--text-on-dark-muted);
	}

	kbd {
		font-size: 12px;
		color: var(--fog);
		border: 1px solid #3a3a3a;
		border-radius: 8px;
		padding: 4px 8px;
	}

	@media (max-width: 720px) {
		.search {
			flex: 1 1 100%;
		}

		kbd {
			display: none;
		}
	}
</style>
