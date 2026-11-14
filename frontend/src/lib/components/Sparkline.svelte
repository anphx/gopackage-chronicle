<script lang="ts">
	let {
		values,
		width = 150,
		height = 36,
		label
	}: { values: number[]; width?: number; height?: number; label: string } = $props();

	const geometry = $derived.by(() => {
		const max = Math.max(1, ...values);
		const step = values.length > 1 ? (width - 8) / (values.length - 1) : 0;
		const points = values.map((v, i) => `${(4 + i * step).toFixed(1)},${(height - 4 - (v / max) * (height - 10)).toFixed(1)}`);
		return {
			line: points.join(' '),
			area: `4,${height - 2} ${points.join(' ')} ${(4 + (values.length - 1) * step).toFixed(1)},${height - 2}`
		};
	});
</script>

<svg {width} {height} viewBox="0 0 {width} {height}" role="img" aria-label={label}>
	<polygon points={geometry.area} fill="var(--sand)" fill-opacity="0.35" />
	<polyline
		points={geometry.line}
		fill="none"
		stroke="var(--ember)"
		stroke-width="2"
		stroke-linejoin="round"
		stroke-linecap="round"
	/>
</svg>
