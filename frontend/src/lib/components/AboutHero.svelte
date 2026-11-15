<!--
	The about page hero: the home hero's construction grid, inverted onto Ink. Same 90px cells;
	the mark is two cells wide and its top and bottom sit on the dashed guides.
-->
<section class="hero">
	<div class="art" aria-hidden="true">
		<div class="grid"></div>
		<div class="guide v text-edge"></div>
		<div class="guide h top"></div>
		<div class="guide h bottom"></div>
		<svg class="mark" width="180" height="180" viewBox="0 0 270 270">
			<circle cx="135" cy="135" r="135" fill="#D2622A" />
			<path
				d="M-30 78 C 30 78, 70 168, 135 168 S 228 64, 300 64 S 370 110, 420 110"
				fill="none"
				stroke="#0B0B0B"
				stroke-width="44"
				stroke-linecap="round"
			/>
		</svg>
	</div>

	<span class="pill">↗ About go/chronicle</span>
	<h1>A logbook for the Go ecosystem<span class="accent">↙</span></h1>
	<p>
		go/chronicle records every version of every public Go module the moment the Go module mirror sees it, so you
		can look back at how any module has shipped: what changed, how big the change was, and how often it happens.
	</p>
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
		background: var(--ink);
		color: var(--white);
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
		border-bottom: 1px solid rgba(255, 255, 255, 0.12);
		background-image:
			linear-gradient(rgba(255, 255, 255, 0.12) 1px, transparent 1px),
			linear-gradient(90deg, rgba(255, 255, 255, 0.12) 1px, transparent 1px);
		background-size: 90px 90px;
	}

	.guide {
		position: absolute;
		border: 0 dashed rgba(210, 98, 42, 0.7);
	}

	.guide.v {
		top: 12px;
		bottom: 12px;
		border-left-width: 1px;
	}

	.guide.h {
		left: 12px;
		width: 720px;
		border-top-width: 1px;
	}

	.text-edge {
		left: 76px;
	}

	.top {
		top: 126px;
	}

	.bottom {
		top: 306px;
	}

	.mark {
		position: absolute;
		left: 580px;
		top: 126px;
		overflow: visible;
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
		max-width: 470px;
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
		max-width: 420px;
		margin: 0;
		font-size: 16px;
		line-height: 1.5;
		color: var(--fog);
	}

	/* The mark ends 760px from the card's left edge (about 680px inside the padding);
	   below that keep the grid and the text guide only. */
	@container (max-width: 679px) {
		.mark,
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
