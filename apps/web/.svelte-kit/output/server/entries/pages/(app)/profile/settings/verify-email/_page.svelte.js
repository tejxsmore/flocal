import "../../../../../../chunks/index-server.js";
import { a as head } from "../../../../../../chunks/server.js";
import "../../../../../../chunks/state.js";
import "../../../../../../chunks/navigation.js";
import "../../../../../../chunks/api.js";
//#region src/routes/(app)/profile/settings/verify-email/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		head("s9frvu", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Flocal — Verify email</title>`);
			});
		});
		$$renderer.push(`<div class="min-h-screen bg-[#BADF96] px-5 py-7 text-[#1C1124] svelte-s9frvu"><div class="mx-auto flex min-h-[calc(100vh-3.5rem)] w-full max-w-5xl flex-col svelte-s9frvu"><header class="flex items-center justify-between svelte-s9frvu"><a href="/" class="text-2xl font-black tracking-tight transition-transform duration-200 hover:-translate-y-0.5 svelte-s9frvu">Flocal</a> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></header> <main class="flex flex-1 items-center justify-center py-16 svelte-s9frvu"><div class="w-full max-w-xl svelte-s9frvu">`);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`<section class="relative overflow-hidden rounded-4xl border-2 border-[#1C1124] bg-[#F7FFCD] p-8 text-center shadow-[9px_9px_0px_#1C1124] sm:p-12 svelte-s9frvu"><div class="pointer-events-none absolute -right-14 -top-14 h-36 w-36 rounded-full border-2 border-[#1C1124] bg-[#9FA1FF] svelte-s9frvu"></div> <div class="pointer-events-none absolute -bottom-16 -left-12 h-32 w-32 rounded-full border-2 border-[#1C1124] bg-[#ff94d0] svelte-s9frvu"></div> <div class="relative svelte-s9frvu"><div class="mx-auto flex h-20 w-20 items-center justify-center rounded-3xl border-2 border-[#1C1124] bg-[#9FA1FF] text-3xl font-black svelte-s9frvu"><span class="animate-pulse svelte-s9frvu">✦</span></div> <p class="mt-8 text-xs font-black uppercase tracking-[0.16em] text-[#4d2a3a]/45 svelte-s9frvu">Email verification</p> <h1 class="mt-2 text-4xl font-black tracking-tighter sm:text-5xl svelte-s9frvu">Checking your email.</h1> <p class="mx-auto mt-4 max-w-md text-sm leading-relaxed text-[#4d2a3a]/60 sm:text-base svelte-s9frvu">We're verifying your new email address. This should only take a moment.</p> <div class="mx-auto mt-8 h-2 w-36 overflow-hidden rounded-full border-2 border-[#1C1124] bg-white svelte-s9frvu"><div class="h-full w-1/2 animate-[loading_1.2s_ease-in-out_infinite] rounded-full bg-[#ff94d0] svelte-s9frvu"></div></div></div></section>`);
		$$renderer.push(`<!--]--></div></main> <footer class="pb-4 text-center svelte-s9frvu"><p class="text-[11px] font-bold text-[#4d2a3a]/35 svelte-s9frvu">Flocal · Speak better, one session at a time.</p></footer></div></div>`);
	});
}
//#endregion
export { _page as default };
