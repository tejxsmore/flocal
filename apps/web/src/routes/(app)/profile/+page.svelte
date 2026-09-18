<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import {
		getMyStats,
		listMyDailyActivity,
		listSessionHistory,
		getCurrentPlan,
		listLinkedAccounts,
		linkAccountUrl,
		ApiRequestError
	} from '$lib/api';
	import { FORMAT_OPTIONS } from '$lib/types';
	import type {
		DailyActivity,
		UserStats,
		SessionHistoryItem,
		UserCurrentPlan,
		LinkedAccount,
		OAuthProvider
	} from '$lib/types';

	let user = $derived(page.data.user);

	function initials(name: string): string {
		return name
			.split(' ')
			.map((part) => part[0])
			.filter(Boolean)
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}

	let stats = $state<UserStats | null>(null);
	let activity = $state<DailyActivity[]>([]);
	let plan = $state<UserCurrentPlan | null>(null);
	let recentHistory = $state<SessionHistoryItem[]>([]);
	let linkedAccounts = $state<LinkedAccount[]>([]);

	let loading = $state(true);
	let noStatsYet = $state(false);
	let errorMessage = $state('');

	const PROVIDER_LABELS: Record<OAuthProvider, string> = {
		google: 'Google',
		github: 'GitHub'
	};

	const OAUTH_PROVIDERS: OAuthProvider[] = ['google', 'github'];

	function isConnected(provider: OAuthProvider): boolean {
		return linkedAccounts.some((a) => a.provider === provider && a.connected);
	}

	function formatIcon(format: string): string {
		return FORMAT_OPTIONS.find((f) => f.value === format)?.icon ?? '🗣️';
	}

	function formatScore(value: number | null | undefined): string {
		if (value === null || value === undefined) return '—';
		return Math.round(value).toString();
	}

	function formatDate(value: string): string {
		return new Date(value).toLocaleDateString(undefined, {
			month: 'short',
			day: 'numeric'
		});
	}

	function statusLabel(status: SessionHistoryItem['status']): string {
		if (status === 'processing') return 'Processing';
		if (status === 'failed') return 'Failed';
		if (status === 'pending') return 'Pending';
		return '';
	}

	function buildHeatmapWeeks(days: DailyActivity[]) {
		const byDate = new Map(days.map((d) => [d.activityDate, d.sessionsCount]));
		const cells: { date: string; count: number }[] = [];
		const today = new Date();

		for (let i = 90; i >= 0; i--) {
			const d = new Date(today);
			d.setDate(d.getDate() - i);

			const key = d.toISOString().slice(0, 10);

			cells.push({
				date: key,
				count: byDate.get(key) ?? 0
			});
		}

		const weeks: { date: string; count: number }[][] = [];

		for (let i = 0; i < cells.length; i += 7) {
			weeks.push(cells.slice(i, i + 7));
		}

		return weeks;
	}

	function heatColor(count: number): string {
		if (count <= 0) return '#26224d';
		if (count === 1) return '#38255b';
		if (count === 2) return '#5C3E94';
		return '#F25912';
	}

	let heatmapWeeks = $derived(buildHeatmapWeeks(activity));

	onMount(async () => {
		try {
			stats = await getMyStats();
		} catch (err) {
			if (err instanceof ApiRequestError && err.status === 404) {
				noStatsYet = true;
			} else {
				errorMessage =
					err instanceof ApiRequestError ? err.message : 'Could not load stats.';
			}
		}

		try {
			activity = await listMyDailyActivity(91);
		} catch {
			activity = [];
		}

		try {
			plan = await getCurrentPlan();
		} catch {
			plan = null;
		}

		try {
			recentHistory = await listSessionHistory(5, 0);
		} catch {
			recentHistory = [];
		}

		try {
			linkedAccounts = await listLinkedAccounts();
		} catch {
			linkedAccounts = [];
		}

		loading = false;
	});

	const primaryButtonClass =
		'flex h-13.5 cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#c64c13] bg-[#F25912] px-6 font-bold text-[#181C14] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:translate-y-0';

	const secondaryButtonClass =
		'flex h-11 cursor-pointer items-center justify-center gap-2 rounded-xl border border-[#412B6B] bg-[#38255b] px-4 text-xs font-bold text-[#f4f0e4] transition-transform duration-200 hover:-translate-y-0.5 sm:text-sm';

	const cardClass = 'rounded-3xl border border-[#412B6B] bg-[#171022] p-6 sm:p-7';
	const sectionLabelClass = 'text-xs font-black uppercase tracking-[0.14em] text-[#f4f0e4]/45';
