import { n as onDestroy } from "../../../../../chunks/index-server.js";
import { a as head, r as derived } from "../../../../../chunks/server.js";
import { t as page } from "../../../../../chunks/state.js";
import "../../../../../chunks/api.js";
import "posthog-js";
//#region src/routes/(app)/session/[id]/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		derived(() => page.params.id);
		derived(() => []);
		derived(() => false);
		let mediaRecorder = null;
		let mediaStream = null;
		let prepTimer = null;
		let speakTimer = null;
		let pollTimer = null;
		function clearTimers() {
			if (prepTimer) clearInterval(prepTimer);
			if (speakTimer) clearInterval(speakTimer);
			if (pollTimer) clearInterval(pollTimer);
			prepTimer = null;
			speakTimer = null;
			pollTimer = null;
		}
		function stopMedia() {
			if (mediaRecorder && mediaRecorder.state !== "inactive") mediaRecorder.stop();
			mediaRecorder = null;
			if (mediaStream) for (const track of mediaStream.getTracks()) track.stop();
			mediaStream = null;
		}
		onDestroy(() => {
			clearTimers();
			stopMedia();
		});
		head("6sjbkm", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Flocal — Session</title>`);
			});
		});
		$$renderer.push(`<div class="min-h-screen bg-[#BADF96] px-5 py-8 text-[#1C1124] sm:px-8 sm:py-12"><div class="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-5xl flex-col"><header class="flex items-center justify-between"><a href="/" class="text-2xl font-black tracking-tight text-[#1C1124] transition-transform duration-200 hover:-translate-y-0.5">Flocal</a> <a href="/session" class="rounded-xl border-2 border-[#1C1124] bg-[#F7FFCD] px-4 py-2 text-xs font-black transition-transform duration-200 hover:-translate-y-0.5 sm:text-sm">All sessions</a></header> <main class="flex flex-1 items-center justify-center py-10 sm:py-14">`);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`<div class="flex w-full max-w-md flex-col items-center text-center"><div class="flex h-16 w-16 items-center justify-center rounded-full border-2 border-[#1C1124] bg-[#F7FFCD] text-2xl shadow-[5px_5px_0px_#1C1124]">…</div> <h1 class="mt-6 text-2xl font-black">Loading your session</h1> <p class="mt-2 text-sm text-[#4d2a3a]">Getting everything ready.</p></div>`);
		$$renderer.push(`<!--]--></main></div></div>`);
	});
}
//#endregion
export { _page as default };
