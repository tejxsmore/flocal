package models

type TopicDifficulty string

const (
	DifficultyEasy   TopicDifficulty = "easy"
	DifficultyMedium TopicDifficulty = "medium"
	DifficultyHard   TopicDifficulty = "hard"
)

func (d TopicDifficulty) Valid() bool {
	switch d {
	case DifficultyEasy, DifficultyMedium, DifficultyHard:
		return true
	default:
		return false
	}
}

func (d TopicDifficulty) String() string {
	return string(d)
}

type TopicFormat string

const (
	TopicFormatWord         TopicFormat = "word"
	TopicFormatQuote        TopicFormat = "quote"
	TopicFormatDebate       TopicFormat = "debate"
	TopicFormatSituation    TopicFormat = "situation"
	TopicFormatStoryStarter TopicFormat = "story_starter"
	TopicFormatImage        TopicFormat = "image"
)

func (f TopicFormat) Valid() bool {
	switch f {
	case TopicFormatWord,
		TopicFormatQuote,
		TopicFormatDebate,
		TopicFormatSituation,
		TopicFormatStoryStarter,
		TopicFormatImage:
		return true
	default:
		return false
	}
}

func (f TopicFormat) String() string {
	return string(f)
}

type SessionStatus string

const (
	SessionStatusPending    SessionStatus = "pending"
	SessionStatusProcessing SessionStatus = "processing"
	SessionStatusCompleted  SessionStatus = "completed"
	SessionStatusFailed     SessionStatus = "failed"
)

func (s SessionStatus) Valid() bool {
	switch s {
	case SessionStatusPending,
		SessionStatusProcessing,
		SessionStatusCompleted,
		SessionStatusFailed:
		return true
	default:
		return false
	}
}

func (s SessionStatus) String() string {
	return string(s)
}

func (s SessionStatus) Terminal() bool {
	return s == SessionStatusCompleted || s == SessionStatusFailed
}

type XPReason string

const (
	XPReasonSessionComplete XPReason = "session_complete"
	XPReasonDailyChallenge  XPReason = "daily_challenge"
	XPReasonStreakBonus     XPReason = "streak_bonus"
	XPReasonFirstSession    XPReason = "first_session"
	XPReasonPerfectScore    XPReason = "perfect_score"
	XPReasonBadgeEarned     XPReason = "badge_earned"
	XPReasonAdjustment      XPReason = "xp_adjustment"
	XPReasonReversal        XPReason = "xp_reversal"
)

func (r XPReason) Valid() bool {
	switch r {
	case XPReasonSessionComplete,
		XPReasonDailyChallenge,
		XPReasonStreakBonus,
		XPReasonFirstSession,
		XPReasonPerfectScore,
		XPReasonBadgeEarned,
		XPReasonAdjustment,
		XPReasonReversal:
		return true
	default:
		return false
	}
}

func (r XPReason) String() string {
	return string(r)
}

type PlanType string

const (
	PlanTypeFree PlanType = "free"
	PlanTypePro  PlanType = "pro"
)

func (p PlanType) Valid() bool {
	switch p {
	case PlanTypeFree, PlanTypePro:
		return true
	default:
		return false
	}
}

func (p PlanType) String() string {
	return string(p)
}

type BillingInterval string

const (
	BillingIntervalMonthly BillingInterval = "monthly"
	BillingIntervalAnnual  BillingInterval = "annual"
)

func (b BillingInterval) Valid() bool {
	switch b {
	case BillingIntervalMonthly, BillingIntervalAnnual:
		return true
	default:
		return false
	}
}

func (b BillingInterval) String() string {
	return string(b)
}

type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
	SubscriptionStatusPaused   SubscriptionStatus = "paused"
)

func (s SubscriptionStatus) Valid() bool {
	switch s {
	case SubscriptionStatusActive,
		SubscriptionStatusTrialing,
		SubscriptionStatusPastDue,
		SubscriptionStatusCanceled,
		SubscriptionStatusExpired,
		SubscriptionStatusPaused:
		return true
	default:
		return false
	}
}

func (s SubscriptionStatus) String() string {
	return string(s)
}

func (s SubscriptionStatus) IsEntitled() bool {
	return s == SubscriptionStatusActive || s == SubscriptionStatusTrialing
}

type PaymentStatus string

