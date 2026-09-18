package models

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var currencyCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)

type SubscriptionPlan struct {
	ID                uuid.UUID        `db:"id" json:"id"`
	Slug              string           `db:"slug" json:"slug"`
	Name              string           `db:"name" json:"name"`
	PlanType          PlanType         `db:"plan_type" json:"planType"`
	BillingInterval   *BillingInterval `db:"billing_interval" json:"billingInterval,omitempty"`
	PriceSubunits     int              `db:"price_subunits" json:"priceSubunits"`
	Currency          string           `db:"currency" json:"currency"`
	DailySessionLimit *int             `db:"daily_session_limit" json:"dailySessionLimit,omitempty"`
	CanViewAnalysis   bool             `db:"can_view_analysis" json:"canViewAnalysis"`
	DodoProductID     *string          `db:"dodo_product_id" json:"dodoProductId,omitempty"`
	IsActive          bool             `db:"is_active" json:"isActive"`
	CreatedAt         time.Time        `db:"created_at" json:"createdAt"`
}

func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}

func (p SubscriptionPlan) Unlimited() bool {
	return p.DailySessionLimit == nil
}

func (p SubscriptionPlan) IsFree() bool {
	return p.PlanType == PlanTypeFree
}

func (p SubscriptionPlan) IsPro() bool {
	return p.PlanType == PlanTypePro
}

func (p SubscriptionPlan) Validate() error {
	var errs []error

	if p.Slug == "" {
		errs = append(errs, errors.New("slug is required"))
	}

	if p.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}

	if !p.PlanType.Valid() {
		errs = append(errs, errors.New("invalid planType"))
	}

	if p.BillingInterval != nil && !p.BillingInterval.Valid() {
		errs = append(errs, errors.New("invalid billingInterval"))
	}

	if p.PriceSubunits < 0 {
		errs = append(errs, errors.New("priceSubunits must be >= 0"))
	}

	if !currencyCodePattern.MatchString(p.Currency) {
		errs = append(errs, errors.New("invalid currency"))
	}

	if p.DailySessionLimit != nil && *p.DailySessionLimit < 0 {
		errs = append(errs, errors.New("dailySessionLimit must be >= 0"))
	}

	switch p.PlanType {
	case PlanTypeFree:
		if p.PriceSubunits != 0 {
			errs = append(
				errs,
				errors.New("free plan priceSubunits must be 0"),
			)
		}

		if p.BillingInterval != nil {
			errs = append(
				errs,
				errors.New("free plan must not have billingInterval"),
			)
		}

	case PlanTypePro:
		if p.PriceSubunits <= 0 {
			errs = append(
				errs,
				errors.New("pro plan priceSubunits must be > 0"),
			)
		}

		if p.BillingInterval == nil {
			errs = append(
				errs,
				errors.New("pro plan requires billingInterval"),
			)
		}
	}

	return errors.Join(errs...)
}

