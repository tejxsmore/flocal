import gsap from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';

gsap.registerPlugin(ScrollTrigger);

type ScoreState = {
	current: number;
	min: number;
	max: number;
};

type HomepageAnimationOptions = {
	root: HTMLElement;
	getTopics: () => string[];
	setTopic: (topic: string) => void;
};

/**
 * Initialize homepage GSAP animations.
 *
 * Svelte owns application state.
 * GSAP owns visual/high-frequency animation.
 */
export function initHomepageAnimations({
	root,
	getTopics,
	setTopic
}: HomepageAnimationOptions): () => void {
	const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

	const ctx = gsap.context(() => {
		/*
		 * --------------------------------------------------
		 * ELEMENTS
		 * --------------------------------------------------
		 */

		const heroElements = Array.from(root.querySelectorAll<HTMLElement>('[data-hero-enter]'));

		const demoCard = root.querySelector<HTMLElement>('[data-demo-card]');

		const floatingPink = root.querySelector<HTMLElement>('[data-floating-pink]');

		const floatingYellow = root.querySelector<HTMLElement>('[data-floating-yellow]');

		const recordingDot = root.querySelector<HTMLElement>('[data-recording-dot]');

		const waveform = Array.from(root.querySelectorAll<HTMLElement>('.voice-bar'));

		const waveformContainer = root.querySelector<HTMLElement>('[data-waveform]');

		const fluencyElement = root.querySelector<HTMLElement>('[data-score="fluency"]');

		const clarityElement = root.querySelector<HTMLElement>('[data-score="clarity"]');

		const structureElement = root.querySelector<HTMLElement>('[data-score="structure"]');

		const topicElement = root.querySelector<HTMLElement>('[data-topic]');

		const discoverySection = root.querySelector<HTMLElement>('[data-discovery-section]');

		const discoveryEnter = Array.from(root.querySelectorAll<HTMLElement>('[data-discovery-enter]'));

		const discoveryCards = Array.from(root.querySelectorAll<HTMLElement>('[data-discovery-card]'));

		const discoveryCallout = root.querySelector<HTMLElement>('[data-discovery-callout]');

		const finalEnter = root.querySelector<HTMLElement>('[data-final-enter]');

		/*
		 * --------------------------------------------------
		 * REDUCED MOTION
		 * --------------------------------------------------
		 */

		if (reducedMotion) {
			const staticElements: HTMLElement[] = [...heroElements, ...discoveryEnter, ...discoveryCards];

			if (demoCard) {
				staticElements.push(demoCard);
			}

			if (discoveryCallout) {
				staticElements.push(discoveryCallout);
			}

			if (finalEnter) {
				staticElements.push(finalEnter);
			}

			gsap.set(staticElements, {
				opacity: 1,
				x: 0,
				y: 0,
				scale: 1,
				rotate: 0
			});

			return;
		}

		/*
		 * --------------------------------------------------
		 * HERO ENTRANCE
		 * --------------------------------------------------
		 */

		gsap.from(heroElements, {
			opacity: 0,
			y: 22,
			duration: 0.75,
			stagger: 0.075,
			ease: 'power3.out'
		});

		if (demoCard) {
			gsap.from(demoCard, {
				opacity: 0,
				x: 30,
				y: 8,
				rotate: 1,
				duration: 0.9,
				delay: 0.12,
				ease: 'power3.out'
			});
		}

		/*
		 * --------------------------------------------------
		 * FLOATING HERO DECORATIONS
		 * --------------------------------------------------
		 */

		if (floatingPink) {
			gsap.to(floatingPink, {
				y: -9,
				x: 5,
				rotate: 7,
				duration: 3.4,
				repeat: -1,
				yoyo: true,
				ease: 'sine.inOut'
			});
		}

		if (floatingYellow) {
			gsap.to(floatingYellow, {
				y: 8,
				x: -4,
				rotate: -6,
				duration: 3.8,
				repeat: -1,
				yoyo: true,
				ease: 'sine.inOut'
			});
		}

		/*
		 * --------------------------------------------------
		 * RECORDING INDICATOR
		 * --------------------------------------------------
		 *
		 * Soft breathing pulse rather than a harsh flash.
		 */

		if (recordingDot) {
			gsap.to(recordingDot, {
				scale: 1.25,
				opacity: 0.55,
				duration: 0.85,
				repeat: -1,
				yoyo: true,
				ease: 'sine.inOut'
			});
		}

		/*
		 * --------------------------------------------------
		 * SPEECH WAVEFORM
		 * --------------------------------------------------
		 *
		 * The bars intentionally remain visually simple.
		 *
		 * Speech itself is not a perfectly uniform equalizer:
		 *
		 * - pauses
		 * - quiet syllables
		 * - consonants
		 * - stressed syllables
		 * - longer voiced sounds
		 *
		 * The envelope below creates those larger speech phrases.
		 */

		const speechEnvelope: number[] = [
			0.08, 0.11, 0.18, 0.13, 0.07, 0.04,

			0.16, 0.28, 0.46, 0.62, 0.71, 0.64, 0.48, 0.31,

			0.14, 0.08,

			0.22, 0.38, 0.59, 0.77, 0.88, 0.78, 0.61, 0.44, 0.29,

			0.12, 0.07,

			0.19, 0.34, 0.53, 0.68, 0.73, 0.61, 0.46, 0.31,

			0.12, 0.06,

			0.19, 0.33
		];

		const barEnergy = waveform.map((_bar: HTMLElement, index: number): number => {
			const envelope = speechEnvelope[index] ?? 0.12;

			const localVariation = 0.88 + Math.random() * 0.24;

			return Math.min(1, envelope * localVariation);
		});

		const animateSpeechBar = (bar: HTMLElement, index: number): void => {
			const energy = barEnergy[index] ?? 0.12;

			/*
			 * Quiet / pause region.
			 */

			if (energy < 0.1) {
				const targetHeight = 10 + Math.random() * 7;

				gsap.to(bar, {
					height: `${targetHeight}px`,
					duration: 0.22 + Math.random() * 0.18,
					ease: 'sine.inOut',

					onComplete: () => {
						animateSpeechBar(bar, index);
					}
				});

				return;
			}

			/*
			 * Normal speech region.
			 */

			const baseHeight = 10 + energy * 57;

			const variation = (Math.random() - 0.5) * 10;

			const targetHeight = Math.max(9, Math.min(70, baseHeight + variation));

			const duration = 0.13 + (1 - energy) * 0.1 + Math.random() * 0.1;

			gsap.to(bar, {
				height: `${targetHeight}px`,
				duration,
				ease: 'sine.inOut',

				onComplete: () => {
					/*
					 * Small occasional pauses.
					 */

					if (Math.random() < 0.055 && energy > 0.25) {
						gsap.delayedCall(0.08 + Math.random() * 0.18, () => {
							animateSpeechBar(bar, index);
						});

						return;
					}

					animateSpeechBar(bar, index);
				}
			});
		};

		waveform.forEach((bar: HTMLElement, index: number): void => {
			const energy = barEnergy[index] ?? 0.12;

			const initialHeight = Math.max(9, 10 + energy * 52);

			gsap.set(bar, {
				height: `${initialHeight}px`,
				transformOrigin: 'center'
			});

			gsap.delayedCall(index * 0.018, () => {
				animateSpeechBar(bar, index);
			});
		});

		/*
		 * Extremely subtle phrase-level movement.
		 */

		if (waveformContainer) {
			gsap.to(waveformContainer, {
				scaleY: 1.018,
				duration: 1.9,
				repeat: -1,
				yoyo: true,
				ease: 'sine.inOut'
			});
		}

		/*
		 * --------------------------------------------------
		 * SCORE SYSTEM
		 * --------------------------------------------------
		 */

		const scoreValues: Record<'fluency' | 'clarity' | 'structure', ScoreState> = {
			fluency: {
				current: 82,
				min: 78,
				max: 94
			},

			clarity: {
				current: 76,
				min: 73,
				max: 92
			},

			structure: {
				current: 89,
				min: 82,
				max: 96
			}
		};

		const animateScore = (element: HTMLElement | null, score: ScoreState): void => {
			if (!element) {
				return;
			}

			const next = score.min + Math.floor(Math.random() * (score.max - score.min + 1));

			gsap.to(score, {
				current: next,
				duration: 1.05,
				ease: 'power2.out',

				onUpdate: () => {
					element.textContent = String(Math.round(score.current));
				}
			});

			gsap.fromTo(
				element,
				{
					scale: 1
				},
				{
					scale: 1.055,
					duration: 0.18,
					repeat: 1,
					yoyo: true,
					ease: 'power2.inOut'
				}
			);
		};

		const scoreTimeline = gsap.timeline({
			repeat: -1,
			repeatDelay: 0.8
		});

		scoreTimeline
			.to(
				{},
				{
					duration: 4.5
				}
			)

			.call(() => {
				animateScore(fluencyElement, scoreValues.fluency);
			})

			.to(
				{},
				{
					duration: 0.3
				}
			)

			.call(() => {
				animateScore(clarityElement, scoreValues.clarity);
			})

			.to(
				{},
				{
					duration: 0.3
				}
			)

			.call(() => {
				animateScore(structureElement, scoreValues.structure);
			})

			.to(
				{},
				{
					duration: 2.5
				}
			);

		/*
		 * --------------------------------------------------
		 * TOPIC ROTATION
		 * --------------------------------------------------
		 */

		if (topicElement) {
			const topicTimeline = gsap.timeline({
				repeat: -1,
				repeatDelay: 0.5
			});

			topicTimeline
				.to(
					{},
					{
						duration: 7.5
					}
				)

				.to(topicElement, {
					opacity: 0,
					y: 9,
					duration: 0.35,
					ease: 'power2.in'
				})

				.call(() => {
					const topics = getTopics();

					if (topics.length === 0) {
						return;
					}

					const currentTopic = topicElement.textContent?.trim() ?? '';

					const currentIndex = topics.indexOf(currentTopic);

					const nextIndex = currentIndex === -1 ? 0 : (currentIndex + 1) % topics.length;

					setTopic(topics[nextIndex]);
				})

				.set(topicElement, {
					y: -9
				})

				.to(topicElement, {
					opacity: 1,
					y: 0,
					duration: 0.5,
					ease: 'power3.out'
				})

				.to(
					{},
					{
						duration: 0.5
					}
				);
		}

		/*
		 * --------------------------------------------------
		 * DISCOVERY SECTION REVEAL
		 * --------------------------------------------------
		 */

		if (discoverySection) {
			gsap.from(discoveryEnter, {
				scrollTrigger: {
					trigger: discoverySection,
					start: 'top 78%',
					once: true
				},

				opacity: 0,
				y: 28,
				duration: 0.75,
				stagger: 0.08,
				ease: 'power3.out'
			});

			/*
			 * Cards animate INTO position once.
			 *
			 * They do NOT animate on hover.
			 */

			gsap.from(discoveryCards, {
				scrollTrigger: {
					trigger: discoverySection,
					start: 'top 70%',
					once: true
				},

				opacity: 0,
				y: 34,
				rotate: 0,
				duration: 0.7,
				stagger: 0.11,
				ease: 'power3.out'
			});

			if (discoveryCallout) {
				gsap.from(discoveryCallout, {
					scrollTrigger: {
						trigger: discoveryCallout,
						start: 'top 86%',
						once: true
					},

					opacity: 0,
					y: 22,
					scale: 0.98,
					duration: 0.7,
					ease: 'power3.out'
				});
			}
		}

		/*
		 * --------------------------------------------------
		 * FINAL SECTION
		 * --------------------------------------------------
		 */

		if (finalEnter) {
			gsap.from(finalEnter, {
				scrollTrigger: {
					trigger: finalEnter,
					start: 'top 88%',
					once: true
				},

				opacity: 0,
				y: 24,
				duration: 0.7,
				ease: 'power3.out'
			});
		}
	}, root);

	/*
	 * Cleanup all GSAP animations and ScrollTriggers.
	 */

	return (): void => {
		ctx.revert();
	};
}
