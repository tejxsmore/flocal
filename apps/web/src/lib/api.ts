import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type {
	ApiEnvelope,
	LinkedAccount,
	OAuthProvider,
	Payment,
	SpeakingSession,
	Subscription,
	SubscriptionPlan,
	Topic,
	TopicCategory,
	TopicFormat,
	TopicUsageStats,
	UserCurrentPlan,
	AIQuotaStatus,
	DailyActivity,
	UserStats,
	UserSkillStats,
	Badge,
	UserBadge,
	LeaderboardEntry,
	CountryLeaderboardEntry,
	SessionHistoryItem,
	SessionHistoryFilters,
	User,
	SessionSummary,
	XPTransaction,
	SessionReport,
	SessionResponse,
	CheckoutSession,
	UpdatePaymentMethodResult,
	DebateStance,
	OnboardingGoal,
	DailyTimeCommitment,
	FocusArea,
	OnboardingPreferencesResponse,
	AuditLog,
	VocabularyListResponse,
	VocabularyEntry,
	NotificationPreferences,
	Notification
} from '$lib/types';

export class ApiRequestError extends Error {
	code: string;
	status: number;

	constructor(status: number, code: string, message: string) {
		super(message);
		this.name = 'ApiRequestError';
		this.status = status;
		this.code = code;
	}
}

const CSRF_COOKIE = 'flocal_csrf';
const CSRF_HEADER = 'X-CSRF-Token';
const MUTATING_METHODS = new Set(['POST', 'PUT', 'PATCH', 'DELETE']);

function readCsrfCookie(): string | null {
	if (typeof document === 'undefined') return null;

	const match = document.cookie.match(new RegExp(`(?:^|;\\s*)${CSRF_COOKIE}=([^;]+)`));

	return match ? decodeURIComponent(match[1]) : null;
}

async function apiFetch<T>(
	path: string,
	init: RequestInit = {},
	fetchFn: typeof fetch = fetch
): Promise<T> {
	const method = (init.method ?? 'GET').toUpperCase();
	const headers = new Headers(init.headers);

	if (!headers.has('Content-Type') && !(init.body instanceof FormData)) {
		headers.set('Content-Type', 'application/json');
	}

	if (MUTATING_METHODS.has(method)) {
		const csrfToken = readCsrfCookie();
		if (csrfToken) headers.set(CSRF_HEADER, csrfToken);
	}

	const res = await fetchFn(`${PUBLIC_API_BASE_URL}${path}`, {
		...init,
		method,
		credentials: 'include',
		headers
	});

	let body: ApiEnvelope<T>;

	try {
		body = await res.json();
	} catch {
		throw new ApiRequestError(
			res.status,
			'invalid_response',
			`Request failed with status ${res.status}`
		);
	}

	if (!res.ok || !body.success || body.data === undefined) {
		throw new ApiRequestError(
			res.status,
			body.error?.code ?? 'unknown_error',
			body.error?.message ?? `Request failed with status ${res.status}`
		);
	}

	return body.data;
}

export function googleLoginUrl(): string {
	return `${PUBLIC_API_BASE_URL}/api/v1/auth/google`;
}

export function githubLoginUrl(): string {
	return `${PUBLIC_API_BASE_URL}/api/v1/auth/github`;
}

export function linkAccountUrl(provider: OAuthProvider): string {
	return `${PUBLIC_API_BASE_URL}/api/v1/auth/${provider}/link`;
}

export function requestMagicLink(
	email: string,
	fetchFn: typeof fetch = fetch
): Promise<{ message: string }> {
	return apiFetch(
		'/api/v1/auth/magic-link',
		{
			method: 'POST',
			body: JSON.stringify({ email })
		},
		fetchFn
	);
}

export function consumeMagicLink(
	email: string,
	token: string,
	fetchFn: typeof fetch = fetch
): Promise<SessionResponse> {
	return apiFetch<SessionResponse>(
		'/api/v1/auth/magic-link/consume',
		{
			method: 'POST',
			body: JSON.stringify({ email, token })
		},
		fetchFn
	);
}

