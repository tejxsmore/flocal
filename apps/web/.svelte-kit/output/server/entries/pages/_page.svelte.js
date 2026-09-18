import "../../chunks/index-server.js";
import { a as head, b as attr, i as ensure_array_like, n as attr_style, r as derived, s as stringify, t as attr_class, x as escape_html } from "../../chunks/server.js";
import "../../chunks/navigation.js";
import "../../chunks/api.js";
import { t as Navbar } from "../../chunks/Navbar.js";
import { t as FORMAT_OPTIONS } from "../../chunks/types.js";
import "posthog-js";
//#region src/routes/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let categories = [];
		let selectedCategoryId = null;
		let selectedFormat = null;
		let spinning = false;
		derived(() => categories.find((c) => c.id === void 0) ?? null);
		derived(() => FORMAT_OPTIONS.find((f) => f.value === void 0) ?? null);
		head("1uha8ag", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Flocal — Practice</title>`);
			});
		});
		Navbar($$renderer, {});
		$$renderer.push(`<!----> <div class="min-h-screen bg-[#BADF96] px-5 py-10 text-[#1C1124] sm:px-8 sm:py-14 lg:px-12"><div class="mx-auto flex w-full max-w-5xl flex-col gap-8"><div class="mx-auto w-full max-w-3xl text-center"><div class="mx-auto inline-flex items-center gap-2 rounded-full border-2 border-[#1C1124] bg-[#F7FFCD] px-3.5 py-1.5 text-xs font-bold text-[#4d2a3a] sm:px-4 sm:py-2 sm:text-sm"><span class="h-2.5 w-2.5 rounded-full bg-[#4d2a3a]"></span> One minute. One topic. Your voice.</div> <h1 class="mt-5 text-4xl font-black leading-[0.95] tracking-[-0.045em] sm:text-5xl md:text-6xl">What do you think?</h1> <p class="mx-auto mt-4 max-w-xl text-sm leading-relaxed text-[#4d2a3a] sm:text-base">Choose what you want to practise, spin for a topic, and put your thoughts into words.</p></div> <div class="mx-auto flex w-full max-w-3xl flex-col gap-6"><div class="rounded-3xl border-2 border-[#1C1124] bg-[#F7FFCD] p-5 shadow-[6px_6px_0px_#1C1124] sm:p-6"><div class="grid gap-6 md:grid-cols-2"><div class="flex flex-col gap-3"><div class="flex items-center justify-between"><p class="text-xs font-black uppercase tracking-[0.14em] text-[#4d2a3a]">Format</p> <span class="text-xs font-semibold text-[#4d2a3a]/70">Optional</span></div> <div class="flex flex-wrap gap-2"><button type="button"${attr_class(`cursor-pointer rounded-full border-2 border-[#1C1124] px-3.5 py-2 text-sm font-bold transition-transform duration-200 hover:-translate-y-0.5 bg-[#ff94d0]`)}>Any format</button> <!--[-->`);
		const each_array = ensure_array_like(FORMAT_OPTIONS);
		for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
			let option = each_array[$$index];
			$$renderer.push(`<button type="button"${attr_class(`cursor-pointer rounded-full border-2 border-[#1C1124] px-3.5 py-2 text-sm font-bold transition-transform duration-200 hover:-translate-y-0.5 ${selectedFormat === option.value ? "bg-[#ff94d0]" : "bg-white"}`)}>${escape_html(option.icon)} ${escape_html(option.label)}</button>`);
		}
		$$renderer.push(`<!--]--></div></div> <div class="flex flex-col gap-3"><div class="flex items-center justify-between"><p class="text-xs font-black uppercase tracking-[0.14em] text-[#4d2a3a]">Category</p> <span class="text-xs font-semibold text-[#4d2a3a]/70">Optional</span></div> <div class="flex flex-wrap gap-2"><button type="button"${attr_class(`cursor-pointer rounded-full border-2 border-[#1C1124] px-3.5 py-2 text-sm font-bold transition-transform duration-200 hover:-translate-y-0.5 bg-[#9FA1FF]`)}>Any category</button> <!--[-->`);
		const each_array_1 = ensure_array_like(categories);
		for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
			let category = each_array_1[$$index_1];
			$$renderer.push(`<button type="button"${attr_class(`cursor-pointer rounded-full border-2 border-[#1C1124] px-3.5 py-2 text-sm font-bold transition-transform duration-200 hover:-translate-y-0.5 ${selectedCategoryId === category.id ? "bg-[#9FA1FF]" : "bg-white"}`)}>${escape_html(category.icon ?? "")} ${escape_html(category.name)}</button>`);
		}
		$$renderer.push(`<!--]--></div></div></div></div> <div class="relative overflow-hidden rounded-3xl border-2 border-[#1C1124] bg-[#F7FFCD] shadow-[8px_8px_0px_#1C1124] sm:shadow-[10px_10px_0px_#1C1124]"><div class="flex items-center justify-between border-b-2 border-[#1C1124] px-5 py-4 sm:px-6"><div><p class="text-xs font-black uppercase tracking-[0.16em] text-[#4d2a3a]">Your next topic</p> <p class="mt-1 text-sm font-semibold text-[#4d2a3a]/70">${escape_html("Spin the reel to get started.")}</p></div> <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border-2 border-[#1C1124] bg-[#BADF96] text-lg font-black">?</div></div> <div class="reel-shell relative w-full overflow-hidden px-5 sm:px-10"${attr_style(`--reel-item-height: ${stringify(112)}px;`)}><div class="reel-highlight pointer-events-none absolute inset-x-0 svelte-1uha8ag"></div> <div class="reel-window overflow-hidden svelte-1uha8ag"></div> <div class="reel-fade reel-fade--top pointer-events-none absolute inset-x-0 top-0 svelte-1uha8ag"></div> <div class="reel-fade reel-fade--bottom pointer-events-none absolute inset-x-0 bottom-0 svelte-1uha8ag"></div></div> <div class="border-t-2 border-[#1C1124] px-5 py-5 sm:px-6"><button type="button"${attr("disabled", spinning, true)} class="group flex min-h-13 w-full cursor-pointer items-center justify-center gap-3 rounded-xl border-2 border-[#1C1124] bg-[#1C1124] px-6 py-3 text-base font-bold text-[#F7FFCD] transition-transform duration-200 hover:-translate-y-1 disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"><span>${escape_html("Spin a topic")}</span> <span class="text-xl transition-transform duration-200 group-hover:translate-x-1">→</span></button></div></div> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div> <div class="mx-auto flex max-w-xl flex-col items-center text-center"><p class="text-sm font-bold text-[#4d2a3a]">One minute is enough to start.</p> <p class="mt-1 text-xs text-[#4d2a3a]/70">Discover a thought. Form an opinion. Say it out loud.</p></div></div></div>`);
	});
}
//#endregion
export { _page as default };
