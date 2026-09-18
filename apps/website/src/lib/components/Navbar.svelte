<script>
    import { onMount } from 'svelte';

    let mobileMenuOpen = $state(false);
    let mobileFeaturesOpen = $state(false);
    let mobileResourcesOpen = $state(false);

    const features = [
        {
            title: 'AI Speech Analysis',
            description: 'Understand how you actually speak.',
            href: '/features/speech-analysis'
        },
        {
            title: 'Fluency Scoring',
            description: 'Track your speaking performance.',
            href: '/features/fluency'
        },
        {
            title: 'Grammar Feedback',
            description: 'Spot patterns and improve naturally.',
            href: '/features/grammar'
        },
        {
            title: 'Vocabulary',
            description: 'Build a stronger speaking vocabulary.',
            href: '/features/vocabulary'
        },
        {
            title: 'Session History',
            description: 'See how your speaking evolves.',
            href: '/features/history'
        },
        {
            title: 'Daily Challenges',
            description: 'Build a consistent speaking habit.',
            href: '/features/challenges'
        }
    ];

    const resources = [
        {
            title: 'Blog',
            description: 'Ideas for becoming a better speaker.',
            href: '/blog'
        },
        {
            title: 'Speaking Tips',
            description: 'Practical techniques for better speech.',
            href: '/resources/speaking-tips'
        },
        {
            title: 'Guides',
            description: 'Deep dives into communication skills.',
            href: '/resources/guides'
        },
        {
            title: 'Topics',
            description: 'Find something interesting to talk about.',
            href: '/topics'
        }
    ];

    function lockBodyScroll() {
        document.body.style.overflow = 'hidden';
    }

    function unlockBodyScroll() {
        document.body.style.overflow = '';
    }

    function toggleMobileMenu() {
        mobileMenuOpen = !mobileMenuOpen;
        mobileFeaturesOpen = false;
        mobileResourcesOpen = false;

        if (mobileMenuOpen) {
            lockBodyScroll();
        } else {
            unlockBodyScroll();
        }
    }

    function closeMobileMenu() {
        mobileMenuOpen = false;
        mobileFeaturesOpen = false;
        mobileResourcesOpen = false;
        unlockBodyScroll();
    }

    onMount(() => {
        const handleResize = () => {
            if (window.innerWidth >= 768) {
                mobileMenuOpen = false;
                mobileFeaturesOpen = false;
                mobileResourcesOpen = false;
                unlockBodyScroll();
            }
        };

        window.addEventListener('resize', handleResize);

        return () => {
            window.removeEventListener('resize', handleResize);
            unlockBodyScroll();
        };
    });
</script>

