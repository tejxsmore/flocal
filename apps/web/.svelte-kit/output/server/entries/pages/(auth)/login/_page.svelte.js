import { a as head, b as attr, r as derived, t as attr_class, x as escape_html } from "../../../../chunks/server.js";
import "../../../../chunks/state.js";
import "../../../../chunks/navigation.js";
import { t as PUBLIC_API_BASE_URL } from "../../../../chunks/public.js";
//#region src/routes/(auth)/login/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let { data, form } = $$props;
		let email = "";
		derived(() => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email));
		function oauthUrl(provider) {
			return `${PUBLIC_API_BASE_URL}/api/v1/auth/${provider}`;
		}
		head("8k30lk", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Flocal - Sign in</title>`);
			});
		});
		$$renderer.push(`<div class="flex min-h-screen items-center justify-center bg-[#BADF96] px-5 py-12 text-[#1C1124] sm:px-8"><div class="w-full max-w-sm"><div class="text-center"><h1 class="text-5xl font-black leading-[0.95] tracking-[-0.04em] sm:text-6xl">Sign in</h1> <p class="mx-auto mt-5 max-w-xs text-base leading-relaxed text-[#4d2a3a] sm:text-lg">A small step today toward becoming a more confident speaker.</p></div> <div class="mt-12 flex w-full flex-col gap-3"><a${attr("href", oauthUrl("google"))} class="flex h-13.5 w-full cursor-pointer items-center justify-center gap-3 rounded-xl border-2 border-[#1C1124] bg-[#F7FFCD] px-5 font-bold text-[#1C1124] transition-transform duration-200 hover:-translate-y-1"><svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 48 48" class="shrink-0"><path fill="#FFC107" d="M43.611,20.083H42V20H24v8h11.303c-1.649,4.657-6.08,8-11.303,8c-6.627,0-12-5.373-12-12c0-6.627,5.373-12,12-12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C12.955,4,4,12.955,4,24c0,11.045,8.955,20,20,20c11.045,0,20-8.955,20-20C44,22.659,43.862,21.35,43.611,20.083z"></path><path fill="#FF3D00" d="M6.306,14.691l6.571,4.819C14.655,15.108,18.961,12,24,12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C16.318,4,9.656,8.337,6.306,14.691z"></path><path fill="#4CAF50" d="M24,44c5.166,0,9.86-1.977,13.409-5.192l-6.19-5.238C29.211,35.091,26.715,36,24,36c-5.202,0-9.619-3.317-11.283-7.946l-6.522,5.025C9.505,39.556,16.227,44,24,44z"></path><path fill="#1976D2" d="M43.611,20.083H42V20H24v8h11.303c-.792,2.237-2.231,4.166-4.087,5.571l.003-.002l6.19,5.238C36.971,39.205,44,34,44,24C44,22.659,43.862,21.35,43.611,20.083z"></path></svg> Continue with Google</a> <a${attr("href", oauthUrl("github"))} class="flex h-13.5 w-full cursor-pointer items-center justify-center gap-3 rounded-xl border-2 border-[#1C1124] bg-[#1C1124] px-5 font-bold text-white transition-transform duration-200 hover:-translate-y-1"><svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 32 32" class="shrink-0"><g fill="#ffffff"><path d="M16,2.345c7.735,0,14,6.265,14,14-.002,6.015-3.839,11.359-9.537,13.282-.7,.14-.963-.298-.963-.665,0-.473,.018-1.978,.018-3.85,0-1.312-.437-2.152-.945-2.59,3.115-.35,6.388-1.54,6.388-6.912,0-1.54-.543-2.783-1.435-3.762,.14-.35,.63-1.785-.14-3.71,0,0-1.173-.385-3.85,1.435-1.12-.315-2.31-.472-3.5-.472s-2.38,.157-3.5,.472c-2.677-1.802-3.85-1.435-3.85-1.435-.77,1.925-.28,3.36-.14,3.71-.892,.98-1.435,2.24-1.435,3.762,0,5.355,3.255,6.563,6.37,6.913-.403,.35-.77,.963-.893,1.872-.805,.368-2.818,.963-4.077-1.155-.263-.42-1.05-1.452-2.152-1.435-1.173,.018-.472,.665,.017,.927,.595,.332,1.277,1.575,1.435,1.978,.28,.787,1.19,2.293,4.707,1.645,0,1.173,.018,2.275,.018,2.607,0,.368-.263,.787-.963,.665-5.719-1.904-9.576-7.255-9.573-13.283,0-7.735,6.265-14,14-14Z"></path></g></svg> Continue with GitHub</a> <div class="flex items-center gap-4 py-3"><div class="h-px flex-1 bg-[#1C1124]/15"></div> <span class="text-xs font-bold uppercase tracking-[0.14em] text-[#4d2a3a]/55">or</span> <div class="h-px flex-1 bg-[#1C1124]/15"></div></div> <form method="POST" action="?/requestMagicLink" class="flex flex-col gap-3"><div class="relative w-full"><input id="email" name="email" type="email" autocomplete="email" placeholder="you@example.com"${attr("value", email)} required=""${attr_class(`h-13.5 w-full rounded-xl border-2 bg-white px-10 text-center font-medium text-[#1C1124] placeholder:text-[#4d2a3a]/40 focus:outline-none transition-colors duration-200 border-[#4d2a3a]`)}/> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div> <button type="submit" class="flex h-13.5 w-full cursor-pointer items-center justify-center gap-3 rounded-xl border-2 border-[#4d2a3a] bg-[#ff94d0] px-5 font-bold text-[#4d2a3a] transition-transform duration-200 hover:-translate-y-1">`);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="shrink-0"><path d="m22 7-8.991 5.727a2 2 0 0 1-2.009 0L2 7"></path><rect x="2" y="4" width="20" height="16" rx="3.25"></rect></svg>${escape_html("Continue with Email")}`);
		$$renderer.push(`<!--]--></button> <p${attr_class(`min-h-5 text-center text-xs font-semibold leading-relaxed invisible`)}>${escape_html("\xA0")}</p></form></div></div></div>`);
	});
}
//#endregion
export { _page as default };