</script>

<svelte:head>
	<title>Flocal — Profile</title>
</svelte:head>

<div class="min-h-screen bg-[#1b1833] px-5 py-7 text-[#f4f0e4] sm:px-8 sm:py-10">
	<div class="mx-auto w-full max-w-3xl">
		<!-- NAV -->
		<header class="flex items-center justify-between">
			<a
				href="/"
				class="text-2xl font-black tracking-tight transition-transform duration-200 hover:-translate-y-0.5"
			>
				Flocal
			</a>

			<div class="flex items-center gap-2">
				<a href="/session" class={secondaryButtonClass}>Sessions</a>
				<a href="/profile/settings" class="hidden sm:flex {secondaryButtonClass}">Settings</a>
			</div>
		</header>

		{#if loading}
			<!-- LOADING -->
			<div class="flex min-h-[70vh] items-center justify-center">
				<div class="rounded-3xl border border-[#412B6B] bg-[#171022] px-8 py-6 text-center">
					<div
						class="mx-auto h-5 w-5 animate-spin rounded-full border border-[#f4f0e4]/25 border-t-[#f4f0e4]"
					></div>
					<p class="mt-4 text-sm font-bold text-[#f4f0e4]/70">Loading your profile…</p>
				</div>
			</div>
		{:else if errorMessage}
			<!-- ERROR -->
			<div class="flex min-h-[70vh] items-center justify-center">
				<div class="w-full max-w-md rounded-3xl border border-[#412B6B] bg-[#171022] p-8 text-center">
					<div
						class="mx-auto flex h-12 w-12 items-center justify-center rounded-full border border-[#FF6B7A] font-black text-[#FF6B7A]"
					>
						!
					</div>

					<h1 class="mt-5 text-xl font-black">Couldn't load your profile</h1>

					<p class="mt-2 text-sm text-[#f4f0e4]/60">
						{errorMessage}
					</p>

					<a href="/" class={`mt-6 inline-flex ${primaryButtonClass}`}>Back home</a>
				</div>
			</div>
		{:else}
			<main class="mt-10 flex flex-col gap-6 pb-14">
				<!-- PROFILE HERO -->
				<section class="relative overflow-hidden rounded-3xl border border-[#412B6B] bg-[#171022] p-6 sm:p-8">
					<div
						class="pointer-events-none absolute inset-x-0 top-1/2 h-40 -translate-y-1/2 bg-[#38255b] opacity-40"
					></div>

					<div class="relative flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
						<div class="flex items-center gap-5">
							{#if user?.image}
								<img
									src={user.image}
									alt={user.name}
									class="h-20 w-20 rounded-2xl border border-[#412B6B] object-cover sm:h-24 sm:w-24"
								/>
							{:else if user}
								<div
									class="flex h-20 w-20 items-center justify-center rounded-2xl border border-[#412B6B] bg-[#38255b] text-2xl font-black sm:h-24 sm:w-24 sm:text-3xl"
								>
									{initials(user.name)}
								</div>
							{/if}

							<div class="min-w-0">
								<p class={sectionLabelClass}>Your Flocal profile</p>

								<h1 class="mt-1 truncate text-3xl font-black tracking-[-0.02em] sm:text-4xl">
									{user?.name}
								</h1>

								<p class="mt-1 truncate text-sm text-[#f4f0e4]/55">
									{user?.email}
								</p>
							</div>
						</div>

						<a href="/" class={`relative shrink-0 ${primaryButtonClass}`}>Start a session →</a>
					</div>
				</section>

				<!-- PLAN -->
				{#if plan}
					<section
						class="grid overflow-hidden rounded-3xl border border-[#412B6B] bg-[#38255b] sm:grid-cols-[1fr_auto]"
					>
						<div class="p-6 sm:p-7">
							<div class="flex items-center gap-3">
								<span
									class="rounded-full border border-[#c64c13] bg-[#F25912] px-3 py-1 text-[10px] font-black uppercase tracking-[0.12em] text-[#181C14]"
								>
									Current plan
								</span>

								<span class="text-xs font-bold text-[#f4f0e4]/50">
									{plan.planSlug === 'free' ? 'Getting started' : 'Pro member'}
								</span>
							</div>

							<h2 class="mt-4 text-2xl font-black capitalize sm:text-3xl">
								{plan.planSlug}
							</h2>

							<p class="mt-1 text-sm text-[#f4f0e4]/55">
								{#if plan.dailySessionLimit !== null && plan.dailySessionLimit !== undefined}
									{plan.dailySessionLimit} AI sessions per day
								{:else}
									Unlimited AI sessions
								{/if}
							</p>
						</div>

						<div class="flex items-center border-t border-[#412B6B] p-5 sm:border-l sm:border-t-0 sm:px-7">
							{#if plan.planSlug === 'free'}
								<a href="/plan" class={`w-full text-center sm:w-auto ${primaryButtonClass}`}>
									Upgrade to Pro
								</a>
							{:else}
								<a
									href="/plan"
									class="flex h-13.5 w-full cursor-pointer items-center justify-center rounded-xl border border-[#412B6B] bg-transparent px-6 text-sm font-bold text-[#f4f0e4] transition-transform duration-200 hover:-translate-y-1 sm:w-auto"
								>
									Manage plan
								</a>
							{/if}
						</div>
					</section>
				{/if}

				{#if noStatsYet}
					<!-- EMPTY STATE -->
					<section class="rounded-3xl border border-[#412B6B] bg-[#171022] p-8 text-center sm:p-12">
						<div
							class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl border border-[#c64c13] bg-[#F25912] text-2xl font-black text-[#181C14]"
						>
							1
						</div>

						<h2 class="mt-6 text-2xl font-black">Your speaking story starts here.</h2>

						<p class="mx-auto mt-2 max-w-md text-sm leading-relaxed text-[#f4f0e4]/55">
							Finish your first session and Flocal will start tracking your streak,
							scores, XP, and progress.
						</p>

						<a href="/" class={`mt-7 inline-flex ${primaryButtonClass}`}>Spin a topic →</a>
					</section>
				{:else if stats}
					<!-- STATS -->
					<section class="flex flex-col gap-3">
						<div class="px-1">
							<p class={sectionLabelClass}>Your progress</p>
							<h2 class="mt-1 text-2xl font-black tracking-[-0.02em]">Keep showing up.</h2>
						</div>

						<div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
							<div class={cardClass}>
								<p class={sectionLabelClass}>Current streak</p>
								<div class="mt-5 text-5xl font-black tabular-nums tracking-tighter text-[#F25912]">
									{stats.currentStreakDays}
								</div>
								<p class="mt-1 text-xs font-bold text-[#f4f0e4]/45">
									{stats.currentStreakDays === 1 ? 'day' : 'days'}
								</p>
							</div>

							<div class={cardClass}>
								<p class={sectionLabelClass}>Total XP</p>
								<div class="mt-5 text-5xl font-black tabular-nums tracking-tighter text-[#87A2FF]">
									{stats.totalXp}
								</div>
								<p class="mt-1 text-xs font-bold text-[#f4f0e4]/45">earned so far</p>
							</div>

							<div class={cardClass}>
								<p class={sectionLabelClass}>Sessions</p>
								<div class="mt-5 text-5xl font-black tabular-nums tracking-tighter text-[#F13E93]">
									{stats.totalSessions}
								</div>
								<p class="mt-1 text-xs font-bold text-[#f4f0e4]/45">completed</p>
							</div>

							<div class={cardClass}>
								<p class={sectionLabelClass}>Average</p>
								<div class="mt-5 text-5xl font-black tabular-nums tracking-tighter">
									{stats.averageScore.toFixed(1)}
								</div>
								<p class="mt-1 text-xs font-bold text-[#f4f0e4]/45">
									best {stats.bestScore.toFixed(1)}
								</p>
							</div>
						</div>
					</section>

					<!-- SECONDARY STATS -->
					<div class="grid grid-cols-2 gap-3">
						<div class="flex items-center justify-between rounded-2xl border border-[#412B6B] bg-[#171022] px-5 py-4">
							<span class="text-xs font-black uppercase tracking-wide text-[#f4f0e4]/45">
								Best streak
							</span>
							<span class="text-xl font-black tabular-nums">{stats.longestStreakDays}d</span>
						</div>

						<div class="flex items-center justify-between rounded-2xl border border-[#412B6B] bg-[#171022] px-5 py-4">
							<span class="text-xs font-black uppercase tracking-wide text-[#f4f0e4]/45">
								Best score
							</span>
							<span class="text-xl font-black tabular-nums">{stats.bestScore.toFixed(1)}</span>
						</div>
					</div>
				{/if}

				<!-- ACTIVITY -->
				{#if activity.length > 0}
					<section class="rounded-3xl border border-[#412B6B] bg-[#171022] p-5 sm:p-7">
						<div class="flex items-end justify-between">
							<div>
								<p class={sectionLabelClass}>Consistency</p>
								<h2 class="mt-1 text-xl font-black">Last 90 days</h2>
							</div>

							<span class="text-xs font-bold text-[#f4f0e4]/35">Practice activity</span>
						</div>

						<div class="mt-6 overflow-x-auto pb-1">
							<div class="flex w-max gap-1.5">
								{#each heatmapWeeks as week, i (i)}
									<div class="flex flex-col gap-1.5">
										{#each week as cell (cell.date)}
											<div
												title="{cell.date}: {cell.count} session{cell.count === 1 ? '' : 's'}"
												class="h-3.5 w-3.5 rounded-sm border border-[#f4f0e4]/5"
												style="background-color: {heatColor(cell.count)}"
											></div>
										{/each}
									</div>
								{/each}
							</div>
						</div>

						<div class="mt-5 flex items-center justify-end gap-2 text-[10px] font-bold text-[#f4f0e4]/35">
							Less
							<span class="h-3.5 w-3.5 rounded-sm bg-[#26224d]"></span>
							<span class="h-3.5 w-3.5 rounded-sm bg-[#38255b]"></span>
							<span class="h-3.5 w-3.5 rounded-sm bg-[#5C3E94]"></span>
							<span class="h-3.5 w-3.5 rounded-sm bg-[#F25912]"></span>
							More
						</div>
					</section>
				{/if}

				<!-- RECENT SESSIONS -->
				<section class="flex flex-col gap-3">
					<div class="flex items-end justify-between px-1">
						<div>
							<p class={sectionLabelClass}>Your practice</p>
							<h2 class="mt-1 text-xl font-black">Recent sessions</h2>
						</div>

						<a
							href="/session"
							class="text-xs font-black text-[#f4f0e4]/70 underline decoration-2 underline-offset-4 transition-colors hover:text-[#f4f0e4]"
						>
							View all
						</a>
					</div>

					{#if recentHistory.length === 0}
						<div class="rounded-3xl border border-[#412B6B] bg-[#171022] p-7 text-center">
							<p class="text-sm font-bold text-[#f4f0e4]/70">No sessions yet.</p>

							<a
								href="/"
								class="mt-4 inline-block text-sm font-black underline decoration-2 underline-offset-4"
							>
								Start speaking →
							</a>
						</div>
					{:else}
						<div class="overflow-hidden rounded-3xl border border-[#412B6B] bg-[#171022]">
							{#each recentHistory as item, index (item.id)}
								<a
									href="/session/{item.id}"
									class="group flex items-center gap-4 px-5 py-4 transition-colors duration-200 hover:bg-[#38255b] sm:px-6"
								>
									<div
										class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-[#412B6B] bg-[#26224d] text-lg"
									>
										{formatIcon(item.topicFormat)}
									</div>

									<div class="min-w-0 flex-1">
										<p class="truncate text-sm font-black">{item.topicTitle}</p>

										<p class="mt-1 text-xs text-[#f4f0e4]/40">
											{formatDate(item.createdAt)}

											{#if item.wordsPerMinute !== null && item.wordsPerMinute !== undefined}
												· {Math.round(item.wordsPerMinute)} wpm
											{/if}
										</p>
									</div>

									<div class="flex shrink-0 items-center gap-3">
										{#if item.status === 'completed'}
											<span class="text-xl font-black tabular-nums">
												{formatScore(item.overallScore)}
											</span>
										{:else}
											<span
												class="text-[10px] font-black uppercase tracking-wide {item.status ===
												'failed'
													? 'text-[#FF6B7A]'
													: 'text-[#f4f0e4]/40'}"
											>
												{statusLabel(item.status)}
											</span>
										{/if}

										<span class="text-lg font-black transition-transform duration-200 group-hover:translate-x-1">
											→
										</span>
									</div>
								</a>

								{#if index < recentHistory.length - 1}
									<div class="mx-5 h-px bg-[#412B6B] sm:mx-6"></div>
								{/if}
							{/each}
						</div>
					{/if}
				</section>

				<!-- ACCOUNTS -->
				<section class="grid gap-3 md:grid-cols-2">
					<div class="rounded-3xl border border-[#412B6B] bg-[#171022] p-6">
						<p class={sectionLabelClass}>Account</p>
						<h2 class="mt-1 text-xl font-black">Connected accounts</h2>

						<p class="mt-1 text-xs leading-relaxed text-[#f4f0e4]/45">
							Manage the services connected to your Flocal account.
						</p>

						<div class="mt-5 flex flex-col gap-2">
							{#each OAUTH_PROVIDERS as provider (provider)}
								<div class="flex items-center justify-between gap-3 rounded-2xl border border-[#412B6B] bg-[#26224d] px-4 py-3">
									<div class="flex items-center gap-3">
										<div class="flex h-9 w-9 items-center justify-center rounded-lg border border-[#412B6B] bg-[#38255b] text-xs font-black">
											{provider === 'google' ? 'G' : 'GH'}
										</div>

										<span class="text-sm font-black">{PROVIDER_LABELS[provider]}</span>
									</div>

									{#if isConnected(provider)}
										<span class="rounded-full border border-[#412B6B] bg-[#38255b] px-2.5 py-1 text-[10px] font-black">
											Connected
										</span>
									{:else}
										<a
											href={linkAccountUrl(provider)}
											class="rounded-lg border border-[#c64c13] bg-[#F25912] px-3 py-1.5 text-xs font-black text-[#181C14] transition-transform duration-200 hover:-translate-y-0.5"
										>
											Connect
										</a>
									{/if}
								</div>
							{/each}
						</div>
					</div>

					<div class="flex flex-col justify-between rounded-3xl border border-[#412B6B] bg-[#38255b] p-6">
						<div>
							<p class={sectionLabelClass}>Keep improving</p>

							<h2 class="mt-2 max-w-xs text-2xl font-black leading-tight">
								One minute of speaking is better than zero.
							</h2>

							<p class="mt-3 max-w-sm text-sm leading-relaxed text-[#f4f0e4]/60">
								Build the habit, collect the reps, and let the scores follow.
							</p>
						</div>

						<a href="/" class={`mt-7 inline-flex w-fit ${primaryButtonClass}`}>Practice now →</a>
					</div>
				</section>
			</main>
		{/if}
	</div>
</div>