<!-- src/routes/leaderboard/+page.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getLeaderboard, getCountryLeaderboard, getMyLeaderboardRank } from '$lib/api';
	import type { LeaderboardEntry, CountryLeaderboardEntry } from '$lib/types';

	let user = $derived(page.data.user);

	type Tab = 'global' | 'country';
	type Entry = LeaderboardEntry | CountryLeaderboardEntry;

	let tab = $state<Tab>('global');
	let entries = $state<LeaderboardEntry[]>([]);
	let countryEntries = $state<CountryLeaderboardEntry[]>([]);
	let myRank = $state<LeaderboardEntry | null>(null);

	let loading = $state(true);
	let loadingMore = $state(false);
	let offset = $state(0);
	let hasMore = $state(true);

	const PAGE_SIZE = 20;

	function displayName(entry: Entry): string {
		return entry.username ?? `Speaker #${entry.rank}`;
	}

	function initials(entry: Entry): string {
		const name = displayName(entry);

		return name
			.split(' ')
			.map((part) => part[0])
			.filter(Boolean)
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}

	function medalColor(rank: number): string {
		if (rank === 1) return '#F7E396';
		if (rank === 2) return '#E6EABF';
		if (rank === 3) return '#ff94d0';
		return '#F7FFCD';
	}

	async function loadGlobal(reset: boolean) {
		if (reset) {
			loading = true;
			offset = 0;
		} else {
			loadingMore = true;
		}

		try {
			const next = await getLeaderboard(PAGE_SIZE, reset ? 0 : offset);

			entries = reset ? next : [...entries, ...next];
			offset = (reset ? 0 : offset) + next.length;
			hasMore = next.length === PAGE_SIZE;
		} finally {
			loading = false;
			loadingMore = false;
		}
	}

	async function loadCountry(reset: boolean) {
		if (!user?.countryCode) return;

		if (reset) {
			loading = true;
			offset = 0;
		} else {
			loadingMore = true;
		}

		try {
			const next = await getCountryLeaderboard(
				user.countryCode,
				PAGE_SIZE,
				reset ? 0 : offset
			);

			countryEntries = reset ? next : [...countryEntries, ...next];
			offset = (reset ? 0 : offset) + next.length;
			hasMore = next.length === PAGE_SIZE;
		} finally {
			loading = false;
			loadingMore = false;
		}
	}

	function switchTab(next: Tab) {
		if (tab === next) return;

		tab = next;

		if (next === 'global' && entries.length === 0) {
			loadGlobal(true);
		} else if (next === 'country' && countryEntries.length === 0) {
			loadCountry(true);
		}
	}

	function loadMore() {
		if (loadingMore || !hasMore) return;

		if (tab === 'global') loadGlobal(false);
		else loadCountry(false);
	}

	onMount(async () => {
		await loadGlobal(true);

		try {
			myRank = await getMyLeaderboardRank();
		} catch {
			myRank = null;
		}
	});

	let visibleEntries = $derived<Entry[]>(tab === 'global' ? entries : countryEntries);
	let podium = $derived(visibleEntries.slice(0, 3));
	let rest = $derived(visibleEntries.slice(3));
</script>

<svelte:head>
	<title>Flocal — Leaderboard</title>
</svelte:head>

