<!--
	The home page hero: copy set on a construction grid. The grid, the dashed guides and the
	mark share one geometry (90px cells): the circle's centre sits on a grid intersection, the
	guides touch the circle's top, centre and bottom, and one guide runs down the text's left edge.
-->
<section class="hero">
	<div class="art" aria-hidden="true">
		<div class="grid"></div>
		<div class="guide v text-edge"></div>
		<div class="guide v mark-edge"></div>
		<div class="guide h top"></div>
		<div class="guide h middle"></div>
		<div class="guide h bottom"></div>
		<svg class="mark" width="270" height="270" viewBox="0 0 270 270">
			<circle cx="135" cy="135" r="135" fill="#D2622A" />
			<path
				d="M-30 78 C 30 78, 70 168, 135 168 S 228 64, 300 64 S 370 110, 420 110"
				fill="none"
				stroke="#ECECEA"
				stroke-width="44"
				stroke-linecap="round"
			/>
		</svg>
		<span class="caption"><span class="num">1</span>Release curve</span>
	</div>

	<span class="pill">↗ Indexed daily from index.golang.org</span>
	<h1>Every Go release, in motion<span class="accent">↙</span></h1>
	<p>Follow how the Go ecosystem ships: who released, what kind of change it was, and how fast each module moves.</p>
</section>

<style>
	.hero {
		position: relative;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: 18px;
		min-height: 432px;
		min-width: 0;
		padding: 40px 40px 56px 76px;
		border-radius: var(--radius-xl);
		border: 1px solid rgba(11, 11, 11, 0.06);
		background: #cdcdcb;
		container-type: inline-size;
	}

	.art {
		position: absolute;
		inset: 0;
		pointer-events: none;
	}

	.grid {
		position: absolute;
		left: 40px;
		top: 36px;
		right: -1px;
		height: 361px;
		border-bottom: 1px solid rgba(11, 11, 11, 0.26);
		background-image:
			linear-gradient(rgba(11, 11, 11, 0.26) 1px, transparent 1px),
			linear-gradient(90deg, rgba(11, 11, 11, 0.26) 1px, transparent 1px);
		background-size: 90px 90px;
	}

	.guide {
		position: absolute;
		border-color: rgba(210, 98, 42, 0.7);
		border-style: dashed;
		border-width: 0;
	}

	.guide.v {
		top: 12px;
		bottom: 12px;
		border-left-width: 1px;
	}

	.guide.h {
		left: 12px;
		border-top-width: 1px;
	}

	.text-edge {
		left: 76px;
	}

	.mark-edge {
		left: 535px;
	}

	.top {
		top: 81px;
		width: 560px;
	}

	.middle {
		top: 216px;
		width: 720px;
	}

	.bottom {
		top: 351px;
		width: 560px;
	}

	.mark {
		position: absolute;
		left: 535px;
		top: 81px;
		overflow: visible;
	}

	.caption {
		position: absolute;
		left: 535px;
		top: 372px;
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
		font-weight: 500;
		color: #2e2e2e;
	}

	.num {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 16px;
		height: 16px;
		border: 1px solid #2e2e2e;
		border-radius: 50%;
		font-size: 10px;
	}

	.pill,
	h1,
	p {
		position: relative;
	}

	.pill {
		align-self: flex-start;
		padding: 8px 14px;
		border-radius: var(--radius-sm);
		background: var(--ember);
		color: var(--ink);
		font-size: 14px;
		font-weight: 600;
	}

	h1 {
		max-width: 430px;
		margin: 0;
		font-size: 60px;
		font-weight: 600;
		line-height: 0.96;
		letter-spacing: -0.045em;
	}

	.accent {
		color: var(--ember);
	}

	p {
		max-width: 400px;
		margin: 0;
		font-size: 16px;
		line-height: 1.5;
		color: #2e2e2e;
	}

	/* The mark needs about 830px of hero width (about 715px inside the padding); below that
	   keep the grid and the text guide only. */
	@container (max-width: 699px) {
		.mark,
		.caption,
		.mark-edge,
		.guide.h {
			display: none;
		}
	}

	@media (max-width: 640px) {
		.hero {
			min-height: 0;
			padding: 32px 24px 40px 40px;
			border-radius: var(--radius-lg);
		}

		.grid {
			left: 16px;
			top: 16px;
			height: auto;
			bottom: 16px;
			background-size: 64px 64px;
		}

		.text-edge {
			left: 40px;
		}

		h1 {
			font-size: 44px;
		}
	}
</style>
