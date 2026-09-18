import { a as head, r as derived } from "../../../chunks/server.js";
import { t as page } from "../../../chunks/state.js";
import { t as Navbar } from "../../../chunks/Navbar.js";
import "posthog-js";
//#region src/routes/(app)/+layout.svelte
function _layout($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let { children } = $$props;
		derived(() => page.data.user);
		head("1v2axqk", $$renderer, ($$renderer) => {});
		Navbar($$renderer, {});
		$$renderer.push(`<!----> `);
		children($$renderer);
		$$renderer.push(`<!---->`);
	});
}
//#endregion
export { _layout as default };
