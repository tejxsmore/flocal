
// this file is generated — do not edit it


declare module "svelte/elements" {
	export interface HTMLAttributes<T> {
		'data-sveltekit-keepfocus'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-noscroll'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-preload-code'?:
			| true
			| ''
			| 'eager'
			| 'viewport'
			| 'hover'
			| 'tap'
			| 'off'
			| undefined
			| null;
		'data-sveltekit-preload-data'?: true | '' | 'hover' | 'tap' | 'off' | undefined | null;
		'data-sveltekit-reload'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-replacestate'?: true | '' | 'off' | undefined | null;
	}
}

export {};


declare module "$app/types" {
	type MatcherParam<M> = M extends (param : string) => param is (infer U extends string) ? U : string;

	export interface AppTypes {
		RouteId(): "/(auth)" | "/(app)" | "/" | "/(app)/leaderboard" | "/(auth)/login" | "/onboarding" | "/plan" | "/(app)/profile" | "/(app)/profile/billing" | "/(app)/profile/settings" | "/(app)/profile/settings/verify-email" | "/(app)/session" | "/(app)/session/[id]" | "/(auth)/verify";
		RouteParams(): {
			"/(app)/session/[id]": { id: string }
		};
		LayoutParams(): {
			"/(auth)": Record<string, never>;
			"/(app)": { id?: string | undefined };
			"/": { id?: string | undefined };
			"/(app)/leaderboard": Record<string, never>;
			"/(auth)/login": Record<string, never>;
			"/onboarding": Record<string, never>;
			"/plan": Record<string, never>;
			"/(app)/profile": Record<string, never>;
			"/(app)/profile/billing": Record<string, never>;
			"/(app)/profile/settings": Record<string, never>;
			"/(app)/profile/settings/verify-email": Record<string, never>;
			"/(app)/session": { id?: string | undefined };
			"/(app)/session/[id]": { id: string };
			"/(auth)/verify": Record<string, never>
		};
		Pathname(): "/" | "/leaderboard" | "/login" | "/onboarding" | "/plan" | "/profile" | "/profile/billing" | "/profile/settings" | "/profile/settings/verify-email" | "/session" | `/session/${string}` & {} | "/verify";
		ResolvedPathname(): `${"" | `/${string}`}${ReturnType<AppTypes['Pathname']>}`;
		Asset(): "/.DS_Store" | "/assets/.DS_Store" | "/assets/font/.DS_Store" | "/assets/font/Switzer/Switzer-Variable.woff2" | "/assets/font/Switzer/Switzer-VariableItalic.woff2" | "/robots.txt" | string & {};
	}
}