<script lang="ts">
	import { page } from '$app/state';
	import posthog from 'posthog-js';

	let { children } = $props();

	let user = $derived(page.data.user);

	$effect(() => {
		if (user) {
			posthog.identify(user.id, {
				email: user.email,
				name: user.name,
				username: user.username,
				role: user.role
			});
		}
	});
</script>

<svelte:head>
	<!-- <link rel="icon" href="/favicon.ico" /> -->
</svelte:head>

{@render children()}
