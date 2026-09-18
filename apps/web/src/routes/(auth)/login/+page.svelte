<script lang="ts">
    import { untrack } from 'svelte';
    import { enhance } from '$app/forms';
    import { replaceState } from '$app/navigation';
    import { page } from '$app/state';
    import { PUBLIC_API_BASE_URL } from '$env/static/public';
    import type { ActionData, PageData } from './$types';

    let { data, form }: { data: PageData; form: ActionData } = $props();

    let submitting = $state(false);
    let email = $state('');
    let emailInput: HTMLInputElement | undefined = $state();
    let dismissedError = $state(false);

    let emailValid = $derived(/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email));

    let emailStatus = $state<'idle' | 'success' | 'error'>('idle');
    let errorMessage = $state('');

    function oauthUrl(provider: 'google' | 'github') {
        return `${PUBLIC_API_BASE_URL}/api/v1/auth/${provider}`;
    }

    function resetStatus() {
        if (emailStatus !== 'idle') {
            emailStatus = 'idle';
            errorMessage = '';
        }
    }

    function clearEmail(event: MouseEvent) {
        event.stopPropagation();
        email = '';
        emailStatus = 'idle';
        errorMessage = '';
        dismissedError = true;
        emailInput?.focus();
    }

    function handleEmailSubmit(event: SubmitEvent) {
        if (!email.trim()) {
            event.preventDefault();

            emailStatus = 'error';
            errorMessage = 'Enter your email address.';
            dismissedError = false;

            emailInput?.focus();
            return;
        }

        if (!emailValid) {
            event.preventDefault();

            emailStatus = 'error';
            errorMessage = 'Enter a valid email address.';
            dismissedError = false;

            emailInput?.focus();
        }
    }

    $effect(() => {
        if (data.errorMessage && !dismissedError) {
            emailStatus = 'error';
            errorMessage = data.errorMessage;
            dismissedError = true;

            untrack(() => {
                const url = new URL(page.url);
                url.searchParams.delete('error');
                replaceState(url, page.state);
            });
        }
    });

    $effect(() => {
        if (form?.sent) {
            emailStatus = 'success';
            errorMessage = '';
            dismissedError = false;
        } else if (form?.error) {
            emailStatus = 'error';
            errorMessage = form.error;
            dismissedError = false;

            if (form.email) {
                email = form.email;
            }
        }
    });

    $effect(() => {
        emailInput?.focus();
    });
</script>

<svelte:head>
    <title>Flocal - Sign in</title>
</svelte:head>

<div
    class="flex min-h-screen items-center justify-center px-5 py-10 sm:px-8"
