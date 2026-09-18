// ── Auth ──────────────────────────────────────────────────────────────

export type UserRole = 'user' | 'moderator' | 'admin';

export interface User {
	id: string;
	name: string;
	email: string;
	emailVerified: boolean;
	emailVerifiedAt?: string | null;
	image?: string | null;
	username?: string | null;
	defaultPrepTimeSeconds: number;
	timezone?: string | null;
	countryCode?: string | null;
	role: UserRole;
	isActive: boolean;
	onboardingCompleted: boolean;
	lastLoginAt?: string | null;
	deletedAt?: string | null;
	createdAt: string;
	updatedAt: string;
}

export interface ApiError {
	code: string;
	message: string;
	details?: unknown;
}

export interface ApiEnvelope<T> {
	success: boolean;
	data?: T;
	error?: ApiError;
}

export type OAuthProvider = 'google' | 'github';

export interface LinkedAccount {
	provider: OAuthProvider;
	connected: boolean;
}

export interface SessionResponse {
	user: User;
	token: string;
	expiresAt: string;
	csrfToken: string;
}

export interface AuthSession {
	id: string;
	expiresAt: string;
	createdAt: string;
	updatedAt: string;
	ipAddress?: string | null;
	userAgent?: string | null;
	userId: string;
}

export interface SessionSummary extends AuthSession {
	current: boolean;
}

// ── Onboarding ────────────────────────────────────────────────────────

export type OnboardingGoal =
	'confidence' | 'conversations' | 'public_speaking' | 'work_interviews' | 'vocabulary' | 'habit';

export type DailyTimeCommitment = 'min_5' | 'min_10' | 'min_15' | 'min_30_plus';

export type FocusArea =
	| 'work_interviews'
	| 'conversations'
	| 'college_presentations'
	| 'travel'
	| 'everyday_situations'
	| 'public_speaking'
	| 'online_meetings'
	| 'anywhere';

export const GOAL_OPTIONS: { label: string; value: OnboardingGoal }[] = [
	{ label: 'Build confidence', value: 'confidence' },
	{ label: 'Better conversations', value: 'conversations' },
	{ label: 'Public speaking', value: 'public_speaking' },
	{ label: 'Work interviews', value: 'work_interviews' },
	{ label: 'Grow vocabulary', value: 'vocabulary' },
	{ label: 'Build a habit', value: 'habit' }
];

export const TIME_COMMITMENT_OPTIONS: { label: string; value: DailyTimeCommitment }[] = [
	{ label: '5 min/day', value: 'min_5' },
	{ label: '10 min/day', value: 'min_10' },
	{ label: '15 min/day', value: 'min_15' },
	{ label: '30+ min/day', value: 'min_30_plus' }
];

export const FOCUS_AREA_OPTIONS: { label: string; value: FocusArea }[] = [
	{ label: 'Work interviews', value: 'work_interviews' },
	{ label: 'Conversations', value: 'conversations' },
	{ label: 'College presentations', value: 'college_presentations' },
	{ label: 'Travel', value: 'travel' },
	{ label: 'Everyday situations', value: 'everyday_situations' },
	{ label: 'Public speaking', value: 'public_speaking' },
	{ label: 'Online meetings', value: 'online_meetings' },
	{ label: 'Anywhere', value: 'anywhere' }
];

// export type OnboardingGoal =
// 	| 'speak_confidently'
// 	| 'speak_fluently'
// 	| 'start_conversations'
// 	| 'express_ideas'
// 	| 'professional_english'
// 	| 'public_speaking'
// 	| 'vocabulary';

// export type DailyTimeCommitment =
// 	| 'min_5'
// 	| 'min_10'
// 	| 'min_15'
// 	| 'min_20'
// 	| 'min_30_plus';

// export type FocusArea =
// 	| 'job_interviews'
// 	| 'workplace'
// 	| 'college'
// 	| 'social_conversations'
// 	| 'meeting_new_people'
// 	| 'travel'
// 	| 'daily_life'
// 	| 'public_speaking'
// 	| 'online_meetings'
// 	| 'customer_interactions'
// 	| 'anywhere';