<div class="min-h-screen bg-[#BADF96] px-5 py-8 text-[#1C1124] sm:px-8 sm:py-12 lg:px-12">
	<div class="mx-auto flex w-full max-w-4xl flex-col gap-8">
		<header>
			<a
				href="/profile"
				class="mb-4 inline-flex items-center gap-1.5 text-sm font-bold text-[#4d2a3a] transition-transform duration-200 hover:-translate-x-1"
			>
				<span class="text-lg">←</span>
				Profile
			</a>

			<div
				class="inline-flex rounded-full border-2 border-[#1C1124] bg-[#F7FFCD] px-3.5 py-1.5 text-xs font-black uppercase tracking-[0.12em]"
			>
				Leaderboard
			</div>

			<h1 class="mt-4 text-4xl font-black tracking-[-0.045em] sm:text-5xl md:text-6xl">
				Rankings
			</h1>

			<p class="mt-3 max-w-xl text-sm leading-relaxed text-[#4d2a3a] sm:text-base">
				See how your XP stacks up against everyone else practising on Flocal.
			</p>
		</header>

		{#if myRank}
			<section
				class="flex items-center justify-between gap-4 rounded-3xl border-2 border-[#1C1124] bg-[#1C1124] p-5 text-[#F7FFCD] shadow-[7px_7px_0px_#F7FFCD] sm:p-6"
			>
				<div class="flex items-center gap-4">
					{#if myRank.image}
						<img
							src={myRank.image}
							alt={displayName(myRank)}
							class="h-12 w-12 rounded-xl border-2 border-[#F7FFCD] object-cover"
						/>
					{:else}
						<div
							class="flex h-12 w-12 items-center justify-center rounded-xl border-2 border-[#F7FFCD] bg-[#9FA1FF] text-sm font-black text-[#1C1124]"
						>
							{initials(myRank)}
						</div>
					{/if}

					<div>
						<p class="text-xs font-black uppercase tracking-[0.14em] text-[#F7FFCD]/55">
							Your rank
						</p>
						<p class="mt-1 text-2xl font-black">#{myRank.rank}</p>
					</div>
				</div>

				<div class="flex items-center gap-6">
					{#if myRank.currentStreakDays > 0}
						<div class="text-right">
							<p class="text-xs font-black uppercase tracking-[0.14em] text-[#F7FFCD]/55">
								Streak
							</p>
							<p class="mt-1 text-2xl font-black tabular-nums">🔥 {myRank.currentStreakDays}</p>
						</div>
					{/if}

					<div class="text-right">
						<p class="text-xs font-black uppercase tracking-[0.14em] text-[#F7FFCD]/55">XP</p>
						<p class="mt-1 text-2xl font-black tabular-nums">{myRank.totalXp}</p>
					</div>
				</div>
			</section>
		{/if}

		<div class="flex gap-2">
			<button
				type="button"
				onclick={() => switchTab('global')}
				class="cursor-pointer rounded-xl border-2 border-[#1C1124] px-4 py-2.5 text-xs font-black transition-transform duration-200 hover:-translate-y-0.5 {tab ===
				'global'
					? 'bg-[#9FA1FF]'
					: 'bg-[#F7FFCD]'}"
			>
				🌍 Global
			</button>

			<button
				type="button"
				onclick={() => switchTab('country')}
				disabled={!user?.countryCode}
				class="cursor-pointer rounded-xl border-2 border-[#1C1124] px-4 py-2.5 text-xs font-black transition-transform duration-200 hover:-translate-y-0.5 disabled:cursor-not-allowed disabled:opacity-40 {tab ===
				'country'
					? 'bg-[#9FA1FF]'
					: 'bg-[#F7FFCD]'}"
			>
				{user?.countryCode ? `🏳️ ${user.countryCode}` : '🏳️ Set country'}
			</button>
		</div>

		{#if loading}
			<div
				class="rounded-3xl border-2 border-[#1C1124] bg-[#F7FFCD] px-6 py-12 text-center shadow-[6px_6px_0px_#1C1124]"
			>
				<p class="text-sm font-bold">Loading rankings…</p>
			</div>
		{:else if visibleEntries.length === 0}
			<div
				class="rounded-3xl border-2 border-[#1C1124] bg-[#F7FFCD] px-6 py-14 text-center shadow-[6px_6px_0px_#1C1124]"
			>
				<h2 class="text-xl font-black">No rankings yet</h2>
				<p class="mx-auto mt-2 max-w-sm text-sm leading-relaxed text-[#4d2a3a]">
					{tab === 'country'
						? 'No one from your country has ranked yet — be the first.'
						: 'Complete a session to appear on the leaderboard.'}
				</p>
			</div>
		{:else}
			{#if podium.length > 0}
				<div class="grid grid-cols-3 gap-3">
					{#each podium as entry (entry.userId)}
						<div
							class="flex flex-col items-center rounded-2xl border-2 border-[#1C1124] p-4 shadow-[5px_5px_0px_#1C1124] {entry.userId ===
							user?.id
								? 'ring-4 ring-[#ff94d0]'
								: ''}"
							style="background-color: {medalColor(entry.rank)}"
						>
							{#if entry.image}
								<img
									src={entry.image}
									alt={displayName(entry)}
									class="h-14 w-14 rounded-xl border-2 border-[#1C1124] object-cover"
								/>
							{:else}
								<div
									class="flex h-14 w-14 items-center justify-center rounded-xl border-2 border-[#1C1124] bg-white text-sm font-black"
								>
									{initials(entry)}
								</div>
							{/if}

							<p class="mt-3 text-lg font-black">#{entry.rank}</p>
							<p class="mt-1 truncate text-xs font-bold">{displayName(entry)}</p>
							<p class="mt-1 text-sm font-black tabular-nums">{entry.totalXp} XP</p>

							{#if entry.currentStreakDays > 0}
								<p class="mt-0.5 text-[10px] font-bold text-[#4d2a3a]/60">
									🔥 {entry.currentStreakDays}d
								</p>
							{/if}
						</div>
					{/each}
				</div>
			{/if}

			{#if rest.length > 0}
				<div
					class="overflow-hidden rounded-3xl border-2 border-[#1C1124] bg-[#F7FFCD] shadow-[6px_6px_0px_#1C1124]"
				>
					{#each rest as entry, index (entry.userId)}
						<div
							class="flex items-center gap-4 px-5 py-4 sm:px-6 {entry.userId === user?.id
								? 'bg-[#F7E396]'
								: ''}"
						>
							<span class="w-8 shrink-0 text-sm font-black tabular-nums text-[#4d2a3a]/60">
								#{entry.rank}
							</span>

							{#if entry.image}
								<img
									src={entry.image}
									alt={displayName(entry)}
									class="h-9 w-9 shrink-0 rounded-lg border-2 border-[#1C1124] object-cover"
								/>
							{:else}
								<div
									class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border-2 border-[#1C1124] bg-[#9FA1FF] text-[10px] font-black"
								>
									{initials(entry)}
								</div>
							{/if}

							<span class="min-w-0 flex-1 truncate text-sm font-bold">{displayName(entry)}</span>

							{#if entry.currentStreakDays > 0}
								<span class="shrink-0 text-xs font-bold text-[#4d2a3a]/50">
									🔥{entry.currentStreakDays}
								</span>
							{/if}

							<span class="shrink-0 text-sm font-black tabular-nums">{entry.totalXp} XP</span>
						</div>

						{#if index < rest.length - 1}
							<div class="mx-5 h-px bg-[#1C1124]/10 sm:mx-6"></div>
						{/if}
					{/each}
				</div>

				{#if hasMore}
					<button
						type="button"
						onclick={loadMore}
						disabled={loadingMore}
						class="mx-auto mt-1 cursor-pointer rounded-xl border-2 border-[#1C1124] bg-[#1C1124] px-6 py-3 text-sm font-black text-[#F7FFCD] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-60"
					>
						{loadingMore ? 'Loading…' : 'Load more'}
					</button>
				{/if}
			{/if}
		{/if}
	</div>
</div>