export function listLinkedAccounts(fetchFn: typeof fetch = fetch): Promise<LinkedAccount[]> {
	return apiFetch<LinkedAccount[]>('/api/v1/auth/accounts', {}, fetchFn);
}

export function logout(fetchFn: typeof fetch = fetch): Promise<{ message: string }> {
	return apiFetch('/api/v1/auth/logout', { method: 'POST' }, fetchFn);
}

export function updateProfile(
	input: {
		name: string;
		username?: string | null;
		image?: string | null;
		defaultPrepTimeSeconds: number;
		timezone?: string | null;
	},
	fetchFn: typeof fetch = fetch
): Promise<User> {
	return apiFetch<User>(
		'/api/v1/auth/me',
		{
			method: 'PATCH',
			body: JSON.stringify(input)
		},
		fetchFn
	);
}

export function requestEmailChange(
	newEmail: string,
	fetchFn: typeof fetch = fetch
): Promise<{ message: string }> {
	return apiFetch(
		'/api/v1/auth/email/change',
		{
			method: 'POST',
			body: JSON.stringify({ newEmail })
		},
		fetchFn
	);
}

export function consumeEmailChange(
	newEmail: string,
	token: string,
	fetchFn: typeof fetch = fetch
): Promise<User> {
	return apiFetch<User>(
		'/api/v1/auth/email/verify/confirm',
		{
			method: 'POST',
			body: JSON.stringify({ newEmail, token })
		},
		fetchFn
	);
}

export function uploadAvatar(file: File, fetchFn: typeof fetch = fetch): Promise<User> {
	const formData = new FormData();
	formData.append('avatar', file);

	return apiFetch<User>(
		'/api/v1/auth/me/avatar',
		{
			method: 'POST',
			body: formData
		},
		fetchFn
	);
}

export function deleteAccount(fetchFn: typeof fetch = fetch): Promise<{ message: string }> {
	return apiFetch('/api/v1/auth/me', { method: 'DELETE' }, fetchFn);
}

export function listMySessions(fetchFn: typeof fetch = fetch): Promise<SessionSummary[]> {
	return apiFetch<SessionSummary[]>('/api/v1/auth/sessions', {}, fetchFn);
}

export function revokeSession(
	id: string,
	fetchFn: typeof fetch = fetch
): Promise<{ message: string }> {
	return apiFetch(
		`/api/v1/auth/sessions/${id}`,
		{
			method: 'DELETE'
		},
		fetchFn
	);
}

export function completeProfileOnboarding(
	input: {
		name: string;
		username?: string | null;
		image?: string | null;
		countryCode?: string | null;
	},
	fetchFn: typeof fetch = fetch
): Promise<User> {
	return apiFetch<User>(
		'/api/v1/onboarding/profile',
		{
			method: 'POST',
			body: JSON.stringify(input)
		},
		fetchFn
	);
}

export function completePreferencesOnboarding(
	input: {
		primaryGoal: OnboardingGoal | null;
		dailyTimeCommitment: DailyTimeCommitment | null;
		focusAreas: FocusArea[];
	},
	fetchFn: typeof fetch = fetch
): Promise<OnboardingPreferencesResponse> {
	return apiFetch<OnboardingPreferencesResponse>(
		'/api/v1/onboarding/preferences',
		{
			method: 'POST',
			body: JSON.stringify(input)
		},
		fetchFn
	);
}

export function getOnboardingPreferences(
	fetchFn: typeof fetch = fetch
): Promise<OnboardingPreferencesResponse> {
	return apiFetch<OnboardingPreferencesResponse>('/api/v1/onboarding/preferences', {}, fetchFn);
}

export function listCategories(fetchFn: typeof fetch = fetch): Promise<TopicCategory[]> {
	return apiFetch<TopicCategory[]>('/api/v1/topics/categories', {}, fetchFn);
}

export function spinTopic(
	categoryId: string | null,
	format: TopicFormat | null,
	fetchFn: typeof fetch = fetch
): Promise<Topic> {
	const params = new URLSearchParams();

	if (categoryId) params.set('categoryId', categoryId);
	if (format) params.set('format', format);

	const query = params.toString() ? `?${params.toString()}` : '';

	return apiFetch<Topic>(`/api/v1/topics/spin${query}`, {}, fetchFn);
}