// export type SpeakingChallenge =
// 	| 'finding_words'
// 	| 'forming_sentences'
// 	| 'thinking_in_english'
// 	| 'confidence'
// 	| 'pronunciation'
// 	| 'keeping_conversations'
// 	| 'understanding_others'
// 	| 'speaking_under_pressure';

// export const GOAL_OPTIONS: { label: string; value: OnboardingGoal }[] = [
// 	{ label: 'Speak with more confidence', value: 'speak_confidently' },
// 	{ label: 'Speak more fluently', value: 'speak_fluently' },
// 	{ label: 'Start and keep conversations going', value: 'start_conversations' },
// 	{ label: 'Express my thoughts more clearly', value: 'express_ideas' },
// 	{ label: 'Improve my professional English', value: 'professional_english' },
// 	{ label: 'Become a better public speaker', value: 'public_speaking' },
// 	{ label: 'Build a stronger vocabulary', value: 'vocabulary' }
// ];

// export const TIME_COMMITMENT_OPTIONS: {
// 	label: string;
// 	value: DailyTimeCommitment;
// }[] = [
// 	{ label: '5 minutes', value: 'min_5' },
// 	{ label: '10 minutes', value: 'min_10' },
// 	{ label: '15 minutes', value: 'min_15' },
// 	{ label: '20 minutes', value: 'min_20' },
// 	{ label: '30+ minutes', value: 'min_30_plus' }
// ];

// export const FOCUS_AREA_OPTIONS: { label: string; value: FocusArea }[] = [
// 	{ label: 'Job interviews', value: 'job_interviews' },
// 	{ label: 'Work & professional communication', value: 'workplace' },
// 	{ label: 'College & classroom', value: 'college' },
// 	{ label: 'Friends & social conversations', value: 'social_conversations' },
// 	{ label: 'Meeting new people', value: 'meeting_new_people' },
// 	{ label: 'Travel & navigating new places', value: 'travel' },
// 	{ label: 'Everyday situations', value: 'daily_life' },
// 	{ label: 'Presentations & public speaking', value: 'public_speaking' },
// 	{ label: 'Online meetings & calls', value: 'online_meetings' },
// 	{ label: 'Customer-facing situations', value: 'customer_interactions' },
// 	{ label: 'I want to be comfortable anywhere', value: 'anywhere' }
// ];

// export const SPEAKING_CHALLENGE_OPTIONS: {
// 	label: string;
// 	value: SpeakingChallenge;
// }[] = [
// 	{ label: 'Finding the right words', value: 'finding_words' },
// 	{ label: 'Putting my thoughts into sentences', value: 'forming_sentences' },
// 	{ label: 'Thinking directly in English', value: 'thinking_in_english' },
// 	{ label: 'Feeling confident while speaking', value: 'confidence' },
// 	{ label: 'Pronouncing words clearly', value: 'pronunciation' },
// 	{ label: 'Keeping conversations going', value: 'keeping_conversations' },
// 	{ label: 'Understanding what others say', value: 'understanding_others' },
// 	{
// 		label: 'Speaking when I feel nervous or under pressure',
// 		value: 'speaking_under_pressure'
// 	}
// ];

export interface UserPreferences {
	userId: string;
	primaryGoal?: OnboardingGoal | null;
	dailyTimeCommitment?: DailyTimeCommitment | null;
	completedAt?: string | null;
	createdAt: string;
	updatedAt: string;
}

export interface UserFocusAreaEntry {
	userId: string;
	focusArea: FocusArea;
	createdAt: string;
}

export interface OnboardingPreferencesResponse {
	preferences: UserPreferences | null;
	focusAreas: UserFocusAreaEntry[];
}

// ── Topics ────────────────────────────────────────────────────────────

export interface TopicCategory {
	id: string;
	slug: string;
	name: string;
	icon?: string | null;
	isActive: boolean;
	sortOrder: number;
	createdAt: string;
}

export type TopicFormat = 'word' | 'quote' | 'debate' | 'situation' | 'story_starter' | 'image';
export type TopicDifficulty = 'easy' | 'medium' | 'hard';
export type TopicSource = 'manual' | 'ai_generated' | 'community';

