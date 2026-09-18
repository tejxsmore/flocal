<script>
    import { onMount } from 'svelte';
    import gsap from 'gsap';

    let annual = $state(true);

    const plans = [
        {
            name: 'Free',
            description: 'Build the habit and see what Flocal can do.',
            monthly: 0,
            annual: 0,
            cta: 'Start for free',
            featured: false,
            features: [
                '5 AI sessions to get started',
                '1 AI session per day after that',
                'Unlimited recording',
                '7 days of session history',
                'AI transcript for eligible sessions',
                'Core speech scores from AI sessions'
            ]
        },
        {
            name: 'Pro',
            description: 'Unlimited practice for serious improvement.',
            monthly: 5,
            annual: 48,
            cta: 'Get Pro',
            featured: true,
            features: [
                'Unlimited AI sessions',
                'Unlimited recording',
                'Unlimited session history',
                'Re-analyze sessions anytime',
                'Download your audio files',
                'Full speech analysis',
                'Grammar and vocabulary insights',
                'Advanced progress tracking'
            ]
        }
    ];

    const faqs = [
        {
            question: 'What counts as an AI session?',
            answer:
                'One AI session includes your speech recording, transcript, and the AI analysis generated from it. Simply recording your speech does not consume an AI session.'
        },
        {
            question: 'How does the Free plan work?',
            answer:
                'You get 5 AI sessions when you start. After those are used, you get 1 AI session every day. You can continue recording without limits, but recordings made after your daily AI session has been used will not receive AI analysis until another session becomes available.'
        },
        {
            question: 'Do I lose my recordings after using my AI session?',
            answer:
                'No. Recording is unlimited on the Free plan. Your recordings remain available even after you have used your available AI session.'
        },
        {
            question: 'What happens to my session history on Free?',
            answer:
                'Free users have access to 7 days of session history. Pro users get unlimited session history, so they can look back at their speaking progress for as long as they keep their subscription.'
        },
        {
            question: 'What do I get with Pro?',
            answer:
                'Pro gives you unlimited AI sessions, unlimited session history, re-analysis of previous sessions, downloadable audio, full speech analysis, grammar and vocabulary insights, and advanced progress tracking.'
        },
        {
            question: 'Why is the annual plan $4 per month?',
            answer:
                'The annual plan costs $48 for the year. That works out to $4 per month, which is 20% less than the $5 monthly plan.'
        },
        {
            question: 'Are taxes included?',
            answer:
                'Yes. All prices shown on this page are tax inclusive, so applicable taxes are already included in the displayed price.'
        },
        {
            question: 'Can I cancel Pro?',
            answer:
                'Yes. You can cancel your Pro subscription at any time. Your Pro access will continue according to your current billing period.'
        }
    ];

    onMount(() => {
        const ctx = gsap.context(() => {
            gsap.from('.pricing-hero > *', {
                opacity: 0,
                y: 22,
                duration: 0.75,
                stagger: 0.08,
                ease: 'power3.out'
            });

            gsap.from('.pricing-card', {
                opacity: 0,
                y: 35,
                duration: 0.85,
                stagger: 0.12,
                delay: 0.18,
                ease: 'power3.out'
            });

            gsap.from('.pricing-info, .faq-section', {
                opacity: 0,
                y: 25,
                duration: 0.8,
                stagger: 0.1,
                delay: 0.35,
                ease: 'power3.out'
            });

            document.querySelectorAll('.faq-item').forEach((item) => {
                const summary = item.querySelector('summary');
                const content = item.querySelector('.faq-content');
                const inner = item.querySelector('.faq-content-inner');

                if (!summary || !content || !inner) return;

                let animation = $state();

                summary.addEventListener('click', (event) => {
                    event.preventDefault();

                    if (animation) {
                        animation.kill();
                    }

                    const isOpen = item.hasAttribute('open');

                    if (isOpen) {
                        const currentHeight = content.scrollHeight;

                        gsap.set(content, {
                            height: currentHeight,
                            overflow: 'hidden'
                        });

                        animation = gsap.timeline({
                            onComplete: () => {
                                item.removeAttribute('open');

                                gsap.set(content, {
                                    height: 0,
                                    overflow: 'hidden'
                                });
                            }
                        });

                        animation
                            .to(inner, {
                                opacity: 0,
                                y: -8,
                                duration: 0.18,
                                ease: 'power2.in'
                            })
                            .to(
                                content,
                                {
                                    height: 0,
                                    duration: 0.35,
                                    ease: 'power3.inOut'
                                },
                                '-=0.05'
                            );
                    } else {
                        item.setAttribute('open', '');

                        gsap.set(content, {
                            height: 0,
                            overflow: 'hidden'
                        });

                        gsap.set(inner, {
                            opacity: 0,
                            y: -8
                        });

                        const targetHeight = content.scrollHeight;

                        animation = gsap.timeline({
                            onComplete: () => {
                                gsap.set(content, {
                                    height: 'auto',
                                    overflow: 'visible'
                                });
                            }
                        });

                        animation
                            .to(content, {
                                height: targetHeight,
                                duration: 0.4,
                                ease: 'power3.out'
                            })
                            .to(
                                inner,
                                {
                                    opacity: 1,
                                    y: 0,
                                    duration: 0.3,
                                    ease: 'power3.out'
                                },
                                '-=0.22'
                            );
                    }
                });
            });
        });

        return () => ctx.revert();
    });