>
    <div class="w-full max-w-sm">

        <div class="text-center">
            <h1
                class="text-5xl font-black leading-[0.95] tracking-[-0.04em] sm:text-6xl"
            >
                Sign in
            </h1>

            <p
                class="mx-auto mt-5 max-w-xs text-base leading-relaxed text-[#d8d8d8] sm:text-lg"
            >
                One step closer to speaking with confidence.
            </p>
        </div>

        <div class="mt-12 flex w-full flex-col gap-3">

            <a
                href={oauthUrl('google')}
                class="flex h-13.5 w-full cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#464040] bg-[#3A3535] px-5 font-bold transition-transform duration-200 hover:-translate-y-1"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="22"
                    height="22"
                    viewBox="0 0 48 48"
                    class="shrink-0"
                >
                    <path
                        fill="#FFC107"
                        d="M43.611,20.083H42V20H24v8h11.303c-1.649,4.657-6.08,8-11.303,8c-6.627,0-12-5.373-12-12c0-6.627,5.373-12,12-12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C12.955,4,4,12.955,4,24c0,11.045,8.955,20,20,20c11.045,0,20-8.955,20-20C44,22.659,43.862,21.35,43.611,20.083z"
                    />
                    <path
                        fill="#FF3D00"
                        d="M6.306,14.691l6.571,4.819C14.655,15.108,18.961,12,24,12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C16.318,4,9.656,8.337,6.306,14.691z"
                    />
                    <path
                        fill="#4CAF50"
                        d="M24,44c5.166,0,9.86-1.977,13.409-5.192l-6.19-5.238C29.211,35.091,26.715,36,24,36c-5.202,0-9.619-3.317-11.283-7.946l-6.522,5.025C9.505,39.556,16.227,44,24,44z"
                    />
                    <path
                        fill="#1976D2"
                        d="M43.611,20.083H42V20H24v8h11.303c-.792,2.237-2.231,4.166-4.087,5.571l.003-.002l6.19,5.238C36.971,39.205,44,34,44,24C44,22.659,43.862,21.35,43.611,20.083z"
                    />
                </svg>

                Continue with Google
            </a>

            <a
                href={oauthUrl('github')}
                class="flex h-13.5 w-full cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#464040] bg-[#3A3535] px-5 font-bold transition-transform duration-200 hover:-translate-y-1"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="22"
                    height="22"
                    viewBox="0 0 32 32"
                    class="shrink-0"
                >
                    <g fill="#ffffff">
                        <path
                            d="M16,2.345c7.735,0,14,6.265,14,14-.002,6.015-3.839,11.359-9.537,13.282-.7,.14-.963-.298-.963-.665,0-.473,.018-1.978,.018-3.85,0-1.312-.437-2.152-.945-2.59,3.115-.35,6.388-1.54,6.388-6.912,0-1.54-.543-2.783-1.435-3.762,.14-.35,.63-1.785-.14-3.71,0,0-1.173-.385-3.85,1.435-1.12-.315-2.31-.472-3.5-.472s-2.38,.157-3.5,.472c-2.677-1.802-3.85-1.435-3.85-1.435-.77,1.925-.28,3.36-.14,3.71-.892,.98-1.435,2.24-1.435,3.762,0,5.355,3.255,6.563,6.37,6.913-.403,.35-.77,.963-.893,1.872-.805,.368-2.818,.963-4.077-1.155-.263-.42-1.05-1.452-2.152-1.435-1.173,.018-.472,.665,.017,.927,.595,.332,1.277,1.575,1.435,1.978,.28,.787,1.19,2.293,4.707,1.645,0,1.173,.018,2.275,.018,2.607,0,.368-.263,.787-.963,.665-5.719-1.904-9.576-7.255-9.573-13.283,0-7.735,6.265-14,14-14Z"
                        />
                    </g>
                </svg>

                Continue with GitHub
            </a>

            <div class="flex items-center gap-4 py-3">
                <div class="h-px flex-1 bg-[#464040]/40"></div>

                <span
                    class="text-xs font-bold uppercase tracking-[0.14em] text-[#464040]"
                >
                    or
                </span>

                <div class="h-px flex-1 bg-[#464040]/40"></div>
            </div>

            <form
                method="POST"
                action="?/requestMagicLink"
                class="flex flex-col gap-3"
                onsubmit={handleEmailSubmit}
                use:enhance={() => {
                    submitting = true;

                    return async ({ update }) => {
                        submitting = false;

                        await update({
                            reset: false,
                            invalidateAll: false
                        });
                    };
                }}
            >
                <div class="relative w-full">
                    <input
                        bind:this={emailInput}
                        id="email"
                        name="email"
                        type="email"
                        autocomplete="email"
                        placeholder="you@example.com"
                        bind:value={email}
                        oninput={resetStatus}
                        required
                        aria-invalid={emailStatus === 'error'}
                        aria-describedby="email-error"
                        class="h-13.5 w-full rounded-xl border bg-[#1a1818] px-10 text-center font-medium placeholder:text-[#464040] focus:outline-none transition-colors duration-200
                        {emailStatus === 'error'
                            ? 'border-[#FF6B7A]'
                            : 'border-[#464040]'}"
                    />

                    {#if email.length > 0}
                        <button
                            type="button"
                            aria-label="Clear email input"
                            onclick={clearEmail}
                            class="absolute right-4 top-1/2 -translate-y-1/2 cursor-pointer text-[#F4F4F4]/55 transition-colors duration-200 hover:text-[#F4F4F4]"
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
                    {/if}
                </div>

                <button
                    type="submit"
                    disabled={submitting}
                    class="flex h-13.5 w-full cursor-pointer items-center justify-center gap-3 rounded-xl border border-[#ff6803] bg-[#FF7315] px-5 font-bold text-[#232020] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-70 disabled:hover:translate-y-0"
                >
                    {#if submitting}
                        <div
                            class="h-5 w-5 shrink-0 animate-spin rounded-full border-2 border-[#232020]/25 border-t-[#232020]"
                        ></div>
                        Sending…
                    {:else if emailStatus === 'success'}
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2.5"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            class="shrink-0"
                        >
                            <path d="M20 6 9 17l-5-5" />
                        </svg>
                        Magic Link sent
                    {:else}
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="22"
                            height="22"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            class="shrink-0"
                        >
                            <path d="m22 7-8.991 5.727a2 2 0 0 1-2.009 0L2 7" />
                            <rect x="2" y="4" width="20" height="16" rx="3.25" />
                        </svg>
                        {emailStatus === 'error' && !emailValid ? 'Try again' : 'Continue with Email'}
                    {/if}
                </button>

                <p
                    id="email-error"
                    role="alert"
                    class="min-h-5 text-center text-xs font-semibold leading-relaxed {emailStatus ===
                    'error'
                        ? 'text-[#FF6B7A]'
                        : 'invisible'}"
                >
                    {errorMessage || '\u00A0'}
                </p>
            </form>
        </div>
    </div>
</div>