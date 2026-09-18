import { b as attr, r as derived, x as escape_html } from "./server.js";
import { t as page } from "./state.js";
import "./navigation.js";
//#region src/lib/components/Navbar.svelte
function Navbar($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let user = derived(() => page.data.user);
		let open = false;
		function initials(name) {
			return name.split(" ").map((part) => part[0]).filter(Boolean).slice(0, 2).join("").toUpperCase();
		}
		$$renderer.push(`<nav class="relative z-50 w-full px-5 py-5 sm:px-6"><div class="mx-auto flex max-w-7xl items-center justify-between"><a href="/" aria-label="Flocal home" class="text-2xl font-black tracking-tight text-[#1C1124] sm:text-3xl">Flocal</a> <div class="flex items-center gap-5 sm:gap-7"><a href="/session" class="rounded-lg px-2 py-2 text-sm font-bold text-[#1C1124] transition-transform duration-200 hover:-translate-y-0.5 sm:text-base">Session</a> `);
		if (user()) {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<div class="relative"><button type="button" aria-label="Open account menu"${attr("aria-expanded", open)} class="flex h-10 w-10 cursor-pointer items-center justify-center overflow-hidden rounded-full border-2 border-[#1C1124] bg-[#F7FFCD] text-xs font-black text-[#1C1124] transition-transform duration-200 hover:-translate-y-0.5">`);
			if (user().image) {
				$$renderer.push("<!--[0-->");
				$$renderer.push(`<img${attr("src", user().image)}${attr("alt", user().name)} class="h-full w-full object-cover"/>`);
			} else {
				$$renderer.push("<!--[-1-->");
				$$renderer.push(`${escape_html(initials(user().name))}`);
			}
			$$renderer.push(`<!--]--></button> `);
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--></div>`);
		} else $$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--></div></div></nav>`);
	});
}
//#endregion
export { Navbar as t };