export function getTopicUsage(
	topicId: string,
	fetchFn: typeof fetch = fetch
): Promise<TopicUsageStats> {
	return apiFetch<TopicUsageStats>(`/api/v1/topics/${topicId}/usage`, {}, fetchFn);
}

export function listTopTopics(
	limit = 10,
	fetchFn: typeof fetch = fetch
): Promise<TopicUsageStats[]> {
	return apiFetch<TopicUsageStats[]>(`/api/v1/topic-stats/top?limit=${limit}`, {}, fetchFn);
}

export function createSession(
	topicId: string,
	prepTimeSeconds: number,
	options: {
		debateStance?: DebateStance;
		fetchFn?: typeof fetch;
	} = {}
): Promise<SpeakingSession> {
	const { debateStance, fetchFn = fetch } = options;

	return apiFetch<SpeakingSession>(
		'/api/v1/sessions',
		{
			method: 'POST',
			body: JSON.stringify({
				topicId,
				prepTimeSeconds,
				...(debateStance ? { debateStance } : {})
			})
		},
		fetchFn
	);
}

export function getSession(id: string, fetchFn: typeof fetch = fetch): Promise<SpeakingSession> {
	return apiFetch<SpeakingSession>(`/api/v1/sessions/${id}`, {}, fetchFn);
}

export function listSessions(
	limit = 20,
	offset = 0,
	fetchFn: typeof fetch = fetch
): Promise<SpeakingSession[]> {
	return apiFetch<SpeakingSession[]>(
		`/api/v1/sessions?limit=${limit}&offset=${offset}`,
		{},
		fetchFn
	);
}

export function deleteSession(
	id: string,
	fetchFn: typeof fetch = fetch
): Promise<{ status: string }> {
	return apiFetch(
		`/api/v1/sessions/${id}`,
		{
			method: 'DELETE'
		},
		fetchFn
	);
}

export function listSessionHistory(
	limit = 20,
	offset = 0,
	filters: SessionHistoryFilters = {},
	fetchFn: typeof fetch = fetch
): Promise<SessionHistoryItem[]> {
	const params = new URLSearchParams();

	params.set('limit', String(limit));
	params.set('offset', String(offset));

	if (filters.status?.length) params.set('status', filters.status.join(','));
	if (filters.format?.length) params.set('format', filters.format.join(','));
	if (filters.from) params.set('from', filters.from);
	if (filters.to) params.set('to', filters.to);
	if (filters.q) params.set('q', filters.q);
	if (filters.sort) params.set('sort', filters.sort);

	return apiFetch<SessionHistoryItem[]>(
		`/api/v1/sessions/history?${params.toString()}`,
		{},
		fetchFn
	);
}

export function getAIQuota(fetchFn: typeof fetch = fetch): Promise<AIQuotaStatus> {
	return apiFetch<AIQuotaStatus>('/api/v1/sessions/quota', {}, fetchFn);
}

export function reanalyzeSession(
	id: string,
	fetchFn: typeof fetch = fetch
): Promise<{ status: string }> {
	return apiFetch(
		`/api/v1/sessions/${id}/reanalyze`,
		{
			method: 'POST'
		},
		fetchFn
	);
}

export function getSessionReport(
	id: string,
	fetchFn: typeof fetch = fetch
): Promise<SessionReport> {
	return apiFetch<SessionReport>(`/api/v1/sessions/${id}/report`, {}, fetchFn);
}

export function sessionStreamUrl(id: string): string {
	const wsBase = PUBLIC_API_BASE_URL.replace(/^http/, 'ws');
	return `${wsBase}/api/v1/sessions/${id}/stream`;
}

export function getCurrentPlan(fetchFn: typeof fetch = fetch): Promise<UserCurrentPlan> {
	return apiFetch<UserCurrentPlan>('/api/v1/plan', {}, fetchFn);
}

