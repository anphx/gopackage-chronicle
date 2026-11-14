<script lang="ts">
	import type { ReleaseWithPackage } from '$lib/api/client';
	import { splitModulePath, timeAgo } from '$lib/format';
	import { KIND_LABEL, releaseKind } from '$lib/semver';
	import Skeleton from '$lib/components/Skeleton.svelte';

	let { releases, loading = false }: { releases: ReleaseWithPackage[]; loading?: boolean } = $props();

	const items = $derived(
		releases.map((r) => {
			const kind = releaseKind(r.version);
			return { ...r, ...splitModulePath(r.package_path), kind, kindLabel: KIND_LABEL[kind] };
		})
	);
</script>

<div class="feed card-dark">
	<div class="head">
		<h2 class="section-title">Just released</h2>
		<span class="live"><span class="dot"></span>Live</span>
	</div>

	<ul>
		{#if loading}
			{#each Array(7) as _, i (i)}
				<li><Skeleton height="34px" /></li>
			{/each}
		{:else if items.length === 0}
			<li class="empty">No releases indexed yet.</li>
		{:else}
			{#each items as item (item.id)}
				<li>
					<span class="kind {item.kind}" title={item.kindLabel}><span class="visually-hidden">{item.kindLabel}</span></span>
					<a class="name" href="/packages/{encodeURIComponent(item.package_path)}">
						<span class="pkg">{item.name}</span>
						<span class="prefix">{item.prefix}</span>
					</a>
					<span class="meta">
						<span class="version">{item.version}</span>
						<span class="ago">{timeAgo(item.released_at)}</span>
					</span>
				</li>
			{/each}
		{/if}
	</ul>

	<a href="#modules" class="more">Browse recently updated modules <span>↗</span></a>
</div>

<style>
	.feed {
		display: flex;
		flex-direction: column;
		gap: 18px;
		min-width: 0;
	}

	.head {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
	}

	.live {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
		color: var(--fog);
	}

	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--ember);
	}

	ul {
		display: flex;
		flex-direction: column;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	li {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 0;
		border-top: 1px solid var(--graphite-line);
	}

	.empty {
		color: var(--fog);
		font-size: 14px;
	}

	.kind {
		width: 10px;
		height: 10px;
		flex-shrink: 0;
		border-radius: 3px;
	}

	.kind.major {
		background: var(--white);
	}

	.kind.minor {
		background: var(--ember);
	}

	.kind.patch {
		background: var(--sand);
	}

	.kind.prerelease {
		background: repeating-linear-gradient(45deg, #8f8f8f 0 2px, #3a3a3a 2px 4px);
	}

	.name {
		flex: 1 1 auto;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
		color: var(--white);
		text-decoration: none;
	}

	.name:hover .pkg {
		color: var(--sand);
	}

	.pkg,
	.prefix {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.pkg {
		font-size: 15px;
		font-weight: 600;
	}

	.prefix {
		font-size: 12px;
		color: var(--text-on-dark-muted);
	}

	.meta {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 2px;
		flex-shrink: 0;
	}

	.version {
		max-width: 150px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-family: var(--font-mono);
		font-size: 13px;
		color: var(--sand);
	}

	.ago {
		font-size: 12px;
		color: var(--text-on-dark-muted);
	}

	.more {
		margin-top: auto;
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 14px 18px;
		border-radius: var(--radius-md);
		background: var(--graphite);
		color: var(--white);
		text-decoration: none;
		font-size: 15px;
		font-weight: 600;
	}

	.more span {
		color: var(--ember);
	}

	.more:hover {
		color: var(--sand);
	}
</style>