export const FORMAT_OPTIONS: { label: string; value: TopicFormat; icon: string }[] = [
	{ label: 'Word', value: 'word', icon: '📝' },
	{ label: 'Quote', value: 'quote', icon: '💭' },
	{ label: 'Debate', value: 'debate', icon: '⚖️' },
	{ label: 'Situation', value: 'situation', icon: '🎭' },
	{ label: 'Story Starter', value: 'story_starter', icon: '📖' },
	{ label: 'Image', value: 'image', icon: '📸' }
];

export interface Topic {
	id: string;
	categoryId: string;
	title: string;
	format: TopicFormat;
	difficulty: TopicDifficulty;
	recommendedPrepSeconds: number;
	tags?: string[] | null;
	metadata?: Record<string, unknown> | null;
	isActive: boolean;
	isPremium: boolean;
	source: TopicSource;
	languageCode: string;
	submittedBy?: string | null;
	createdAt: string;
	updatedAt: string;
}

export interface TopicUsageStats {
	topicId: string;
	topicTitle: string;
	spinCount: number;
}

// ── Sessions ──────────────────────────────────────────────────────────

export type PrepTimeSeconds = 0 | 300 | 600 | 900;

export const PREP_TIME_OPTIONS: { label: string; value: PrepTimeSeconds }[] = [
	{ label: 'No prep', value: 0 },
	{ label: '5 min', value: 300 },
	{ label: '10 min', value: 600 },
	{ label: '15 min', value: 900 }
];

export type SessionStatus = 'pending' | 'processing' | 'completed' | 'failed';
export type DebateStance = 'for' | 'against';

export interface SpeakingSession {
	id: string;
	userId: string;
	topicId: string;
	dailyChallengeId?: string | null;
	debateStance?: DebateStance | null;
	prepTimeSeconds: number;
	speakTimeSeconds: number;
	status: SessionStatus;
	processingStep?: string | null;
	processingStartedAt?: string | null;
	failureReason?: string | null;
	retryCount: number;
	lastError?: string | null;
	lastRetryAt?: string | null;
	nextRetryAt?: string | null;
	audioDurationSeconds?: number | null;
	audioFormat?: string | null;
	audioSizeBytes?: number | null;
	audioExpiresAt?: string | null;
	audioDeletedAt?: string | null;
	shareEnabled: boolean;
	shareExpiresAt?: string | null;
	startedAt?: string | null;
	submittedAt?: string | null;
	completedAt?: string | null;
	createdAt: string;
	updatedAt: string;
}

export interface GrammarCorrection {
	id: string;
	sessionId: string;
	transcriptId: string;
	originalText: string;
	correctedText: string;
	explanation?: string | null;
	errorType?: string | null;
	startChar?: number | null;
	endChar?: number | null;
	createdAt: string;
}

export interface VocabularySuggestion {
	id: string;
	sessionId: string;
	transcriptId: string;
	originalWord: string;
	suggestedWords: string[];
	reason?: string | null;
	contextSnippet?: string | null;
	startChar?: number | null;
	endChar?: number | null;
	createdAt: string;
}

export interface SessionHistoryItem {
	id: string;
	topicId: string;
	topicTitle: string;
	topicFormat: TopicFormat;
	status: SessionStatus;
	prepTimeSeconds: number;
	speakTimeSeconds: number;
	overallScore?: number | null;
	wordsPerMinute?: number | null;
	createdAt: string;
	completedAt?: string | null;
}

export type SessionSortOption = 'newest' | 'oldest' | 'score_desc' | 'score_asc';

export const SORT_OPTIONS: { label: string; value: SessionSortOption }[] = [
	{ label: 'Newest first', value: 'newest' },
	{ label: 'Oldest first', value: 'oldest' },
	{ label: 'Highest score', value: 'score_desc' },
	{ label: 'Lowest score', value: 'score_asc' }
];

export const STATUS_OPTIONS: { label: string; value: SessionStatus }[] = [
	{ label: 'Completed', value: 'completed' },
	{ label: 'Processing', value: 'processing' },
	{ label: 'Pending', value: 'pending' },
	{ label: 'Failed', value: 'failed' }
];

export interface SessionHistoryFilters {
	status?: SessionStatus[];
	format?: TopicFormat[];
	from?: string;
	to?: string;
	q?: string;
	sort?: SessionSortOption;
}

