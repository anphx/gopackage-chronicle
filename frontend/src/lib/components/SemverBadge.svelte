<script lang="ts">
	import { KIND_LABEL, type ReleaseKind } from '$lib/semver';

	let { kind }: { kind: ReleaseKind } = $props();
</script>

<span class="badge {kind}">{KIND_LABEL[kind]}</span>

<style>
	.badge {
		display: inline-block;
		font-size: 12px;
		font-weight: 600;
		line-height: 1.3;
		padding: 4px 10px;
		border-radius: 999px;
		white-space: nowrap;
	}

	.major {
		background: var(--kind-major);
		color: var(--white);
	}

	.minor {
		background: var(--kind-minor);
		color: var(--ink);
	}

	.patch {
		background: var(--kind-patch);
		color: var(--ink);
	}

	.prerelease {
		padding: 3px 9px;
		background: var(--white);
		color: var(--text-muted);
		border: 1px dashed #6a6a6a;
	}
</style>
