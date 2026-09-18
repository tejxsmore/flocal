<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { consumeEmailChange, ApiRequestError } from '$lib/api';

	let status = $state<'verifying' | 'done' | 'error'>('verifying');
	let errorMessage = $state('');

	onMount(async () => {
		const email = page.url.searchParams.get('email');
		const token = page.url.searchParams.get('token');

		if (!email || !token) {
			status = 'error';
			errorMessage = 'This verification link is missing information.';
			return;
		}

		try {
			await consumeEmailChange(email, token);
			status = 'done';

			setTimeout(() => goto('/profile/settings'), 1500);
		} catch (err) {
			status = 'error';

			errorMessage =
				err instanceof ApiRequestError
					? err.message
					: 'Could not verify this email change.';
		}
	});

	const primaryButtonClass =
		'flex h-13.5 cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#c64c13] bg-[#F25912] px-6 font-bold text-[#181C14] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:translate-y-0';

	const secondaryButtonClass =
		'flex h-11 cursor-pointer items-center justify-center gap-2 rounded-xl border border-[#412B6B] bg-[#38255b] px-4 text-xs font-bold text-[#f4f0e4] transition-transform duration-200 hover:-translate-y-0.5 sm:text-sm';
</script>

<svelte:head>
	<title>Flocal — Verify email</title>
</svelte:head>

<div class="min-h-screen bg-[#1b1833] px-5 py-7 text-[#f4f0e4]">
	<div class="mx-auto flex min-h-[calc(100vh-3.5rem)] w-full max-w-3xl flex-col">
		<!-- TOP BAR -->
		<header class="flex items-center justify-between">
			<a
				href="/"
				class="text-2xl font-black tracking-tight transition-transform duration-200 hover:-translate-y-0.5"
			>
				Flocal
			</a>

			{#if status === 'error'}
				<a href="/profile/settings" class={secondaryButtonClass}>Settings</a>
			{/if}
		</header>

		<!-- CENTER -->
		<main class="flex flex-1 items-center justify-center py-16">
			<div class="w-full max-w-xl">
				{#if status === 'verifying'}
					<!-- VERIFYING -->
					<section class="relative overflow-hidden rounded-3xl border border-[#412B6B] bg-[#171022] p-8 text-center sm:p-12">
						<div
							class="pointer-events-none absolute inset-x-0 top-1/2 h-40 -translate-y-1/2 bg-[#38255b] opacity-40"
						></div>

						<div class="relative">
							<div class="mx-auto flex h-20 w-20 items-center justify-center rounded-3xl border border-[#412B6B] bg-[#38255b] text-3xl font-black">
								<span class="animate-pulse">✦</span>
							</div>

							<p class="mt-8 text-xs font-black uppercase tracking-[0.16em] text-[#f4f0e4]/45">
								Email verification
							</p>

							<h1 class="mt-2 text-4xl font-black tracking-tighter sm:text-5xl">
								Checking your email.
							</h1>

							<p class="mx-auto mt-4 max-w-md text-sm leading-relaxed text-[#f4f0e4]/60 sm:text-base">
								We're verifying your new email address. This should only take a moment.
							</p>

							<div class="mx-auto mt-8 h-2 w-36 overflow-hidden rounded-full border border-[#412B6B] bg-[#26224d]">
								<div class="h-full w-1/2 animate-[loading_1.2s_ease-in-out_infinite] rounded-full bg-[#F25912]"></div>
							</div>
						</div>
					</section>
				{:else if status === 'done'}
					<!-- SUCCESS -->
					<section class="relative overflow-hidden rounded-3xl border border-[#412B6B] bg-[#171022] p-8 text-center sm:p-12">
						<div
							class="pointer-events-none absolute inset-x-0 top-1/2 h-40 -translate-y-1/2 bg-[#38255b] opacity-40"
						></div>

						<div class="relative">
							<div class="mx-auto flex h-20 w-20 items-center justify-center rounded-3xl border border-[#5FCB7A]/50 bg-[#171022] text-3xl font-black text-[#5FCB7A]">
								✓
							</div>

							<p class="mt-8 text-xs font-black uppercase tracking-[0.16em] text-[#f4f0e4]/45">
								All set
							</p>

							<h1 class="mt-2 text-4xl font-black tracking-tighter sm:text-5xl">
								Email updated.
							</h1>

							<p class="mx-auto mt-4 max-w-md text-sm leading-relaxed text-[#f4f0e4]/60 sm:text-base">
								Your new email address has been verified successfully.
								We're taking you back to settings.
							</p>

							<div class="mt-8 inline-flex items-center gap-2 rounded-xl border border-[#5FCB7A]/40 bg-[#171022] px-4 py-2 text-xs font-black text-[#5FCB7A]">
								<span class="h-2 w-2 rounded-full bg-[#5FCB7A]"></span>
								Redirecting…
							</div>
						</div>
					</section>
				{:else}
					<!-- ERROR -->
					<section class="relative overflow-hidden rounded-3xl border border-[#FF6B7A]/40 bg-[#2a1620] p-8 sm:p-12">
						<div
							class="pointer-events-none absolute inset-x-0 top-1/2 h-40 -translate-y-1/2 bg-[#FF6B7A]/10"
						></div>

						<div class="relative text-center">
							<div class="mx-auto flex h-20 w-20 items-center justify-center rounded-3xl border border-[#FF6B7A]/50 bg-[#171022] text-3xl font-black text-[#FF6B7A]">
								!
							</div>

							<p class="mt-8 text-xs font-black uppercase tracking-[0.16em] text-[#FF6B7A]/70">
								Verification failed
							</p>

							<h1 class="mt-2 text-4xl font-black tracking-tighter sm:text-5xl">
								That link didn't work.
							</h1>

							<p class="mx-auto mt-4 max-w-md text-sm leading-relaxed text-[#f4f0e4]/60 sm:text-base">
								{errorMessage}
							</p>

							<div class="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
								<a href="/profile/settings" class={primaryButtonClass}>
									Back to settings
								</a>

								<a
									href="/"
									class="flex h-13.5 cursor-pointer items-center justify-center rounded-xl border border-[#412B6B] bg-transparent px-6 text-sm font-bold text-[#f4f0e4] transition-transform duration-200 hover:-translate-y-1"
								>
									Go home
								</a>
							</div>
						</div>
					</section>
				{/if}
			</div>
		</main>

		<!-- FOOTER -->
		<footer class="pb-4 text-center">
			<p class="text-[11px] font-bold text-[#f4f0e4]/30">
				Flocal · Speak better, one session at a time.
			</p>
		</footer>
	</div>
</div>

<style>
	@keyframes loading {
		0% {
			transform: translateX(-100%);
		}

		50% {
			transform: translateX(100%);
		}

		100% {
			transform: translateX(220%);
		}
	}
</style>