<nav class="relative z-50 px-5 py-5 sm:px-6">
    <div class="mx-auto flex max-w-7xl items-center justify-between">
        <a
            href="/"
            aria-label="Flocal home"
            class="text-2xl font-black tracking-tight text-[#1C1124] sm:text-3xl"
        >
            Flocal
        </a>

        <div
            class="hidden items-center gap-8 font-semibold text-[#1C1124] md:flex lg:gap-10"
        >
            <div class="group relative">
                <button
                    type="button"
                    class="flex cursor-pointer items-center gap-1.5 py-3 text-[#1C1124]"
                >
                    Features

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
                        class="h-4 w-4 transition-transform duration-200 group-hover:rotate-180 group-focus-within:rotate-180"
                        aria-hidden="true"
                    >
                        <path d="m6 9 6 6 6-6" />
                    </svg>
                </button>

                <div
                    class="pointer-events-none invisible absolute left-1/2 top-full w-130 -translate-x-1/2 translate-y-2 opacity-0 transition-all duration-200 ease-out group-hover:pointer-events-auto group-hover:visible group-hover:translate-y-0 group-hover:opacity-100 group-focus-within:pointer-events-auto group-focus-within:visible group-focus-within:translate-y-0 group-focus-within:opacity-100"
                >
                    <div class="pt-3">
                        <div
                            class="grid grid-cols-2 gap-1 rounded-2xl border-2 border-[#1C1124] bg-[#F7FFCD] p-3 shadow-[6px_6px_0px_#1C1124]"
                        >
                            {#each features as feature}
                                <a
                                    href={feature.href}
                                    class="rounded-xl p-4 transition-colors hover:bg-[#eff9bd]"
                                >
                                    <p class="font-bold text-[#1C1124]">
                                        {feature.title}
                                    </p>

                                    <p
                                        class="mt-1 text-sm font-medium leading-snug text-[#4d2a3a]"
                                    >
                                        {feature.description}
                                    </p>
                                </a>
                            {/each}
                        </div>
                    </div>
                </div>
            </div>

            <div class="group relative">
                <button
                    type="button"
                    class="flex cursor-pointer items-center gap-1.5 py-3 text-[#1C1124]"
                >
                    Resources

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
                        class="h-4 w-4 transition-transform duration-200 group-hover:rotate-180 group-focus-within:rotate-180"
                        aria-hidden="true"
                    >
                        <path d="m6 9 6 6 6-6" />
                    </svg>
                </button>

                <div
                    class="pointer-events-none invisible absolute left-1/2 top-full w-100 -translate-x-1/2 translate-y-2 opacity-0 transition-all duration-200 ease-out group-hover:pointer-events-auto group-hover:visible group-hover:translate-y-0 group-hover:opacity-100 group-focus-within:pointer-events-auto group-focus-within:visible group-focus-within:translate-y-0 group-focus-within:opacity-100"
                >
                    <div class="pt-3">
                        <div
                            class="grid grid-cols-2 gap-1 rounded-2xl border-2 border-[#1C1124] bg-[#F7FFCD] p-3 shadow-[6px_6px_0px_#1C1124]"
                        >
                            {#each resources as resource}
                                <a
                                    href={resource.href}
                                    class="rounded-xl p-4 transition-colors hover:bg-[#eff9bd]"
                                >
                                    <p class="font-bold text-[#1C1124]">
                                        {resource.title}
                                    </p>

                                    <p
                                        class="mt-1 text-sm font-medium leading-snug text-[#4d2a3a]"
                                    >
                                        {resource.description}
                                    </p>
                                </a>
                            {/each}
                        </div>
                    </div>
                </div>
            </div>

            <a
                href="/pricing"
                class="py-3 text-[#1C1124]"
            >
                Pricing
            </a>
        </div>

        <div class="hidden md:block">
            <a
                href="https://web.flocal.app/login"
                class="block cursor-pointer rounded-[0.75em] border-2 border-[#4d2a3a] bg-[#ff94d0] px-4 py-1.5 text-lg font-semibold text-[#4d2a3a] transition-transform duration-200 hover:-translate-y-1"
            >
                Login
            </a>
        </div>

        <button
            type="button"
            aria-label={mobileMenuOpen ? 'Close navigation' : 'Open navigation'}
            aria-expanded={mobileMenuOpen}
            onclick={toggleMobileMenu}
            class="relative z-60 flex h-11 w-11 cursor-pointer items-center justify-center rounded-xl border-2 border-[#1C1124] bg-[#F7FFCD] text-[#1C1124] md:hidden"
        >
            {#if mobileMenuOpen}
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
                    class="h-6 w-6"
                    aria-hidden="true"
                >
                    <path d="M18 6 6 18" />
                    <path d="m6 6 12 12" />
                </svg>
            {:else}
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
                    class="h-6 w-6"
                    aria-hidden="true"
                >
                    <path d="M4 5h16" />
                    <path d="M4 12h16" />
                    <path d="M4 19h16" />
                </svg>
            {/if}
        </button>
    </div>

    {#if mobileMenuOpen}
        <div
            class="fixed inset-x-0 bottom-0 top-20.25 z-40 overflow-y-auto overscroll-contain bg-[#BADF96] px-5 pb-8 pt-6 sm:px-6 md:hidden"
        >
            <div class="mx-auto max-w-7xl">
                <div class="flex flex-col text-[#1C1124]">
                    <button
                        type="button"
                        onclick={() => {
                            mobileFeaturesOpen = !mobileFeaturesOpen;
                            mobileResourcesOpen = false;
                        }}
                        class="flex items-center justify-between py-4 text-left text-lg font-semibold"
                    >
                        <span>Features</span>

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
                            class={`h-5 w-5 transition-transform duration-200 ${
                                mobileFeaturesOpen ? 'rotate-180' : ''
                            }`}
                            aria-hidden="true"
                        >
                            <path d="m6 9 6 6 6-6" />
                        </svg>
                    </button>

                    {#if mobileFeaturesOpen}
                        <div
                            class="mb-3 grid grid-cols-1 gap-1 rounded-2xl border-2 border-[#1C1124] bg-[#F7FFCD] p-3 shadow-[6px_6px_0px_#1C1124]"
                        >
                            {#each features as feature}
                                <a
                                    href={feature.href}
                                    onclick={closeMobileMenu}
                                    class="rounded-xl p-4 transition-colors hover:bg-[#eff9bd]"
                                >
                                    <p class="font-bold text-[#1C1124]">
                                        {feature.title}
                                    </p>

                                    <p
                                        class="mt-1 text-sm font-medium leading-snug text-[#4d2a3a]"
                                    >
                                        {feature.description}
                                    </p>
                                </a>
                            {/each}
                        </div>
                    {/if}

                    <button
                        type="button"
                        onclick={() => {
                            mobileResourcesOpen = !mobileResourcesOpen;
                            mobileFeaturesOpen = false;
                        }}
                        class="flex items-center justify-between py-4 text-left text-lg font-semibold"
                    >
                        <span>Resources</span>

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
                            class={`h-5 w-5 transition-transform duration-200 ${
                                mobileResourcesOpen ? 'rotate-180' : ''
                            }`}
                            aria-hidden="true"
                        >
                            <path d="m6 9 6 6 6-6" />
                        </svg>
                    </button>

                    {#if mobileResourcesOpen}
                        <div
                            class="mb-3 grid grid-cols-1 gap-1 rounded-2xl border-2 border-[#1C1124] bg-[#F7FFCD] p-3 shadow-[6px_6px_0px_#1C1124]"
                        >
                            {#each resources as resource}
                                <a
                                    href={resource.href}
                                    onclick={closeMobileMenu}
                                    class="rounded-xl p-4 transition-colors hover:bg-[#eff9bd]"
                                >
                                    <p class="font-bold text-[#1C1124]">
                                        {resource.title}
                                    </p>

                                    <p
                                        class="mt-1 text-sm font-medium leading-snug text-[#4d2a3a]"
                                    >
                                        {resource.description}
                                    </p>
                                </a>
                            {/each}
                        </div>
                    {/if}

                    <a
                        href="/pricing"
                        onclick={closeMobileMenu}
                        class="py-4 text-lg font-semibold text-[#1C1124]"
                    >
                        Pricing
                    </a>

                    <a
                        href="https://web.flocal.app/login"
                        onclick={closeMobileMenu}
                        class="mt-3 flex w-full cursor-pointer items-center justify-center rounded-[0.75em] border-2 border-[#4d2a3a] bg-[#ff94d0] px-5 py-3.5 text-lg font-bold text-[#4d2a3a] transition-transform duration-200 hover:-translate-y-1"
                    >
                        Login
                    </a>
                </div>
            </div>
        </div>
    {/if}
</nav>