export function listPlans(fetchFn: typeof fetch = fetch): Promise<SubscriptionPlan[]> {
	return apiFetch<SubscriptionPlan[]>('/api/v1/billing/plans', {}, fetchFn);
}

export function getMySubscription(fetchFn: typeof fetch = fetch): Promise<Subscription> {
	return apiFetch<Subscription>('/api/v1/billing/subscription', {}, fetchFn);
}

export function listMyPayments(
	limit = 20,
	offset = 0,
	fetchFn: typeof fetch = fetch
): Promise<Payment[]> {
	return apiFetch<Payment[]>(
		`/api/v1/billing/payments?limit=${limit}&offset=${offset}`,
		{},
		fetchFn
	);
}

export function createCheckout(
	planSlug: string,
	country?: string,
	fetchFn: typeof fetch = fetch
): Promise<CheckoutSession> {
	return apiFetch<CheckoutSession>(
		'/api/v1/billing/checkout',
		{
			method: 'POST',
			body: JSON.stringify({
				planSlug,
				...(country ? { country } : {})
			})
		},
		fetchFn
	);
}

export function cancelSubscription(fetchFn: typeof fetch = fetch): Promise<{ canceled: true }> {
	return apiFetch(
		'/api/v1/billing/subscription/cancel',
		{
			method: 'POST'
		},
		fetchFn
	);
}

export function resumeSubscription(fetchFn: typeof fetch = fetch): Promise<{ resumed: true }> {
	return apiFetch(
		'/api/v1/billing/subscription/resume',
		{
			method: 'POST'
		},
		fetchFn
	);
}

export function changePlan(
	planSlug: string,
	fetchFn: typeof fetch = fetch
): Promise<{ changing: true }> {
	return apiFetch(
		'/api/v1/billing/subscription/change-plan',
		{
			method: 'POST',
			body: JSON.stringify({ planSlug })
		},
		fetchFn
	);
}

export function updatePaymentMethod(
	fetchFn: typeof fetch = fetch
): Promise<UpdatePaymentMethodResult> {
	return apiFetch<UpdatePaymentMethodResult>(
		'/api/v1/billing/subscription/update-payment-method',
		{ method: 'POST' },
		fetchFn
	);
}

export function getMyStats(fetchFn: typeof fetch = fetch): Promise<UserStats> {
	return apiFetch<UserStats>('/api/v1/gamification/stats', {}, fetchFn);
}

export function getMySkillStats(fetchFn: typeof fetch = fetch): Promise<UserSkillStats> {
	return apiFetch<UserSkillStats>('/api/v1/gamification/skills', {}, fetchFn);
}

export function listMyDailyActivity(
	limit = 30,
	fetchFn: typeof fetch = fetch
): Promise<DailyActivity[]> {
	return apiFetch<DailyActivity[]>(`/api/v1/gamification/activity?limit=${limit}`, {}, fetchFn);
}

export function listMyXPTransactions(
	limit = 20,
	offset = 0,
	fetchFn: typeof fetch = fetch
): Promise<XPTransaction[]> {
	return apiFetch<XPTransaction[]>(
		`/api/v1/gamification/xp?limit=${limit}&offset=${offset}`,
		{},
		fetchFn
	);
}

export function listBadges(fetchFn: typeof fetch = fetch): Promise<Badge[]> {
	return apiFetch<Badge[]>('/api/v1/gamification/badges', {}, fetchFn);
}

export function listMyBadges(fetchFn: typeof fetch = fetch): Promise<UserBadge[]> {
	return apiFetch<UserBadge[]>('/api/v1/gamification/badges/me', {}, fetchFn);
}

export function getLeaderboard(
	limit = 20,
	offset = 0,
	fetchFn: typeof fetch = fetch
): Promise<LeaderboardEntry[]> {
	return apiFetch<LeaderboardEntry[]>(
		`/api/v1/gamification/leaderboard?limit=${limit}&offset=${offset}`,
		{},
		fetchFn
	);
}

export function getCountryLeaderboard(
	countryCode: string,
	limit = 20,
	offset = 0,
	fetchFn: typeof fetch = fetch
): Promise<CountryLeaderboardEntry[]> {
	return apiFetch<CountryLeaderboardEntry[]>(
		`/api/v1/gamification/leaderboard/country/${countryCode}?limit=${limit}&offset=${offset}`,
		{},
		fetchFn
	);
}

