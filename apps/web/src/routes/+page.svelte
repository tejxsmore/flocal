<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import posthog from 'posthog-js';

	import {
		listCategories,
		spinTopic,
		createSession,
		ApiRequestError
	} from '$lib/api';

	import {
		initReel,
		spinReel,
		unlockAudio
	} from '$lib/spin';

	import {
		FORMAT_OPTIONS,
		PREP_TIME_OPTIONS,
		type PrepTimeSeconds,
		type Topic,
		type TopicCategory,
		type TopicFormat
	} from '$lib/types';

	let categories = $state<TopicCategory[]>([]);
	let selectedCategoryId = $state<string | null>(null);
	let selectedFormat = $state<TopicFormat | null>(null);
	let formatMenuOpen = $state(false);
	let formatMenuEl = $state<HTMLDivElement>();

	let topic = $state<Topic | null>(null);
	let prepTimeSeconds = $state<PrepTimeSeconds>(0);
	let debateStance = $state<'for' | 'against'>('for');

	let spinning = $state(false);
	let starting = $state(false);
	let errorMessage = $state('');

	let reelEl = $state<HTMLDivElement>();

	// how many extra topics to fetch for the reel pool (above/below the winner)
	const REEL_POOL_SIZE = 8;

	let topicCategory = $derived(
		categories.find((category) => category.id === topic?.categoryId) ?? null
	);

	let topicFormatOption = $derived(
		FORMAT_OPTIONS.find((format) => format.value === topic?.format) ?? null
	);

	let selectedFormatOption = $derived(
		FORMAT_OPTIONS.find((format) => format.value === selectedFormat) ?? null
	);

	const primaryButtonClass =
		'flex h-13.5 w-full cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#ff6803] bg-[#FF7315] px-5 font-bold text-[#232020] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:translate-y-0';

	const generateButtonClass =
		'flex h-13.5 cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#464040] bg-[#3A3535] px-8 font-bold text-[#f4f4f4] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:translate-y-0';

	const dropdownButtonClass =
		'flex h-13.5 w-full cursor-pointer items-center justify-between gap-3 rounded-xl border border-[#464040] bg-[#1a1818] px-5 text-left font-bold text-[#f4f4f4] transition-colors duration-200 hover:border-[#5C5555]';

	function pillClass(selected: boolean) {
		return `cursor-pointer rounded-full border px-4 py-2 text-sm font-bold transition-colors duration-200 ${
			selected
				? 'border-[#464040] bg-[#3A3535] text-[#f4f4f4]'
				: 'border-[#464040] bg-transparent text-[#f4f4f4]/60 hover:text-[#f4f4f4]'
		}`;
	}

	onMount(() => {
		if (reelEl) {
			initReel(
				reelEl,
				'',
				'Tap generate to get a topic',
				''
			);
		}

		listCategories()
			.then((result) => {
				categories = result;
			})
			.catch((err) => {
				errorMessage =
					err instanceof ApiRequestError
						? err.message
						: 'Could not load categories.';
			});

		function handleClickOutside(event: MouseEvent) {
			if (
				formatMenuOpen &&
				formatMenuEl &&
				!formatMenuEl.contains(event.target as Node)
			) {
				formatMenuOpen = false;
			}
		}

		function handleEscape(event: KeyboardEvent) {
			if (event.key === 'Escape') {
				formatMenuOpen = false;
			}
		}

		document.addEventListener('click', handleClickOutside);
		document.addEventListener('keydown', handleEscape);

		return () => {
			document.removeEventListener('click', handleClickOutside);
			document.removeEventListener('keydown', handleEscape);
		};
	});

	function selectFormat(format: TopicFormat | null) {
		selectedFormat = format;
		formatMenuOpen = false;
	}

	function selectCategory(id: string | null) {
		selectedCategoryId = id;
	}

	// Fetch real topics matching the current category/format filters to
	// populate the reel above/below the winning topic. No static/dummy data.
	async function fetchReelPool(excludeTopicId: string): Promise<string[]> {
		const settled = await Promise.allSettled(
			Array.from({ length: REEL_POOL_SIZE }, () =>
				spinTopic(selectedCategoryId, selectedFormat)
			)
		);

		const seenIds = new Set<string>([excludeTopicId]);
		const seenTitles = new Set<string>();
		const titles: string[] = [];

		for (const outcome of settled) {
			if (outcome.status !== 'fulfilled') continue;

			const candidate = outcome.value;

			if (seenIds.has(candidate.id) || seenTitles.has(candidate.title)) continue;

			seenIds.add(candidate.id);
			seenTitles.add(candidate.title);
			titles.push(candidate.title);
		}

		return titles;
	}

	async function spin() {
		if (spinning || !reelEl) return;

		unlockAudio();

		spinning = true;
		errorMessage = '';
		debateStance = 'for';

		try {
			const result = await spinTopic(
				selectedCategoryId,
				selectedFormat
			);

			const pool = await fetchReelPool(result.id);

			await spinReel(
				reelEl,
				pool,
				result.title
			);

			topic = result;

			posthog.capture('topic_spun', {
				topic_format: result.format,
				topic_difficulty: result.difficulty,
				filter_category: selectedCategoryId ?? null,
				filter_format: selectedFormat ?? null
			});

			const recommended = result.recommendedPrepSeconds;

			prepTimeSeconds = PREP_TIME_OPTIONS.some(
				(option) => option.value === recommended
			)
				? (recommended as PrepTimeSeconds)
				: 0;
		} catch (err) {
			errorMessage =
				err instanceof ApiRequestError
					? err.message
					: 'Could not spin a topic.';
		} finally {
			spinning = false;
		}
	}

	async function start() {
		if (!topic) return;

		starting = true;
		errorMessage = '';

		try {
			const session = await createSession(
				topic.id,
				prepTimeSeconds,
				{
					debateStance:
						topic.format === 'debate'
							? debateStance
							: undefined
				}
			);

			sessionStorage.setItem(
				`flocal:topic:${session.id}`,
				JSON.stringify(topic)
			);

			posthog.capture('session_started', {
				topic_format: topic.format,
				topic_difficulty: topic.difficulty,
				prep_time_seconds: prepTimeSeconds,
				debate_stance:
					topic.format === 'debate'
						? debateStance
						: null
			});

			await goto(`/session/${session.id}`);
		} catch (err) {
			errorMessage =
				err instanceof ApiRequestError
					? err.message
					: 'Could not start session.';

			starting = false;
		}
	}
