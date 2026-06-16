<script lang="ts">
	import { fade } from 'svelte/transition';
	import { onMount } from 'svelte';
	import { shellStore } from '$lib/stores/shell.svelte';
	import { settingsStore } from '$lib/stores/settings.svelte';

	// ──────────────────────────────────────────────────────────────────────────
	// Cometline first-run intro.
	// Aesthetic: elegant + vintage (warm cream/gold title card, film grain,
	// vignette, hairline double-rule frame) over heavy-tech motion (a warp
	// starfield and a comet that streaks in and ignites the wordmark, ringed by
	// the user's configured hero-glow color).
	//
	// The sequence is timeline-driven via requestAnimationFrame so every beat is
	// eased; text reveals layer on top with CSS. Honors prefers-reduced-motion,
	// is skippable (Esc / click), and replayable from Settings → About.
	// ──────────────────────────────────────────────────────────────────────────

	// Beat timing (ms) — the spine of the cinematic.
	const T = {
		spaceIn: 700, // deep-space + grain fade in
		comet: 1500, // comet streaks toward center and ignites
		ring: 2300, // orbital ring forms around the mark
		wordmark: 2300, // "Cometline" resolves
		tagline: 3100, // tagline types in
		hold: 4500, // hold the title card
		total: 5400 // begin graceful exit
	};

	let canvas = $state<HTMLCanvasElement | null>(null);
	let phase = $state<'run' | 'exit'>('run');

	// Reactive text reveals are driven off a single elapsed clock.
	let elapsed = $state(0);
	let reducedMotion = false;
	let raf = 0;
	let finished = false;

	const GLOW = () => settingsStore.settings.appearance.heroComposer.glowColor || '#72c0ff';

	function readCssVar(name: string, fallback: string): string {
		if (typeof window === 'undefined') return fallback;
		const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
		return v || fallback;
	}

	// Convert any CSS color to {r,g,b} via a throwaway canvas pixel.
	function toRgb(color: string): { r: number; g: number; b: number } {
		const c = document.createElement('canvas');
		c.width = c.height = 1;
		const ctx = c.getContext('2d');
		if (!ctx) return { r: 114, g: 192, b: 255 };
		ctx.fillStyle = color;
		ctx.fillRect(0, 0, 1, 1);
		const [r, g, b] = ctx.getImageData(0, 0, 1, 1).data;
		return { r, g, b };
	}

	// easeOutCubic / easeInOutCubic for cinematic deceleration.
	const easeOut = (t: number) => 1 - Math.pow(1 - t, 3);
	const easeInOut = (t: number) =>
		t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
	const clamp01 = (t: number) => Math.max(0, Math.min(1, t));
	// Normalize an absolute elapsed time into a 0..1 progress across [start,end].
	const seg = (now: number, start: number, end: number) =>
		clamp01((now - start) / (end - start));

	type Star = { x: number; y: number; z: number; px: number; py: number };

	function complete() {
		if (finished) return;
		finished = true;
		cancelAnimationFrame(raf);
		phase = 'exit';
		// Persist the "seen" flag (no-op if already seen / replay).
		void settingsStore.markIntroSeen().catch(() => {});
		// Let the fade-out transition play before unmounting.
		setTimeout(() => shellStore.closeIntro(), reducedMotion ? 0 : 760);
	}

	function skip() {
		complete();
	}

	onMount(() => {
		reducedMotion =
			typeof window !== 'undefined' &&
			window.matchMedia('(prefers-reduced-motion: reduce)').matches;

		const onKey = (e: KeyboardEvent) => {
			if (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') {
				e.preventDefault();
				skip();
			}
		};
		window.addEventListener('keydown', onKey, true);

		if (reducedMotion) {
			// Reduced motion: show the static title card briefly, then leave.
			elapsed = T.tagline + 200;
			const t = setTimeout(complete, 1800);
			return () => {
				clearTimeout(t);
				window.removeEventListener('keydown', onKey, true);
			};
		}

		const el = canvas;
		const ctx0 = el?.getContext('2d', { alpha: false }) ?? null;
		if (!el || !ctx0) {
			const t = setTimeout(complete, 1800);
			return () => {
				clearTimeout(t);
				window.removeEventListener('keydown', onKey, true);
			};
		}

		const ctx: CanvasRenderingContext2D = ctx0;
		const dpr = Math.min(window.devicePixelRatio || 1, 2);
		let W = 0;
		let H = 0;
		const resize = () => {
			W = window.innerWidth;
			H = window.innerHeight;
			el.width = Math.floor(W * dpr);
			el.height = Math.floor(H * dpr);
			el.style.width = W + 'px';
			el.style.height = H + 'px';
			ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
		};
		resize();
		window.addEventListener('resize', resize);

		const glow = toRgb(GLOW());
		const gold = toRgb(readCssVar('--intro-gold', '#d8c08a'));
		const bg = readCssVar('--intro-bg', '#07090e');

		// Warp starfield seed.
		const STAR_COUNT = Math.min(420, Math.floor((W * H) / 2600));
		const stars: Star[] = Array.from({ length: STAR_COUNT }, () => ({
			x: (Math.random() - 0.5) * W,
			y: (Math.random() - 0.5) * H,
			z: Math.random() * W,
			px: 0,
			py: 0
		}));

		// Precompute a film-grain tile for cheap reuse.
		const grain = document.createElement('canvas');
		grain.width = grain.height = 140;
		const gctx = grain.getContext('2d');
		if (gctx) {
			const img = gctx.createImageData(140, 140);
			for (let i = 0; i < img.data.length; i += 4) {
				const v = 120 + Math.random() * 135;
				img.data[i] = img.data[i + 1] = img.data[i + 2] = v;
				img.data[i + 3] = 255;
			}
			gctx.putImageData(img, 0, 0);
		}

		const start = performance.now();

		function frame(nowAbs: number) {
			const now = nowAbs - start;
			elapsed = now;
			const cx = W / 2;
			const cy = H / 2;

			// 1. Deep space base + radial nebula in the configured glow color.
			ctx.fillStyle = bg;
			ctx.fillRect(0, 0, W, H);

			const spaceFade = seg(now, 0, T.spaceIn);
			const neb = ctx.createRadialGradient(cx, cy, 0, cx, cy, Math.max(W, H) * 0.62);
			const ringForm = seg(now, T.comet, T.ring);
			const nebA = 0.06 + 0.16 * easeOut(ringForm);
			neb.addColorStop(0, `rgba(${glow.r},${glow.g},${glow.b},${nebA * spaceFade})`);
			neb.addColorStop(0.5, `rgba(${glow.r},${glow.g},${glow.b},${nebA * 0.25 * spaceFade})`);
			neb.addColorStop(1, 'rgba(0,0,0,0)');
			ctx.fillStyle = neb;
			ctx.fillRect(0, 0, W, H);

			// 2. Warp starfield — speed ramps up during the comet beat, then settles.
			const warp =
				0.4 + 7.5 * easeInOut(seg(now, 200, T.comet)) * (1 - 0.7 * easeOut(seg(now, T.comet, T.ring)));
			ctx.globalAlpha = spaceFade;
			for (const s of stars) {
				s.px = cx + (s.x / s.z) * W;
				s.py = cy + (s.y / s.z) * H;
				s.z -= warp * 6;
				if (s.z < 1) {
					s.z = W;
					s.x = (Math.random() - 0.5) * W;
					s.y = (Math.random() - 0.5) * H;
					s.px = cx + (s.x / s.z) * W;
					s.py = cy + (s.y / s.z) * H;
				}
				const nx = cx + (s.x / s.z) * W;
				const ny = cy + (s.y / s.z) * H;
				const k = (1 - s.z / W) * 2.1;
				ctx.strokeStyle = `rgba(243,240,231,${0.18 + 0.5 * (1 - s.z / W)})`;
				ctx.lineWidth = Math.max(0.4, k);
				ctx.beginPath();
				ctx.moveTo(s.px, s.py);
				ctx.lineTo(nx, ny);
				ctx.stroke();
			}
			ctx.globalAlpha = 1;

			// 3. Comet — streaks from upper-left toward center, then ignites the core.
			const cometP = seg(now, 400, T.comet);
			if (cometP > 0 && cometP < 1) {
				const e = easeOut(cometP);
				const sx = -W * 0.15 + (cx - -W * 0.15) * e;
				const sy = -H * 0.12 + (cy - -H * 0.12) * e;
				const tailLen = 260 * (0.4 + 0.6 * (1 - cometP));
				const ang = Math.atan2(cy - -H * 0.12, cx - -W * 0.15);
				const tx = sx - Math.cos(ang) * tailLen;
				const ty = sy - Math.sin(ang) * tailLen;
				const grad = ctx.createLinearGradient(tx, ty, sx, sy);
				grad.addColorStop(0, 'rgba(0,0,0,0)');
				grad.addColorStop(1, `rgba(${gold.r},${gold.g},${gold.b},0.9)`);
				ctx.strokeStyle = grad;
				ctx.lineWidth = 2.4;
				ctx.lineCap = 'round';
				ctx.beginPath();
				ctx.moveTo(tx, ty);
				ctx.lineTo(sx, sy);
				ctx.stroke();
				ctx.fillStyle = `rgba(${gold.r},${gold.g},${gold.b},0.95)`;
				ctx.beginPath();
				ctx.arc(sx, sy, 3.4, 0, Math.PI * 2);
				ctx.fill();
			}

			// 4. Ignition flash + core bloom when the comet reaches center.
			const ignite = seg(now, T.comet - 120, T.comet + 380);
			if (ignite > 0) {
				const flash = Math.sin(ignite * Math.PI) * (1 - seg(now, T.ring, T.hold));
				const bloom = ctx.createRadialGradient(cx, cy, 0, cx, cy, 220);
				bloom.addColorStop(0, `rgba(${glow.r},${glow.g},${glow.b},${0.55 * flash})`);
				bloom.addColorStop(0.4, `rgba(${gold.r},${gold.g},${gold.b},${0.22 * flash})`);
				bloom.addColorStop(1, 'rgba(0,0,0,0)');
				ctx.fillStyle = bloom;
				ctx.fillRect(0, 0, W, H);
			}

			// 5. Orbital ring — forms around the mark and slowly rotates (heavy tech).
			const ringP = seg(now, T.comet, T.ring);
			if (ringP > 0) {
				const er = easeOut(ringP);
				const radius = 132;
				const sweep = Math.PI * 2 * er;
				const rot = now * 0.00035;
				ctx.save();
				ctx.translate(cx, cy);
				ctx.rotate(rot);
				ctx.lineWidth = 1.4;
				ctx.strokeStyle = `rgba(${glow.r},${glow.g},${glow.b},${0.5 * er})`;
				ctx.shadowBlur = 18;
				ctx.shadowColor = `rgba(${glow.r},${glow.g},${glow.b},0.6)`;
				ctx.beginPath();
				ctx.arc(0, 0, radius, -Math.PI / 2, -Math.PI / 2 + sweep);
				ctx.stroke();
				// Inner hairline gold ring — the vintage rule.
				ctx.shadowBlur = 0;
				ctx.lineWidth = 1;
				ctx.strokeStyle = `rgba(${gold.r},${gold.g},${gold.b},${0.4 * er})`;
				ctx.beginPath();
				ctx.arc(0, 0, radius - 10, -Math.PI / 2, -Math.PI / 2 + sweep);
				ctx.stroke();
				// Orbiting node at the sweep head.
				const hx = Math.cos(-Math.PI / 2 + sweep) * radius;
				const hy = Math.sin(-Math.PI / 2 + sweep) * radius;
				ctx.fillStyle = `rgba(${glow.r},${glow.g},${glow.b},${0.9 * er})`;
				ctx.shadowBlur = 14;
				ctx.shadowColor = `rgba(${glow.r},${glow.g},${glow.b},0.9)`;
				ctx.beginPath();
				ctx.arc(hx, hy, 3, 0, Math.PI * 2);
				ctx.fill();
				ctx.restore();
			}

			// 6. Vignette — pulls focus to the title card.
			const vig = ctx.createRadialGradient(cx, cy, H * 0.2, cx, cy, Math.max(W, H) * 0.75);
			vig.addColorStop(0, 'rgba(0,0,0,0)');
			vig.addColorStop(1, 'rgba(0,0,0,0.62)');
			ctx.fillStyle = vig;
			ctx.fillRect(0, 0, W, H);

			// 7. Film grain — subtle, animated, the vintage texture.
			if (grain) {
				ctx.globalAlpha = 0.045;
				ctx.globalCompositeOperation = 'overlay';
				const ox = (Math.random() * 140) | 0;
				const oy = (Math.random() * 140) | 0;
				for (let x = -ox; x < W; x += 140) {
					for (let y = -oy; y < H; y += 140) {
						ctx.drawImage(grain, x, y);
					}
				}
				ctx.globalCompositeOperation = 'source-over';
				ctx.globalAlpha = 1;
			}

			if (now >= T.total) {
				complete();
				return;
			}
			raf = requestAnimationFrame(frame);
		}

		raf = requestAnimationFrame(frame);

		return () => {
			cancelAnimationFrame(raf);
			window.removeEventListener('resize', resize);
			window.removeEventListener('keydown', onKey, true);
		};
	});

	// CSS-driven text reveals, gated on the same clock as the canvas.
	let showWordmark = $derived(elapsed >= T.wordmark - 250);
	let showTagline = $derived(elapsed >= T.tagline - 150);
	let showHint = $derived(elapsed >= T.spaceIn + 200 && phase === 'run');
