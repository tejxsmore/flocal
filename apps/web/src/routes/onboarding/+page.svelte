<script lang="ts">
	import { tick } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import posthog from 'posthog-js';
	import {
		completeProfileOnboarding,
		completePreferencesOnboarding,
		ApiRequestError
	} from '$lib/api';
	import {
		GOAL_OPTIONS,
		TIME_COMMITMENT_OPTIONS,
		FOCUS_AREA_OPTIONS,
		type OnboardingGoal,
		type DailyTimeCommitment,
		type FocusArea
	} from '$lib/types';

	let user = page.data.user;

	let step = $state(0);
	let maxStepVisited = $state(0);
	const totalSteps = 5;

	let name = $state(user.name ?? '');
	let username = $state(user.username ?? '');
	let image = $state(user.image ?? '');

	let nameInput: HTMLInputElement | undefined = $state();
	let usernameInput: HTMLInputElement | undefined = $state();

	let goals = $state<OnboardingGoal[]>([]);
	let focusAreas = $state<FocusArea[]>([]);
	let dailyTimeCommitment = $state<DailyTimeCommitment | null>(null);

	let submitting = $state(false);
	let submitError = $state('');

	const inputClass =
		'h-13.5 w-full rounded-xl border border-[#464040] bg-[#1a1818] px-10 text-center font-medium placeholder:text-[#f4f4f4]/40 focus:outline-none';
	const primaryButtonClass =
		'flex h-13.5 flex-1 cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#ff6803] bg-[#FF7315] px-5 font-bold text-[#232020] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:translate-y-0';
	const secondaryButtonClass =
		'flex h-13.5 cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#464040] bg-[#3A3535] px-6 font-bold transition-transform duration-200 hover:-translate-y-1';
	const headingClass = 'text-3xl font-black leading-[1.15] tracking-[-0.02em] sm:text-4xl';
	const subtextClass = 'mt-3 text-base leading-relaxed text-[#d8d8d8]';

	function pillClass(selected: boolean) {
		return `cursor-pointer rounded-full border px-4 py-2 text-sm font-bold transition-colors duration-200 ${
			selected
				? 'border-[#464040] bg-[#3A3535] text-[#f4f4f4]'
				: 'border-[#464040] bg-transparent text-[#f4f4f4]/60'
		}`;
	}

	const isInputStep = $derived(step === 0 || step === 1);
	const bodyToActionsGap = $derived(isInputStep ? 'gap-3' : 'gap-8');

	$effect(() => {
		if (step > maxStepVisited) maxStepVisited = step;
	});

	$effect(() => {
		step;
		tick().then(() => {
			if (step === 0) nameInput?.focus();
			else if (step === 1) usernameInput?.focus();
		});
	});

	function canContinue(): boolean {
		if (step === 0) return name.trim().length > 0;

		return true;
	}

	function goToStep(i: number) {
		if (i <= maxStepVisited) {
			submitError = '';
			step = i;
		}
	}

	function next() {
		submitError = '';
		if (step < totalSteps - 1) step += 1;
	}

	function back() {
		submitError = '';
		if (step > 0) step -= 1;
	}

	function toggleGoal(value: OnboardingGoal) {
		goals = goals.includes(value) ? goals.filter((g) => g !== value) : [...goals, value];
	}

	function toggleFocusArea(value: FocusArea) {
		focusAreas = focusAreas.includes(value)
			? focusAreas.filter((f) => f !== value)
			: [...focusAreas, value];
	}

	async function finish() {
		submitting = true;
		submitError = '';

		try {
			await completeProfileOnboarding({
				name,
				username: username.trim() || null,
				image: image.trim() || null
			});

			await completePreferencesOnboarding({
				primaryGoal: goals[0] ?? null,
				dailyTimeCommitment,
				focusAreas
			});

			posthog.capture('onboarding_completed', {
				goal_count: goals.length,
				focus_area_count: focusAreas.length,
				has_daily_time_commitment: dailyTimeCommitment !== null
			});

			await goto('/');
		} catch (err) {
			submitError = err instanceof ApiRequestError ? err.message : 'Could not finish onboarding.';
			submitting = false;
		}
	}

	function handleFormSubmit(event: SubmitEvent) {
		event.preventDefault();

		if (!canContinue()) return;

		if (step === totalSteps - 1) {
			finish();
		} else {
			next();
		}
	}
</script>

