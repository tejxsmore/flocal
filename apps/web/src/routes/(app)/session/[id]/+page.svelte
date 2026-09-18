<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import posthog from 'posthog-js';
	import { getSession, getSessionReport, sessionStreamUrl } from '$lib/api';
	import type { SpeakingSession, SessionReport, StreamMessage, Topic } from '$lib/types';

	type Phase = 'loading' | 'prep' | 'recording' | 'processing' | 'done' | 'error';

	function formatTime(totalSeconds: number): string {
		const clamped = Math.max(0, totalSeconds);
		const minutes = Math.floor(clamped / 60);
		const seconds = clamped % 60;

		return `${minutes}:${seconds.toString().padStart(2, '0')}`;
	}

	function formatScore(value: number | null | undefined): string {
		if (value === null || value === undefined) return '—';

		return Math.round(value).toString();
	}

	let sessionId = $derived(page.params.id!);
	let phase = $state<Phase>('loading');
	let session = $state<SpeakingSession | null>(null);
	let topic = $state<Topic | null>(null);
	let errorMessage = $state('');

	let prepRemaining = $state(0);
	let speakRemaining = $state(0);
	let finalTranscript = $state('');
	let interimTranscript = $state('');
	let fillerCount = $state(0);
	let extraPrepUsed = $state(false);
	let practiceMode = $state(false);

	let report = $state<SessionReport | null>(null);
	let reportError = $state('');

	let audioEl: HTMLAudioElement | undefined = $state();
	let isPlaying = $state(false);

	let scoreCards = $derived(
		report
			? [
					{ label: 'Clarity', value: report.clarityScore },
					{ label: 'Delivery', value: report.deliveryScore },
					{ label: 'Content', value: report.contentScore },
					{ label: 'Vocabulary', value: report.vocabularyScore },
					{ label: 'Grammar', value: report.grammarScore }
				]
			: []
	);

	let statCards = $derived(
		report
			? [
					{
						label: 'Words / min',
						value:
							report.wordsPerMinute !== null && report.wordsPerMinute !== undefined
								? Math.round(report.wordsPerMinute).toString()
								: null
					},
					{
						label: 'Filler words',
						value:
							report.fillerWordCount !== null && report.fillerWordCount !== undefined
								? report.fillerWordCount.toString()
								: null
					},
					{
						label: 'Longest pause',
						value:
							report.longestPauseSeconds !== null &&
							report.longestPauseSeconds !== undefined
								? `${report.longestPauseSeconds.toFixed(1)}s`
								: null
					}
				].filter((stat) => stat.value !== null)
			: []
	);

	let analysisLocked = $derived(
		report !== null && report.status === 'completed' && report.overallScore === null
	);

	let isCenteredScreen = $derived(
		phase === 'loading' || phase === 'error' || (phase === 'done' && !report)
	);

	let outerContainerClass = $derived(
		`flex min-h-screen w-full justify-center p-6 sm:p-8 ${
			isCenteredScreen ? ' items-center' : ''
		}`
	);

	let outerWidthClass = $derived(
		phase === 'done' && report ? 'w-full 2xl:max-w-7xl' : 'w-full max-w-lg'
	);

	let ws: WebSocket | null = null;
	let mediaRecorder: MediaRecorder | null = null;
	let mediaStream: MediaStream | null = null;
	let prepTimer: ReturnType<typeof setInterval> | null = null;
	let speakTimer: ReturnType<typeof setInterval> | null = null;
	let pollTimer: ReturnType<typeof setInterval> | null = null;

	function clearTimers() {
		if (prepTimer) clearInterval(prepTimer);
		if (speakTimer) clearInterval(speakTimer);
		if (pollTimer) clearInterval(pollTimer);

		prepTimer = null;
		speakTimer = null;
		pollTimer = null;
	}

	function stopMedia() {
		if (mediaRecorder && mediaRecorder.state !== 'inactive') {
			mediaRecorder.stop();
		}

		mediaRecorder = null;

		if (mediaStream) {
			for (const track of mediaStream.getTracks()) {
				track.stop();
			}
		}

		mediaStream = null;
	}

	function fail(message: string) {
		clearTimers();
		stopMedia();

		if (ws) ws.close();

		errorMessage = message;
		phase = 'error';
	}

	async function startPrep() {
		if (!session) return;

		if (session.prepTimeSeconds === 0) {
			await startRecording();
			return;
		}

		prepRemaining = session.prepTimeSeconds;
		extraPrepUsed = false;
		phase = 'prep';

		prepTimer = setInterval(() => {
			prepRemaining -= 1;

			if (prepRemaining <= 0) {
				if (prepTimer) clearInterval(prepTimer);

				startRecording();
			}
		}, 1000);
	}

	function addExtraPrep() {
		if (extraPrepUsed || phase !== 'prep') return;

		extraPrepUsed = true;
		prepRemaining += 300;
	}

	function skipPrep() {
		if (prepTimer) clearInterval(prepTimer);

		prepTimer = null;

		startRecording();
	}

	async function startRecording() {
		if (!session) return;

		try {
			mediaStream = await navigator.mediaDevices.getUserMedia({
				audio: true
			});
		} catch {
			fail('Microphone access is required to record.');
			return;
		}

		ws = new WebSocket(sessionStreamUrl(session.id));
		ws.binaryType = 'arraybuffer';

		ws.onopen = () => {
			phase = 'recording';
			speakRemaining = session!.speakTimeSeconds;

			mediaRecorder = new MediaRecorder(mediaStream!, {
				mimeType: 'audio/webm;codecs=opus'
			});

			mediaRecorder.ondataavailable = (event) => {
				if (
					event.data.size > 0 &&
					ws &&
					ws.readyState === WebSocket.OPEN
				) {
					ws.send(event.data);
				}
			};

			mediaRecorder.start(250);

			speakTimer = setInterval(() => {
				speakRemaining -= 1;

				if (speakRemaining <= 0) {
					if (speakTimer) clearInterval(speakTimer);

					finishRecording();
				}
			}, 1000);
		};

		ws.onmessage = (event) => {
			const msg: StreamMessage = JSON.parse(event.data);

			if (msg.type === 'interim') {
				interimTranscript = msg.transcript ?? '';
			} else if (msg.type === 'final') {
				finalTranscript =
					`${finalTranscript} ${msg.transcript ?? ''}`.trim();

				interimTranscript = '';
				fillerCount += msg.fillerCount ?? 0;
			} else if (
				msg.type === 'info' &&
				msg.message === 'practice_mode'
			) {
				practiceMode = true;
			}
		};

		ws.onclose = () => {
			if (phase === 'recording' || phase === 'prep') {
				beginProcessing();
			}
		};

		ws.onerror = () => {
			fail('Lost connection while streaming audio.');
		};
	}

	function finishRecording() {
		stopMedia();

		if (ws && ws.readyState === WebSocket.OPEN) {
			ws.close();
		}

		beginProcessing();
	}

	function beginProcessing() {
		if (phase === 'processing' || phase === 'done') return;

		phase = 'processing';

		pollTimer = setInterval(async () => {
			if (!session) return;

			try {
				const updated = await getSession(session.id);

				session = updated;

				if (updated.status === 'completed') {
					if (pollTimer) clearInterval(pollTimer);

					await loadReport();

					phase = 'done';
				} else if (updated.status === 'failed') {
					if (pollTimer) clearInterval(pollTimer);

					fail(
						updated.failureReason ??
							'Session processing failed.'
					);
				}
			} catch {}
		}, 2000);
	}

	async function loadReport() {
		if (!session) return;

		reportError = '';

		try {
			report = await getSessionReport(session.id);

			posthog.capture('session_completed', {
				topic_format: report.topicFormat,
				topic_category: report.topicCategory,
				overall_score: report.overallScore ?? null,
				words_per_minute: report.wordsPerMinute ?? null,
				filler_word_count: report.fillerWordCount ?? null,
				prep_time_seconds: report.prepTimeSeconds,
				speak_time_seconds: report.speakTimeSeconds
			});
		} catch {
			reportError = 'Could not load your results.';
		}
	}

	function toggleAudioPlayback() {
		if (!audioEl) return;

		if (isPlaying) {
			audioEl.pause();
		} else {
			audioEl.play();
		}
	}

	onMount(async () => {
		try {
			session = await getSession(sessionId);

			const stored = sessionStorage.getItem(
				`flocal:topic:${sessionId}`
			);

			if (stored) {
				topic = JSON.parse(stored);
			}

			if (session.status === 'completed') {
				await loadReport();

				phase = 'done';
			} else if (session.status === 'failed') {
				fail(
					session.failureReason ??
						'Session processing failed.'
				);
			} else {
				await startPrep();
			}
		} catch {
			fail('Could not load this session.');
		}
	});

	onDestroy(() => {
		clearTimers();
		stopMedia();

		if (ws) ws.close();
	});

	const primaryButtonClass =
		'flex h-13.5 w-full cursor-pointer items-center justify-center gap-4 rounded-xl border border-[#ff6803] bg-[#FF7315] px-5 font-bold text-[#232020] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:translate-y-0';

	const secondaryButtonClass =
		'flex h-13.5 w-full cursor-pointer items-center justify-center gap-4 rounded-xl border border-[#464040] bg-[#3A3535] px-6 font-bold text-[#f4f4f4] transition-transform duration-200 hover:-translate-y-1';

	const audioButtonClass =
		'flex h-12 w-12 shrink-0 cursor-pointer items-center justify-center rounded-full border border-[#ff6803] bg-[#FF7315] text-[#232020] transition-transform duration-200 hover:-translate-y-0.5';

	const statCardClass =
		'rounded-xl border border-[#464040] bg-[#232020] px-3 py-4 text-center';

	const eyebrowClass =
		'text-xs font-black uppercase tracking-[0.14em] text-[#f4f4f4]/60';

	const columnLabelClass =
		'text-[10px] font-black uppercase tracking-[0.14em] text-[#f4f4f4]/35';

	// Card section padding is a single source of truth so the divider's
	// negative margins always cancel it out exactly, on every breakpoint.
	const cardSectionClass = 'p-6 sm:p-7';

	const sectionDividerClass =
		'my-6 -mx-6 border-0 border-t border-[#464040] sm:-mx-7';

	const centeredSpinnerClass =
		'h-6 w-6 animate-spin rounded-full border border-[#464040] border-t-[#FF7315]';

	const playIconOffsetClass = '-ml-1';