</script>

<div
	class="intro"
	class:is-exit={phase === 'exit'}
	role="button"
	tabindex="0"
	aria-label="Skip intro"
	onclick={skip}
	onkeydown={() => {}}
	transition:fade={{ duration: phase === 'exit' ? 760 : 200 }}
>
	<canvas bind:this={canvas} class="stage" aria-hidden="true"></canvas>

	<!-- Hairline double-rule frame: old-cinema title card. -->
	<div class="frame" aria-hidden="true"></div>

	<div class="card">
		<h1 class="wordmark" class:in={showWordmark}>
			<span class="lead">Comet</span><span class="trail">line</span>
		</h1>
		<p class="tagline" class:in={showTagline}>A thought, a task, a file — Cometline continues.</p>
	</div>

	<button class="skip" class:in={showHint} onclick={skip}>Press Esc to skip</button>
</div>

<style>
	.intro {
		position: fixed;
		inset: 0;
		z-index: 90;
		background: var(--intro-bg, #07090e);
		overflow: hidden;
		cursor: pointer;
		display: grid;
		place-items: center;
	}

	.stage {
		position: absolute;
		inset: 0;
		display: block;
	}

	/* Vintage double-rule border, inset from the edges. */
	.frame {
		position: absolute;
		inset: 26px;
		border: 1px solid color-mix(in srgb, var(--intro-gold, #d8c08a) 45%, transparent);
		border-radius: 4px;
		pointer-events: none;
		opacity: 0;
		animation: frame-in 1.2s var(--ease-intro, ease) 0.5s forwards;
	}

	.frame::after {
		content: '';
		position: absolute;
		inset: 5px;
		border: 1px solid color-mix(in srgb, var(--intro-gold, #d8c08a) 22%, transparent);
		border-radius: 2px;
	}

	@keyframes frame-in {
		from {
			opacity: 0;
			transform: scale(1.02);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}

	.card {
		position: relative;
		z-index: 2;
		text-align: center;
		transform: translateY(58px);
		pointer-events: none;
		user-select: none;
	}

	.wordmark {
		margin: 0;
		font-family: 'Hoefler Text', 'Iowan Old Style', 'Palatino Linotype', Palatino, Georgia, serif;
		font-weight: 600;
		letter-spacing: 0.12em;
		font-size: clamp(38px, 6.4vw, 76px);
		color: var(--intro-ink, #f3f0e7);
		opacity: 0;
		filter: blur(8px);
		transform: translateY(10px);
		transition:
			opacity 0.9s var(--ease-intro, ease),
			filter 0.9s var(--ease-intro, ease),
			transform 0.9s var(--ease-intro, ease);
		text-shadow: 0 0 24px rgba(216, 192, 138, 0.18);
	}

	.wordmark.in {
		opacity: 1;
		filter: blur(0);
		transform: translateY(0);
	}

	.wordmark .lead {
		color: var(--intro-ink, #f3f0e7);
	}

	.wordmark .trail {
		color: var(--intro-gold, #d8c08a);
	}

	.tagline {
		margin: 18px 0 0;
		font-family: 'Hoefler Text', 'Iowan Old Style', Georgia, serif;
		font-style: italic;
		font-size: clamp(13px, 1.7vw, 17px);
		letter-spacing: 0.04em;
		color: var(--intro-ink-soft, rgba(243, 240, 231, 0.62));
		opacity: 0;
		transform: translateY(8px);
		transition:
			opacity 1s var(--ease-intro, ease),
			transform 1s var(--ease-intro, ease);
	}

	.tagline.in {
		opacity: 1;
		transform: translateY(0);
	}

	.skip {
		position: absolute;
		bottom: 42px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 3;
		border: 0;
		background: transparent;
		color: var(--intro-ink-soft, rgba(243, 240, 231, 0.55));
		font-size: 11px;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		opacity: 0;
		transition: opacity 0.8s var(--ease-intro, ease);
	}

	.skip.in {
		opacity: 0.7;
	}

	.skip:hover {
		opacity: 1;
		color: var(--intro-ink, #f3f0e7);
	}

	.is-exit .card {
		transition:
			transform 0.76s var(--ease-intro, ease),
			opacity 0.6s var(--ease-intro, ease);
		transform: translateY(58px) scale(1.04);
		opacity: 0;
	}

	@media (prefers-reduced-motion: reduce) {
		.frame,
		.wordmark,
		.tagline,
		.skip {
			animation: none !important;
			transition: opacity 0.3s linear !important;
		}
		.wordmark,
		.tagline {
			filter: none;
			transform: none;
		}
	}
</style>
