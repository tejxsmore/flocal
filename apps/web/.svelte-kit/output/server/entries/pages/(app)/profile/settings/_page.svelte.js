import "../../../../../chunks/index-server.js";
import { a as head, b as attr, i as ensure_array_like, t as attr_class, x as escape_html } from "../../../../../chunks/server.js";
import { t as page } from "../../../../../chunks/state.js";
import "../../../../../chunks/navigation.js";
import "../../../../../chunks/api.js";
import { n as PREP_TIME_OPTIONS } from "../../../../../chunks/types.js";
import "posthog-js";
//#region src/routes/(app)/profile/settings/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let initialUser = page.data.user;
		let name = initialUser.name;
		let username = initialUser.username ?? "";
		let image = initialUser.image ?? "";
		let defaultPrepTimeSeconds = initialUser.defaultPrepTimeSeconds;
		let timezone = initialUser.timezone ?? "";
		let savingProfile = false;
		let newEmail = "";
		let sendingEmailChange = false;
		function initials(value) {
			return value.split(" ").map((part) => part[0]).filter(Boolean).slice(0, 2).join("").toUpperCase();
		}
		head("65f5pd", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Flocal — Settings</title>`);
			});
		});
		$$renderer.push(`<div class="min-h-screen bg-[#BADF96] px-5 py-7 text-[#1C1124] sm:px-8 sm:py-10"><div class="mx-auto w-full max-w-5xl"><header class="flex items-center justify-between"><a href="/" class="text-2xl font-black tracking-tight transition-transform duration-200 hover:-translate-y-0.5">Flocal</a> <div class="flex items-center gap-2"><a href="/profile" class="rounded-xl border-2 border-[#1C1124] bg-[#F7FFCD] px-4 py-2 text-xs font-black transition-transform duration-200 hover:-translate-y-0.5 sm:text-sm">Profile</a> <a href="/session" class="hidden rounded-xl border-2 border-[#1C1124] bg-[#F7FFCD] px-4 py-2 text-xs font-black transition-transform duration-200 hover:-translate-y-0.5 sm:block sm:text-sm">Sessions</a></div></header> <section class="relative mt-12 overflow-hidden rounded-4xl border-2 border-[#1C1124] bg-[#F7FFCD] p-7 shadow-[9px_9px_0px_#1C1124] sm:p-10"><div class="pointer-events-none absolute -right-16 -top-20 h-48 w-48 rounded-full border-2 border-[#1C1124] bg-[#ff94d0]"></div> <div class="pointer-events-none absolute -bottom-16 right-32 h-28 w-28 rounded-full border-2 border-[#1C1124] bg-[#9FA1FF]"></div> <div class="relative max-w-2xl"><p class="text-xs font-black uppercase tracking-[0.16em] text-[#4d2a3a]/50">Account settings</p> <h1 class="mt-2 text-4xl font-black tracking-tighter sm:text-5xl">Make Flocal yours.</h1> <p class="mt-4 max-w-xl text-sm leading-relaxed text-[#4d2a3a]/65 sm:text-base">Update your profile, speaking preferences, email, and active sessions.
					Everything in one little control room.</p></div></section> <main class="mt-8 grid gap-6 pb-14 lg:grid-cols-[1.25fr_0.75fr]"><div class="flex flex-col gap-6"><form class="rounded-4xl border-2 border-[#1C1124] bg-[#F7FFCD] p-6 shadow-[7px_7px_0px_#1C1124] sm:p-8"><div class="flex items-start justify-between gap-4"><div><p class="text-xs font-black uppercase tracking-[0.15em] text-[#4d2a3a]/50">01</p> <h2 class="mt-1 text-2xl font-black">Profile</h2> <p class="mt-1 text-sm text-[#4d2a3a]/55">Your public-facing Flocal identity.</p></div> <div class="hidden h-12 w-12 items-center justify-center rounded-2xl border-2 border-[#1C1124] bg-[#9FA1FF] text-lg font-black sm:flex">✦</div></div> <div class="mt-7 flex flex-col gap-4 rounded-3xl border-2 border-[#1C1124] bg-white p-4 sm:flex-row sm:items-center">`);
		if (image) {
			$$renderer.push("<!--[0-->");
			$$renderer.push(`<img${attr("src", image)}${attr("alt", name)} class="h-20 w-20 shrink-0 rounded-2xl border-2 border-[#1C1124] object-cover"/>`);
		} else {
			$$renderer.push("<!--[-1-->");
			$$renderer.push(`<div class="flex h-20 w-20 shrink-0 items-center justify-center rounded-2xl border-2 border-[#1C1124] bg-[#ff94d0] text-2xl font-black">${escape_html(initials(name || "U"))}</div>`);
		}
		$$renderer.push(`<!--]--> <div class="min-w-0 flex-1"><label for="image" class="text-xs font-black uppercase tracking-wide text-[#4d2a3a]/50">Avatar URL</label> <input id="image" type="url"${attr("value", image)} placeholder="https://…" class="mt-2 w-full rounded-xl border-2 border-[#1C1124]/15 bg-[#F7FFCD] px-4 py-3 text-sm outline-none transition-colors focus:border-[#1C1124]"/> <p class="mt-2 text-[11px] text-[#4d2a3a]/45">Leave empty to use your initials.</p></div></div> <div class="mt-5 grid gap-5 sm:grid-cols-2"><div class="flex flex-col gap-2"><label for="name" class="text-xs font-black uppercase tracking-wide text-[#4d2a3a]/50">Name</label> <input id="name" type="text" required=""${attr("value", name)} class="rounded-xl border-2 border-[#1C1124]/15 bg-white px-4 py-3 text-sm outline-none transition-colors focus:border-[#1C1124]"/></div> <div class="flex flex-col gap-2"><label for="username" class="text-xs font-black uppercase tracking-wide text-[#4d2a3a]/50">Username</label> <input id="username" type="text"${attr("value", username)} placeholder="@username" class="rounded-xl border-2 border-[#1C1124]/15 bg-white px-4 py-3 text-sm outline-none transition-colors focus:border-[#1C1124]"/></div></div> <div class="mt-6 rounded-3xl border-2 border-[#1C1124] bg-[#E6EABF] p-5"><div><p class="text-xs font-black uppercase tracking-[0.14em] text-[#4d2a3a]/50">Speaking preference</p> <h3 class="mt-1 text-lg font-black">Default prep time</h3> <p class="mt-1 text-xs text-[#4d2a3a]/55">How long you get to think before speaking.</p></div> <div class="mt-4 flex flex-wrap gap-2"><!--[-->`);
		const each_array = ensure_array_like(PREP_TIME_OPTIONS);
		for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
			let option = each_array[$$index];
			$$renderer.push(`<button type="button"${attr_class(`cursor-pointer rounded-xl border-2 px-4 py-2.5 text-sm font-black transition-transform duration-200 hover:-translate-y-0.5 ${defaultPrepTimeSeconds === option.value ? "border-[#1C1124] bg-[#1C1124] text-[#F7FFCD]" : "border-[#1C1124]/20 bg-[#F7FFCD] text-[#1C1124]"}`)}>${escape_html(option.label)}</button>`);
		}
		$$renderer.push(`<!--]--></div></div> <div class="mt-5"><label for="timezone" class="text-xs font-black uppercase tracking-wide text-[#4d2a3a]/50">Timezone</label> <input id="timezone" type="text"${attr("value", timezone)} placeholder="e.g. Asia/Kolkata" class="mt-2 w-full rounded-xl border-2 border-[#1C1124]/15 bg-white px-4 py-3 text-sm outline-none transition-colors focus:border-[#1C1124]"/></div> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> <button type="submit"${attr("disabled", savingProfile, true)} class="mt-6 cursor-pointer rounded-xl border-2 border-[#1C1124] bg-[#ff94d0] px-6 py-3 text-sm font-black shadow-[4px_4px_0px_#1C1124] transition-all duration-200 hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-[2px_2px_0px_#1C1124] disabled:cursor-not-allowed disabled:opacity-50">${escape_html("Save changes")}</button></form> <form class="rounded-4xl border-2 border-[#1C1124] bg-[#9FA1FF] p-6 shadow-[7px_7px_0px_#1C1124] sm:p-8"><div class="flex items-start justify-between gap-4"><div><p class="text-xs font-black uppercase tracking-[0.15em] text-[#1C1124]/50">02</p> <h2 class="mt-1 text-2xl font-black">Email address</h2> <p class="mt-1 text-sm text-[#1C1124]/60">Change where Flocal sends account emails.</p></div> <div class="hidden h-12 w-12 items-center justify-center rounded-2xl border-2 border-[#1C1124] bg-[#F7E396] text-xl sm:flex">@</div></div> <div class="mt-6 rounded-2xl border-2 border-[#1C1124] bg-[#F7FFCD] px-4 py-3"><p class="text-[10px] font-black uppercase tracking-[0.14em] text-[#4d2a3a]/45">Current email</p> <p class="mt-1 truncate text-sm font-black">${escape_html(initialUser.email)}</p></div> <div class="mt-5"><label for="newEmail" class="text-xs font-black uppercase tracking-wide text-[#1C1124]/55">New email</label> <input id="newEmail" type="email" required=""${attr("value", newEmail)} placeholder="you@example.com" class="mt-2 w-full rounded-xl border-2 border-[#1C1124]/20 bg-white px-4 py-3 text-sm outline-none transition-colors focus:border-[#1C1124]"/></div> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> `);
		$$renderer.push("<!--[-1-->");
		$$renderer.push(`<!--]--> <button type="submit"${attr("disabled", sendingEmailChange, true)} class="mt-6 cursor-pointer rounded-xl border-2 border-[#1C1124] bg-[#1C1124] px-6 py-3 text-sm font-black text-[#F7FFCD] shadow-[4px_4px_0px_#F7FFCD] transition-all duration-200 hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-[2px_2px_0px_#F7FFCD] disabled:cursor-not-allowed disabled:opacity-50">${escape_html("Send verification email")}</button></form></div> <div class="flex flex-col gap-6"><section class="rounded-4xl border-2 border-[#1C1124] bg-[#F7FFCD] p-6 shadow-[7px_7px_0px_#1C1124] sm:p-7"><div class="flex items-start justify-between gap-4"><div><p class="text-xs font-black uppercase tracking-[0.15em] text-[#4d2a3a]/50">03</p> <h2 class="mt-1 text-2xl font-black">Active sessions</h2> <p class="mt-1 text-sm leading-relaxed text-[#4d2a3a]/55">Devices currently signed into your account.</p></div> <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border-2 border-[#1C1124] bg-[#BADF96] font-black">↗</div></div> <div class="mt-6">`);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`<div class="rounded-2xl border-2 border-[#1C1124]/10 bg-white p-5"><p class="text-sm font-bold text-[#4d2a3a]/45">Loading sessions…</p></div>`);
		$$renderer.push(`<!--]--></div></section> <section class="rounded-4xl border-2 border-[#1C1124] bg-[#F7E396] p-6 shadow-[7px_7px_0px_#1C1124] sm:p-7"><p class="text-xs font-black uppercase tracking-[0.15em] text-[#4d2a3a]/50">Account</p> <h2 class="mt-2 text-2xl font-black leading-tight">Your account, <br/> your rules.</h2> <p class="mt-3 text-sm leading-relaxed text-[#4d2a3a]/60">Need to check something else? Head back to your profile or jump
						straight into another speaking session.</p> <div class="mt-6 flex flex-col gap-2"><a href="/profile" class="rounded-xl border-2 border-[#1C1124] bg-[#F7FFCD] px-5 py-3 text-center text-sm font-black transition-transform duration-200 hover:-translate-y-0.5">View profile</a> <a href="/" class="rounded-xl border-2 border-[#1C1124] bg-[#1C1124] px-5 py-3 text-center text-sm font-black text-[#F7FFCD] transition-transform duration-200 hover:-translate-y-0.5">Start a session →</a></div></section> <section class="rounded-4xl border-2 border-[#A15668] bg-[#FBE4E8] p-6 shadow-[7px_7px_0px_#A15668] sm:p-7"><div class="flex items-start gap-4"><div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border-2 border-[#A15668] bg-white font-black text-[#A15668]">!</div> <div><p class="text-xs font-black uppercase tracking-[0.15em] text-[#A15668]/70">Danger zone</p> <h2 class="mt-1 text-2xl font-black text-[#7c3e50]">Delete account</h2> <p class="mt-2 text-sm leading-relaxed text-[#7c3e50]/65">This permanently deactivates your account and signs you out
								everywhere.</p></div></div> `);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`<button type="button" class="mt-6 cursor-pointer rounded-xl border-2 border-[#A15668] bg-white px-5 py-3 text-sm font-black text-[#A15668] transition-colors hover:bg-[#A15668] hover:text-white">Delete account</button>`);
		$$renderer.push(`<!--]--></section></div></main></div></div>`);
	});
}
//#endregion
export { _page as default };