</script>

<svelte:head>
    <title>Pricing | Flocal</title>
    <meta
        name="description"
        content="Practice speaking for free with Flocal or unlock unlimited AI-powered speech coaching with Pro."
    />
</svelte:head>

<div class="min-h-screen bg-[#BADF96] px-5 sm:px-10 md:px-20 py-10 pb-20 text-[#1C1124]">
    <section class="pricing-hero mx-auto max-w-3xl text-center space-y-15">
        <div
            class="mx-auto inline-flex items-center gap-2 rounded-full
            border-2 border-[#1C1124] bg-[#F7FFCD] px-4 py-2
            font-semibold text-[#4d2a3a]"
        >
            Simple pricing
        </div>

        <h1
            class="text-5xl font-black leading-[0.95]
            tracking-[-0.04em] sm:text-6xl md:text-7xl"
        >
            Practice for free.
            <span class="relative z-10 inline-block">
                <span class="relative z-10">
                    Improve without limits.
                </span>
            </span>
        </h1>

        <p
            class="mx-auto max-w-2xl text-lg leading-relaxed
            text-[#4d2a3a] sm:text-xl"
        >
            Start with 5 AI-powered sessions, then keep practicing
            every day for free. Go Pro when you want unlimited analysis
            and your complete speaking history.
        </p>

        <div class="space-y-5">
            <div
                class="inline-flex items-center rounded-xl
                border-2 border-[#1C1124] bg-[#9FA1FF] p-1"
            >
                <button
                    type="button"
                    onclick={() => (annual = false)}
                    class={`cursor-pointer rounded-lg px-5 py-2 font-bold
                    transition-colors ${
                        !annual
                            ? 'bg-[#1C1124] text-white'
                            : 'text-[#1C1124]'
                    }`}
                >
                    Monthly
                </button>

                <button
                    type="button"
                    onclick={() => (annual = true)}
                    class={`cursor-pointer rounded-lg px-5 py-2 font-bold
                    transition-colors ${
                        annual
                            ? 'bg-[#1C1124] text-white'
                            : 'text-[#1C1124]'
                    }`}
                >
                    Annual
                    <span
                        class={`ml-1 text-xs font-black ${
                            annual
                                ? 'text-[#F7E396]'
                                : 'text-[#4d2a3a]'
                        }`}
                    >
                        Save 20%
                    </span>
                </button>
            </div>

            <p class="mt-3 text-sm font-semibold text-[#4d2a3a]">
                Tax inclusive pricing
            </p>
        </div>
    </section>

    <section
        class="mx-auto mt-14 grid max-w-5xl items-stretch gap-6
        lg:grid-cols-2"
    >
        {#each plans as plan}
            <div
                class={`pricing-card relative flex h-full flex-col
                rounded-4xl border-2 border-[#1C1124] p-7
                shadow-[8px_8px_0px_#1C1124] sm:p-9 ${
                    plan.featured
                        ? 'bg-[#F7E396]'
                        : 'bg-white'
                }`}
            >
                {#if plan.featured}
                    <div
                        class="absolute -top-4 left-7 rounded-full
                        border-2 border-[#1C1124] bg-[#ff94d0]
                        px-4 py-1.5 text-sm font-black"
                    >
                        Most popular
                    </div>
                {/if}

                <div
                    class="flex min-h-34.5 items-start
                    justify-between gap-4"
                >
                    <div class="min-w-0">
                        <h2 class="text-3xl font-black">
                            {plan.name}
                        </h2>

                        <p
                            class="mt-2 max-w-xs text-[#4d2a3a]
                            line-clamp-2 leading-relaxed"
                        >
                            {plan.description}
                        </p>
                    </div>

                    {#if plan.monthly > 0}
                        <div class="shrink-0 text-right">
                            <div
                                class="flex items-baseline gap-1
                                whitespace-nowrap"
                            >
                                <span class="text-5xl font-black">
                                    ${annual ? 4 : 5}
                                </span>

                                <span
                                    class="text-sm font-semibold
                                    text-[#4d2a3a]"
                                >
                                    /month
                                </span>
                            </div>

                            {#if annual}
                                <p
                                    class="mt-1 text-xs font-bold
                                    text-[#4d2a3a]"
                                >
                                    billed $48/year
                                </p>
                            {/if}
                        </div>
                    {/if}
                </div>

                <a
                    href={plan.featured ? '/checkout' : '/signup'}
                    class={`mt-8 flex w-full cursor-pointer
                    items-center justify-center rounded-[0.75em]
                    border-2 px-5 py-3.5 text-lg font-bold
                    transition-transform duration-200
                    hover:-translate-y-1 ${
                        plan.featured
                            ? 'border-[#4d2a3a] bg-[#ff94d0] text-[#4d2a3a]'
                            : 'border-[#1C1124] bg-[#BADF96] text-[#1C1124]'
                    }`}
                >
                    {plan.cta}
                </a>

                <div class="my-8 h-px bg-[#1C1124]/15"></div>

                <p class="mb-4 font-black">
                    What's included
                </p>

                <ul class="space-y-3">
                    {#each plan.features as feature}
                        <li
                            class="flex items-start gap-3 text-[#4d2a3a]"
                        >
                            <span
                                class="mt-0.5 flex h-5 w-5 shrink-0
                                items-center justify-center rounded-full
                                border-2 border-[#1C1124]
                                bg-[#BADF96] text-xs font-black"
                            >
                                ✓
                            </span>

                            <span>
                                {feature}
                            </span>
                        </li>
                    {/each}
                </ul>

                <div class="mt-auto pt-7">
                    {#if plan.featured}
                        <div
                            class="rounded-xl border-2 border-[#1C1124]
                            bg-[#ffe684] px-4 py-3 text-sm font-semibold
                            leading-relaxed text-[#1C1124]"
                        >
                            Unlimited AI sessions, unlimited history,
                            and unlimited opportunities to practice
                            and improve.
                        </div>
                    {:else}
                        <div
                            class="rounded-xl border-2 border-[#1C1124]
                            bg-[#F0F0F0] px-4 py-3 text-sm font-semibold
                            leading-relaxed text-[#1C1124]"
                        >
                            You can always keep recording. AI
                            session allowance only controls when Flocal
                            analyzes your speech.
                        </div>
                    {/if}
                </div>
            </div>
        {/each}
    </section>

    <section
        class="pricing-info mx-auto mt-20 max-w-4xl rounded-2xl
        border-2 border-[#1C1124] bg-[#9FA1FF] p-6
        shadow-[6px_6px_0px_#1C1124] sm:p-8"
    >
        <h2 class="text-xl font-black">
            What counts as an AI session?
        </h2>

        <p
            class="mt-3 max-w-xl leading-relaxed
            text-[#4d2a3a]"
        >
            One AI session includes your speech recording, transcript,
            and the AI analysis generated from it. Recording by itself
            does not consume an AI session.
        </p>

        <div class="mt-6 grid gap-4 sm:grid-cols-2">
            <div
                class="rounded-xl border-2 border-[#1C1124]
                bg-white p-5"
            >
                <p class="font-black">
                    Free
                </p>

                <p
                    class="mt-2 text-sm leading-relaxed
                    text-[#4d2a3a]"
                >
                    Get 5 AI sessions to start, followed by 1 AI
                    session every day. You can continue recording
                    even after your daily AI session is used.
                </p>
            </div>

            <div
                class="rounded-xl border-2 border-[#1C1124]
                bg-[#F7E396] p-5"
            >
                <p class="font-black">
                    Pro
                </p>

                <p
                    class="mt-2 text-sm leading-relaxed
                    text-[#4d2a3a]"
                >
                    Every eligible recording can be analyzed, with
                    unlimited AI sessions and unlimited session history.
                </p>
            </div>
        </div>
    </section>

    <section class="faq-section mx-auto max-w-4xl mt-20 space-y-15">
        <div class="mx-auto max-w-3xl text-center space-y-10">
            <div
                class="mx-auto inline-flex items-center gap-2 rounded-full
                border-2 border-[#1C1124] bg-[#F7FFCD] px-4 py-2
                font-semibold text-[#4d2a3a]"
            >
                FAQs
            </div>

            <h2
                class="text-4xl font-black tracking-[-0.03em]
                sm:text-5xl"
            >
                Questions, answered.
            </h2>
        </div>

        <div class="space-y-4">
            {#each faqs as faq}
                <details
                    class="faq-item group rounded-2xl
                    border-2 border-[#1C1124] bg-white
                    shadow-[4px_4px_0px_#1C1124]"
                >
                    <summary
                        class="flex cursor-pointer list-none
                        items-center justify-between gap-5 px-6 py-5
                        text-lg font-black
                        [&::-webkit-details-marker]:hidden"
                    >
                        <span>
                            {faq.question}
                        </span>

                        <span
                            class="flex h-8 w-8 shrink-0 items-center
                            justify-center rounded-full border-2
                            border-[#1C1124] bg-[#BADF96]
                            transition-transform duration-300
                            group-open:rotate-45"
                        >
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
                                class="lucide lucide-plus-icon lucide-plus"
                            >
                                <path d="M5 12h14" />
                                <path d="M12 5v14" />
                            </svg>
                        </span>
                    </summary>

                    <div
                        class="faq-content h-0 overflow-hidden"
                    >
                        <div class="faq-content-inner px-6 pb-6">
                            <div
                                class="h-px bg-[#1C1124]/10"
                            ></div>

                            <p
                                class="pt-5 leading-relaxed
                                text-[#4d2a3a]"
                            >
                                {faq.answer}
                            </p>
                        </div>
                    </div>
                </details>
            {/each}
        </div>
    </section>
</div>