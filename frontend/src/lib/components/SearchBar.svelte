<script lang="ts">
	import { goto } from '$app/navigation';

	let searchQuery = '';

	const handleSubmit = (e: Event) => {
		e.preventDefault();

		if (!searchQuery.trim()) {
			return;
		}

		// Navigate to the package detail page
		const encodedPath = encodeURIComponent(searchQuery.trim());
		goto(`/packages/${encodedPath}`);

		// Clear the search after navigation
		searchQuery = '';
	};
</script>

<form class="search-form" on:submit={handleSubmit}>
	<div class="search-container">
		<input
			type="text"
			class="search-input"
			placeholder="Search for a Go package (e.g., github.com/gin-gonic/gin)"
			bind:value={searchQuery}
		/>
		<button type="submit" class="search-button">
			Search
		</button>
	</div>
</form>

<style>
	.search-form {
		width: 100%;
		max-width: 800px;
		margin: 0 auto 2rem;
	}

	.search-container {
		display: flex;
		gap: 0.5rem;
	}

	.search-input {
		flex: 1;
		padding: 0.75rem 1rem;
		font-size: 1rem;
		border: 2px solid #e0e0e0;
		border-radius: 8px;
		outline: none;
		transition: border-color 0.2s ease;
	}

	.search-input:focus {
		border-color: #1a73e8;
	}

	.search-button {
		padding: 0.75rem 1.5rem;
		font-size: 1rem;
		font-weight: 600;
		color: #fff;
		background-color: #1a73e8;
		border: none;
		border-radius: 8px;
		cursor: pointer;
		transition: background-color 0.2s ease;
	}

	.search-button:hover {
		background-color: #1557b0;
	}

	.search-button:active {
		background-color: #0d47a1;
	}
</style>