export interface AIQuotaStatus {
	eligible: boolean;
	unlimited: boolean;
	lifetimeUsed: number;
	lifetimeLimit: number;
	dailyUsed: number;
	dailyLimit: number;
}

export interface SessionReport {
	sessionId: string;
	userId: string;
	topicId: string;
	topicTitle: string;
	topicFormat: TopicFormat;
	topicCategory: string;
	status: SessionStatus;
	debateStance?: DebateStance | null;
	prepTimeSeconds: number;
	speakTimeSeconds: number;
	audioExpiresAt?: string | null;
	shareEnabled: boolean;
	shareExpiresAt?: string | null;
	audioPlaybackUrl?: string | null;
	transcript?: string | null;
	wordsPerMinute?: number | null;
	fillerWordCount?: number | null;
	fillerRate?: number | null;
	pauseRate?: number | null;
	longestPauseSeconds?: number | null;
	uniqueWordCount?: number | null;
	lexicalDiversity?: number | null;
	sentenceCount?: number | null;
	averageSentenceLength?: number | null;
	overallScore?: number | null;
	clarityScore?: number | null;
	deliveryScore?: number | null;
	contentScore?: number | null;
	vocabularyScore?: number | null;
	grammarScore?: number | null;
	strengths?: string[] | null;
	improvements?: string[] | null;
	coachMessage?: string | null;
	grammarCorrections?: GrammarCorrection[];
	vocabularySuggestions?: VocabularySuggestion[];
	createdAt: string;
	completedAt?: string | null;
}

export interface StreamMessage {
	type: 'interim' | 'final' | 'error' | 'info';
	transcript?: string;
	fillerCount?: number;
	message?: string;
}

// ── Plan ──────────────────────────────────────────────────────────────

export interface UserCurrentPlan {
	userId: string;
	planSlug: string;
	dailySessionLimit?: number | null;
	canViewAnalysis: boolean;
	planRenewsAt?: string | null;
	subscriptionStatus?: SubscriptionStatus | null;
}

// ── Billing ───────────────────────────────────────────────────────────

export type PlanType = 'free' | 'pro';
export type BillingInterval = 'monthly' | 'annual';
export type SubscriptionStatus =
	'active' | 'trialing' | 'past_due' | 'canceled' | 'expired' | 'paused';
export type PaymentStatus = 'created' | 'captured' | 'failed' | 'refunded';

export interface SubscriptionPlan {
	id: string;
	slug: string;
	name: string;
	planType: PlanType;
	billingInterval?: BillingInterval | null;
	priceSubunits: number;
	currency: string;
	dailySessionLimit?: number | null;
	canViewAnalysis: boolean;
	dodoProductId?: string | null;
	isActive: boolean;
	createdAt: string;
}

export interface Subscription {
	id: string;
	userId: string;
	planId: string;
	status: SubscriptionStatus;
	dodoSubscriptionId?: string | null;
	dodoCustomerId?: string | null;
	priceSubunitsAtPurchase: number;
	currencyAtPurchase: string;
	billingIntervalAtPurchase?: BillingInterval | null;
	currentPeriodStart?: string | null;
	currentPeriodEnd?: string | null;
	cancelAtPeriodEnd: boolean;
	canceledAt?: string | null;
	createdAt: string;
	updatedAt: string;
}

export interface Payment {
	id: string;
	userId: string;
	invoiceId?: string | null;
	dodoPaymentId?: string | null;
	dodoCheckoutSessionId?: string | null;
	amountSubunits: number;
	refundedAmountSubunits: number;
	currency: string;
	paymentMethod?: string | null;
	status: PaymentStatus;
	paidAt?: string | null;
	createdAt: string;
}

export interface CheckoutSession {
	sessionId: string;
	checkoutUrl: string;
}

export interface UpdatePaymentMethodResult {
	paymentId?: string;
	paymentLink?: string;
	clientSecret?: string;
}

// ── Gamification ──────────────────────────────────────────────────────

export type XPReason =
	| 'session_complete'
	| 'daily_challenge'
	| 'streak_bonus'
	| 'first_session'
	| 'perfect_score'
	| 'badge_earned'
	| 'xp_adjustment'
	| 'xp_reversal';

