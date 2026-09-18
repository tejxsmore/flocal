<script lang="ts">
	import { onMount } from 'svelte';
	import posthog from 'posthog-js';
	import { listSessionHistory, deleteSession } from '$lib/api';
	import { FORMAT_OPTIONS, STATUS_OPTIONS, SORT_OPTIONS } from '$lib/types';
	import type {
		SessionHistoryItem,
		SessionStatus,
		TopicFormat,
		SessionSortOption
	} from '$lib/types';

	function formatScore(value: number | null | undefined): string {
		if (value === null || value === undefined) return '—';
		return Math.round(value).toString();
	}

	function formatDate(value: string): string {
		return new Date(value).toLocaleDateString(undefined, {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function statusLabel(status: SessionHistoryItem['status']): string {
		if (status === 'processing') return 'Processing';
		if (status === 'failed') return 'Failed';
		if (status === 'pending') return 'Pending';
		return '';
	}

	function statusClasses(status: SessionHistoryItem['status']): string {
		if (status === 'failed') {
			return 'border-[#FF6B7A]/40 bg-[#FF6B7A]/10 text-[#FF6B7A]';
		}

		return 'border-[#464040] bg-[#3A3535] text-[#f4f4f4]';
	}

	const PAGE_SIZE = 20;

	let searchInput = $state('');
	let debouncedSearch = $state('');
	let statuses = $state<SessionStatus[]>([]);
	let formats = $state<TopicFormat[]>([]);
	let sort = $state<SessionSortOption>('newest');

	let searchInputEl: HTMLInputElement | undefined = $state();

	let items = $state<SessionHistoryItem[]>([]);
	let offset = $state(0);
	let hasMore = $state(true);
	let loading = $state(true);
	let loadingMore = $state(false);
	let deletingId = $state<string | null>(null);
	let deleteModalId = $state<string | null>(null);

	let filterMenuOpen = $state(false);
	let filterMenuEl = $state<HTMLDivElement>();

	let sortMenuOpen = $state(false);
	let sortMenuEl = $state<HTMLDivElement>();

	let searchTimer: ReturnType<typeof setTimeout> | null = null;

	let filterCount = $derived(statuses.length + formats.length);
	let sortActive = $derived(sort !== 'newest');

	let deleteModalItem = $derived(
		deleteModalId ? (items.find((item) => item.id === deleteModalId) ?? null) : null
	);

	function pillClass(selected: boolean) {
		return `cursor-pointer rounded-full border px-4 py-2 text-sm font-bold transition-colors duration-200 ${
			selected
				? 'border-[#464040] bg-[#3A3535] text-[#f4f4f4]'
				: 'border-[#464040] bg-transparent text-[#f4f4f4]/60'
		}`;
	}

	onMount(() => {
		function handleClickOutside(event: MouseEvent) {
			if (
				filterMenuOpen &&
				filterMenuEl &&
				!filterMenuEl.contains(event.target as Node)
			) {
				filterMenuOpen = false;
			}

			if (
				sortMenuOpen &&
				sortMenuEl &&
				!sortMenuEl.contains(event.target as Node)
			) {
				sortMenuOpen = false;
			}
		}

		function handleEscape(event: KeyboardEvent) {
			if (event.key === 'Escape') {
				if (deleteModalId) {
					deleteModalId = null;
				} else {
					filterMenuOpen = false;
					sortMenuOpen = false;
				}
			}
		}

		document.addEventListener('click', handleClickOutside);
		document.addEventListener('keydown', handleEscape);

		return () => {
			document.removeEventListener('click', handleClickOutside);
			document.removeEventListener('keydown', handleEscape);
		};
	});

	function onSearchInput() {
		if (searchTimer) clearTimeout(searchTimer);

		searchTimer = setTimeout(() => {
			debouncedSearch = searchInput.trim();
		}, 400);
	}

	function clearSearch() {
		if (searchTimer) clearTimeout(searchTimer);

		searchInput = '';
		debouncedSearch = '';
		searchInputEl?.focus();
	}

	function toggleStatus(value: SessionStatus) {
		statuses = statuses.includes(value)
			? statuses.filter((s) => s !== value)
			: [...statuses, value];
	}

	function toggleFormat(value: TopicFormat) {
		formats = formats.includes(value)
			? formats.filter((f) => f !== value)
			: [...formats, value];
	}

	function clearFilters() {
		statuses = [];
		formats = [];
	}

	function selectSort(value: SessionSortOption) {
		sort = value;
		sortMenuOpen = false;
	}

	function requestParams() {
		return {
			status: statuses.length ? statuses : undefined,
			format: formats.length ? formats : undefined,
			q: debouncedSearch || undefined,
			sort
		};
	}

	async function reload() {
		loading = true;

		try {
			const next = await listSessionHistory(PAGE_SIZE, 0, requestParams());

			items = next;
			offset = next.length;
			hasMore = next.length === PAGE_SIZE;
		} catch {
			items = [];
			hasMore = false;
		} finally {
			loading = false;
		}
	}

	async function loadMore() {
		if (loadingMore || !hasMore) return;

		loadingMore = true;

		try {
			const next = await listSessionHistory(PAGE_SIZE, offset, requestParams());

			items = [...items, ...next];
			offset += next.length;
			hasMore = next.length === PAGE_SIZE;
		} catch {
			hasMore = false;
		} finally {
			loadingMore = false;
		}
	}

	function requestDelete(id: string) {
		deleteModalId = id;
	}

	function cancelDelete() {
		deleteModalId = null;
	}

	async function confirmDelete() {
		if (!deleteModalId) return;

		const id = deleteModalId;

		deletingId = id;

		try {
			await deleteSession(id);
			items = items.filter((item) => item.id !== id);
			posthog.capture('session_deleted', { session_id: id });
		} finally {
			deletingId = null;
			deleteModalId = null;
		}
	}

	$effect(() => {
		reload();
	});

	const primaryButtonClass =
		'flex h-13.5 cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#ff6803] bg-[#FF7315] px-5 font-bold text-[#232020] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:translate-y-0';

	const secondaryButtonClass =
		'flex h-13.5 cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#464040] bg-[#3A3535] px-6 font-bold text-[#f4f4f4] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:translate-y-0';

	const eyebrowClass = 'text-xs font-black uppercase tracking-[0.14em] text-[#f4f4f4]/60';

	const centeredSpinnerClass =
		'h-6 w-6 animate-spin rounded-full border border-[#464040] border-t-[#FF7315]';

	const dropdownPanelClass =
		'absolute right-0 z-10 mt-2 w-[min(22rem,90vw)] rounded-2xl border border-[#464040] bg-[#232020] p-5 shadow-lg';
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
	<title>Flocal — Sessions</title>
</svelte:head>

<div class="min-h-screen w-full p-6 sm:p-8">
	<div class="mx-auto w-full 2xl:max-w-7xl">

		<div class="flex flex-col gap-4 sm:flex-row">
			<div class="relative flex-1">
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
					class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-[#f4f4f4]/40"
				>
					<circle cx="11" cy="11" r="8" />
					<path d="m21 21-4.35-4.35" />
				</svg>

				<input
					bind:this={searchInputEl}
					type="text"
					placeholder="Search by topic…"
					bind:value={searchInput}
					oninput={onSearchInput}
					class="h-13.5 w-full rounded-xl border border-[#464040] bg-[#1a1818] pl-11 pr-10 text-sm font-medium text-[#f4f4f4] placeholder:text-[#f4f4f4]/40 focus:outline-none"
				/>

				{#if searchInput.length > 0}
					{@render clearButton(clearSearch, 'Clear search')}
				{/if}
			</div>

			<div class="flex gap-4">
				<div class="relative flex-1 sm:flex-none" bind:this={filterMenuEl}>
					<button
						type="button"
						onclick={() => (filterMenuOpen = !filterMenuOpen)}
						class={`${secondaryButtonClass} w-full sm:w-auto`}
					>
						Filters

						{#if filterCount > 0}
							<span class="flex h-5 w-5 items-center justify-center rounded-full bg-[#FF7315] text-[10px] font-black text-[#232020]">
								{filterCount}
							</span>
						{/if}

						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="16"
							height="16"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							class="shrink-0 text-[#f4f4f4]/60 transition-transform duration-200 {filterMenuOpen
								? 'rotate-180'
								: ''}"
						>
							<path d="m6 9 6 6 6-6" />
						</svg>
					</button>

					{#if filterMenuOpen}
						<div class={dropdownPanelClass}>
							<div class="flex flex-col gap-5">

								<div>
									<p class="mb-2.5 {eyebrowClass}">Status</p>

									<div class="flex flex-wrap gap-2">
										{#each STATUS_OPTIONS as opt (opt.value)}
											<button
												type="button"
												onclick={() => toggleStatus(opt.value)}
												class={pillClass(statuses.includes(opt.value))}
											>
												{opt.label}
											</button>
										{/each}
									</div>
								</div>

								<div>
									<p class="mb-2.5 {eyebrowClass}">Format</p>

									<div class="flex flex-wrap gap-2">
										{#each FORMAT_OPTIONS as opt (opt.value)}
											<button
												type="button"
												onclick={() => toggleFormat(opt.value)}
												class={pillClass(formats.includes(opt.value))}
											>
												{opt.icon} {opt.label}
											</button>
										{/each}
									</div>
								</div>

								<button
									type="button"
									onclick={clearFilters}
									disabled={filterCount === 0}
									class="w-full cursor-pointer rounded-xl border border-[#464040] bg-transparent py-2.5 text-xs font-black text-[#f4f4f4]/70 transition-colors duration-200 hover:text-[#f4f4f4] disabled:cursor-not-allowed disabled:opacity-40"
								>
									Clear filters
								</button>
							</div>
						</div>
					{/if}
				</div>

				<div class="relative flex-1 sm:flex-none" bind:this={sortMenuEl}>
					<button
						type="button"
						onclick={() => (sortMenuOpen = !sortMenuOpen)}
						class={`${secondaryButtonClass} w-full sm:w-auto`}
					>
						Sort

						{#if sortActive}
							<span class="flex h-5 w-5 items-center justify-center rounded-full bg-[#FF7315] text-[10px] font-black text-[#232020]">
								1
							</span>
						{/if}

						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="16"
							height="16"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							class="shrink-0 text-[#f4f4f4]/60 transition-transform duration-200 {sortMenuOpen
								? 'rotate-180'
								: ''}"
						>
							<path d="m6 9 6 6 6-6" />
						</svg>
					</button>

					{#if sortMenuOpen}
						<div class="{dropdownPanelClass} w-[min(16rem,90vw)]">
							<div class="flex flex-col gap-2">
								{#each SORT_OPTIONS as opt (opt.value)}
									<button
										type="button"
										onclick={() => selectSort(opt.value)}
										class="flex cursor-pointer items-center justify-between rounded-xl border px-4 py-2.5 text-left text-sm font-bold transition-colors duration-200 {sort ===
										opt.value
											? 'border-[#464040] bg-[#3A3535] text-[#f4f4f4]'
											: 'border-transparent bg-transparent text-[#f4f4f4]/60'}"
									>
										{opt.label}

										{#if sort === opt.value}
											<svg
												xmlns="http://www.w3.org/2000/svg"
												width="16"
												height="16"
												viewBox="0 0 24 24"
												fill="none"
												stroke="currentColor"
												stroke-width="3"
												stroke-linecap="round"
												stroke-linejoin="round"
												class="shrink-0 text-[#FF7315]"
											>
												<path d="M20 6 9 17l-5-5" />
											</svg>
										{/if}
									</button>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>

		<div class="mt-8 flex flex-col gap-4">
			<p class={eyebrowClass}>
				{#if loading}
					Loading your practice history…
				{:else}
					{items.length} session{items.length === 1 ? '' : 's'} shown
				{/if}
			</p>

			{#if loading}
				<div class="rounded-2xl border border-[#464040] bg-[#3A3535] px-6 py-14 text-center">
					<div class={`mx-auto ${centeredSpinnerClass}`}></div>

					<p class="mt-4 text-sm font-bold text-[#f4f4f4]">Loading sessions…</p>
				</div>
			{:else if items.length === 0}
				<div class="rounded-2xl border border-[#464040] bg-[#3A3535] px-6 py-14 text-center">
					<h2 class="text-xl font-black text-[#f4f4f4]">No sessions found</h2>

					<p class="mx-auto mt-2 max-w-sm text-sm leading-relaxed text-[#d8d8d8]">
						Nothing matches these filters yet. Try clearing them or start a new
						practice session.
					</p>

					<a href="/" class={`${primaryButtonClass} mt-6 inline-flex px-8`}>
						Start practising →
					</a>
				</div>
			{:else}
				<div class="flex w-full flex-col gap-4">
					{#each items as item (item.id)}
						<div
							class="flex w-full flex-col gap-4 rounded-2xl border border-[#464040] bg-[#232020] p-5 sm:flex-row sm:items-center sm:justify-between sm:gap-6 sm:px-6"
						>
							<a href="/session/{item.id}" class="flex min-w-0 flex-1 flex-col gap-2">
								<div class="flex min-w-0 items-center gap-2">
									<h2 class="min-w-0 flex-1 truncate text-base font-black text-[#f4f4f4]">
										{item.topicTitle}
									</h2>

									{#if item.status !== 'completed'}
										<span
											class="shrink-0 rounded-full border px-2.5 py-1 text-[10px] font-black uppercase tracking-wide {statusClasses(
												item.status
											)}"
										>
											{statusLabel(item.status)}
										</span>
									{/if}
								</div>

								<div class="flex flex-wrap items-center gap-1.5 text-xs font-semibold text-[#f4f4f4]/45">
									<span>{formatDate(item.createdAt)}</span>

									{#if item.wordsPerMinute !== null && item.wordsPerMinute !== undefined}
										<span>·</span>
										<span>{Math.round(item.wordsPerMinute)} wpm</span>
									{/if}
								</div>
							</a>

							<div class="flex shrink-0 items-center justify-between gap-3 sm:justify-end">
								{#if item.status === 'completed'}
									<div class="flex min-w-14 items-center justify-center rounded-xl border border-[#464040] bg-[#3A3535] px-3 py-2">
										<span class="text-base font-black tabular-nums text-[#f4f4f4]">
											{formatScore(item.overallScore)}
										</span>
									</div>
								{:else}
									<span></span>
								{/if}

								<button
									type="button"
									onclick={() => requestDelete(item.id)}
									aria-label="Delete session"
									class="flex h-9 w-9 shrink-0 cursor-pointer items-center justify-center rounded-xl text-[#f4f4f4]/50 transition-colors duration-200 hover:text-[#FF6B7A]"
								>
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
									>
										<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6" />
										<path d="M3 6h18" />
										<path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
									</svg>
								</button>
							</div>
						</div>
					{/each}
				</div>

				{#if hasMore}
					<button
						type="button"
						onclick={loadMore}
						disabled={loadingMore}
						class={`${secondaryButtonClass} mx-auto mt-2 px-8`}
					>
						{loadingMore ? 'Loading…' : 'Load more'}
					</button>
				{/if}
			{/if}
		</div>
	</div>
</div>

{#if deleteModalId && deleteModalItem}
	<div class="fixed inset-0 z-50 flex items-center justify-center px-5">
		<button
			type="button"
			tabindex="-1"
			aria-label="Close"
			onclick={cancelDelete}
			class="absolute inset-0 h-full w-full cursor-default appearance-none border-0 bg-black/60 p-0"
		></button>

		<div class="relative w-full max-w-sm rounded-2xl border border-[#464040] bg-[#232020] p-6">
			<h2 class="text-xl font-black text-[#f4f4f4]">Delete session?</h2>

			<p class="mt-2 text-sm leading-relaxed text-[#d8d8d8]">
				This will permanently remove "{deleteModalItem.topicTitle}" and its results. This
				can't be undone.
			</p>

			<div class="mt-6 flex gap-3">
				<button type="button" onclick={cancelDelete} class={`${secondaryButtonClass} flex-1`}>
					Cancel
				</button>

				<button
					type="button"
					onclick={confirmDelete}
					disabled={deletingId === deleteModalId}
					class="flex h-13.5 flex-1 cursor-pointer items-center justify-center rounded-xl border border-[#FF6B7A]/50 bg-[#FF6B7A]/10 font-bold text-[#FF6B7A] transition-colors duration-200 hover:bg-[#FF6B7A]/20 disabled:cursor-not-allowed disabled:opacity-50"
				>
					{deletingId === deleteModalId ? 'Deleting…' : 'Delete'}
				</button>
			</div>
		</div>
	</div>
{/if}