{#snippet clearButton(onClear: () => void, label: string)}
	<button
		type="button"
		aria-label={label}
		onclick={onClear}
		class="absolute right-4 top-1/2 -translate-y-1/2 cursor-pointer text-[#f4f4f4]/55 transition-colors duration-200 hover:text-[#f4f4f4]"
	>
		<svg
			xmlns="http://www.w3.org/2000/svg"
			width="20"
			height="20"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
		>
			<path d="M18 6 6 18" />
			<path d="m6 6 12 12" />
		</svg>
	</button>
{/snippet}

<svelte:head>
	<title>Flocal — Welcome</title>
</svelte:head>

<div class="flex min-h-screen justify-center px-5 pt-16 pb-10 sm:px-8 sm:pt-24">
	<div class="w-full max-w-sm">

		<div class="mb-12 flex justify-center gap-1.5">
			{#each Array(totalSteps) as _, i (i)}
				<button
					type="button"
					aria-label={`Go to step ${i + 1}`}
					disabled={i > maxStepVisited}
					onclick={() => goToStep(i)}
					class="h-2 w-8 rounded-full border transition-colors duration-200 {i <= step
						? 'border-[#ff6803] bg-[#FF7315]'
						: 'border-[#464040] bg-[#232020]'} {i <= maxStepVisited
						? 'cursor-pointer'
						: 'cursor-default'}"
				></button>
			{/each}
		</div>

		<form onsubmit={handleFormSubmit}>
			<div class="flex flex-col justify-center text-center sm:min-h-28 min-h-26">
				{#if step === 0}
					<h1 class={headingClass}>What should we call you?</h1>
					<p class={subtextClass}>This is how you'll appear across Flocal.</p>
				{:else if step === 1}
					<h1 class={headingClass}>Pick a username</h1>
					<p class={subtextClass}>Optional — you can set it later.</p>
				{:else if step === 2}
					<h1 class={headingClass}>What's your speaking goal?</h1>
					<p class={subtextClass}>Optional — pick as many as apply.</p>
				{:else if step === 3}
					<h1 class={headingClass}>Where do you want to sound confident?</h1>
					<p class={subtextClass}>Optional — pick as many as apply.</p>
				{:else if step === 4}
					<h1 class={headingClass}>How much time can you give it?</h1>
					<p class={subtextClass}>A little each day beats a lot once in a while.</p>
				{/if}
			</div>

			<div class="mt-12 flex flex-col {bodyToActionsGap}">
				<div>
					{#if step === 0}
						<div class="relative w-full">
							<input
								bind:this={nameInput}
								id="name"
								type="text"
								autocomplete="name"
								required
								bind:value={name}
								class={inputClass}
							/>
							{#if name.length > 0}
								{@render clearButton(() => {
									name = '';
									nameInput?.focus();
								}, 'Clear name')}
							{/if}
						</div>
					{:else if step === 1}
						<div class="relative w-full">
							<input
								bind:this={usernameInput}
								id="username"
								type="text"
								autocomplete="username"
								bind:value={username}
								class={inputClass}
							/>
							{#if username.length > 0}
								{@render clearButton(() => {
									username = '';
									usernameInput?.focus();
								}, 'Clear username')}
							{/if}
						</div>
					{:else if step === 2}
						<div class="flex flex-wrap justify-center gap-2.5">
							{#each GOAL_OPTIONS as opt (opt.value)}
								<button
									type="button"
									onclick={() => toggleGoal(opt.value)}
									class={pillClass(goals.includes(opt.value))}
								>
									{opt.label}
								</button>
							{/each}
						</div>
					{:else if step === 3}
						<div class="flex flex-wrap justify-center gap-2.5">
							{#each FOCUS_AREA_OPTIONS as opt (opt.value)}
								<button
									type="button"
									onclick={() => toggleFocusArea(opt.value)}
									class={pillClass(focusAreas.includes(opt.value))}
								>
									{opt.label}
								</button>
							{/each}
						</div>
					{:else if step === 4}
						<div class="flex flex-col gap-3">
							<div class="grid grid-cols-2 gap-3">
								{#each TIME_COMMITMENT_OPTIONS as opt (opt.value)}
									<button
										type="button"
										onclick={() => (dailyTimeCommitment = opt.value)}
										class={`${pillClass(dailyTimeCommitment === opt.value)} rounded-xl px-4 py-3 text-left`}
									>
										{opt.label}
									</button>
								{/each}
							</div>

							<p
								class="text-center text-xs font-semibold leading-relaxed {submitError
									? 'text-[#FF6B7A]'
									: 'invisible'}"
							>
								{submitError || '\u00A0'}
							</p>
						</div>
					{/if}
				</div>

				<div class="flex gap-3">
					{#if step > 0}
						<button type="button" onclick={back} class={secondaryButtonClass}>Back</button>
					{/if}

					{#if step === 0}
						<button type="submit" disabled={!canContinue()} class={primaryButtonClass}>
							Continue
						</button>
					{:else if step < totalSteps - 1}
						<button type="submit" class={primaryButtonClass}>Continue</button>
					{:else}
						<button type="submit" disabled={submitting} class={primaryButtonClass}>
							{#if submitting}
								<div
									class="h-5 w-5 shrink-0 animate-spin rounded-full border border-[#232020]/25 border-t-[#232020]"
								></div>
								Finishing…
							{:else}
								Let's go
							{/if}
						</button>
					{/if}
				</div>
			</div>
		</form>
	</div>
</div>