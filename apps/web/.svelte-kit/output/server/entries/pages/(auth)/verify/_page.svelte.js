import { a as head } from "../../../../chunks/server.js";
//#region src/routes/(auth)/verify/+page.svelte
function _page($$renderer) {
	head("f20z5s", $$renderer, ($$renderer) => {
		$$renderer.title(($$renderer) => {
			$$renderer.push(`<title>Flocal - Sign in</title>`);
		});
	});
	$$renderer.push(`<div class="flex min-h-screen items-center justify-center bg-[#BADF96] px-5 py-12 text-[#1C1124] sm:px-8"><div class="w-full max-w-sm"><div class="text-center"><h1 class="text-5xl font-black leading-[0.95] tracking-[-0.04em] sm:text-6xl">Signing you in</h1> <p class="mx-auto mt-5 max-w-xs text-base leading-relaxed text-[#4d2a3a] sm:text-lg">Just a moment while we finish signing you in.</p></div> <div class="mt-12 flex w-full flex-col items-center">`);
	$$renderer.push("<!--[0-->");
	$$renderer.push(`<div class="h-8 w-8 animate-spin rounded-full border-[3px] border-[#4d2a3a]/20 border-t-[#1C1124]"></div>`);
	$$renderer.push(`<!--]--></div></div></div>`);
}
//#endregion
export { _page as default };