const (
	PaymentStatusCreated  PaymentStatus = "created"
	PaymentStatusCaptured PaymentStatus = "captured"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

func (p PaymentStatus) Valid() bool {
	switch p {
	case PaymentStatusCreated,
		PaymentStatusCaptured,
		PaymentStatusFailed,
		PaymentStatusRefunded:
		return true
	default:
		return false
	}
}

func (p PaymentStatus) String() string {
	return string(p)
}

type WebhookProcessingStatus string

const (
	WebhookStatusReceived   WebhookProcessingStatus = "received"
	WebhookStatusProcessing WebhookProcessingStatus = "processing"
	WebhookStatusProcessed  WebhookProcessingStatus = "processed"
	WebhookStatusFailed     WebhookProcessingStatus = "failed"
)

func (w WebhookProcessingStatus) Valid() bool {
	switch w {
	case WebhookStatusReceived,
		WebhookStatusProcessing,
		WebhookStatusProcessed,
		WebhookStatusFailed:
		return true
	default:
		return false
	}
}

func (w WebhookProcessingStatus) String() string {
	return string(w)
}

func (w WebhookProcessingStatus) Terminal() bool {
	return w == WebhookStatusProcessed || w == WebhookStatusFailed
}

type InvoiceStatus string

const (
	InvoiceStatusOpen          InvoiceStatus = "open"
	InvoiceStatusPaid          InvoiceStatus = "paid"
	InvoiceStatusVoid          InvoiceStatus = "void"
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
	InvoiceStatusRefunded      InvoiceStatus = "refunded"
)

func (s InvoiceStatus) Valid() bool {
	switch s {
	case InvoiceStatusOpen,
		InvoiceStatusPaid,
		InvoiceStatusVoid,
		InvoiceStatusUncollectible,
		InvoiceStatusRefunded:
		return true
	default:
		return false
	}
}

func (s InvoiceStatus) String() string {
	return string(s)
}

func (s InvoiceStatus) Paid() bool {
	return s == InvoiceStatusPaid
}

type SubscriptionEventType string

const (
	SubscriptionEventCreated          SubscriptionEventType = "created"
	SubscriptionEventUpgraded         SubscriptionEventType = "upgraded"
	SubscriptionEventDowngraded       SubscriptionEventType = "downgraded"
	SubscriptionEventRenewed          SubscriptionEventType = "renewed"
	SubscriptionEventCanceled         SubscriptionEventType = "canceled"
	SubscriptionEventResumed          SubscriptionEventType = "resumed"
	SubscriptionEventPaused           SubscriptionEventType = "paused"
	SubscriptionEventTrialStarted     SubscriptionEventType = "trial_started"
	SubscriptionEventTrialEnded       SubscriptionEventType = "trial_ended"
	SubscriptionEventPaymentSucceeded SubscriptionEventType = "payment_succeeded"
	SubscriptionEventPaymentFailed    SubscriptionEventType = "payment_failed"
	SubscriptionEventPaymentRefunded  SubscriptionEventType = "payment_refunded"
	SubscriptionEventExpired          SubscriptionEventType = "expired"
)

func (e SubscriptionEventType) Valid() bool {
	switch e {
	case SubscriptionEventCreated,
		SubscriptionEventUpgraded,
		SubscriptionEventDowngraded,
		SubscriptionEventRenewed,
		SubscriptionEventCanceled,
		SubscriptionEventResumed,
		SubscriptionEventPaused,
		SubscriptionEventTrialStarted,
		SubscriptionEventTrialEnded,
		SubscriptionEventPaymentSucceeded,
		SubscriptionEventPaymentFailed,
		SubscriptionEventPaymentRefunded,
		SubscriptionEventExpired:
		return true
	default:
		return false
	}
}

func (e SubscriptionEventType) String() string {
	return string(e)
}

type DisputeStatus string

const (
	DisputeStatusOpened     DisputeStatus = "opened"
	DisputeStatusExpired    DisputeStatus = "expired"
	DisputeStatusAccepted   DisputeStatus = "accepted"
	DisputeStatusCancelled  DisputeStatus = "cancelled"
	DisputeStatusChallenged DisputeStatus = "challenged"
	DisputeStatusWon        DisputeStatus = "won"
	DisputeStatusLost       DisputeStatus = "lost"
)

func (d DisputeStatus) Valid() bool {
	switch d {
	case DisputeStatusOpened,
		DisputeStatusExpired,
		DisputeStatusAccepted,
		DisputeStatusCancelled,
		DisputeStatusChallenged,
		DisputeStatusWon,
		DisputeStatusLost:
		return true
	default:
		return false
	}
}

func (d DisputeStatus) String() string {
	return string(d)
}

func (d DisputeStatus) Resolved() bool {
	switch d {
	case DisputeStatusWon, DisputeStatusLost, DisputeStatusCancelled, DisputeStatusExpired:
		return true
	default:
		return false
	}
}

type OnboardingGoal string

const (
	OnboardingGoalConfidence     OnboardingGoal = "confidence"
	OnboardingGoalConversations  OnboardingGoal = "conversations"
	OnboardingGoalPublicSpeaking OnboardingGoal = "public_speaking"
	OnboardingGoalWorkInterviews OnboardingGoal = "work_interviews"
	OnboardingGoalVocabulary     OnboardingGoal = "vocabulary"
	OnboardingGoalHabit          OnboardingGoal = "habit"
)

func (g OnboardingGoal) Valid() bool {
	switch g {
	case OnboardingGoalConfidence,
		OnboardingGoalConversations,
		OnboardingGoalPublicSpeaking,
		OnboardingGoalWorkInterviews,
		OnboardingGoalVocabulary,
		OnboardingGoalHabit:
		return true
	default:
		return false
	}
}

func (g OnboardingGoal) String() string {
	return string(g)
}

type FocusArea string

const (
	FocusAreaWorkInterviews       FocusArea = "work_interviews"
	FocusAreaConversations        FocusArea = "conversations"
	FocusAreaCollegePresentations FocusArea = "college_presentations"
	FocusAreaTravel               FocusArea = "travel"
	FocusAreaEverydaySituations   FocusArea = "everyday_situations"
	FocusAreaPublicSpeaking       FocusArea = "public_speaking"
	FocusAreaOnlineMeetings       FocusArea = "online_meetings"
	FocusAreaAnywhere             FocusArea = "anywhere"
)

func (f FocusArea) Valid() bool {
	switch f {
	case FocusAreaWorkInterviews,
		FocusAreaConversations,
		FocusAreaCollegePresentations,
		FocusAreaTravel,
		FocusAreaEverydaySituations,
		FocusAreaPublicSpeaking,
		FocusAreaOnlineMeetings,
		FocusAreaAnywhere:
		return true
	default:
		return false
	}
}

func (f FocusArea) String() string {
	return string(f)
}

type DailyTimeCommitment string

const (
	DailyTimeCommitmentMin5      DailyTimeCommitment = "min_5"
	DailyTimeCommitmentMin10     DailyTimeCommitment = "min_10"
	DailyTimeCommitmentMin15     DailyTimeCommitment = "min_15"
	DailyTimeCommitmentMin30Plus DailyTimeCommitment = "min_30_plus"
)

func (d DailyTimeCommitment) Valid() bool {
	switch d {
	case DailyTimeCommitmentMin5,
		DailyTimeCommitmentMin10,
		DailyTimeCommitmentMin15,
		DailyTimeCommitmentMin30Plus:
		return true
	default:
		return false
	}
}

func (d DailyTimeCommitment) String() string {
	return string(d)
}

type BadgeCriteriaType string

const (
	BadgeCriteriaTypeStreakDays     BadgeCriteriaType = "streak_days"
	BadgeCriteriaTypeSessionCount   BadgeCriteriaType = "session_count"
	BadgeCriteriaTypeXPTotal        BadgeCriteriaType = "xp_total"
	BadgeCriteriaTypeScoreThreshold BadgeCriteriaType = "score_threshold"
	BadgeCriteriaTypeManual         BadgeCriteriaType = "manual"
)

func (b BadgeCriteriaType) Valid() bool {
	switch b {
	case BadgeCriteriaTypeStreakDays,
		BadgeCriteriaTypeSessionCount,
		BadgeCriteriaTypeXPTotal,
		BadgeCriteriaTypeScoreThreshold,
		BadgeCriteriaTypeManual:
		return true
	default:
		return false
	}
}

func (b BadgeCriteriaType) String() string {
	return string(b)
}

func (b BadgeCriteriaType) RequiresValue() bool {
	return b != BadgeCriteriaTypeManual
}

type TopicSubmissionStatus string

const (
	TopicSubmissionStatusPending  TopicSubmissionStatus = "pending"
	TopicSubmissionStatusApproved TopicSubmissionStatus = "approved"
	TopicSubmissionStatusRejected TopicSubmissionStatus = "rejected"
)

func (s TopicSubmissionStatus) Valid() bool {
	switch s {
	case TopicSubmissionStatusPending,
		TopicSubmissionStatusApproved,
		TopicSubmissionStatusRejected:
		return true
	default:
		return false
	}
}

func (s TopicSubmissionStatus) String() string {
	return string(s)
}

func (s TopicSubmissionStatus) Terminal() bool {
	return s == TopicSubmissionStatusApproved ||
		s == TopicSubmissionStatusRejected
}

type NotificationType string

const (
	NotificationTypeDailyReminder NotificationType = "daily_reminder"
	NotificationTypeStreakRisk    NotificationType = "streak_risk"
	NotificationTypeSessionReady  NotificationType = "session_ready"
	NotificationTypeBadgeEarned   NotificationType = "badge_earned"
	NotificationTypeSubscription  NotificationType = "subscription"
)

func (n NotificationType) Valid() bool {
	switch n {
	case NotificationTypeDailyReminder,
		NotificationTypeStreakRisk,
		NotificationTypeSessionReady,
		NotificationTypeBadgeEarned,
		NotificationTypeSubscription:
		return true
	default:
		return false
	}
}

func (n NotificationType) String() string {
	return string(n)
}

type NotificationChannel string

const (
	NotificationChannelInApp NotificationChannel = "in_app"
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelPush  NotificationChannel = "push"
)

func (c NotificationChannel) Valid() bool {
	switch c {
	case NotificationChannelInApp,
		NotificationChannelEmail,
		NotificationChannelPush:
		return true
	default:
		return false
	}
}

func (c NotificationChannel) String() string {
	return string(c)
}

type NotificationDeliveryStatus string

const (
	NotificationDeliveryPending   NotificationDeliveryStatus = "pending"
	NotificationDeliverySent      NotificationDeliveryStatus = "sent"
	NotificationDeliveryDelivered NotificationDeliveryStatus = "delivered"
	NotificationDeliveryFailed    NotificationDeliveryStatus = "failed"
)

func (s NotificationDeliveryStatus) Valid() bool {
	switch s {
	case NotificationDeliveryPending,
		NotificationDeliverySent,
		NotificationDeliveryDelivered,
		NotificationDeliveryFailed:
		return true
	default:
		return false
	}
}

func (s NotificationDeliveryStatus) String() string {
	return string(s)
}

func (s NotificationDeliveryStatus) Terminal() bool {
	return s == NotificationDeliveryDelivered || s == NotificationDeliveryFailed
}

type LeaderboardPeriod string

const (
	LeaderboardPeriodAllTime LeaderboardPeriod = "all_time"
	LeaderboardPeriodWeekly  LeaderboardPeriod = "weekly"
	LeaderboardPeriodMonthly LeaderboardPeriod = "monthly"
)

func (p LeaderboardPeriod) Valid() bool {
	switch p {
	case LeaderboardPeriodAllTime,
		LeaderboardPeriodWeekly,
		LeaderboardPeriodMonthly:
		return true
	default:
		return false
	}
}

func (p LeaderboardPeriod) String() string {
	return string(p)
}

type UserRole string

const (
	UserRoleUser      UserRole = "user"
	UserRoleModerator UserRole = "moderator"
	UserRoleAdmin     UserRole = "admin"
)

func (r UserRole) Valid() bool {
	switch r {
	case UserRoleUser,
		UserRoleModerator,
		UserRoleAdmin:
		return true
	default:
		return false
	}
}

func (r UserRole) String() string {
	return string(r)
}

type TopicSource string

const (
	TopicSourceManual      TopicSource = "manual"
	TopicSourceAIGenerated TopicSource = "ai_generated"
	TopicSourceCommunity   TopicSource = "community"
)

func (s TopicSource) Valid() bool {
	switch s {
	case TopicSourceManual,
		TopicSourceAIGenerated,
		TopicSourceCommunity:
		return true
	default:
		return false
	}
}

func (s TopicSource) String() string {
	return string(s)
}

type AIProviderLogSource string

const (
	AIProviderLogSourceDeepgram AIProviderLogSource = "deepgram"
	AIProviderLogSourceOpenAI   AIProviderLogSource = "openai"
)

func (s AIProviderLogSource) Valid() bool {
	switch s {
	case AIProviderLogSourceDeepgram,
		AIProviderLogSourceOpenAI:
		return true
	default:
		return false
	}
}

func (s AIProviderLogSource) String() string {
	return string(s)
}
