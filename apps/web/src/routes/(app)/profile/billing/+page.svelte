<script lang="ts">
	import { onMount } from 'svelte';
	import posthog from 'posthog-js';

	import {
		getCurrentPlan,
		getMySubscription,
		listMyPayments,
		listPlans,
		createCheckout,
		cancelSubscription,
		resumeSubscription,
		changePlan,
		updatePaymentMethod,
		ApiRequestError
	} from '$lib/api';

	import type {
		Payment,
		Subscription,
		SubscriptionPlan,
		UserCurrentPlan
	} from '$lib/types';

	let loading = $state(true);
	let actionLoading = $state(false);
	let errorMessage = $state('');
	let successMessage = $state('');
	let showCancelConfirm = $state(false);

	let currentPlan = $state<UserCurrentPlan | null>(null);
	let subscription = $state<Subscription | null>(null);
	let plans = $state<SubscriptionPlan[]>([]);
	let payments = $state<Payment[]>([]);

	function formatMoney(subunits: number, currency: string): string {
		try {
			return new Intl.NumberFormat(undefined, {
				style: 'currency',
				currency
			}).format(subunits / 100);
		} catch {
			return `${(subunits / 100).toFixed(2)} ${currency}`;
		}
	}

	function formatDate(value: string | null | undefined): string {
		if (!value) return '—';

		const date = new Date(value);

		if (Number.isNaN(date.getTime())) return '—';

		return date.toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function formatDateTime(value: string | null | undefined): string {
		if (!value) return '—';

		const date = new Date(value);

		if (Number.isNaN(date.getTime())) return '—';

		return date.toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function titleCase(value: string | null | undefined): string {
		if (!value) return '';

		return value
			.replaceAll('_', ' ')
			.replace(/\b\w/g, (letter) => letter.toUpperCase());
	}

	function isCurrentPlan(plan: SubscriptionPlan): boolean {
		return plan.slug === currentPlan?.planSlug;
	}

	function isPaidPlan(plan: SubscriptionPlan): boolean {
		return plan.planType === 'pro';
	}

	function aiLimitLabel(
		canViewAnalysis: boolean,
		dailySessionLimit: number | null | undefined
	): string {
		if (canViewAnalysis) {
			return 'Unlimited AI analysis';
		}

		if (dailySessionLimit === null || dailySessionLimit === undefined) {
			return 'Unlimited AI analysis';
		}

		return `${dailySessionLimit} AI-analyzed session${
			dailySessionLimit === 1 ? '' : 's'
		} per day`;
	}

	function paymentStatusClass(status: Payment['status']): string {
		switch (status) {
			case 'captured':
				return 'border-[#5FCB7A]/50 text-[#5FCB7A]';

			case 'created':
				return 'border-[#F7E396]/50 text-[#F7E396]';

			case 'failed':
			case 'refunded':
				return 'border-[#FF6B7A]/50 text-[#FF6B7A]';

			default:
				return 'border-[#412B6B] text-[#f4f0e4]/70';
		}
	}

	function subscriptionStatusClass(
		status: Subscription['status'] | undefined
	): string {
		switch (status) {
			case 'active':
			case 'trialing':
				return 'border-[#5FCB7A]/50 text-[#5FCB7A]';

			case 'past_due':
				return 'border-[#F7E396]/50 text-[#F7E396]';

			case 'paused':
			case 'canceled':
			case 'expired':
				return 'border-[#FF6B7A]/50 text-[#FF6B7A]';

			default:
				return 'border-[#412B6B] text-[#f4f0e4]/70';
		}
	}

	function getPlanFeatures(plan: SubscriptionPlan): string[] {
		if (plan.planType === 'free') {
			return [
				'5 AI-analyzed sessions to start',
				`${plan.dailySessionLimit ?? 1} AI-analyzed session${
					(plan.dailySessionLimit ?? 1) === 1 ? '' : 's'
				} per day after that`,
				'Unlimited practice sessions'
			];
		}

		return [
			'Unlimited AI-analyzed sessions',
			'Unlimited practice sessions',
			'Full speech analysis',
			'Grammar & vocabulary insights',
			'Progress tracking',
			'Unlimited session history'
		];
	}

	function clearMessages() {
		errorMessage = '';
		successMessage = '';
	}

	async function loadBilling() {
		loading = true;
		errorMessage = '';

		const results = await Promise.allSettled([
			getCurrentPlan(),
			getMySubscription(),
			listPlans(),
			listMyPayments(20, 0)
		]);

		const [planResult, subscriptionResult, plansResult, paymentsResult] =
			results;

		if (planResult.status === 'fulfilled') {
			currentPlan = planResult.value;
		} else if (
			!(
				planResult.reason instanceof ApiRequestError &&
				planResult.reason.status === 404
			)
		) {
			errorMessage = 'Could not load your current plan.';
		}

		if (subscriptionResult.status === 'fulfilled') {
			subscription = subscriptionResult.value;
		} else if (
			!(
				subscriptionResult.reason instanceof ApiRequestError &&
				subscriptionResult.reason.status === 404
			)
		) {
			errorMessage = 'Could not load your subscription.';
		}

		if (plansResult.status === 'fulfilled') {
			plans = plansResult.value.filter((plan) => plan.isActive);
		} else {
			errorMessage = 'Could not load available plans.';
		}

		if (paymentsResult.status === 'fulfilled') {
			payments = paymentsResult.value;
		}

		loading = false;
	}

	async function startCheckout(plan: SubscriptionPlan) {
		if (actionLoading || isCurrentPlan(plan)) return;

		clearMessages();
		actionLoading = true;

		try {
			const checkout = await createCheckout(plan.slug);

			if (!checkout.checkoutUrl) {
				throw new Error('Checkout URL was not returned.');
			}

			posthog.capture('checkout_started', { plan_type: plan.planType });
			window.location.href = checkout.checkoutUrl;
		} catch (error) {
			errorMessage =
				error instanceof ApiRequestError
					? error.message
					: error instanceof Error
						? error.message
						: 'Could not start checkout. Please try again.';

			actionLoading = false;
		}
	}

	async function handleChangePlan(plan: SubscriptionPlan) {
		if (actionLoading || isCurrentPlan(plan)) return;

		clearMessages();
		actionLoading = true;

		try {
			await changePlan(plan.slug);

			successMessage = `Your plan has been changed to ${plan.name}.`;
			posthog.capture('plan_changed', { plan_type: plan.planType });

			await loadBilling();
		} catch (error) {
			errorMessage =
				error instanceof ApiRequestError
					? error.message
					: error instanceof Error
						? error.message
						: 'Could not change your plan. Please try again.';
		} finally {
			actionLoading = false;
		}
	}

	async function handlePlanAction(plan: SubscriptionPlan) {
		if (actionLoading || isCurrentPlan(plan)) return;

		const hasActiveSubscription =
			subscription &&
			(subscription.status === 'active' || subscription.status === 'trialing');

		if (!hasActiveSubscription && isPaidPlan(plan)) {
			await startCheckout(plan);
			return;
		}

		if (hasActiveSubscription) {
			await handleChangePlan(plan);
		}
	}

	async function handleCancel() {
		if (actionLoading) return;

		clearMessages();
		actionLoading = true;

		try {
			await cancelSubscription();

			showCancelConfirm = false;
			successMessage =
				'Your subscription will cancel at the end of your current billing period.';
			posthog.capture('subscription_cancelled');

			await loadBilling();
		} catch (error) {
			errorMessage =
				error instanceof ApiRequestError
					? error.message
					: error instanceof Error
						? error.message
						: 'Could not cancel your subscription. Please try again.';
		} finally {
			actionLoading = false;
		}
	}

	async function handleResume() {
		if (actionLoading) return;

		clearMessages();
		actionLoading = true;

		try {
			await resumeSubscription();

			successMessage = 'Your subscription has been resumed.';
			posthog.capture('subscription_resumed');

			await loadBilling();
		} catch (error) {
			errorMessage =
				error instanceof ApiRequestError
					? error.message
					: error instanceof Error
						? error.message
						: 'Could not resume your subscription. Please try again.';
		} finally {
			actionLoading = false;
		}
	}

	async function handlePaymentMethod() {
		if (actionLoading) return;

		clearMessages();
		actionLoading = true;

		try {
			const result = await updatePaymentMethod();

			if (!result.paymentLink) {
				throw new Error(
					'Payment method management link was not returned.'
				);
			}

			window.location.href = result.paymentLink;
		} catch (error) {
			errorMessage =
				error instanceof ApiRequestError
					? error.message
					: error instanceof Error
						? error.message
						: 'Could not open payment method management.';

			actionLoading = false;
		}
	}

	function currentPlanDetails(): SubscriptionPlan | null {
		return (
			plans.find((plan) => plan.slug === currentPlan?.planSlug) ?? null
		);
	}

	function currentPlanName(): string {
		const plan = currentPlanDetails();

		if (plan) return plan.name;

		if (!currentPlan?.planSlug) return 'Free';

		return currentPlan.planSlug
			.split('-')
			.map((part) => part.charAt(0).toUpperCase() + part.slice(1))
			.join(' ');
	}

	onMount(loadBilling);

	const primaryButtonClass =
		'flex h-13.5 cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#c64c13] bg-[#F25912] px-6 font-bold text-[#181C14] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:translate-y-0';

	const secondaryButtonClass =
		'flex h-11 cursor-pointer items-center justify-center gap-2 rounded-xl border border-[#412B6B] bg-[#38255b] px-4 text-xs font-bold text-[#f4f0e4] transition-transform duration-200 hover:-translate-y-0.5 sm:text-sm';

	const sectionLabelClass = 'text-xs font-black uppercase tracking-[0.14em] text-[#f4f0e4]/45';
</script>

<svelte:head>
	<title>Plan & Billing | Flocal</title>
	<meta
		name="description"
		content="Manage your Flocal subscription and billing."
	/>
</svelte:head>

<div class="min-h-screen bg-[#1b1833] px-5 py-7 text-[#f4f0e4] sm:px-8 sm:py-10">
	<div class="mx-auto w-full max-w-4xl">
		<!-- NAV -->
		<header class="flex items-center justify-between">
			<a
				href="/"
				class="text-2xl font-black tracking-tight transition-transform duration-200 hover:-translate-y-0.5"
			>
				Flocal
			</a>

			<a href="/profile" class={secondaryButtonClass}>
				<span class="text-base">←</span>
				Profile
			</a>
		</header>

		<!-- PAGE HEADER -->
		<section class="relative mt-10 overflow-hidden rounded-3xl border border-[#412B6B] bg-[#171022] p-7 sm:p-10">
			<div
				class="pointer-events-none absolute inset-x-0 top-1/2 h-40 -translate-y-1/2 bg-[#38255b] opacity-40"
			></div>

			<div class="relative max-w-2xl">
				<p class={sectionLabelClass}>Account billing</p>

				<h1 class="mt-2 text-4xl font-black tracking-tighter sm:text-5xl">
					Your plan.
					<br />
					Your progress.
				</h1>

				<p class="mt-4 max-w-xl text-sm leading-relaxed text-[#f4f0e4]/60 sm:text-base">
					Manage your Flocal plan, payment method and subscription.
				</p>
			</div>
		</section>

		{#if loading}
			<div class="mt-8 rounded-3xl border border-[#412B6B] bg-[#171022] p-12 text-center">
				<div
					class="mx-auto h-8 w-8 animate-spin rounded-full border border-[#f4f0e4]/25 border-t-[#f4f0e4]"
				></div>

				<p class="mt-4 font-bold text-[#f4f0e4]/70">Loading your billing…</p>
			</div>
		{:else}
			<!-- Alerts -->
			{#if errorMessage}
				<div class="mt-8 rounded-2xl border border-[#FF6B7A]/40 bg-[#2a1620] px-5 py-4">
					<p class="font-black text-[#FF6B7A]">Something went wrong</p>
					<p class="mt-1 text-sm text-[#f4f0e4]/70">{errorMessage}</p>
				</div>
			{/if}

			{#if successMessage}
				<div class="mt-8 rounded-2xl border border-[#412B6B] bg-[#26224d] px-5 py-4">
					<p class="font-black text-[#f4f0e4]">Done</p>
					<p class="mt-1 text-sm text-[#f4f0e4]/70">{successMessage}</p>
				</div>
			{/if}

			<!-- Current plan -->
			<section
				class="mt-8 rounded-3xl border border-[#412B6B] bg-[#38255b] p-6 sm:p-8"
			>
				<div class="flex flex-col gap-7 lg:flex-row lg:items-start lg:justify-between">
					<div>
						<p class={sectionLabelClass}>Current plan</p>

						<div class="mt-3 flex flex-wrap items-center gap-3">
							<h2 class="text-3xl font-black sm:text-4xl">
								{currentPlanName()}
							</h2>

							{#if currentPlan?.subscriptionStatus}
								<span
									class={`rounded-full border px-3 py-1 text-xs font-black ${subscriptionStatusClass(
										currentPlan.subscriptionStatus
									)}`}
								>
									{titleCase(currentPlan.subscriptionStatus)}
								</span>
							{/if}
						</div>

						{#if currentPlanDetails()}
							<p class="mt-3 text-sm text-[#f4f0e4]/60">
								{aiLimitLabel(
									currentPlanDetails()?.canViewAnalysis ?? false,
									currentPlanDetails()?.dailySessionLimit
								)}
							</p>
						{:else}
							<p class="mt-3 text-sm text-[#f4f0e4]/60">
								You're on the Free plan.
							</p>
						{/if}
					</div>

					{#if subscription && subscription.status !== 'canceled' && subscription.status !== 'expired'}
						<div class="flex flex-wrap gap-2">
							<button
								type="button"
								onclick={handlePaymentMethod}
								disabled={actionLoading}
								class="rounded-xl border border-[#412B6B] bg-[#171022] px-4 py-2.5 text-sm font-black text-[#f4f0e4] transition hover:-translate-y-0.5 disabled:cursor-not-allowed disabled:opacity-50"
							>
								Manage payment
							</button>

							{#if subscription.cancelAtPeriodEnd}
								<button
									type="button"
									onclick={handleResume}
									disabled={actionLoading}
									class="rounded-xl border border-[#5FCB7A]/50 bg-[#171022] px-4 py-2.5 text-sm font-black text-[#5FCB7A] transition hover:-translate-y-0.5 disabled:cursor-not-allowed disabled:opacity-50"
								>
									Resume subscription
								</button>
							{:else}
								<button
									type="button"
									onclick={() => (showCancelConfirm = true)}
									disabled={actionLoading}
									class="rounded-xl border border-[#FF6B7A]/50 bg-[#171022] px-4 py-2.5 text-sm font-black text-[#FF6B7A] transition hover:-translate-y-0.5 disabled:cursor-not-allowed disabled:opacity-50"
								>
									Cancel subscription
								</button>
							{/if}
						</div>
					{/if}
				</div>

				<!-- Current plan stats -->
				<div class="mt-8 grid gap-3 sm:grid-cols-3">
					<div class="rounded-2xl border border-[#412B6B] bg-[#171022] p-4">
						<p class="text-[11px] font-black uppercase tracking-wide text-[#f4f0e4]/45">
							AI analysis
						</p>

						<p class="mt-2 text-lg font-black">
							{currentPlan?.canViewAnalysis
								? 'Unlimited'
								: `${currentPlan?.dailySessionLimit ?? 1}/day`}
						</p>
					</div>

					<div class="rounded-2xl border border-[#412B6B] bg-[#171022] p-4">
						<p class="text-[11px] font-black uppercase tracking-wide text-[#f4f0e4]/45">
							Practice
						</p>

						<p class="mt-2 text-lg font-black">Unlimited</p>
					</div>

					<div class="rounded-2xl border border-[#412B6B] bg-[#171022] p-4">
						<p class="text-[11px] font-black uppercase tracking-wide text-[#f4f0e4]/45">
							{subscription?.cancelAtPeriodEnd ? 'Access until' : 'Next renewal'}
						</p>

						<p class="mt-2 text-lg font-black">
							{formatDate(
								subscription?.currentPeriodEnd ?? currentPlan?.planRenewsAt
							)}
						</p>
					</div>
				</div>

				{#if subscription?.cancelAtPeriodEnd}
					<div class="mt-5 rounded-2xl border border-[#FF6B7A]/40 bg-[#2a1620] p-4">
						<p class="font-black text-[#FF6B7A]">
							Your subscription is scheduled to cancel.
						</p>

						<p class="mt-1 text-sm text-[#f4f0e4]/70">
							You'll keep your current benefits until
							{formatDate(subscription.currentPeriodEnd)}.
						</p>
					</div>
				{/if}
			</section>

			<!-- Cancel confirmation -->
			{#if showCancelConfirm}
				<section class="mt-6 rounded-2xl border border-[#412B6B] bg-[#171022] p-5">
					<p class="text-lg font-black">Cancel your subscription?</p>

					<p class="mt-2 max-w-xl text-sm leading-relaxed text-[#f4f0e4]/60">
						You won't lose access immediately. Your current plan will
						remain active until the end of your billing period.
					</p>

					<div class="mt-5 flex flex-wrap gap-2">
						<button
							type="button"
							onclick={handleCancel}
							disabled={actionLoading}
							class="rounded-xl border border-[#FF6B7A]/50 bg-[#2a1620] px-5 py-2.5 text-sm font-black text-[#FF6B7A] disabled:opacity-50"
						>
							{actionLoading ? 'Cancelling…' : 'Yes, cancel'}
						</button>

						<button
							type="button"
							onclick={() => (showCancelConfirm = false)}
							disabled={actionLoading}
							class="rounded-xl border border-[#412B6B] bg-[#26224d] px-5 py-2.5 text-sm font-black text-[#f4f0e4]"
						>
							Keep subscription
						</button>
					</div>
				</section>
			{/if}

			<!-- Plans -->
			<section class="mt-14">
				<div class="mb-6">
					<p class={sectionLabelClass}>Plans</p>

					<h2 class="mt-1 text-3xl font-black tracking-tight">
						Choose what works for you.
					</h2>

					<p class="mt-2 max-w-xl text-sm text-[#f4f0e4]/55">
						You can change your plan anytime. Your current subscription
						is always shown above.
					</p>
				</div>

				<div class="grid gap-5 lg:grid-cols-2">
					{#each plans as plan (plan.id)}
						{@const current = isCurrentPlan(plan)}

						<div
							class={`relative flex flex-col rounded-3xl border p-6 sm:p-7 ${
								current
									? 'border-[#F25912] bg-[#38255b]'
									: plan.planType === 'pro'
										? 'border-[#412B6B] bg-[#26224d]'
										: 'border-[#412B6B] bg-[#171022]'
							}`}
						>
							{#if current}
								<div
									class="absolute -top-3 left-6 rounded-full border border-[#c64c13] bg-[#F25912] px-3 py-1 text-[11px] font-black uppercase tracking-wide text-[#181C14]"
								>
									Current plan
								</div>
							{:else if plan.planType === 'pro'}
								<div
									class="absolute -top-3 left-6 rounded-full border border-[#412B6B] bg-[#38255b] px-3 py-1 text-[11px] font-black uppercase tracking-wide text-[#f4f0e4]"
								>
									Recommended
								</div>
							{/if}

							<div class="mt-2 flex items-start justify-between gap-5">
								<div>
									<h3 class="text-2xl font-black">{plan.name}</h3>

									<p class="mt-2 text-sm leading-relaxed text-[#f4f0e4]/55">
										{plan.planType === 'pro'
											? 'For people who want the full Flocal experience.'
											: 'Build your speaking habit and get started with Flocal.'}
									</p>
								</div>

								<div class="shrink-0 text-right">
									{#if plan.priceSubunits === 0}
										<p class="text-3xl font-black">Free</p>
									{:else}
										<p class="text-3xl font-black">
											{formatMoney(plan.priceSubunits, plan.currency)}
										</p>

										<p class="text-xs font-bold text-[#f4f0e4]/45">
											{plan.billingInterval === 'annual' ? 'per year' : 'per month'}
										</p>
									{/if}
								</div>
							</div>

							<div class="my-6 h-px bg-[#412B6B]"></div>

							<ul class="space-y-3">
								{#each getPlanFeatures(plan) as feature}
									<li class="flex items-start gap-3 text-sm text-[#f4f0e4]/70">
										<span
											class="mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full border border-[#5FCB7A]/50 text-[10px] font-black text-[#5FCB7A]"
										>
											✓
										</span>

										<span>{feature}</span>
									</li>
								{/each}
							</ul>

							<div class="mt-auto pt-7">
								<button
									type="button"
									onclick={() => handlePlanAction(plan)}
									disabled={current || actionLoading}
									class={current
										? 'w-full rounded-xl border border-[#412B6B] bg-[#171022] px-5 py-3.5 text-sm font-black text-[#f4f0e4]/60 disabled:cursor-not-allowed'
										: `w-full ${primaryButtonClass}`}
								>
									{#if current}
										Current plan
									{:else if actionLoading}
										Processing…
									{:else if plan.planType === 'pro'}
										Get {plan.name}
									{:else}
										Switch to {plan.name}
									{/if}
								</button>
							</div>
						</div>
					{/each}
				</div>
			</section>

			<!-- Subscription details -->
			{#if subscription}
				<section class="mt-14">
					<div class="mb-6">
						<p class={sectionLabelClass}>Subscription</p>
						<h2 class="mt-1 text-2xl font-black">Subscription details</h2>
					</div>

					<div class="rounded-3xl border border-[#412B6B] bg-[#171022] p-6">
						<div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
							<div>
								<p class="text-[11px] font-black uppercase tracking-wide text-[#f4f0e4]/45">
									Status
								</p>

								<div class="mt-2 flex items-center gap-2">
									<span
										class={`rounded-full border px-3 py-1 text-xs font-black ${subscriptionStatusClass(
											subscription.status
										)}`}
									>
										{titleCase(subscription.status)}
									</span>
								</div>
							</div>

							<div>
								<p class="text-[11px] font-black uppercase tracking-wide text-[#f4f0e4]/45">
									Billing cycle
								</p>

								<p class="mt-2 font-black">
									{subscription.billingIntervalAtPurchase === 'annual'
										? 'Annual'
										: 'Monthly'}
								</p>
							</div>

							<div>
								<p class="text-[11px] font-black uppercase tracking-wide text-[#f4f0e4]/45">
									Started
								</p>

								<p class="mt-2 font-black">{formatDate(subscription.createdAt)}</p>
							</div>

							<div>
								<p class="text-[11px] font-black uppercase tracking-wide text-[#f4f0e4]/45">
									Period ends
								</p>

								<p class="mt-2 font-black">
									{formatDate(subscription.currentPeriodEnd)}
								</p>
							</div>
						</div>
					</div>
				</section>
			{/if}

			<!-- Payment history -->
			<section class="mt-14">
				<div class="mb-6">
					<p class={sectionLabelClass}>Payments</p>
					<h2 class="mt-1 text-2xl font-black">Payment history</h2>

					<p class="mt-2 text-sm text-[#f4f0e4]/55">
						Your recent Flocal payments.
					</p>
				</div>

				<div class="overflow-hidden rounded-3xl border border-[#412B6B] bg-[#171022]">
					{#if payments.length === 0}
						<div class="px-6 py-10 text-center">
							<p class="font-black">No payments yet.</p>

							<p class="mt-1 text-sm text-[#f4f0e4]/55">
								Your payment history will appear here after your
								first payment.
							</p>
						</div>
					{:else}
						<div class="divide-y divide-[#412B6B]">
							{#each payments as payment (payment.id)}
								<div
									class="flex flex-col gap-3 px-5 py-5 sm:flex-row sm:items-center sm:justify-between sm:px-6"
								>
									<div>
										<p class="font-black">
											{formatMoney(payment.amountSubunits, payment.currency)}
										</p>

										<p class="mt-1 text-xs text-[#f4f0e4]/45">
											{formatDateTime(payment.paidAt ?? payment.createdAt)}
										</p>

										{#if payment.paymentMethod}
											<p class="mt-1 text-xs text-[#f4f0e4]/35">
												{payment.paymentMethod}
											</p>
										{/if}
									</div>

									<span
										class={`w-fit rounded-full border px-3 py-1 text-xs font-black ${paymentStatusClass(
											payment.status
										)}`}
									>
										{titleCase(payment.status)}
									</span>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</section>

			<div class="pb-6 pt-10 text-center text-xs font-semibold text-[#f4f0e4]/40">
				Your billing information is managed securely through Flocal's payment provider.
			</div>
		{/if}
	</div>
</div>