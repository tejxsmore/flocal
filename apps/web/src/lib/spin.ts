import gsap from 'gsap';

export const REEL_ITEM_HEIGHT = 160;

const ROLL_STEPS = 20;
const SPIN_DURATION = 3;

let audioCtx: AudioContext | null = null;

export function unlockAudio() {
	if (typeof window === 'undefined') return;

	if (!audioCtx) {
		const AudioCtor =
			window.AudioContext ||
			(window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;

		if (!AudioCtor) return;

		audioCtx = new AudioCtor();
	}

	if (audioCtx.state === 'suspended') {
		void audioCtx.resume();
	}
}

function playTick(step: number) {
	if (!audioCtx || audioCtx.state !== 'running') return;

	const ctx = audioCtx;
	const now = ctx.currentTime;

	const osc = ctx.createOscillator();
	const gain = ctx.createGain();

	osc.type = 'square';
	osc.frequency.setValueAtTime(180 + Math.min(step, ROLL_STEPS) * 3, now);

	gain.gain.setValueAtTime(0.035, now);
	gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.045);

	osc.connect(gain);
	gain.connect(ctx.destination);

	osc.start(now);
	osc.stop(now + 0.05);
}

function playLand() {
	if (!audioCtx || audioCtx.state !== 'running') return;

	const ctx = audioCtx;
	const now = ctx.currentTime;

	const notes = [440, 554.37, 659.25];

	notes.forEach((frequency, index) => {
		const start = now + index * 0.06;

		const osc = ctx.createOscillator();
		const gain = ctx.createGain();

		osc.type = 'sine';
		osc.frequency.setValueAtTime(frequency, start);

		gain.gain.setValueAtTime(0.0001, start);
		gain.gain.exponentialRampToValueAtTime(0.06, start + 0.02);
		gain.gain.exponentialRampToValueAtTime(0.0001, start + 0.28);

		osc.connect(gain);
		gain.connect(ctx.destination);

		osc.start(start);
		osc.stop(start + 0.32);
	});
}

function randomFrom(pool: string[]): string {
	if (pool.length === 0) return '';

	return pool[Math.floor(Math.random() * pool.length)] ?? '';
}

function escapeHtml(text: string): string {
	const div = document.createElement('div');
	div.textContent = text;

	return div.innerHTML;
}

function buildItem(text: string, variant: 'current' | 'dim'): string {
	return `
		<div class="reel-item reel-item--${variant}">
			${escapeHtml(text)}
		</div>
	`;
}

export function initReel(
	el: HTMLElement,
	previous = '',
	current = 'Tap generate to get a topic',
	next = ''
) {
	const existingTrack = el.querySelector<HTMLElement>('.reel-track');

	if (existingTrack) {
		gsap.killTweensOf(existingTrack);
	}

	el.innerHTML = `
		<div class="reel-track">
			${buildItem(previous, 'dim')}
			${buildItem(current, 'current')}
			${buildItem(next, 'dim')}
		</div>
	`;

	const track = el.querySelector<HTMLElement>('.reel-track');

	if (!track) return;

	gsap.set(track, { y: 0 });
}

export function spinReel(el: HTMLElement, pool: string[], finalText: string): Promise<void> {
	const track = el.querySelector<HTMLElement>('.reel-track');

	if (!track) {
		initReel(el, '', finalText, '');
		return Promise.resolve();
	}

	gsap.killTweensOf(track);

	const rollItems: string[] = [];

	for (let i = 0; i < ROLL_STEPS; i++) {
		let item = randomFrom(pool);

		if (pool.length > 1 && item === rollItems[rollItems.length - 1]) {
			item = randomFrom(pool.filter((value) => value !== item));
		}

		rollItems.push(item);
	}

	const previous = rollItems[rollItems.length - 1] ?? '';
	const next = randomFrom(pool.filter((item) => item !== finalText));

	const items = ['', ...rollItems, finalText, next, ''];

	const finalIndex = items.length - 3;

	track.innerHTML = items
		.map((text, index) => buildItem(text, index === finalIndex ? 'current' : 'dim'))
		.join('');

	gsap.set(track, { y: 0 });

	const targetY = -(finalIndex - 1) * REEL_ITEM_HEIGHT;

	const finalEl = track.children[finalIndex] as HTMLElement | undefined;

	if (!finalEl) return Promise.resolve();

	return new Promise<void>((resolve) => {
		let lastStep = 0;

		gsap.to(track, {
			y: targetY,
			duration: SPIN_DURATION,
			ease: 'power4.out',
			overwrite: true,

			onUpdate() {
				const currentY = Number(gsap.getProperty(track, 'y'));

				const rawStep = Math.floor(Math.abs(currentY) / REEL_ITEM_HEIGHT);

				const step = Math.min(finalIndex - 1, rawStep);

				if (step > lastStep) {
					playTick(step);
					lastStep = step;
				}
			},

			onComplete() {
				gsap.set(track, {
					y: targetY
				});

				playLand();

				gsap.fromTo(
					finalEl,
					{
						scale: 1.04
					},
					{
						scale: 1,
						duration: 0.3,
						ease: 'back.out(2.2)',
						clearProps: 'transform',
						onComplete: resolve
					}
				);
			}
		});
	});
}