</script>

{#snippet processingScreen(title: string, description: string)}
	<div class="flex flex-col items-center gap-4 text-center">
		<div class={centeredSpinnerClass}></div>

		<h1 class="text-2xl font-black leading-tight tracking-tight">
			{title}
		</h1>

		<p class="max-w-xs text-sm leading-relaxed text-[#d8d8d8]">
			{description}
		</p>
	</div>
{/snippet}

<svelte:head>
	<title>Flocal — Session</title>
</svelte:head>

<div class={outerContainerClass}>
	<div class={outerWidthClass}>

		{#if phase === 'loading'}
			<div class="flex flex-col items-center gap-4 text-center">
				<div class={centeredSpinnerClass}></div>

				<p class="text-base leading-relaxed text-[#d8d8d8]">
					Getting everything ready.
				</p>
			</div>

		{:else if phase === 'error'}
			<div class="flex flex-col items-center text-center">
				<h1 class="text-3xl font-black leading-[1.15] tracking-[-0.02em] sm:text-4xl">
					Something went wrong
				</h1>

				<p class="mt-3 text-base leading-relaxed text-[#d8d8d8]">
					{errorMessage}
				</p>

				<a href="/" class={`${primaryButtonClass} mt-8`}>
					Back to home
				</a>
			</div>

		{:else}

			{#if topic && phase !== 'done'}
				<div class="flex flex-col items-center gap-4 text-center">
					<span class="inline-flex rounded-full border border-[#464040] bg-[#3A3535] px-3.5 py-1.5 text-xs font-bold text-[#f4f4f4]">
						{topic.difficulty}
					</span>

					<h1 class="mx-auto max-w-md text-2xl font-black leading-tight tracking-tight sm:text-3xl">
						{topic.title}
					</h1>
				</div>
			{/if}

			{#if phase === 'prep'}
				<div class="mt-8 flex flex-col items-center gap-6 border-y border-[#464040] py-10 text-center">
					<p class="text-xs font-black uppercase tracking-[0.14em] text-[#f4f4f4]/60">
						Prep time
					</p>

					<div class="text-7xl font-black tabular-nums tracking-[-0.06em] sm:text-8xl">
						{formatTime(prepRemaining)}
					</div>

					<p class="max-w-xs text-sm leading-relaxed text-[#d8d8d8]">
						Think about your answer. Structure your thoughts, then say what you mean.
					</p>
				</div>

				<div class="mt-8 flex flex-col gap-4 sm:flex-row">
					<button
						type="button"
						onclick={addExtraPrep}
						disabled={extraPrepUsed}
						class={secondaryButtonClass}
					>
						{extraPrepUsed ? '+5 min used' : '+5 min'}
					</button>

					<button
						type="button"
						onclick={skipPrep}
						class={primaryButtonClass}
					>
						I'm ready, start now →
					</button>
				</div>

			{:else if phase === 'recording'}
				<div class="mt-8 flex flex-col items-center gap-6">
					<div class="flex items-center gap-2 rounded-full border border-[#464040] bg-[#3A3535] px-4 py-2">
						<span class="h-2 w-2 animate-pulse rounded-full bg-[#FF7315]"></span>

						<span class="text-xs font-black uppercase tracking-[0.12em] text-[#f4f4f4]">
							Recording
						</span>
					</div>

					<div class="w-full border-y border-[#464040] py-10 text-center">
						<div class="text-7xl font-black tabular-nums tracking-[-0.06em] sm:text-8xl">
							{formatTime(speakRemaining)}
						</div>

						<p class="mt-3 text-xs font-black uppercase tracking-[0.15em] text-[#f4f4f4]/50">
							Keep going
						</p>
					</div>

					{#if practiceMode}
						<div class="flex w-full flex-col items-center gap-4 text-center">
							<span class="inline-flex rounded-full border border-[#464040] bg-[#3A3535] px-3.5 py-1.5 text-xs font-bold text-[#f4f4f4]">
								∞ Practice mode
							</span>

							<p class="text-xs leading-relaxed text-[#d8d8d8]">
								No live transcript this session.
							</p>

							<p class="text-xs font-semibold text-[#f4f4f4]/50">
								Upgrade for unlimited AI transcript &amp; scoring.
							</p>
						</div>
					{:else}
						<div class="w-full text-left">
							<div class="mb-3 flex items-center justify-between">
								<span class="text-[10px] font-black uppercase tracking-[0.15em] text-[#f4f4f4]/50">
									Live transcript
								</span>

								{#if fillerCount > 0}
									<span class="rounded-full border border-[#464040] bg-[#3A3535] px-2.5 py-1 text-[10px] font-black text-[#f4f4f4]">
										{fillerCount} filler{fillerCount === 1 ? '' : 's'}
									</span>
								{/if}
							</div>

							<p class="min-h-24 text-sm leading-7 text-[#f4f4f4]">
								{finalTranscript}
								<span class="text-[#f4f4f4]/40">
									{interimTranscript}
								</span>
							</p>
						</div>
					{/if}
				</div>

			{:else if phase === 'processing'}
				{@render processingScreen(
					practiceMode ? 'Saving your session…' : 'Analyzing your response…',
					practiceMode
						? 'Your practice session is being saved.'
						: 'We’re looking at your delivery, clarity, content, and language.'
				)}

			{:else if phase === 'done'}

				{#if reportError}
					<div class="flex flex-col items-center gap-4 text-center">
						<p class="text-sm font-semibold leading-relaxed text-[#FF6B7A]">
							{reportError}
						</p>
					</div>

				{:else if report}
					<div class="flex flex-col gap-6 sm:gap-8">
						<div class="grid gap-6 sm:gap-8 lg:grid-cols-3 lg:items-start">

							<div class="flex flex-col gap-6 sm:gap-8 lg:top-8 lg:sticky lg:col-span-1">

								{#if analysisLocked}
									<div class="relative overflow-hidden rounded-2xl">
										<div class="pointer-events-none select-none blur-sm">
											<div class="rounded-2xl border border-[#464040] bg-[#3A3535] px-6 py-10 text-center">
												<p class="text-xs font-black uppercase tracking-[0.15em] text-[#f4f4f4]/60">
													Overall score
												</p>

												<div class="mt-2 text-7xl font-black tabular-nums tracking-[-0.06em] text-[#FF7315]">
													82
												</div>
											</div>

											<div class="mt-8 grid grid-cols-3 gap-4 sm:grid-cols-5 lg:grid-cols-3">
												{#each ['Clarity', 'Delivery', 'Content', 'Vocabulary', 'Grammar'] as label}
													<div class={statCardClass}>
														<div class="text-lg font-black text-[#f4f4f4]">
															78
														</div>

														<div class="mt-1 text-[9px] font-black uppercase tracking-wide text-[#f4f4f4]/50">
															{label}
														</div>
													</div>
												{/each}
											</div>
										</div>

										<div class="absolute inset-0 flex flex-col items-center justify-center px-6 text-center backdrop-blur-sm">
											<span class="inline-flex rounded-full border border-[#464040] bg-[#3A3535] px-3.5 py-1.5 text-xs font-bold text-[#f4f4f4]">
												🔒 Locked
											</span>

											<h2 class="mt-5 text-xl font-black leading-tight tracking-tight">
												AI analysis locked
											</h2>

											<p class="mt-2 max-w-xs text-sm leading-relaxed text-[#d8d8d8]">
												Free members get 5 AI-analyzed sessions, then 1 per day. You've used today's.
												Practice is still unlimited, but scoring, tips, and transcript are Pro features.
											</p>

											<a
												href="/plan"
												onclick={() =>
													posthog.capture('upgrade_cta_clicked', {
														source: 'session_locked'
													})}
												class={`${primaryButtonClass} mt-6 w-auto px-8`}
											>
												Upgrade to Pro →
											</a>
										</div>
									</div>

								{:else}
									<div class="rounded-2xl border border-[#464040] bg-[#3A3535] px-6 py-10 text-center">
										<p class="text-xs font-black uppercase tracking-[0.15em] text-[#f4f4f4]/60">
											Overall score
										</p>

										<div class="mt-2 text-7xl font-black tabular-nums tracking-[-0.06em] text-[#FF7315] sm:text-8xl">
											{formatScore(report.overallScore)}
										</div>
									</div>

									<div class="grid grid-cols-3 gap-4 sm:grid-cols-5 lg:grid-cols-3">
										{#each scoreCards as card (card.label)}
											<div class={statCardClass}>
												<div class="text-xl font-black tabular-nums text-[#f4f4f4]">
													{formatScore(card.value)}
												</div>

												<div class="mt-1 text-[9px] font-black uppercase tracking-wide text-[#f4f4f4]/50">
													{card.label}
												</div>
											</div>
										{/each}
									</div>

									{#if statCards.length > 0}
										<div class="overflow-hidden rounded-2xl border border-[#464040] bg-[#232020]">
											{#each statCards as stat, index (stat.label)}
												<div
													class="flex items-center justify-between gap-6 px-6 py-4"
													class:border-b={index !== statCards.length - 1}
													class:border-[#464040]={index !== statCards.length - 1}
												>
													<p class="text-[11px] font-black uppercase tracking-[0.14em] text-[#f4f4f4]/45">
														{stat.label}
													</p>

													<span class="shrink-0 text-xl font-black tabular-nums tracking-tight text-[#f4f4f4]">
														{stat.value}
													</span>
												</div>
											{/each}
										</div>
									{/if}
								{/if}

								<a href="/" class={primaryButtonClass}>
									Spin another topic →
								</a>
							</div>

							<div class="flex flex-col gap-6 sm:gap-8 lg:col-span-2">

								{#if report.audioPlaybackUrl}
									<div class="flex items-center gap-4 rounded-2xl border border-[#464040] bg-[#3A3535] p-5">
										<button
											type="button"
											onclick={toggleAudioPlayback}
											aria-label={isPlaying ? 'Pause recording' : 'Play recording'}
											class={audioButtonClass}
										>
											{#if isPlaying}
												<svg
													xmlns="http://www.w3.org/2000/svg"
													width="24"
													height="24"
													viewBox="0 0 24 24"
													fill="currentColor"
													aria-hidden="true"
												>
													<rect x="5" y="4" width="5" height="16" rx="1" />
													<rect x="14" y="4" width="5" height="16" rx="1" />
												</svg>
											{:else}
												<svg
													xmlns="http://www.w3.org/2000/svg"
													width="24"
													height="24"
													viewBox="0 0 24 24"
													fill="currentColor"
													class={playIconOffsetClass}
													aria-hidden="true"
												>
													<path d="M8 5.5v13a1.5 1.5 0 0 0 2.3 1.27l9.5-6.5a1.5 1.5 0 0 0 0-2.54l-9.5-6.5A1.5 1.5 0 0 0 8 5.5Z" />
												</svg>
											{/if}
										</button>

										<div>
											<p class="text-sm font-bold text-[#f4f4f4]">
												Your recording
											</p>

											<p class="mt-0.5 text-xs font-medium text-[#f4f4f4]/45">
												Listen back to your response
											</p>
										</div>

										<audio
											bind:this={audioEl}
											src={report.audioPlaybackUrl}
											onplay={() => (isPlaying = true)}
											onpause={() => (isPlaying = false)}
											onended={() => (isPlaying = false)}
											class="hidden"
										></audio>
									</div>
								{/if}

								{#if !analysisLocked}

									{#if report.transcript}
										<div class="rounded-2xl border border-[#464040] bg-[#232020] {cardSectionClass}">
											<p class={eyebrowClass}>
												Transcript
											</p>

											<p class="mt-3 text-sm leading-7 text-[#f4f4f4]/80">
												{report.transcript}
											</p>
										</div>
									{/if}

									{#if report.coachMessage || report.strengths?.length || report.improvements?.length}
										<div class="overflow-hidden rounded-2xl border border-[#464040] bg-[#232020]">

											{#if report.coachMessage}
												<section class={cardSectionClass}>
													<div class="flex items-center gap-4">
														<span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-[#3A3535] text-[#FF7315]">
															<svg
																xmlns="http://www.w3.org/2000/svg"
																width="18"
																height="18"
																viewBox="0 0 24 24"
																fill="currentColor"
																aria-hidden="true"
															>
																<path d="M11.017 2.814a1 1 0 0 1 1.966 0l1.051 5.558a2 2 0 0 0 1.594 1.594l5.558 1.051a1 1 0 0 1 0 1.966l-5.558 1.051a2 2 0 0 0-1.594 1.594l-1.051 5.558a1 1 0 0 1-1.966 0l-1.051-5.558a2 2 0 0 0-1.594-1.594l-5.558-1.051a1 1 0 0 1 0-1.966l5.558-1.051a2 2 0 0 0 1.594-1.594z" />
															</svg>
														</span>

														<div>
															<h2 class="text-base font-black text-[#f4f4f4]">
																Coach's note
															</h2>

															<p class="mt-0.5 text-xs font-medium text-[#f4f4f4]/40">
																A quick take on your overall response
															</p>
														</div>
													</div>

													<p class="mt-5 text-sm font-semibold leading-7 text-[#f4f4f4]/80">
														{report.coachMessage}
													</p>
												</section>
											{/if}

											{#if report.coachMessage && report.strengths?.length}
												<hr class="border-0 border-t border-[#464040]" />
											{/if}

											{#if report.strengths?.length}
												<section class={cardSectionClass}>
													<div class="flex items-center gap-4">
														<span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-[#3A3535] text-[#FF7315]">
															<svg
																xmlns="http://www.w3.org/2000/svg"
																width="18"
																height="18"
																viewBox="0 0 24 24"
																fill="none"
																stroke="currentColor"
																stroke-width="3"
																stroke-linecap="round"
																stroke-linejoin="round"
																aria-hidden="true"
															>
																<path d="M20 6 9 17l-5-5" />
															</svg>
														</span>

														<div>
															<h2 class="text-base font-black text-[#f4f4f4]">
																What went well
															</h2>

															<p class="mt-0.5 text-xs font-medium text-[#f4f4f4]/40">
																Strengths to keep building on
															</p>
														</div>
													</div>

													<div class="mt-5 flex flex-col gap-4">
														{#each report.strengths as strength}
															<p class="text-sm font-medium leading-7 text-[#f4f4f4]/75">
																{strength}
															</p>
														{/each}
													</div>
												</section>
											{/if}

											{#if report.strengths?.length && report.improvements?.length}
												<hr class="border-0 border-t border-[#464040]" />
											{/if}

											{#if report.improvements?.length}
												<section class={cardSectionClass}>
													<div class="flex items-center gap-4">
														<span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-[#3A3535] text-[#FF7315]">
															<svg
																xmlns="http://www.w3.org/2000/svg"
																width="18"
																height="18"
																viewBox="0 0 24 24"
																fill="none"
																stroke="currentColor"
																stroke-width="3"
																stroke-linecap="round"
																stroke-linejoin="round"
																aria-hidden="true"
															>
																<path d="M5 12h14" />
																<path d="m12 5 7 7-7 7" />
															</svg>
														</span>

														<div>
															<h2 class="text-base font-black text-[#f4f4f4]">
																Try next time
															</h2>

															<p class="mt-0.5 text-xs font-medium text-[#f4f4f4]/40">
																A few things to improve in your next response
															</p>
														</div>
													</div>

													<div class="mt-5 flex flex-col gap-4">
														{#each report.improvements as tip}
															<p class="text-sm font-medium leading-7 text-[#f4f4f4]/75">
																{tip}
															</p>
														{/each}
													</div>
												</section>
											{/if}
										</div>
									{/if}

									{#if report.grammarCorrections?.length}
										<section class="overflow-hidden rounded-2xl border border-[#464040] bg-[#232020] {cardSectionClass}">
											<div class="flex items-center gap-4">
												<svg
													xmlns="http://www.w3.org/2000/svg"
													width="24"
													height="24"
													viewBox="0 0 24 24"
													fill="none"
													stroke="currentColor"
													stroke-width="2"
													stroke-linecap="round"
													stroke-linejoin="round"
													class="shrink-0 text-[#f4f4f4]"
													aria-hidden="true"
												>
													<path d="M12 5v16" />
													<path d="M20.001 19A2 2 0 0022 17V5a2 2 0 00-1.999-2L16 3.002A5 5 0 0012 5a5 5 0 00-4-2H4a2 2 0 00-2 2v12a2 2 0 001.999 2H8a5 5 0 014 2 5 5 0 014-2z" />
												</svg>

												<h2 class="text-base font-black text-[#f4f4f4]">
													Grammar
												</h2>
											</div>

											<hr class={sectionDividerClass} />

											<div>
												<div class="hidden gap-5 sm:grid sm:grid-cols-[minmax(0,1fr)_28px_minmax(0,1fr)]">
													<p class={columnLabelClass}>Original</p>
													<span></span>
													<p class={columnLabelClass}>Better</p>
												</div>

												<div class="mt-3 flex flex-col gap-5">
													{#each report.grammarCorrections as correction}
														<div>
															<div class="grid gap-4 sm:grid-cols-[minmax(0,1fr)_28px_minmax(0,1fr)] sm:items-center sm:gap-5">
																<div class="sm:hidden">
																	<p class={columnLabelClass}>Original</p>
																</div>

																<span class="text-sm font-semibold text-[#f4f4f4]/45 line-through decoration-[#f4f4f4]/30">
																	{correction.originalText}
																</span>

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
																	class="hidden shrink-0 text-[#FF7315] sm:block"
																	aria-hidden="true"
																>
																	<path d="M5 12h14" />
																	<path d="m12 5 7 7-7 7" />
																</svg>

																<div class="sm:hidden">
																	<p class={columnLabelClass}>Better</p>
																</div>

																<span class="text-sm font-bold text-[#f4f4f4]">
																	{correction.correctedText}
																</span>
															</div>

															{#if correction.explanation}
																<p class="mt-4 rounded-r-xl border-l-2 border-[#FF7315] bg-[#3A3535] px-4 py-3 text-xs leading-relaxed text-[#f4f4f4]/55">
																	{correction.explanation}
																</p>
															{/if}
														</div>
													{/each}
												</div>
											</div>
										</section>
									{/if}

									{#if report.vocabularySuggestions?.length}
										<section class="overflow-hidden rounded-2xl border border-[#464040] bg-[#232020] {cardSectionClass}">
											<div class="flex items-center gap-4">
												<svg
													xmlns="http://www.w3.org/2000/svg"
													width="24"
													height="24"
													viewBox="0 0 24 24"
													fill="none"
													stroke="currentColor"
													stroke-width="2"
													stroke-linecap="round"
													stroke-linejoin="round"
													class="shrink-0 text-[#f4f4f4]"
													aria-hidden="true"
												>
													<rect width="8" height="18" x="3" y="3" rx="1" />
													<path d="M7 3v18" />
													<path d="M20.4 18.9c.2.5-.1 1.1-.6 1.3l-1.9.7c-.5.2-1.1-.1-1.3-.6L11.1 5.1c-.2-.5.1-1.1.6-1.3l1.9-.7c.5-.2 1.1.1 1.3.6Z" />
												</svg>

												<h2 class="text-base font-black text-[#f4f4f4]">
													Vocabulary
												</h2>
											</div>

											<hr class={sectionDividerClass} />

											<div>
												<div class="hidden gap-5 sm:grid sm:grid-cols-[minmax(0,0.8fr)_28px_minmax(0,1.6fr)]">
													<p class={columnLabelClass}>You said</p>
													<span></span>
													<p class={columnLabelClass}>Try instead</p>
												</div>

												<div class="mt-3 flex flex-col gap-5">
													{#each report.vocabularySuggestions as suggestion}
														<div class="grid gap-4 sm:grid-cols-[minmax(0,0.8fr)_28px_minmax(0,1.6fr)] sm:items-center sm:gap-5">
															<div class="sm:hidden">
																<p class={columnLabelClass}>You said</p>
															</div>

															<p class="text-sm font-semibold text-[#f4f4f4]/45 line-through decoration-[#f4f4f4]/30">
																{suggestion.originalWord}
															</p>

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
																class="hidden shrink-0 text-[#FF7315] sm:block"
																aria-hidden="true"
															>
																<path d="M5 12h14" />
																<path d="m12 5 7 7-7 7" />
															</svg>

															<div class="sm:hidden">
																<p class={columnLabelClass}>Try instead</p>
															</div>

															<div class="flex flex-wrap gap-2">
																{#each suggestion.suggestedWords as word}
																	<span class="rounded-full border border-[#464040] px-3 py-1.5 text-xs font-bold text-[#f4f4f4] transition-colors hover:border-[#FF7315] hover:text-[#FF7315]">
																		{word}
																	</span>
																{/each}
															</div>
														</div>
													{/each}
												</div>
											</div>
										</section>
									{/if}

								{/if}
							</div>
						</div>
					</div>
				{/if}
			{/if}
		{/if}
	</div>
</div>