</script>

<svelte:head>
	<title>Flocal — Practice</title>
</svelte:head>

<div class="flex min-h-screen justify-center px-5 py-10 sm:px-8 sm:py-14">
	<div class="w-full max-w-3xl">

		<div class="mx-auto max-w-md text-center">
			<h1 class="text-3xl font-black leading-[1.15] tracking-[-0.02em] sm:text-4xl">
				Ready to practice?
			</h1>

			<p class="mt-3 text-base leading-relaxed text-[#d8d8d8]">
				Pick a format and category, then spin for a topic.
			</p>
		</div>

		<div class="mt-12 flex flex-col gap-6">

			<div class="relative mx-auto w-full max-w-md" bind:this={formatMenuEl}>
				<button
					type="button"
					onclick={() => (formatMenuOpen = !formatMenuOpen)}
					class={dropdownButtonClass}
					aria-haspopup="listbox"
					aria-expanded={formatMenuOpen}
				>
					<span class="flex items-center gap-2">
						{#if selectedFormatOption}
							<span>{selectedFormatOption.icon}</span>
							<span>{selectedFormatOption.label}</span>
						{:else}
							<span class="text-[#f4f4f4]/60">
								Any format
							</span>
						{/if}
					</span>

					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="18"
						height="18"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						class="shrink-0 text-[#f4f4f4]/60 transition-transform duration-200 {formatMenuOpen
							? 'rotate-180'
							: ''}"
					>
						<path d="m6 9 6 6 6-6" />
					</svg>
				</button>

				{#if formatMenuOpen}
					<div
						role="listbox"
						class="absolute z-10 mt-2 w-full overflow-hidden rounded-xl border border-[#464040] bg-[#1a1818] shadow-lg"
					>
						<button
							type="button"
							role="option"
							aria-selected={selectedFormat === null}
							onclick={() => selectFormat(null)}
							class="flex w-full cursor-pointer items-center px-5 py-3 text-left font-semibold text-[#f4f4f4]/70 transition-colors duration-200 hover:bg-[#3A3535] {selectedFormat === null
								? 'bg-[#3A3535] text-[#f4f4f4]'
								: ''}"
						>
							Any format
						</button>

						{#each FORMAT_OPTIONS as option (option.value)}
							<button
								type="button"
								role="option"
								aria-selected={selectedFormat === option.value}
								onclick={() => selectFormat(option.value)}
								class="flex w-full cursor-pointer items-center gap-2 px-5 py-3 text-left font-semibold text-[#f4f4f4]/70 transition-colors duration-200 hover:bg-[#3A3535] {selectedFormat === option.value
									? 'bg-[#3A3535] text-[#f4f4f4]'
									: ''}"
							>
								<span>{option.icon}</span>
								<span>{option.label}</span>
							</button>
						{/each}
					</div>
				{/if}
			</div>

			<div class="flex flex-col gap-3">
				<p class="text-xs font-black uppercase tracking-[0.14em] text-[#d8d8d8]/60">
					Category
				</p>

				<div class="flex flex-wrap gap-2.5">
					<button
						type="button"
						onclick={() => selectCategory(null)}
						class={pillClass(selectedCategoryId === null)}
					>
						Any category
					</button>

					{#each categories as category (category.id)}
						<button
							type="button"
							onclick={() => selectCategory(category.id)}
							class={pillClass(selectedCategoryId === category.id)}
						>
							{#if category.icon}
								<span>{category.icon}</span>
							{/if}
							<span>{category.name}</span>
						</button>
					{/each}
				</div>
			</div>
		</div>

		<div class="relative mt-12 overflow-hidden rounded-3xl border border-[#464040] bg-[#1a1818]">
			<div
				class="pointer-events-none absolute inset-x-0 top-1/2 h-40 -translate-y-1/2 bg-[#3A3535]"
			></div>

			<div
				bind:this={reelEl}
				class="slot-machine relative h-120 overflow-hidden"
				aria-live="polite"
				aria-label="Generated speaking topic"
			></div>
		</div>

		{#if topic}
			<div class="mt-6 flex flex-wrap justify-center gap-2">
				{#if topicFormatOption}
					<span
						class="rounded-full border border-[#464040] bg-[#3A3535] px-3.5 py-1.5 text-xs font-bold text-[#f4f4f4]"
					>
						{topicFormatOption.icon}
						{topicFormatOption.label}
					</span>
				{/if}

				{#if topicCategory}
					<span
						class="rounded-full border border-[#464040] bg-[#3A3535] px-3.5 py-1.5 text-xs font-bold text-[#f4f4f4]"
					>
						{topicCategory.icon ?? ''}
						{topicCategory.name}
					</span>
				{/if}
			</div>
		{/if}

		<div class="mt-8 flex justify-center">
			<button
				type="button"
				onclick={spin}
				disabled={spinning}
				class={generateButtonClass}
			>
				{#if spinning}
					<div
						class="h-5 w-5 shrink-0 animate-spin rounded-full border border-[#f4f4f4]/25 border-t-[#f4f4f4]"
					></div>

					Spinning…
				{:else if topic}
					Spin again
				{:else}
					Generate topic
				{/if}
			</button>
		</div>

		{#if topic}
			<div class="mx-auto mt-12 flex max-w-md flex-col gap-10">

				<div class="flex flex-col gap-3">
					<p class="text-xs font-black uppercase tracking-[0.14em] text-[#d8d8d8]/60">
						Prep time
					</p>

					<div class="grid grid-cols-2 gap-2.5 sm:grid-cols-4">
						{#each PREP_TIME_OPTIONS as option (option.value)}
							<button
								type="button"
								onclick={() => (prepTimeSeconds = option.value)}
								class={`${pillClass(prepTimeSeconds === option.value)} rounded-xl py-3 text-center`}
							>
								{option.label}
							</button>
						{/each}
					</div>
				</div>

				{#if topic.format === 'debate'}
					<div class="flex flex-col gap-3">
						<p class="text-xs font-black uppercase tracking-[0.14em] text-[#d8d8d8]/60">
							Your stance
						</p>

						<div class="grid grid-cols-2 gap-2.5">
							<button
								type="button"
								onclick={() => (debateStance = 'against')}
								class={`${pillClass(debateStance === 'against')} rounded-xl py-3 text-center`}
							>
								Against
							</button>

							<button
								type="button"
								onclick={() => (debateStance = 'for')}
								class={`${pillClass(debateStance === 'for')} rounded-xl py-3 text-center`}
							>
								For
							</button>
						</div>
					</div>
				{/if}

				<button
					type="button"
					onclick={start}
					disabled={starting}
					class={primaryButtonClass}
				>
					{#if starting}
						<div
							class="h-5 w-5 shrink-0 animate-spin rounded-full border border-[#232020]/25 border-t-[#232020]"
						></div>
						Starting…
					{:else}
						Start speaking →
					{/if}
				</button>
			</div>
		{/if}

		{#if errorMessage}
			<p class="mx-auto mt-8 max-w-md text-center text-xs font-semibold leading-relaxed text-[#FF6B7A]">
				{errorMessage}
			</p>
		{/if}
	</div>
</div>

<style>
	:global(.slot-machine .reel-track) {
		will-change: transform;
	}

	:global(.slot-machine .reel-item) {
		box-sizing: border-box;
		display: flex;
		height: 160px;
		align-items: center;
		justify-content: center;
		padding: 1.25rem 1.75rem;
		text-align: center;
		font-size: clamp(0.9rem, 2.6vw, 1rem);
		font-weight: 700;
		line-height: 1.5;
		color: rgba(244, 244, 244, 0.4);
	}

	:global(.slot-machine .reel-item--dim) {
		opacity: 0.5;
	}

	:global(.slot-machine .reel-item--current) {
		font-size: clamp(1.05rem, 3vw, 1.25rem);
		font-weight: 900;
		color: #f4f4f4;
	}

	@media (min-width: 640px) {
		:global(.slot-machine .reel-item) {
			padding: 1.5rem 2.5rem;
		}
	}
</style>