<script lang="ts">
	import { onMount } from 'svelte';
	import posthog from 'posthog-js';
	import { getCurrentPlan, getMySubscription, listMyPayments, listPlans, ApiRequestError } from '$lib/api';
	import type { Payment, Subscription, SubscriptionPlan, UserCurrentPlan } from '$lib/types';

	let loading = $state(true);
	let errorMessage = $state('');

	let currentPlan = $state<UserCurrentPlan | null>(null);
	let subscription = $state<Subscription | null>(null);
	let plans = $state<SubscriptionPlan[]>([]);
	let payments = $state<Payment[]>([]);

	function formatMoney(subunits: number, currency: string): string {
		try {
			return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(
				subunits / 100
			);
		} catch {
			return `${(subunits / 100).toFixed(2)} ${currency}`;
		}
	}

	function formatDate(iso: string | null | undefined): string {
		if (!iso) return '—';
		return new Date(iso).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function statusColor(status: string): string {
		switch (status) {
			case 'active':
			case 'trialing':
			case 'captured':
				return 'text-[#9fb3ff] border-[#9fb3ff]/40';
			case 'past_due':
			case 'created':
				return 'text-yellow-300 border-yellow-300/40';
			default:
				return 'text-[#ff6abe] border-[#ff6abe]/40';
		}
	}

	function aiLimitLabel(canViewAnalysis: boolean, dailySessionLimit: number | null | undefined): string {
		if (canViewAnalysis) return 'Unlimited AI-analyzed sessions';
		if (dailySessionLimit === null || dailySessionLimit === undefined) {
			return 'Unlimited AI-analyzed sessions';
		}
		return `${dailySessionLimit} AI-analyzed session${dailySessionLimit === 1 ? '' : 's'} per day`;
	}

	onMount(async () => {
		posthog.capture('plan_page_viewed');

		const results = await Promise.allSettled([
			getCurrentPlan(),
			getMySubscription(),
			listPlans(),
			listMyPayments()
		]);

		const [planResult, subResult, plansResult, paymentsResult] = results;

		if (planResult.status === 'fulfilled') {
			currentPlan = planResult.value;
		} else if (!(planResult.reason instanceof ApiRequestError && planResult.reason.status === 404)) {
			errorMessage = 'Could not load your plan.';
		}

		if (subResult.status === 'fulfilled') {
			subscription = subResult.value;
		} else if (!(subResult.reason instanceof ApiRequestError && subResult.reason.status === 404)) {
			errorMessage = 'Could not load your subscription.';
		}

		if (plansResult.status === 'fulfilled') {
			plans = plansResult.value;
		} else {
			errorMessage = 'Could not load plans.';
		}

		if (paymentsResult.status === 'fulfilled') {
			payments = paymentsResult.value;
		}

		loading = false;
	});
</script>

<svelte:head>
	<title>Flocal — Plan & Billing</title>
</svelte:head>

<div class="min-h-screen flex flex-col items-center px-5 py-16 gap-10">
	<div class="flex flex-col items-center gap-2 text-center">
		<div class="text-4xl font-bold uppercase text-[#ff6abe]">Plan & Billing</div>
		<p class="text-sm text-white/60">Manage your plan and view payment history.</p>
	</div>

	{#if loading}
		<p class="text-white/50">Loading…</p>
	{:else}
		<div class="w-full max-w-lg flex flex-col gap-8">
			{#if errorMessage}
				<p class="text-xs text-[#e34f5e] text-center">{errorMessage}</p>
			{/if}

			<div class="rounded-2xl border border-[#342e65] bg-[#26224d] p-6 flex flex-col gap-3">
				<div class="text-xs uppercase tracking-wide text-white/40">Current plan</div>

				{#if currentPlan}
					<div class="flex items-center justify-between">
						<div class="text-lg font-semibold capitalize">{currentPlan.planSlug}</div>
						{#if currentPlan.subscriptionStatus}
							<span
								class="rounded-full border px-3 py-1 text-xs capitalize {statusColor(
									currentPlan.subscriptionStatus
								)}"
							>
								{currentPlan.subscriptionStatus.replace('_', ' ')}
							</span>
						{/if}
					</div>

					<div class="text-sm text-white/60">Unlimited practice sessions</div>

					<div class="text-sm text-white/60">
						{aiLimitLabel(currentPlan.canViewAnalysis, currentPlan.dailySessionLimit)}
					</div>

					{#if currentPlan.planSlug === 'free'}
						<div class="text-xs text-white/40">
							Includes 5 free AI-analyzed sessions to start, then
							{currentPlan.dailySessionLimit ?? 1} per day after that.
						</div>
					{/if}

					{#if currentPlan.planRenewsAt}
						<div class="text-xs text-white/40">Renews {formatDate(currentPlan.planRenewsAt)}</div>
					{/if}
				{:else}
					<p class="text-sm text-white/50">You're on the free plan.</p>
				{/if}

				{#if subscription?.cancelAtPeriodEnd}
					<p class="text-xs text-yellow-300">
						Cancels at the end of the current period ({formatDate(subscription.currentPeriodEnd)}).
					</p>
				{/if}
			</div>

			<div class="flex flex-col gap-3">
				<div class="text-xs uppercase tracking-wide text-white/40 px-1">Available plans</div>

				<div class="flex flex-col gap-3">
					{#each plans as plan (plan.id)}
						<div
							class="rounded-2xl border p-5 flex flex-col gap-2 {plan.slug === currentPlan?.planSlug
								? 'border-[#9fb3ff] bg-[#26224d]'
								: 'border-[#342e65] bg-[#151229]'}"
						>
							<div class="flex items-center justify-between">
								<div class="font-semibold">{plan.name}</div>
								<div class="text-sm text-white/70">
									{plan.priceSubunits === 0
										? 'Free'
										: `${formatMoney(plan.priceSubunits, plan.currency)} / ${plan.billingInterval}`}
								</div>
							</div>

							<div class="text-xs text-white/50">
								Unlimited practice · {aiLimitLabel(plan.canViewAnalysis, plan.dailySessionLimit)}
								{plan.planType === 'free' ? ' (+5 free to start)' : ''}
							</div>

							<button
								type="button"
								disabled
								class="mt-2 cursor-not-allowed rounded-full py-2 text-sm font-semibold bg-[#342e65] text-white/40"
							>
								{plan.slug === currentPlan?.planSlug ? 'Current plan' : 'Checkout coming soon'}
							</button>
						</div>
					{/each}
				</div>
			</div>

			<div class="flex flex-col gap-3">
				<div class="text-xs uppercase tracking-wide text-white/40 px-1">Payment history</div>

				{#if payments.length === 0}
					<p class="text-sm text-white/40 px-1">No payments yet.</p>
				{:else}
					<div class="flex flex-col gap-2">
						{#each payments as payment (payment.id)}
							<div
								class="rounded-xl border border-[#342e65] bg-[#151229] px-4 py-3 flex items-center justify-between"
							>
								<div class="flex flex-col">
									<span class="text-sm">{formatMoney(payment.amountSubunits, payment.currency)}</span>
									<span class="text-xs text-white/40">{formatDate(payment.paidAt ?? payment.createdAt)}</span>
								</div>
								<span class="rounded-full border px-3 py-1 text-xs capitalize {statusColor(payment.status)}">
									{payment.status}
								</span>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>