import * as server from '../entries/pages/(auth)/verify/_page.server.ts.js';

export const index = 12;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/(auth)/verify/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/(auth)/verify/+page.server.ts";
export const imports = ["_app/immutable/nodes/12.D09wtdrC.js","_app/immutable/chunks/B9cUK-4N.js","_app/immutable/chunks/xihTtKlq.js"];
export const stylesheets = [];
export const fonts = [];
