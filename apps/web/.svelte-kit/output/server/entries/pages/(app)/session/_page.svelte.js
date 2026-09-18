import { a as head, b as attr, i as ensure_array_like, t as attr_class, x as escape_html } from "../../../../chunks/server.js";
import { i as STATUS_OPTIONS, r as SORT_OPTIONS, t as FORMAT_OPTIONS } from "../../../../chunks/types.js";
import "posthog-js";
//#region src/routes/(app)/session/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let searchInput = "";
		let statuses = [];
		let formats = [];
		let sort = "newest";
		let dateFrom = "";
		let dateTo = "";
		head("1n2hzbi", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Flocal — Sessions</title>`);
			});
		});
		$$renderer.push(`<div class="min-h-screen bg-[#BADF96] px-5 py-8 text-[#1C1124] sm:px-8 sm:py-12 lg:px-12"><div class="mx-auto flex w-full max-w-5xl flex-col gap-8"><header class="flex items-end justify-between gap-5"><div><a href="/profile" class="mb-4 inline-flex items-center gap-1.5 text-sm font-bold text-[#4d2a3a] transition-transform duration-200 hover:-translate-x-1"><span class="text-lg">←</span> Profile</a> <div class="inline-flex rounded-full border-2 border-[#1C1124] bg-[#F7FFCD] px-3.5 py-1.5 text-xs font-black uppercase tracking-[0.12em]">Practice history</div> <h1 class="mt-4 text-4xl font-black tracking-[-0.045em] sm:text-5xl md:text-6xl">Sessions</h1> <p class="mt-3 max-w-xl text-sm leading-relaxed text-[#4d2a3a] sm:text-base">Review what you've practised, find old topics, and see how your speaking is
					progressing.</p></div></header> <section class="rounded-3xl border-2 border-[#1C1124] bg-[#F7FFCD] p-5 shadow-[8px_8px_0px_#1C1124] sm:p-6"><div class="flex flex-col gap-5"><div class="relative"><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-[#4d2a3a]/60"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.35-4.35"></path></svg> <input type="text" placeholder="Search by topic…"${attr("value", searchInput)} class="w-full rounded-xl border-2 border-[#1C1124] bg-white px-11 py-3 text-sm font-semibold text-[#1C1124] placeholder:text-[#4d2a3a]/40 focus:outline-none focus:ring-0"/></div> <div><p class="mb-2.5 text-xs font-black uppercase tracking-[0.14em] text-[#4d2a3a]">Status</p> <div class="flex flex-wrap gap-2"><!--[-->`);
		const each_array = ensure_array_like(STATUS_OPTIONS);
		for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
			let opt = each_array[$$index];
			$$renderer.push(`<button type="button"${attr_class(`cursor-pointer rounded-full border-2 border-[#1C1124] px-3.5 py-2 text-xs font-bold transition-transform duration-200 hover:-translate-y-0.5 ${statuses.includes(opt.value) ? "bg-[#ff94d0]" : "bg-white"}`)}>${escape_html(opt.label)}</button>`);
		}
		$$renderer.push(`<!--]--></div></div> <div><p class="mb-2.5 text-xs font-black uppercase tracking-[0.14em] text-[#4d2a3a]">Format</p> <div class="flex flex-wrap gap-2"><!--[-->`);
		const each_array_1 = ensure_array_like(FORMAT_OPTIONS);
		for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
			let opt = each_array_1[$$index_1];
			$$renderer.push(`<button type="button"${attr_class(`cursor-pointer rounded-full border-2 border-[#1C1124] px-3.5 py-2 text-xs font-bold transition-transform duration-200 hover:-translate-y-0.5 ${formats.includes(opt.value) ? "bg-[#9FA1FF]" : "bg-white"}`)}>${escape_html(opt.icon)} ${escape_html(opt.label)}</button>`);
		}
		$$renderer.push(`<!--]--></div></div> <div><p class="mb-2.5 text-xs font-black uppercase tracking-[0.14em] text-[#4d2a3a]">Date range</p> <div class="grid grid-cols-[1fr_auto_1fr] items-center gap-2"><input type="date"${attr("value", dateFrom)} class="min-w-0 rounded-xl border-2 border-[#1C1124] bg-white px-3 py-3 text-xs font-semibold text-[#1C1124] focus:outline-none"/> <span class="text-xs font-black text-[#4d2a3a]">TO</span> <input type="date"${attr("value", dateTo)} class="min-w-0 rounded-xl border-2 border-[#1C1124] bg-white px-3 py-3 text-xs font-semibold text-[#1C1124] focus:outline-none"/></div></div> <div class="flex flex-wrap items-center justify-between gap-3 border-t-2 border-[#1C1124] pt-5">`);
		$$renderer.select({
			value: sort,
			class: "cursor-pointer rounded-xl border-2 border-[#1C1124] bg-white px-4 py-3 text-xs font-bold text-[#1C1124] focus:outline-none"
		}, ($$renderer) => {
			$$renderer.push(`<!--[-->`);
			const each_array_2 = ensure_array_like(SORT_OPTIONS);
			for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
				let opt = each_array_2[$$index_2];
				$$renderer.option({ value: opt.value }, ($$renderer) => {
					$$renderer.push(`${escape_html(opt.label)}`);
				});
			}
			$$renderer.push(`<!--]-->`);
		});
		$$renderer.push(` <button type="button" class="cursor-pointer rounded-xl border-2 border-[#1C1124] bg-[#BADF96] px-4 py-3 text-xs font-black transition-transform duration-200 hover:-translate-y-0.5">Clear filters</button></div></div></section> <section class="flex flex-col gap-4"><div class="flex items-center justify-between"><div><p class="text-xs font-black uppercase tracking-[0.14em] text-[#4d2a3a]">Your sessions</p> <p class="mt-1 text-sm font-semibold text-[#4d2a3a]/70">`);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`Loading your practice history…`);
		$$renderer.push(`<!--]--></p></div></div> `);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`<div class="rounded-3xl border-2 border-[#1C1124] bg-[#F7FFCD] px-6 py-12 text-center shadow-[6px_6px_0px_#1C1124]"><div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full border-2 border-[#1C1124] bg-[#9FA1FF] text-xl">…</div> <p class="text-sm font-bold">Loading sessions…</p></div>`);
		$$renderer.push(`<!--]--></section></div></div>`);
	});
}
//#endregion
export { _page as default };