export function getMyLeaderboardRank(fetchFn: typeof fetch = fetch): Promise<LeaderboardEntry> {
	return apiFetch<LeaderboardEntry>('/api/v1/gamification/leaderboard/me', {}, fetchFn);
}

export function listAuditLogs(
	options: {
		entityType?: string;
		entityId?: string;
		actorUserId?: string;
		limit?: number;
		offset?: number;
	} = {},
	fetchFn: typeof fetch = fetch
): Promise<AuditLog[]> {
	const params = new URLSearchParams();

	if (options.entityType) params.set('entityType', options.entityType);
	if (options.entityId) params.set('entityId', options.entityId);
	if (options.actorUserId) params.set('actorUserId', options.actorUserId);

	params.set('limit', String(options.limit ?? 50));
	params.set('offset', String(options.offset ?? 0));

	return apiFetch<AuditLog[]>(`/api/v1/audit-logs?${params.toString()}`, {}, fetchFn);
}

export function listMyVocabulary(
	options: {
		mastered?: boolean;
		limit?: number;
		offset?: number;
	} = {},
	fetchFn: typeof fetch = fetch
): Promise<VocabularyListResponse> {
	const params = new URLSearchParams();

	if (options.mastered !== undefined) params.set('mastered', String(options.mastered));
	params.set('limit', String(options.limit ?? 30));
	params.set('offset', String(options.offset ?? 0));

	return apiFetch<VocabularyListResponse>(`/api/v1/vocabulary?${params.toString()}`, {}, fetchFn);
}

export function setWordMastered(
	wordId: string,
	mastered: boolean,
	fetchFn: typeof fetch = fetch
): Promise<VocabularyEntry> {
	return apiFetch<VocabularyEntry>(
		`/api/v1/vocabulary/${wordId}`,
		{
			method: 'PATCH',
			body: JSON.stringify({ mastered })
		},
		fetchFn
	);
}

export function getNotificationPreferences(
	fetchFn: typeof fetch = fetch
): Promise<NotificationPreferences> {
	return apiFetch<NotificationPreferences>('/api/v1/notifications/preferences', {}, fetchFn);
}

export function updateNotificationPreferences(
	input: {
		dailyReminderEnabled: boolean;
		streakRiskEnabled: boolean;
		sessionReadyEnabled: boolean;
		reminderTime?: string | null;
	},
	fetchFn: typeof fetch = fetch
): Promise<NotificationPreferences> {
	return apiFetch<NotificationPreferences>(
		'/api/v1/notifications/preferences',
		{
			method: 'PATCH',
			body: JSON.stringify(input)
		},
		fetchFn
	);
}

export function listNotifications(
	options: {
		unreadOnly?: boolean;
		limit?: number;
		offset?: number;
	} = {},
	fetchFn: typeof fetch = fetch
): Promise<Notification[]> {
	const params = new URLSearchParams();

	if (options.unreadOnly !== undefined) params.set('unreadOnly', String(options.unreadOnly));
	params.set('limit', String(options.limit ?? 20));
	params.set('offset', String(options.offset ?? 0));

	return apiFetch<Notification[]>(`/api/v1/notifications?${params.toString()}`, {}, fetchFn);
}

export function getUnreadNotificationCount(
	fetchFn: typeof fetch = fetch
): Promise<{ unreadCount: number }> {
	return apiFetch<{ unreadCount: number }>('/api/v1/notifications/unread-count', {}, fetchFn);
}

export function markNotificationRead(
	id: string,
	fetchFn: typeof fetch = fetch
): Promise<{ message: string }> {
	return apiFetch(`/api/v1/notifications/${id}/read`, { method: 'POST' }, fetchFn);
}

export function markAllNotificationsRead(
	fetchFn: typeof fetch = fetch
): Promise<{ message: string }> {
	return apiFetch('/api/v1/notifications/read-all', { method: 'POST' }, fetchFn);
}
