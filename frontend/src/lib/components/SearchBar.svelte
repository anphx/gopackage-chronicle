<script lang="ts">
	import { onDestroy, tick } from 'svelte';
	import { goto } from '$app/navigation';
	import { searchPackages, type Package } from '$lib/api/client';
	import { packageHref } from '$lib/format';

	let query = $state('');
	let input: HTMLInputElement | undefined = $state();
	let suggestions = $state<Package[]>([]);
	let activeIndex = $state(-1);
	let suggestionsOpen = $state(false);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let debounce: ReturnType<typeof setTimeout> | undefined;
	let requestId = 0;

	const openPackage = (path: string) => {
		clearTimeout(debounce);
		requestId++;
		suggestionsOpen = false;
		query = '';
		suggestions = [];
		input?.blur();
		goto(packageHref(path));
	};

	const handleSubmit = (e: SubmitEvent) => {
		e.preventDefault();
		const path = query.trim();
		if (!path) return;
		openPackage(activeIndex >= 0 ? suggestions[activeIndex].path : path);
	};

	const handleInput = (value: string) => {
		query = value;
		activeIndex = -1;
		error = null;
		clearTimeout(debounce);
		const id = ++requestId;
		const term = value.trim();
		if (term.length < 2) {
			suggestionsOpen = false;
			suggestions = [];
			loading = false;
			return;
		}

		suggestionsOpen = true;
		loading = true;
		suggestions = [];
		debounce = setTimeout(async () => {
			try {
				const results = await searchPackages(term);
				if (id === requestId) suggestions = results;
			} catch (e) {
				if (id === requestId) {
					error = e instanceof Error ? e.message : 'Search failed';
				}
			} finally {
				if (id === requestId) loading = false;
			}
		}, 180);
	};

	const handleSuggestionKeydown = (e: KeyboardEvent) => {
		if (e.key === 'ArrowDown' && suggestions.length > 0) {
			e.preventDefault();
			suggestionsOpen = true;
			activeIndex = (activeIndex + 1) % suggestions.length;
			void scrollActiveSuggestionIntoView();
		} else if (e.key === 'ArrowUp' && suggestions.length > 0) {
			e.preventDefault();
			suggestionsOpen = true;
			activeIndex = activeIndex <= 0 ? suggestions.length - 1 : activeIndex - 1;
			void scrollActiveSuggestionIntoView();
		} else if (e.key === 'Escape') {
			suggestionsOpen = false;
			activeIndex = -1;
		}
	};

	const scrollActiveSuggestionIntoView = async () => {
		await tick();
		document.getElementById(`module-search-result-${activeIndex}`)?.scrollIntoView({ block: 'nearest' });
	};

	const handleFocusout = (e: FocusEvent) => {
		const nextTarget = e.relatedTarget as Node | null;
		const form = e.currentTarget as HTMLFormElement;
		if (!nextTarget || !form.contains(nextTarget)) {
			suggestionsOpen = false;
			activeIndex = -1;
		}
	};

	// "/" focuses the search from anywhere, unless the user is already typing somewhere.
	const handleKeydown = (e: KeyboardEvent) => {
		if (e.key !== '/' || e.metaKey || e.ctrlKey || e.altKey) return;
		const target = e.target as HTMLElement | null;
		if (target && (target.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName))) return;
		e.preventDefault();
		input?.focus();
	};

	onDestroy(() => clearTimeout(debounce));
</script>

<svelte:window onkeydown={handleKeydown} />

<form class="search" role="search" onsubmit={handleSubmit} onfocusout={handleFocusout}>
	<label class="field">
		<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
			<circle cx="11" cy="11" r="7" />
			<path d="M20 20l-4-4" />
		</svg>
		<span class="visually-hidden">Search indexed modules</span>
		<input
			bind:this={input}
			value={query}
			oninput={(e) => handleInput(e.currentTarget.value)}
			onkeydown={handleSuggestionKeydown}
			onfocus={() => { if (query.trim().length >= 2) suggestionsOpen = true; }}
			type="search"
			placeholder="Search modules, e.g. cobra"
			autocomplete="off"
			spellcheck="false"
			role="combobox"
			aria-autocomplete="list"
			aria-expanded={suggestionsOpen}
			aria-controls="module-search-results"
			aria-activedescendant={activeIndex >= 0 ? `module-search-result-${activeIndex}` : undefined}
		/>
		<kbd aria-hidden="true">/</kbd>
	</label>
	{#if suggestionsOpen}
		<div class="results">
			{#if loading}
				<div class="status" role="status">Searching modules…</div>
			{:else if error}
				<div class="status error" role="alert">{error}</div>
			{:else if suggestions.length === 0}
				<div class="status" role="status">No matching modules</div>
			{:else}
				<ul id="module-search-results" role="listbox" aria-label="Module search results">
					{#each suggestions as item, index (item.id)}
						<li>
							<button
								id="module-search-result-{index}"
								type="button"
								role="option"
								aria-selected={activeIndex === index}
								class:active={activeIndex === index}
								onmousedown={(e) => e.preventDefault()}
								onclick={() => openPackage(item.path)}
							>
								{item.path}
							</button>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	{/if}
</form>

<style>
	.search {
		position: relative;
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

	.results {
		position: absolute;
		z-index: 20;
		top: calc(100% + 8px);
		right: 0;
		width: min(440px, calc(100vw - 24px));
		overflow: hidden;
		border: 1px solid var(--line-strong);
		border-radius: 16px;
		background: var(--white);
		box-shadow: 0 12px 32px rgba(0, 0, 0, 0.18);
		color: var(--ink);
	}

	.results ul {
		max-height: 320px;
		overflow-y: auto;
		margin: 0;
		padding: 6px;
		list-style: none;
	}

	.results button {
		width: 100%;
		overflow: hidden;
		padding: 12px 14px;
		border: 0;
		border-radius: 10px;
		background: transparent;
		color: inherit;
		font-family: var(--font-mono);
		font-size: 13px;
		text-align: left;
		text-overflow: ellipsis;
		white-space: nowrap;
		cursor: pointer;
	}

	.results button:hover,
	.results button:focus-visible {
		background: var(--surface-hover);
	}

	.results button.active {
		background: var(--ember);
		color: var(--ink);
		font-weight: 600;
	}

	.status {
		padding: 14px 16px;
		color: var(--text-subtle);
		font-size: 14px;
	}

	.status.error {
		color: #a3311d;
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

		.results {
			left: 0;
			right: auto;
			width: 100%;
		}

		kbd {
			display: none;
		}
	}
</style>
