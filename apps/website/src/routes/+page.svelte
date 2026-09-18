<script lang="ts">
	import { onMount } from 'svelte';
	import { initHomepageAnimations } from '$lib/animations/homepage';

	const topics: string[] = [
		'Should everyone learn how to cook?',
		'Is social media making us more connected?',
		'What makes someone a good leader?',
		'Should college education be free?',
		'Does technology make life easier?',
		'What is one skill everyone should learn?',
		'Is failure necessary for success?',
		'Should artificial intelligence be taught in schools?'
	];

	let root: HTMLDivElement | null = $state(null);
	let topic: string = $state(topics[0]);

	onMount(() => {
		if (!root) {
			return;
		}

		return initHomepageAnimations({
			root,
			getTopics: () => topics,
			setTopic: (nextTopic: string) => {
				topic = nextTopic;
			}
		});
	});
</script>

<svelte:head>
	<title>Flocal | Become a better speaker</title>

	<meta
		name="description"
		content="Take one conscious step toward better speech every day. Discover ideas, research new topics, speak your mind, and improve with AI-powered feedback."
	/>
</svelte:head>

<div
	bind:this={root}
	class="min-h-screen overflow-hidden bg-[#BADF96] text-[#1C1124]"
>
	<main>
		<!-- ================================================= -->
<!-- HERO -->
<!-- ================================================= -->

<section
	class="px-4 pb-16 pt-6
	sm:px-8 sm:pb-20 sm:pt-10
	lg:px-16 lg:pb-24 lg:pt-10
	xl:px-24"