type Subscription struct {
	ID                        uuid.UUID          `db:"id" json:"id"`
	UserID                    string             `db:"user_id" json:"userId"`
	PlanID                    uuid.UUID          `db:"plan_id" json:"planId"`
	Status                    SubscriptionStatus `db:"status" json:"status"`
	DodoSubscriptionID        *string            `db:"dodo_subscription_id" json:"dodoSubscriptionId,omitempty"`
	DodoCustomerID            *string            `db:"dodo_customer_id" json:"dodoCustomerId,omitempty"`
	PriceSubunitsAtPurchase   int                `db:"price_subunits_at_purchase" json:"priceSubunitsAtPurchase"`
	CurrencyAtPurchase        string             `db:"currency_at_purchase" json:"currencyAtPurchase"`
	BillingIntervalAtPurchase *BillingInterval   `db:"billing_interval_at_purchase" json:"billingIntervalAtPurchase,omitempty"`
	CurrentPeriodStart        *time.Time         `db:"current_period_start" json:"currentPeriodStart,omitempty"`
	CurrentPeriodEnd          *time.Time         `db:"current_period_end" json:"currentPeriodEnd,omitempty"`
	CancelAtPeriodEnd         bool               `db:"cancel_at_period_end" json:"cancelAtPeriodEnd"`
	CanceledAt                *time.Time         `db:"canceled_at" json:"canceledAt,omitempty"`
	CreatedAt                 time.Time          `db:"created_at" json:"createdAt"`
	UpdatedAt                 time.Time          `db:"updated_at" json:"updatedAt"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

func (s Subscription) Active() bool {
	return s.Status.IsEntitled()
}

func (s Subscription) Current() bool {
	switch s.Status {
	case SubscriptionStatusActive,
		SubscriptionStatusTrialing,
		SubscriptionStatusPastDue:
		return true
	default:
		return false
	}
}

func (s Subscription) Paused() bool {
	return s.Status == SubscriptionStatusPaused
}

func (s Subscription) Canceled() bool {
	return s.Status == SubscriptionStatusCanceled
}

func (s Subscription) Expired() bool {
	return s.Status == SubscriptionStatusExpired
}

func (s Subscription) CancelingAtPeriodEnd() bool {
	return s.CancelAtPeriodEnd && s.CurrentPeriodEnd != nil
}

func (s Subscription) Validate() error {
	var errs []error

	if s.UserID == "" {
		errs = append(errs, errors.New("userId is required"))
	}

	if s.PlanID == uuid.Nil {
		errs = append(errs, errors.New("planId is required"))
	}

	if !s.Status.Valid() {
		errs = append(errs, errors.New("invalid status"))
	}

	if s.PriceSubunitsAtPurchase < 0 {
		errs = append(
			errs,
			errors.New("priceSubunitsAtPurchase must be >= 0"),
		)
	}

	if !currencyCodePattern.MatchString(s.CurrencyAtPurchase) {
		errs = append(
			errs,
			errors.New("invalid currencyAtPurchase"),
		)
	}

	if s.CurrentPeriodStart != nil &&
		s.CurrentPeriodEnd != nil &&
		s.CurrentPeriodStart.After(*s.CurrentPeriodEnd) {
		errs = append(
			errs,
			errors.New("currentPeriodStart must be before or equal to currentPeriodEnd"),
		)
	}

	if s.CancelAtPeriodEnd && s.CurrentPeriodEnd == nil {
		errs = append(
			errs,
			errors.New("currentPeriodEnd is required when cancelAtPeriodEnd is true"),
		)
	}

	if s.Current() && s.CurrentPeriodEnd == nil {
		errs = append(
			errs,
			errors.New("currentPeriodEnd is required for active subscriptions"),
		)
	}

	if s.Status == SubscriptionStatusCanceled && s.CanceledAt == nil {
		errs = append(
			errs,
			errors.New("canceledAt is required for canceled subscriptions"),
		)
	}

	return errors.Join(errs...)
}

type SubscriptionEvent struct {
	ID             uuid.UUID             `db:"id" json:"id"`
	SubscriptionID uuid.UUID             `db:"subscription_id" json:"subscriptionId"`
	EventType      SubscriptionEventType `db:"event_type" json:"eventType"`
	FromPlanID     *uuid.UUID            `db:"from_plan_id" json:"fromPlanId,omitempty"`
	ToPlanID       *uuid.UUID            `db:"to_plan_id" json:"toPlanId,omitempty"`
	Metadata       JSONB                 `db:"metadata" json:"metadata"`
	CreatedAt      time.Time             `db:"created_at" json:"createdAt"`
}

func (SubscriptionEvent) TableName() string {
	return "subscription_events"
}

func (e *SubscriptionEvent) ApplyDefaults() {
	if e.Metadata.IsNull() {
		e.Metadata = JSONB("{}")
	}
}

func (e SubscriptionEvent) Validate() error {
	var errs []error

	if e.SubscriptionID == uuid.Nil {
		errs = append(
			errs,
			errors.New("subscriptionId is required"),
		)
	}

	if !e.EventType.Valid() {
		errs = append(
			errs,
			errors.New("invalid eventType"),
		)
	}

	if !e.Metadata.IsNull() && !e.Metadata.IsObject() {
		errs = append(
			errs,
			errors.New("metadata must be a JSON object"),
		)
	}

	return errors.Join(errs...)
}

type Invoice struct {
	ID             uuid.UUID     `db:"id" json:"id"`
	SubscriptionID uuid.UUID     `db:"subscription_id" json:"subscriptionId"`
	PlanID         *uuid.UUID    `db:"plan_id" json:"planId,omitempty"`
	DodoInvoiceID  *string       `db:"dodo_invoice_id" json:"dodoInvoiceId,omitempty"`
	AmountSubunits int           `db:"amount_subunits" json:"amountSubunits"`
	Currency       string        `db:"currency" json:"currency"`
	Status         InvoiceStatus `db:"status" json:"status"`
	IssuedAt       time.Time     `db:"issued_at" json:"issuedAt"`
	CreatedAt      time.Time     `db:"created_at" json:"createdAt"`
}

func (Invoice) TableName() string {
	return "invoices"
}

func (i Invoice) Paid() bool {
	return i.Status == InvoiceStatusPaid
}

func (i Invoice) Open() bool {
	return i.Status == InvoiceStatusOpen
}

func (i Invoice) Void() bool {
	return i.Status == InvoiceStatusVoid
}

func (i Invoice) Uncollectible() bool {
	return i.Status == InvoiceStatusUncollectible
}

func (i Invoice) Refunded() bool {
	return i.Status == InvoiceStatusRefunded
}

func (i Invoice) Validate() error {
	var errs []error

	if i.SubscriptionID == uuid.Nil {
		errs = append(
			errs,
			errors.New("subscriptionId is required"),
		)
	}

	if i.AmountSubunits < 0 {
		errs = append(
			errs,
			errors.New("amountSubunits must be >= 0"),
		)
	}

	if !currencyCodePattern.MatchString(i.Currency) {
		errs = append(
			errs,
			errors.New("invalid currency"),
		)
	}

	if !i.Status.Valid() {
		errs = append(
			errs,
			errors.New("invalid status"),
		)
	}

	return errors.Join(errs...)
}

type Payment struct {
	ID                     uuid.UUID     `db:"id" json:"id"`
	UserID                 string        `db:"user_id" json:"userId"`
	InvoiceID              *uuid.UUID    `db:"invoice_id" json:"invoiceId,omitempty"`
	DodoPaymentID          *string       `db:"dodo_payment_id" json:"dodoPaymentId,omitempty"`
	DodoCheckoutSessionID  *string       `db:"dodo_checkout_session_id" json:"dodoCheckoutSessionId,omitempty"`
	AmountSubunits         int           `db:"amount_subunits" json:"amountSubunits"`
	RefundedAmountSubunits int           `db:"refunded_amount_subunits" json:"refundedAmountSubunits"`
	Currency               string        `db:"currency" json:"currency"`
	PaymentMethod          *string       `db:"payment_method" json:"paymentMethod,omitempty"`
	Status                 PaymentStatus `db:"status" json:"status"`
	IdempotencyKey         *string       `db:"idempotency_key" json:"-"`
	PaidAt                 *time.Time    `db:"paid_at" json:"paidAt,omitempty"`
	CreatedAt              time.Time     `db:"created_at" json:"createdAt"`
}

func (Payment) TableName() string {
	return "payments"
}

func (p Payment) Captured() bool {
	return p.Status == PaymentStatusCaptured
}

func (p Payment) Failed() bool {
	return p.Status == PaymentStatusFailed
}

func (p Payment) Refunded() bool {
	return p.Status == PaymentStatusRefunded
}

func (p Payment) PartiallyRefunded() bool {
	return p.RefundedAmountSubunits > 0 && p.Status != PaymentStatusRefunded
}

func (p Payment) Paid() bool {
	return p.PaidAt != nil && p.Status == PaymentStatusCaptured
}

func (p Payment) Validate() error {
	var errs []error

	if p.UserID == "" {
		errs = append(
			errs,
			errors.New("userId is required"),
		)
	}

	if p.AmountSubunits < 0 {
		errs = append(
			errs,
			errors.New("amountSubunits must be >= 0"),
		)
	}

	if p.RefundedAmountSubunits < 0 || p.RefundedAmountSubunits > p.AmountSubunits {
		errs = append(
			errs,
			errors.New("refundedAmountSubunits must be between 0 and amountSubunits"),
		)
	}

	if !currencyCodePattern.MatchString(p.Currency) {
		errs = append(
			errs,
			errors.New("invalid currency"),
		)
	}

	if !p.Status.Valid() {
		errs = append(
			errs,
			errors.New("invalid status"),
		)
	}

	return errors.Join(errs...)
}

type PaymentDispute struct {
	ID             uuid.UUID     `db:"id" json:"id"`
	PaymentID      uuid.UUID     `db:"payment_id" json:"paymentId"`
	DodoDisputeID  string        `db:"dodo_dispute_id" json:"dodoDisputeId"`
	Status         DisputeStatus `db:"status" json:"status"`
	AmountSubunits *int          `db:"amount_subunits" json:"amountSubunits,omitempty"`
	Currency       *string       `db:"currency" json:"currency,omitempty"`
	Reason         *string       `db:"reason" json:"reason,omitempty"`
	Stage          *string       `db:"stage" json:"stage,omitempty"`
	OpenedAt       time.Time     `db:"opened_at" json:"openedAt"`
	ResolvedAt     *time.Time    `db:"resolved_at" json:"resolvedAt,omitempty"`
	CreatedAt      time.Time     `db:"created_at" json:"createdAt"`
}

func (PaymentDispute) TableName() string {
	return "payment_disputes"
}

func (d PaymentDispute) Open() bool {
	return d.Status == DisputeStatusOpened || d.Status == DisputeStatusChallenged
}

func (d PaymentDispute) Lost() bool {
	return d.Status == DisputeStatusLost
}

func (d PaymentDispute) Won() bool {
	return d.Status == DisputeStatusWon
}

func (d PaymentDispute) Validate() error {
	var errs []error

	if d.PaymentID == uuid.Nil {
		errs = append(errs, errors.New("paymentId is required"))
	}

	if d.DodoDisputeID == "" {
		errs = append(errs, errors.New("dodoDisputeId is required"))
	}

	if !d.Status.Valid() {
		errs = append(errs, errors.New("invalid status"))
	}

	if d.AmountSubunits != nil && *d.AmountSubunits < 0 {
		errs = append(errs, errors.New("amountSubunits must be >= 0"))
	}

	if d.Currency != nil && !currencyCodePattern.MatchString(*d.Currency) {
		errs = append(errs, errors.New("invalid currency"))
	}

	return errors.Join(errs...)
}

type DodoWebhookEvent struct {
	ID               uuid.UUID               `db:"id" json:"id"`
	DodoEventID      string                  `db:"dodo_event_id" json:"dodoEventId"`
	EventType        string                  `db:"event_type" json:"eventType"`
	Payload          JSONB                   `db:"payload" json:"payload"`
	ProcessingStatus WebhookProcessingStatus `db:"processing_status" json:"processingStatus"`
	ProcessingError  *string                 `db:"processing_error" json:"processingError,omitempty"`
	ProcessedAt      *time.Time              `db:"processed_at" json:"processedAt,omitempty"`
	CreatedAt        time.Time               `db:"created_at" json:"createdAt"`
}

func (DodoWebhookEvent) TableName() string {
	return "dodo_webhook_events"
}

func (e DodoWebhookEvent) Processed() bool {
	return e.ProcessingStatus == WebhookStatusProcessed
}

func (e DodoWebhookEvent) Failed() bool {
	return e.ProcessingStatus == WebhookStatusFailed
}

func (e DodoWebhookEvent) Pending() bool {
	return e.ProcessingStatus == WebhookStatusReceived ||
		e.ProcessingStatus == WebhookStatusProcessing
}

func (e DodoWebhookEvent) Validate() error {
	var errs []error

	if e.DodoEventID == "" {
		errs = append(
			errs,
			errors.New("dodoEventId is required"),
		)
	}

	if e.EventType == "" {
		errs = append(
			errs,
			errors.New("eventType is required"),
		)
	}

	if e.Payload.IsNull() {
		errs = append(
			errs,
			errors.New("payload is required"),
		)
	}

	if !e.ProcessingStatus.Valid() {
		errs = append(
			errs,
			errors.New("invalid processingStatus"),
		)
	}

	if e.ProcessingStatus == WebhookStatusProcessed &&
		e.ProcessedAt == nil {
		errs = append(
			errs,
			errors.New("processedAt is required for processed events"),
		)
	}

	if e.ProcessingStatus == WebhookStatusFailed &&
		e.ProcessingError == nil {
		errs = append(
			errs,
			errors.New("processingError is required for failed events"),
		)
	}

	return errors.Join(errs...)
}
