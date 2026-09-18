import "../../../chunks/index-server.js";
import { a as head, b as attr, i as ensure_array_like, t as attr_class } from "../../../chunks/server.js";
import { t as page } from "../../../chunks/state.js";
import "../../../chunks/navigation.js";
import "posthog-js";
//#region src/routes/onboarding/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let user = page.data.user;
		let step = 0;
		let maxStepVisited = 0;
		const totalSteps = 5;
		let name = user.name ?? "";
		user.username;
		user.image;
		function canContinue() {
			return name.trim().length > 0;
		}
		head("fpvdp2", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Flocal — Welcome</title>`);
			});
		});
		$$renderer.push(`<div class="min-h-screen bg-[#BADF96] px-5 sm:px-10 md:px-20 py-16 text-[#1C1124]"><div class="mx-auto flex max-w-md flex-col items-center gap-8"><div class="flex gap-1.5"><!--[-->`);
		const each_array = ensure_array_like(Array(totalSteps));
		for (let i = 0, $$length = each_array.length; i < $$length; i++) {
			each_array[i];
			$$renderer.push(`<button type="button"${attr("aria-label", `Go to step ${i + 1}`)}${attr("disabled", i > maxStepVisited, true)}${attr_class(`h-2 w-8 rounded-full border-2 border-[#1C1124] transition-colors duration-200 ${i <= step ? "bg-[#ff94d0]" : "bg-white"} ${i <= maxStepVisited ? "cursor-pointer" : "cursor-default"}`)}></button>`);
		}
		$$renderer.push(`<!--]--></div> <form class="w-full rounded-4xl border-2 border-[#1C1124] bg-white p-7 shadow-[8px_8px_0px_#1C1124] sm:p-9">`);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`<div class="flex flex-col gap-6"><div class="flex flex-col gap-1"><h2 class="text-2xl font-black">What should we call you?</h2> <p class="text-[#4d2a3a]">This is how you'll appear across Flocal.</p></div> <div class="relative w-full"><input id="name" type="text" autocomplete="name" required=""${attr("value", name)} class="h-13.5 w-full rounded-xl border-2 border-[#4d2a3a] bg-white px-4 pr-10 text-base font-medium outline-none focus:bg-[#F7FFCD]"/> `);
		if (name.length > 0) {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<button type="button" aria-label="Clear name" class="absolute right-4 top-1/2 -translate-y-1/2 cursor-pointer text-[#1C1124]/55 transition-colors duration-200 hover:text-[#1C1124]"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"></path><path d="m6 6 12 12"></path></svg></button>`);
		} else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div> <button type="submit"${attr("disabled", !canContinue(), true)} class="h-13.5 w-full cursor-pointer rounded-xl border-2 border-[#4d2a3a] bg-[#ff94d0] px-6 font-bold text-[#4d2a3a] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:translate-y-0">Continue</button></div>`);
		$$renderer.push(`<!--]--></form></div></div>`);
	});
}
//#endregion
export { _page as default };