>
	<div
		class="mx-auto flex max-w-6xl flex-col
		gap-12
		lg:grid lg:min-h-[calc(100vh-90px)]
		lg:grid-cols-[0.95fr_1.05fr]
		lg:items-center lg:gap-14"
	>
		<!-- LEFT SIDE -->

		<div
			class="w-full max-w-xl
			lg:-translate-y-7"
		>
			<div
				data-hero-enter
				class="inline-flex items-center gap-2 rounded-full
				border-2 border-[#1C1124] bg-[#F7FFCD]
				px-3.5 py-1.5 text-xs font-semibold
				text-[#4d2a3a]
				sm:px-4 sm:py-2 sm:text-sm"
			>
				<span
					class="h-2.5 w-2.5 shrink-0 rounded-full
					bg-[#4d2a3a]"
				></span>

				Your AI speech fluency coach
			</div>

			<h1
				data-hero-enter
				class="mt-5 max-w-xl
				text-[2.85rem] font-black
				leading-[0.94]
				tracking-[-0.045em]
				sm:mt-6 sm:text-5xl
				md:text-6xl
				lg:text-[4.35rem]"
			>
				Be the best
				<span class="block">
					speaker in the room.
				</span>
			</h1>

			<p
				data-hero-enter
				class="mt-5 max-w-lg
				text-[0.95rem] leading-relaxed
				text-[#4d2a3a]
				sm:mt-6 sm:text-lg"
			>
				Take one conscious step toward bettering your
				speech every day. Learn to think on your feet,
				organize your ideas, and express yourself with
				clarity and confidence.
			</p>

			<div
				data-hero-enter
				class="mt-7 flex w-full flex-col gap-3
				sm:mt-8 sm:flex-row sm:items-center"
			>
				<a
					href="/practice"
					class="group inline-flex min-h-12
					w-full items-center justify-center
					gap-3 rounded-[0.75em]
					border-2 border-[#1C1124]
					bg-[#1C1124] px-5 py-3
					text-base font-bold text-[#F7FFCD]
					transition-transform duration-200
					hover:-translate-y-1
					sm:w-auto"
				>
					Start speaking

					<span
						class="text-xl transition-transform
						duration-200 group-hover:translate-x-1"
					>
						→
					</span>
				</a>

				<a
					href="/features"
					class="inline-flex min-h-12
					w-full items-center justify-center
					rounded-[0.75em]
					border-2 border-[#1C1124]
					bg-[#F7FFCD] px-5 py-3
					text-base font-bold text-[#1C1124]
					transition-transform duration-200
					hover:-translate-y-1
					sm:w-auto"
				>
					See how it works
				</a>
			</div>

			<div
				data-hero-enter
				class="mt-6 flex flex-col gap-2
				text-sm font-medium text-[#4d2a3a]
				sm:mt-7 sm:flex-row sm:items-center sm:gap-5"
			>
				<div class="flex items-center gap-2">
					<span>⏱</span>
					1 minute sessions
				</div>

				<div
					class="hidden h-1 w-1 rounded-full
					bg-[#4d2a3a] sm:block"
				></div>

				<div class="flex items-center gap-2">
					<span>✦</span>
					AI-powered feedback
				</div>
			</div>
		</div>

		<!-- RIGHT SIDE -->

		<div
			class="relative flex w-full
			justify-center
			lg:justify-end"
		>
			<!-- FLOATING PINK -->

			<div
				data-floating-pink
				class="absolute -right-1 -top-5
				h-12 w-12 rounded-full
				border-2 border-[#1C1124]
				bg-[#ff94d0]
				sm:-right-5 sm:-top-7
				sm:h-20 sm:w-20"
			></div>

			<!-- FLOATING YELLOW -->

			<div
				data-floating-yellow
				class="absolute -bottom-4 -left-1
				h-11 w-11 rounded-full
				border-2 border-[#1C1124]
				bg-[#F7FFCD]
				sm:-bottom-6 sm:-left-5
				sm:h-16 sm:w-16"
			></div>

			<!-- DEMO CARD -->

			<div
				data-demo-card
				class="relative w-full max-w-lg
				overflow-hidden rounded-3xl
				border-2 border-[#1C1124]
				bg-[#F7FFCD]
				shadow-[6px_6px_0px_#1C1124]
				sm:shadow-[10px_10px_0px_#1C1124]"
			>
				<!-- CARD HEADER -->

				<div
					class="flex items-center
					justify-between
					border-b-2 border-[#1C1124]
					px-4 py-3.5
					sm:px-5 sm:py-4"
				>
					<div class="flex min-w-0 items-center gap-2.5 sm:gap-3">
						<div
							class="flex h-8 w-8 shrink-0
							items-center justify-center
							rounded-full border-2
							border-[#1C1124]
							bg-[#BADF96]
							text-sm font-bold
							sm:h-9 sm:w-9 sm:text-base"
						>
							F
						</div>

						<div class="min-w-0">
							<p class="truncate text-sm font-bold sm:text-base">
								Flocal Practice
							</p>

							<p
								class="text-[11px]
								text-[#4d2a3a]
								sm:text-xs"
							>
								Speech session #24
							</p>
						</div>
					</div>

					<div
						class="shrink-0 rounded-full
						border-2 border-[#1C1124]
						bg-[#BADF96]
						px-2 py-1
						text-[10px] font-bold
						sm:px-2.5 sm:text-xs"
					>
						01:00
					</div>
				</div>

				<div class="p-4 sm:p-6">
					<p
						class="text-[10px] font-bold
						uppercase tracking-widest
						text-[#4d2a3a]
						sm:text-xs"
					>
						Your topic
					</p>

					<div
						class="flex min-h-24 items-center
						sm:min-h-28"
					>
						<h2
							data-topic
							class="text-xl font-black
							leading-tight
							sm:text-3xl"
						>
							{topic}
						</h2>
					</div>

					<!-- VOICE NOTE -->

					<div
						class="mt-3 rounded-2xl
						border-2 border-[#1C1124]
						bg-[#BADF96] p-3
						sm:mt-4 sm:p-4"
					>
						<div
							class="flex items-center
							justify-end"
						>
							<div
								class="flex items-center gap-2
								text-[11px] font-bold
								text-[#4d2a3a]
								sm:text-xs"
							>
								<span
									data-recording-dot
									class="h-2.5 w-2.5
									rounded-full
									bg-[#FF6060]"
								></span>

								<span>
									Recording
								</span>
							</div>
						</div>

						<!-- ORIGINAL STYLE WAVEFORM -->

						<div
							data-waveform
							class="mt-2 flex h-20
							items-center justify-center
							gap-0.5 overflow-hidden px-0
							sm:mt-3 sm:h-24 sm:gap-1 sm:px-1"
						>
							{#each Array(38) as _, index}
								<span
									class="voice-bar h-8
									w-1 shrink-0 rounded-full
									bg-[#4d2a3a]
									sm:w-1.5"
									data-wave-index={index}
								></span>
							{/each}
						</div>
					</div>

					<!-- SCORES -->

					<div
						class="mt-3 grid grid-cols-3
						gap-1.5
						sm:mt-4 sm:gap-3"
					>
						<div
							class="rounded-xl border-2
							border-[#1C1124]
							bg-white px-1.5 py-2.5
							text-center
							sm:px-3 sm:py-4"
						>
							<p
								data-score="fluency"
								class="text-xl font-black
								sm:text-2xl"
							>
								82
							</p>

							<p
								class="text-[9px] font-bold
								text-[#4d2a3a]
								sm:text-xs"
							>
								Fluency
							</p>
						</div>

						<div
							class="rounded-xl border-2
							border-[#1C1124]
							bg-white px-1.5 py-2.5
							text-center
							sm:px-3 sm:py-4"
						>
							<p
								data-score="clarity"
								class="text-xl font-black
								sm:text-2xl"
							>
								76
							</p>

							<p
								class="text-[9px] font-bold
								text-[#4d2a3a]
								sm:text-xs"
							>
								Clarity
							</p>
						</div>

						<div
							class="rounded-xl border-2
							border-[#1C1124]
							bg-white px-1.5 py-2.5
							text-center
							sm:px-3 sm:py-4"
						>
							<p
								data-score="structure"
								class="text-xl font-black
								sm:text-2xl"
							>
								89
							</p>

							<p
								class="text-[9px] font-bold
								text-[#4d2a3a]
								sm:text-xs"
							>
								Structure
							</p>
						</div>
					</div>

					<!-- AI FEEDBACK -->

					<div
						class="mt-3 rounded-xl
						border-2 border-[#1C1124]
						bg-[#ff94d0] p-3
						sm:mt-4 sm:p-3.5"
					>
						<p class="text-sm font-black sm:text-base">
							Flocal says
						</p>

						<p
							class="mt-1 text-xs font-medium
							leading-relaxed
							text-[#4d2a3a]
							sm:text-sm"
						>
							Try slowing down slightly before
							your main point. Your ideas are
							strong, give them room to land.
						</p>
					</div>
				</div>
			</div>
		</div>
	</div>
</section>
		<section
			data-discovery-section
			class="relative overflow-hidden
			bg-[#BADF96] px-6 py-20 sm:px-10
			sm:py-24 lg:px-16 xl:px-24"
		>
			<div class="relative mx-auto max-w-6xl">
				<!-- HEADING -->

				<div
					data-discovery-enter
					class="mx-auto max-w-3xl text-center"
				>
					<div
						class="mx-auto inline-flex items-center
						gap-2 rounded-full border-2
						border-[#1C1124] bg-[#F7FFCD]
						px-4 py-2 text-sm font-semibold
						text-[#4d2a3a]"
					>
						✦ Think. Research. Speak.
					</div>

					<h2
						class="mt-6 text-4xl font-black
						leading-[0.98] tracking-[-0.04em]
						sm:text-5xl md:text-6xl"
					>
						Don't just practice speaking.
						<span class="block">
							Exercise your brain.
						</span>
					</h2>

					<p
						class="mx-auto mt-6 max-w-2xl text-base
						leading-relaxed text-[#3e3151]
						sm:text-lg"
					>
						Discover a new topic. Research it. Form an
						opinion. Then explain it in your own words.
						Flocal turns a one-minute speaking session
						into a daily exercise for thinking, learning,
						and communicating.
					</p>
				</div>

				<!-- THREE CARDS -->

				<div
					class="mt-12 grid gap-5 md:grid-cols-3"
				>
					<!-- DISCOVER -->

					<div
						data-discovery-card
						class="rounded-3xl border-2
						border-[#1C1124] bg-white p-6
						shadow-[6px_6px_0px_#1C1124]
						sm:p-7"
					>
						<div
							class="flex h-11 w-11 items-center
							justify-center rounded-full
							border-2 border-[#1C1124]
							bg-[#BADF96] text-sm font-black"
						>
							01
						</div>

						<h3
							class="mt-6 text-2xl font-black"
						>
							Discover
						</h3>

						<p
							class="mt-3 text-sm leading-relaxed
							text-[#4d2a3a] sm:text-base"
						>
							Encounter ideas you wouldn't normally
							think about. Every new prompt gives your
							brain something unfamiliar to work with.
						</p>
					</div>

					<!-- RESEARCH -->

					<div
						data-discovery-card
						class="rounded-3xl border-2
						border-[#1C1124] bg-[#F7FFCD]
						p-6 shadow-[6px_6px_0px_#1C1124]
						sm:p-7"
					>
						<div
							class="flex h-11 w-11 items-center
							justify-center rounded-full
							border-2 border-[#1C1124]
							bg-[#9FA1FF] text-sm font-black"
						>
							02
						</div>

						<h3
							class="mt-6 text-2xl font-black"
						>
							Research
						</h3>

						<p
							class="mt-3 text-sm leading-relaxed
							text-[#4d2a3a] sm:text-base"
						>
							Look up unfamiliar words, ideas,
							people, and concepts. Build knowledge
							you can actually use instead of simply
							scrolling past it.
						</p>
					</div>

					<!-- SPEAK -->

					<div
						data-discovery-card
						class="rounded-3xl border-2
						border-[#1C1124] bg-[#ff94d0]
						p-6 shadow-[6px_6px_0px_#1C1124]
						sm:p-7"
					>
						<div
							class="flex h-11 w-11 items-center
							justify-center rounded-full
							border-2 border-[#1C1124]
							bg-[#F7FFCD] text-sm font-black"
						>
							03
						</div>

						<h3
							class="mt-6 text-2xl font-black"
						>
							Speak
						</h3>

						<p
							class="mt-3 text-sm leading-relaxed
							text-[#4d2a3a] sm:text-base"
						>
							Turn what you learned into your own
							words. Retrieve the idea, organize your
							thoughts, and explain them without a
							script.
						</p>
					</div>
				</div>

				<!-- PURPLE CALLOUT -->

				<div class="mx-auto mt-8 max-w-3xl">
					<div
						data-discovery-callout
						class="rounded-2xl border-2
						border-[#1C1124] bg-[#9FA1FF]
						px-6 py-7 text-center
						text-[#1C1124]
						shadow-[6px_6px_0px_#1C1124]"
					>
						<p
							class="mx-auto max-w-2xl
							text-lg font-black leading-tight
							sm:text-xl"
						>
							In a world where technology does more
							of the thinking, choosing to think for
							yourself becomes a skill worth training.
						</p>

						<p
							class="mx-auto mt-3 max-w-xl text-sm
							leading-relaxed text-[#3e3151]
							sm:text-base"
						>
							Flocal gives you a reason to pause,
							form your own thoughts, and put them
							into words.
						</p>
					</div>
				</div>
			</div>
		</section>

		<!-- ================================================= -->
		<!-- FINAL CTA -->
		<!-- ================================================= -->

		<section
			class="bg-[#BADF96] px-6 py-20
			sm:px-10 lg:px-16"
		>
			<div
				data-final-enter
				class="mx-auto max-w-3xl text-center"
			>
				<p
					class="text-sm font-bold uppercase
					tracking-[0.18em] text-[#4d2a3a]"
				>
					One minute is enough to start.
				</p>

				<h2
					class="mt-4 text-3xl font-black
					tracking-[-0.035em] sm:text-4xl"
				>
					Your next conversation starts
					with one thought.
				</h2>

				<a
					href="/practice"
					class="mt-7 inline-flex items-center
					gap-3 rounded-[0.75em]
					border-2 border-[#1C1124]
					bg-[#1C1124] px-5 py-3
					text-base font-bold text-[#F7FFCD]
					transition-transform duration-200
					hover:-translate-y-1"
				>
					Start speaking

					<span>→</span>
				</a>
			</div>
		</section>
	</main>
</div>