export const XP_REASON_LABELS: Record<XPReason, string> = {
	session_complete: 'Session complete',
	daily_challenge: 'Daily challenge',
	streak_bonus: 'Streak bonus',
	first_session: 'First session',
	perfect_score: 'Perfect score',
	badge_earned: 'Badge earned',
	xp_adjustment: 'Adjustment',
	xp_reversal: 'Reversal'
};

export interface XPTransaction {
	id: string;
	userId: string;
	sessionId?: string | null;
	amount: number;
	reason: XPReason;
	createdAt: string;
}

export interface DailyActivity {
	id: string;
	userId: string;
	activityDate: string;
	sessionsCount: number;
	xpEarned: number;
	createdAt: string;
	updatedAt: string;
}

export interface UserStats {
	userId: string;
	totalSessions: number;
	averageScore: number;
	bestScore: number;
	currentStreakDays: number;
	longestStreakDays: number;
	lastSessionDate?: string | null;
	totalXp: number;
	updatedAt: string;
}

export interface UserSkillStats {
	userId: string;
	avgClarity?: number | null;
	avgDelivery?: number | null;
	avgContent?: number | null;
	avgVocabulary?: number | null;
	avgGrammar?: number | null;
	sessionsCounted: number;
	updatedAt: string;
}

export type BadgeCriteriaType =
	'streak_days' | 'session_count' | 'xp_total' | 'score_threshold' | 'manual';

export interface Badge {
	id: string;
	slug: string;
	name: string;
	description?: string | null;
	icon?: string | null;
	criteriaType: BadgeCriteriaType;
	criteriaValue?: number | null;
	isActive: boolean;
	sortOrder: number;
	createdAt: string;
}

export interface UserBadge {
	id: string;
	userId: string;
	badgeId: string;
	sessionId?: string | null;
	earnedAt: string;
}

export interface LeaderboardEntry {
	userId: string;
	username?: string | null;
	image?: string | null;
	countryCode?: string | null;
	totalXp: number;
	currentStreakDays: number;
	totalSessions: number;
	rank: number;
}

export type CountryLeaderboardEntry = LeaderboardEntry;

// ── Audit ─────────────────────────────────────────────────────────────

export interface AuditLog {
	id: string;
	actorUserId?: string | null;
	action: string;
	entityType: string;
	entityId: string;
	oldValue?: Record<string, unknown> | null;
	newValue?: Record<string, unknown> | null;
	source?: string | null;
	createdAt: string;
}

// ── Vocabulary ────────────────────────────────────────────────────────

export interface VocabularyEntry {
	wordId: string;
	word: string;
	definition?: string | null;
	difficulty?: TopicDifficulty | null;
	timesSeen: number;
	timesSuggested: number;
	mastered: boolean;
	firstSeenAt: string;
	lastSeenAt: string;
}

export interface VocabularyListResponse {
	words: VocabularyEntry[];
	total: number;
	limit: number;
	offset: number;
}

// ── Notifications ─────────────────────────────────────────────────────

export type NotificationType =
	'daily_reminder' | 'streak_risk' | 'session_ready' | 'badge_earned' | 'subscription';

export type NotificationChannel = 'in_app' | 'email' | 'push';

export type NotificationDeliveryStatus = 'pending' | 'sent' | 'delivered' | 'failed';

export const NOTIFICATION_TYPE_LABELS: Record<NotificationType, string> = {
	daily_reminder: 'Daily reminder',
	streak_risk: 'Streak at risk',
	session_ready: 'Session ready',
	badge_earned: 'Badge earned',
	subscription: 'Subscription'
};

export interface NotificationPreferences {
	userId: string;
	dailyReminderEnabled: boolean;
	streakRiskEnabled: boolean;
	sessionReadyEnabled: boolean;
	reminderTime?: string | null;
	updatedAt: string;
}

export interface Notification {
	id: string;
	userId: string;
	type: NotificationType;
	channel: NotificationChannel;
	deliveryStatus: NotificationDeliveryStatus;
	title: string;
	body?: string | null;
	data?: Record<string, unknown> | null;
	providerMessageId?: string | null;
	readAt?: string | null;
	sentAt?: string | null;
	failedAt?: string | null;
	createdAt: string;
}
