<script lang="ts">
	let status: 'verifying' | 'error' = $state('verifying');
	let errorMessage = $state('');
</script>

<svelte:head>
	<title>Flocal - Sign in</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center px-5 py-10 sm:px-8">
	<div class="w-full max-w-sm">
		<div class="text-center">
			<h1 class="text-5xl font-black leading-[0.95] tracking-[-0.04em] sm:text-6xl">
				{status === 'verifying' ? 'Signing you in' : 'Sign in failed'}
			</h1>

			<p class="mx-auto mt-5 max-w-xs text-base leading-relaxed text-[#d8d8d8] sm:text-lg">
				{status === 'verifying'
					? 'Just a moment while we finish signing you in.'
					: errorMessage || 'This link may have expired.'}
			</p>
		</div>

		<div class="mt-12 flex w-full flex-col items-center">
			{#if status === 'verifying'}
				<div
					class="h-8 w-8 animate-spin rounded-full border-[3px] border-[#464040] border-t-[#FF7315]"
				></div>
			{:else}
				<a
					href="/login"
					class="flex h-13.5 w-full cursor-pointer items-center justify-center rounded-xl border border-[#ff6803] bg-[#FF7315] px-5 font-bold text-[#232020] transition-transform duration-200 hover:-translate-y-1"
				>
					Back to sign in
				</a>
			{/if}
		</div>
	</div>
</div>