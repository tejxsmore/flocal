<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import posthog from 'posthog-js';
	import {
		updateProfile,
		requestEmailChange,
		deleteAccount,
		listMySessions,
		revokeSession,
		ApiRequestError
	} from '$lib/api';
	import { PREP_TIME_OPTIONS } from '$lib/types';
	import type { SessionSummary } from '$lib/types';

	let initialUser = page.data.user;

	let name = $state(initialUser.name);
	let username = $state(initialUser.username ?? '');
	let image = $state(initialUser.image ?? '');
	let defaultPrepTimeSeconds = $state(initialUser.defaultPrepTimeSeconds);
	let timezone = $state(initialUser.timezone ?? '');

	let savingProfile = $state(false);
	let profileMessage = $state('');
	let profileError = $state('');

	let newEmail = $state('');
	let sendingEmailChange = $state(false);
	let emailChangeMessage = $state('');
	let emailChangeError = $state('');

	let sessions = $state<SessionSummary[]>([]);
	let sessionsLoading = $state(true);
	let revokingId = $state('');

	let deleteConfirmOpen = $state(false);
	let deleteConfirmText = $state('');
	let deletingAccount = $state(false);
	let deleteError = $state('');

	function initials(value: string): string {
		return value
			.split(' ')
			.map((part) => part[0])
			.filter(Boolean)
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}

	function formatDate(value: string): string {
		return new Date(value).toLocaleDateString(undefined, {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	async function saveProfile(event: SubmitEvent) {
		event.preventDefault();

		savingProfile = true;
		profileMessage = '';
		profileError = '';

		try {
			await updateProfile({
				name,
				username: username.trim() || null,
				image: image.trim() || null,
				defaultPrepTimeSeconds,
				timezone: timezone.trim() || null
			});

			profileMessage = 'Profile updated.';
			posthog.capture('profile_updated', {
				has_username: !!(username.trim()),
				prep_time_seconds: defaultPrepTimeSeconds
			});
		} catch (err) {
			profileError =
				err instanceof ApiRequestError
					? err.message
					: 'Could not update profile.';
		} finally {
			savingProfile = false;
		}
	}

	async function sendEmailChange(event: SubmitEvent) {
		event.preventDefault();

		sendingEmailChange = true;
		emailChangeMessage = '';
		emailChangeError = '';

		try {
			await requestEmailChange(newEmail);

			emailChangeMessage = `Check ${newEmail} for a verification link.`;
			posthog.capture('email_change_requested');
			newEmail = '';
		} catch (err) {
			emailChangeError =
				err instanceof ApiRequestError
					? err.message
					: 'Could not send verification email.';
		} finally {
			sendingEmailChange = false;
		}
	}

	async function loadSessions() {
		sessionsLoading = true;

		try {
			sessions = await listMySessions();
		} catch {
			sessions = [];
		} finally {
			sessionsLoading = false;
		}
	}

	async function handleRevoke(id: string) {
		revokingId = id;

		try {
			await revokeSession(id);
			sessions = sessions.filter((s) => s.id !== id);
		} finally {
			revokingId = '';
		}
	}

	async function confirmDelete() {
		deletingAccount = true;
		deleteError = '';

		try {
			await deleteAccount();
			posthog.capture('account_deleted');
			posthog.reset();
			await goto('/login');
		} catch (err) {
			deleteError =
				err instanceof ApiRequestError
					? err.message
					: 'Could not delete account.';

			deletingAccount = false;
		}
	}

	onMount(() => {
		loadSessions();
	});

	const inputClass =
		'w-full rounded-xl border border-[#412B6B] bg-[#26224d] px-4 py-3 text-sm text-[#f4f0e4] outline-none transition-colors placeholder:text-[#f4f0e4]/35 focus:border-[#5C3E94]';

	const labelClass = 'text-xs font-black uppercase tracking-wide text-[#f4f0e4]/50';

	const primaryButtonClass =
		'flex h-13.5 cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#c64c13] bg-[#F25912] px-6 font-bold text-[#181C14] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:translate-y-0';

	const secondaryButtonClass =
		'flex h-11 cursor-pointer items-center justify-center gap-2 rounded-xl border border-[#412B6B] bg-[#38255b] px-4 text-xs font-bold text-[#f4f0e4] transition-transform duration-200 hover:-translate-y-0.5 sm:text-sm';

	function pillClass(selected: boolean) {
		return `cursor-pointer rounded-full border px-4 py-2 text-sm font-bold transition-colors duration-200 ${
			selected
				? 'border-[#412B6B] bg-[#38255b] text-[#f4f0e4]'
				: 'border-[#412B6B] bg-transparent text-[#f4f0e4]/60 hover:text-[#f4f0e4]'
		}`;
	}
</script>

<svelte:head>
	<title>Flocal — Settings</title>
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
				<a href="/profile" class={secondaryButtonClass}>Profile</a>
				<a href="/session" class="hidden sm:flex {secondaryButtonClass}">Sessions</a>
			</div>
		</header>

		<!-- PAGE HEADER -->
		<section class="relative mt-10 overflow-hidden rounded-3xl border border-[#412B6B] bg-[#171022] p-7 sm:p-10">
			<div
				class="pointer-events-none absolute inset-x-0 top-1/2 h-40 -translate-y-1/2 bg-[#38255b] opacity-40"
			></div>

			<div class="relative max-w-2xl">
				<p class="text-xs font-black uppercase tracking-[0.14em] text-[#f4f0e4]/45">
					Account settings
				</p>

				<h1 class="mt-2 text-4xl font-black tracking-tighter sm:text-5xl">
					Make Flocal yours.
				</h1>

				<p class="mt-4 max-w-xl text-sm leading-relaxed text-[#f4f0e4]/60 sm:text-base">
					Update your profile, speaking preferences, email, and active sessions.
					Everything in one place.
				</p>
			</div>
		</section>

		<main class="mt-8 grid gap-6 pb-14 lg:grid-cols-[1.25fr_0.75fr]">
			<!-- LEFT COLUMN -->
			<div class="flex flex-col gap-6">
				<!-- PROFILE -->
				<form
					onsubmit={saveProfile}
					class="rounded-3xl border border-[#412B6B] bg-[#171022] p-6 sm:p-8"
				>
					<div class="flex items-start justify-between gap-4">
						<div>
							<p class="text-xs font-black uppercase tracking-[0.14em] text-[#f4f0e4]/40">01</p>
							<h2 class="mt-1 text-2xl font-black">Profile</h2>
							<p class="mt-1 text-sm text-[#f4f0e4]/55">Your public-facing Flocal identity.</p>
						</div>

						<div class="hidden h-12 w-12 items-center justify-center rounded-2xl border border-[#412B6B] bg-[#38255b] text-lg font-black sm:flex">
							✦
						</div>
					</div>

					<!-- AVATAR -->
					<div class="mt-7 flex flex-col gap-4 rounded-2xl border border-[#412B6B] bg-[#26224d] p-4 sm:flex-row sm:items-center">
						{#if image}
							<img
								src={image}
								alt={name}
								class="h-20 w-20 shrink-0 rounded-2xl border border-[#412B6B] object-cover"
							/>
						{:else}
							<div class="flex h-20 w-20 shrink-0 items-center justify-center rounded-2xl border border-[#412B6B] bg-[#38255b] text-2xl font-black">
								{initials(name || 'U')}
							</div>
						{/if}

						<div class="min-w-0 flex-1">
							<label for="image" class={labelClass}>Avatar URL</label>

							<input
								id="image"
								type="url"
								bind:value={image}
								placeholder="https://…"
								class={`mt-2 ${inputClass}`}
							/>

							<p class="mt-2 text-[11px] text-[#f4f0e4]/40">Leave empty to use your initials.</p>
						</div>
					</div>

					<div class="mt-5 grid gap-5 sm:grid-cols-2">
						<div class="flex flex-col gap-2">
							<label for="name" class={labelClass}>Name</label>
							<input id="name" type="text" required bind:value={name} class={inputClass} />
						</div>

						<div class="flex flex-col gap-2">
							<label for="username" class={labelClass}>Username</label>
							<input
								id="username"
								type="text"
								bind:value={username}
								placeholder="@username"
								class={inputClass}
							/>
						</div>
					</div>

					<!-- SPEAKING PREFS -->
					<div class="mt-6 rounded-2xl border border-[#412B6B] bg-[#38255b] p-5">
						<div>
							<p class="text-xs font-black uppercase tracking-[0.14em] text-[#f4f0e4]/50">
								Speaking preference
							</p>
							<h3 class="mt-1 text-lg font-black">Default prep time</h3>
							<p class="mt-1 text-xs text-[#f4f0e4]/55">How long you get to think before speaking.</p>
						</div>

						<div class="mt-4 flex flex-wrap gap-2">
							{#each PREP_TIME_OPTIONS as option (option.value)}
								<button
									type="button"
									onclick={() => (defaultPrepTimeSeconds = option.value)}
									class={pillClass(defaultPrepTimeSeconds === option.value)}
								>
									{option.label}
								</button>
							{/each}
						</div>
					</div>

					<div class="mt-5">
						<label for="timezone" class={labelClass}>Timezone</label>
						<input
							id="timezone"
							type="text"
							bind:value={timezone}
							placeholder="e.g. Asia/Kolkata"
							class={`mt-2 ${inputClass}`}
						/>
					</div>

					{#if profileMessage}
						<div class="mt-5 rounded-xl border border-[#412B6B] bg-[#26224d] px-4 py-3 text-sm font-bold text-[#f4f0e4]">
							{profileMessage}
						</div>
					{/if}

					{#if profileError}
						<div class="mt-5 rounded-xl border border-[#FF6B7A]/40 bg-[#26224d] px-4 py-3 text-sm font-bold text-[#FF6B7A]">
							{profileError}
						</div>
					{/if}

					<button type="submit" disabled={savingProfile} class={`mt-6 ${primaryButtonClass}`}>
						{savingProfile ? 'Saving…' : 'Save changes'}
					</button>
				</form>

				<!-- EMAIL -->
				<form
					onsubmit={sendEmailChange}
					class="rounded-3xl border border-[#412B6B] bg-[#171022] p-6 sm:p-8"
				>
					<div class="flex items-start justify-between gap-4">
						<div>
							<p class="text-xs font-black uppercase tracking-[0.14em] text-[#f4f0e4]/40">02</p>
							<h2 class="mt-1 text-2xl font-black">Email address</h2>
							<p class="mt-1 text-sm text-[#f4f0e4]/55">Change where Flocal sends account emails.</p>
						</div>

						<div class="hidden h-12 w-12 items-center justify-center rounded-2xl border border-[#412B6B] bg-[#38255b] text-xl sm:flex">
							@
						</div>
					</div>

					<div class="mt-6 rounded-2xl border border-[#412B6B] bg-[#26224d] px-4 py-3">
						<p class="text-[10px] font-black uppercase tracking-[0.14em] text-[#f4f0e4]/40">
							Current email
						</p>
						<p class="mt-1 truncate text-sm font-black">{initialUser.email}</p>
					</div>

					<div class="mt-5">
						<label for="newEmail" class={labelClass}>New email</label>
						<input
							id="newEmail"
							type="email"
							required
							bind:value={newEmail}
							placeholder="you@example.com"
							class={`mt-2 ${inputClass}`}
						/>
					</div>

					{#if emailChangeMessage}
						<div class="mt-5 rounded-xl border border-[#412B6B] bg-[#26224d] px-4 py-3 text-sm font-bold text-[#f4f0e4]">
							{emailChangeMessage}
						</div>
					{/if}

					{#if emailChangeError}
						<div class="mt-5 rounded-xl border border-[#FF6B7A]/40 bg-[#26224d] px-4 py-3 text-sm font-bold text-[#FF6B7A]">
							{emailChangeError}
						</div>
					{/if}

					<button type="submit" disabled={sendingEmailChange} class={`mt-6 ${primaryButtonClass}`}>
						{sendingEmailChange ? 'Sending…' : 'Send verification email'}
					</button>
				</form>
			</div>

			<!-- RIGHT COLUMN -->
			<div class="flex flex-col gap-6">
				<!-- ACTIVE SESSIONS -->
				<section class="rounded-3xl border border-[#412B6B] bg-[#171022] p-6 sm:p-7">
					<div class="flex items-start justify-between gap-4">
						<div>
							<p class="text-xs font-black uppercase tracking-[0.14em] text-[#f4f0e4]/40">03</p>
							<h2 class="mt-1 text-2xl font-black">Active sessions</h2>
							<p class="mt-1 text-sm leading-relaxed text-[#f4f0e4]/55">
								Devices currently signed into your account.
							</p>
						</div>

						<div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-[#412B6B] bg-[#38255b] font-black">
							↗
						</div>
					</div>

					<div class="mt-6">
						{#if sessionsLoading}
							<div class="rounded-2xl border border-[#412B6B] bg-[#26224d] p-5">
								<p class="text-sm font-bold text-[#f4f0e4]/40">Loading sessions…</p>
							</div>
						{:else if sessions.length === 0}
							<div class="rounded-2xl border border-[#412B6B] bg-[#26224d] p-5">
								<p class="text-sm font-bold text-[#f4f0e4]/40">No active sessions.</p>
							</div>
						{:else}
							<div class="flex flex-col gap-2">
								{#each sessions as s (s.id)}
									<div class="rounded-2xl border border-[#412B6B] bg-[#26224d] p-4">
										<div class="flex items-start justify-between gap-3">
											<div class="min-w-0 flex-1">
												<div class="flex flex-wrap items-center gap-2">
													<p class="truncate text-sm font-black">
														{s.userAgent ?? 'Unknown device'}
													</p>

													{#if s.current}
														<span class="shrink-0 rounded-full border border-[#412B6B] bg-[#38255b] px-2 py-0.5 text-[9px] font-black uppercase tracking-wide">
															This device
														</span>
													{/if}
												</div>

												<p class="mt-1 text-[11px] leading-relaxed text-[#f4f0e4]/40">
													{s.ipAddress ?? 'Unknown IP'}
													· signed in {formatDate(s.createdAt)}
												</p>
											</div>

											{#if !s.current}
												<button
													type="button"
													onclick={() => handleRevoke(s.id)}
													disabled={revokingId === s.id}
													class="cursor-pointer shrink-0 rounded-lg border border-[#412B6B] bg-[#171022] px-3 py-2 text-[10px] font-black text-[#f4f0e4] transition-colors hover:border-[#FF6B7A] hover:text-[#FF6B7A] disabled:cursor-not-allowed disabled:opacity-50"
												>
													{revokingId === s.id ? 'Revoking…' : 'Revoke'}
												</button>
											{/if}
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</section>

				<!-- QUICK ACCOUNT CARD -->
				<section class="rounded-3xl border border-[#412B6B] bg-[#38255b] p-6 sm:p-7">
					<p class="text-xs font-black uppercase tracking-[0.14em] text-[#f4f0e4]/50">Account</p>

					<h2 class="mt-2 text-2xl font-black leading-tight">
						Your account,
						<br />
						your rules.
					</h2>

					<p class="mt-3 text-sm leading-relaxed text-[#f4f0e4]/60">
						Need to check something else? Head back to your profile or jump
						straight into another speaking session.
					</p>

					<div class="mt-6 flex flex-col gap-2">
						<a
							href="/profile"
							class="rounded-xl border border-[#412B6B] bg-[#171022] px-5 py-3 text-center text-sm font-black text-[#f4f0e4] transition-transform duration-200 hover:-translate-y-0.5"
						>
							View profile
						</a>

						<a href="/" class={`text-center ${primaryButtonClass}`}>Start a session →</a>
					</div>
				</section>

				<!-- DANGER ZONE -->
				<section class="rounded-3xl border border-[#FF6B7A]/40 bg-[#2a1620] p-6 sm:p-7">
					<div class="flex items-start gap-4">
						<div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-[#FF6B7A]/50 bg-[#171022] font-black text-[#FF6B7A]">
							!
						</div>

						<div>
							<p class="text-xs font-black uppercase tracking-[0.14em] text-[#FF6B7A]/70">
								Danger zone
							</p>
							<h2 class="mt-1 text-2xl font-black text-[#f4f0e4]">Delete account</h2>
							<p class="mt-2 text-sm leading-relaxed text-[#f4f0e4]/55">
								This permanently deactivates your account and signs you out
								everywhere.
							</p>
						</div>
					</div>

					{#if !deleteConfirmOpen}
						<button
							type="button"
							onclick={() => (deleteConfirmOpen = true)}
							class="mt-6 cursor-pointer rounded-xl border border-[#FF6B7A]/50 bg-[#171022] px-5 py-3 text-sm font-black text-[#FF6B7A] transition-colors hover:bg-[#FF6B7A] hover:text-[#181C14]"
						>
							Delete account
						</button>
					{:else}
						<div class="mt-6 rounded-2xl border border-[#FF6B7A]/40 bg-[#171022] p-4">
							<p class="text-sm leading-relaxed text-[#f4f0e4]/80">
								Type <span class="font-black text-[#f4f0e4]">DELETE</span> to permanently
								delete your account.
							</p>

							<input
								type="text"
								bind:value={deleteConfirmText}
								placeholder="DELETE"
								class="mt-4 w-full rounded-xl border border-[#FF6B7A]/40 bg-[#2a1620] px-4 py-3 text-sm font-black uppercase text-[#f4f0e4] outline-none focus:border-[#FF6B7A]"
							/>

							{#if deleteError}
								<p class="mt-3 text-sm font-bold text-[#FF6B7A]">{deleteError}</p>
							{/if}

							<div class="mt-4 flex flex-col gap-2 sm:flex-row">
								<button
									type="button"
									onclick={confirmDelete}
									disabled={deleteConfirmText !== 'DELETE' || deletingAccount}
									class="cursor-pointer rounded-xl border border-[#FF6B7A]/60 bg-[#FF6B7A] px-5 py-3 text-sm font-black text-[#181C14] transition-opacity disabled:cursor-not-allowed disabled:opacity-40"
								>
									{deletingAccount ? 'Deleting…' : 'Permanently delete'}
								</button>

								<button
									type="button"
									onclick={() => {
										deleteConfirmOpen = false;
										deleteConfirmText = '';
										deleteError = '';
									}}
									class="cursor-pointer rounded-xl border border-[#412B6B] bg-[#26224d] px-5 py-3 text-sm font-black text-[#f4f0e4] transition-colors hover:border-[#5C3E94]"
								>
									Cancel
								</button>
							</div>
						</div>
					{/if}
				</section>
			</div>
		</main>
	</div>
</div>