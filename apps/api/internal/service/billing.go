package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"flocal/internal/models"
	"flocal/internal/repository"
)

var (
	ErrNoActiveSubscription  = errors.New("service: no active subscription")
	ErrPlanNotPurchasable    = errors.New("service: plan is not purchasable")
	ErrCannotCheckoutFree    = errors.New("service: free plan does not require checkout")
	ErrAlreadySubscribed     = errors.New("service: user already has an active subscription")
	ErrWebhookSignature      = errors.New("service: invalid webhook signature")
	ErrSamePlan              = errors.New("service: subscription is already on this plan")
	ErrSubscriptionNotActive = errors.New("service: subscription is not in a state that allows this action")
	ErrPendingPlanChange     = errors.New("service: a plan change is already pending for this subscription")
)

type BillingService struct {
	billingRepo repository.BillingRepository
	dodo        *DodoClient
	returnURL   string
}

func NewBillingService(billingRepo repository.BillingRepository, dodo *DodoClient, returnURL string) *BillingService {
	return &BillingService{billingRepo: billingRepo, dodo: dodo, returnURL: returnURL}
}

func (s *BillingService) ListPlans(ctx context.Context) ([]models.SubscriptionPlan, error) {
	plans, err := s.billingRepo.ListActivePlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: list plans: %w", err)
	}
	return plans, nil
}

func (s *BillingService) GetMySubscription(ctx context.Context, userID string) (*models.Subscription, error) {
	sub, err := s.billingRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNoActiveSubscription
		}
		return nil, fmt.Errorf("service: get my subscription: %w", err)
	}
	return sub, nil
}

func (s *BillingService) ListMyPayments(ctx context.Context, userID string, limit, offset int) ([]models.Payment, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	payments, err := s.billingRepo.ListPaymentsForUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service: list my payments: %w", err)
	}
	return payments, nil
}

func (s *BillingService) CreateCheckoutSession(ctx context.Context, userID, userEmail, userName, planSlug, billingCountry string) (*CheckoutSession, error) {
	plan, err := s.billingRepo.GetPlanBySlug(ctx, planSlug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrPlanNotFound
		}
		return nil, fmt.Errorf("service: get plan by slug: %w", err)
	}

	if plan.IsFree() {
		return nil, ErrCannotCheckoutFree
	}

	if !plan.IsActive || plan.DodoProductID == nil {
		return nil, ErrPlanNotPurchasable
	}

	_, err = s.billingRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err == nil {
		return nil, ErrAlreadySubscribed
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("service: check existing subscription: %w", err)
	}

	session, err := s.dodo.CreateCheckoutSession(ctx, CreateCheckoutSessionInput{
		ProductID:       *plan.DodoProductID,
		CustomerEmail:   userEmail,
		CustomerName:    userName,
		ReturnURL:       s.returnURL,
		BillingCurrency: plan.Currency,
		BillingCountry:  billingCountry,
		Metadata: map[string]string{
			"user_id": userID,
			"plan_id": plan.ID.String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("service: create checkout session: %w", err)
	}

	return session, nil
}

func (s *BillingService) CancelMySubscription(ctx context.Context, userID string) error {
	sub, err := s.billingRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNoActiveSubscription
		}
		return fmt.Errorf("service: cancel subscription: %w", err)
	}

	if sub.DodoSubscriptionID == nil {
		return fmt.Errorf("service: cancel subscription: subscription has no dodo subscription id")
	}

	if _, err := s.dodo.SetCancelAtNextBillingDate(ctx, *sub.DodoSubscriptionID, true); err != nil {
		return fmt.Errorf("service: cancel subscription at dodo: %w", err)
	}

	if err := s.billingRepo.SetSubscriptionCancelAtPeriodEnd(ctx, sub.ID, true); err != nil {
		return fmt.Errorf("service: cancel subscription: %w", err)
	}

	if err := s.billingRepo.CreateSubscriptionEvent(ctx, &models.SubscriptionEvent{
		SubscriptionID: sub.ID,
		EventType:      models.SubscriptionEventCanceled,
	}); err != nil {
		return fmt.Errorf("service: log cancel subscription event: %w", err)
	}

	return nil
}

func (s *BillingService) ResumeMySubscription(ctx context.Context, userID string) error {
	sub, err := s.billingRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNoActiveSubscription
		}
		return fmt.Errorf("service: resume subscription: %w", err)
	}

	if !sub.CancelAtPeriodEnd {
		return nil
	}

	if sub.DodoSubscriptionID == nil {
		return fmt.Errorf("service: resume subscription: subscription has no dodo subscription id")
	}

	if _, err := s.dodo.SetCancelAtNextBillingDate(ctx, *sub.DodoSubscriptionID, false); err != nil {
		return fmt.Errorf("service: resume subscription at dodo: %w", err)
	}

	if err := s.billingRepo.SetSubscriptionCancelAtPeriodEnd(ctx, sub.ID, false); err != nil {
		return fmt.Errorf("service: resume subscription: %w", err)
	}

	if err := s.billingRepo.CreateSubscriptionEvent(ctx, &models.SubscriptionEvent{
		SubscriptionID: sub.ID,
		EventType:      models.SubscriptionEventResumed,
	}); err != nil {
		return fmt.Errorf("service: log resume subscription event: %w", err)
	}

	return nil
}

func (s *BillingService) ChangeMyPlan(ctx context.Context, userID, newPlanSlug string) error {
	sub, err := s.billingRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNoActiveSubscription
		}
		return fmt.Errorf("service: change plan: %w", err)
	}

	if !sub.Status.IsEntitled() {
		return ErrSubscriptionNotActive
	}

	if sub.DodoSubscriptionID == nil {
		return fmt.Errorf("service: change plan: subscription has no dodo subscription id")
	}

	newPlan, err := s.billingRepo.GetPlanBySlug(ctx, newPlanSlug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrPlanNotFound
		}
		return fmt.Errorf("service: change plan: %w", err)
	}

	if newPlan.IsFree() || !newPlan.IsActive || newPlan.DodoProductID == nil {
		return ErrPlanNotPurchasable
	}

	if sub.PlanID == newPlan.ID {
		return ErrSamePlan
	}

	if err := s.dodo.ChangePlan(ctx, ChangePlanInput{
		SubscriptionID: *sub.DodoSubscriptionID,
		ProductID:      *newPlan.DodoProductID,
	}); err != nil {
		switch {
		case errors.Is(err, ErrDodoConflict):
			return ErrPendingPlanChange
		case errors.Is(err, ErrDodoUnprocessable):
			return ErrSubscriptionNotActive
		default:
			return fmt.Errorf("service: change plan at dodo: %w", err)
		}
	}

	return nil
}

func (s *BillingService) UpdateMyPaymentMethod(ctx context.Context, userID string) (*UpdatePaymentMethodResult, error) {
	sub, err := s.billingRepo.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNoActiveSubscription
		}
		return nil, fmt.Errorf("service: update payment method: %w", err)
	}

	if sub.DodoSubscriptionID == nil {
		return nil, fmt.Errorf("service: update payment method: subscription has no dodo subscription id")
	}

	result, err := s.dodo.InitiatePaymentMethodUpdate(ctx, *sub.DodoSubscriptionID, s.returnURL)
	if err != nil {
		return nil, fmt.Errorf("service: update payment method at dodo: %w", err)
	}

	return result, nil
}

type dodoWebhookPayload struct {
	Type string `json:"type"`
	Data struct {
		PayloadType        string            `json:"payload_type"`
		SubscriptionID     string            `json:"subscription_id"`
		PaymentID          string            `json:"payment_id"`
		CustomerID         string            `json:"customer_id"`
		CheckoutSessionID  string            `json:"checkout_session_id"`
		TotalAmount        int               `json:"total_amount"`
		Currency           string            `json:"currency"`
		CurrentPeriodStart *time.Time        `json:"current_period_start"`
		CurrentPeriodEnd   *time.Time        `json:"current_period_end"`
		Metadata           map[string]string `json:"metadata"`
		RefundID           string            `json:"refund_id"`
		IsPartialRefund    bool              `json:"is_partial"`
		DisputeID          string            `json:"dispute_id"`
		DisputeStage       string            `json:"dispute_stage"`
		Amount             json.RawMessage   `json:"amount"`
		Remarks            string            `json:"remarks"`
	} `json:"data"`
}

func (s *BillingService) HandleWebhook(ctx context.Context, webhookID, timestamp, signature string, body []byte) error {
	if err := s.dodo.VerifyWebhookSignature(webhookID, timestamp, signature, body); err != nil {
		return ErrWebhookSignature
	}

	existing, err := s.billingRepo.GetWebhookEventByDodoEventID(ctx, webhookID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("service: check webhook event: %w", err)
	}
	if existing != nil && existing.Processed() {
		return nil
	}

	var payload dodoWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("service: parse webhook payload: %w", err)
	}

	eventID := uuid.New()
	if existing != nil {
		eventID = existing.ID
	} else {
		event := &models.DodoWebhookEvent{
			ID:               eventID,
			DodoEventID:      webhookID,
			EventType:        payload.Type,
			Payload:          models.JSONB(body),
			ProcessingStatus: models.WebhookStatusProcessing,
		}
		if err := s.billingRepo.CreateWebhookEvent(ctx, event); err != nil {
			return fmt.Errorf("service: store webhook event: %w", err)
		}
	}

	if err := s.applyWebhookEvent(ctx, payload); err != nil {
		errMsg := err.Error()
		_ = s.billingRepo.UpdateWebhookProcessingStatus(ctx, eventID, models.WebhookStatusFailed, &errMsg)
		return fmt.Errorf("service: apply webhook event: %w", err)
	}

	if err := s.billingRepo.MarkWebhookProcessed(ctx, eventID); err != nil {
		return fmt.Errorf("service: mark webhook processed: %w", err)
	}

	return nil
}

func (s *BillingService) applyWebhookEvent(ctx context.Context, payload dodoWebhookPayload) error {
	switch payload.Type {
	case "subscription.active", "subscription.renewed":
		return s.upsertSubscriptionFromWebhook(ctx, payload, models.SubscriptionStatusActive)
	case "subscription.on_hold":
		return s.finalizeSubscriptionStatus(ctx, payload.Data.SubscriptionID, models.SubscriptionStatusPastDue)
	case "subscription.cancelled":
		return s.finalizeSubscriptionStatus(ctx, payload.Data.SubscriptionID, models.SubscriptionStatusCanceled)
	case "subscription.expired":
		return s.finalizeSubscriptionStatus(ctx, payload.Data.SubscriptionID, models.SubscriptionStatusExpired)
	case "subscription.failed":
		return nil
	case "subscription.plan_changed":
		return s.resyncSubscriptionFromDodo(ctx, payload.Data.SubscriptionID)
	case "subscription.paused":
		return s.pauseSubscription(ctx, payload.Data.SubscriptionID)
	case "subscription.unpaused":
		return s.unpauseSubscription(ctx, payload.Data.SubscriptionID)
	case "payment.succeeded":
		return s.recordPayment(ctx, payload, models.PaymentStatusCaptured)
	case "payment.failed":
		return s.recordPayment(ctx, payload, models.PaymentStatusFailed)
	case "refund.succeeded":
		return s.refundPayment(ctx, payload)
	case "refund.failed":
		return nil
	case "dispute.opened":
		return s.upsertDispute(ctx, payload, models.DisputeStatusOpened)
	case "dispute.accepted":
		return s.upsertDispute(ctx, payload, models.DisputeStatusAccepted)
	case "dispute.challenged":
		return s.upsertDispute(ctx, payload, models.DisputeStatusChallenged)
	case "dispute.cancelled":
		return s.upsertDispute(ctx, payload, models.DisputeStatusCancelled)
	case "dispute.expired":
		return s.upsertDispute(ctx, payload, models.DisputeStatusExpired)
	case "dispute.won":
		return s.upsertDispute(ctx, payload, models.DisputeStatusWon)
	case "dispute.lost":
		return s.upsertDispute(ctx, payload, models.DisputeStatusLost)
	default:
		return nil
	}
}

func monthlyEquivalentSubunits(priceSubunits int64, interval models.BillingInterval) int64 {
	if interval == models.BillingIntervalAnnual {
		return priceSubunits / 12
	}
	return priceSubunits
}

func (s *BillingService) resyncSubscriptionFromDodo(ctx context.Context, dodoSubscriptionID string) error {
	if dodoSubscriptionID == "" {
		return errors.New("webhook missing subscription_id for plan change resync")
	}

	sub, err := s.billingRepo.GetSubscriptionByDodoSubscriptionID(ctx, dodoSubscriptionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("lookup subscription for resync: %w", err)
	}

	remote, err := s.dodo.GetSubscription(ctx, dodoSubscriptionID)
	if err != nil {
		return fmt.Errorf("fetch subscription from dodo: %w", err)
	}

	newPlan, err := s.billingRepo.GetPlanByDodoProductID(ctx, remote.ProductID)
	if err != nil {
		return fmt.Errorf("lookup plan for product %s: %w", remote.ProductID, err)
	}

	oldInterval := models.BillingIntervalMonthly
	if sub.BillingIntervalAtPurchase != nil {
		oldInterval = *sub.BillingIntervalAtPurchase
	}
	newInterval := models.BillingIntervalMonthly
	if newPlan.BillingInterval != nil {
		newInterval = *newPlan.BillingInterval
	}

	oldMonthly := monthlyEquivalentSubunits(int64(sub.PriceSubunitsAtPurchase), oldInterval)
	newMonthly := monthlyEquivalentSubunits(remote.RecurringPreTaxAmount, newInterval)

	eventType := models.SubscriptionEventUpgraded
	if newMonthly < oldMonthly {
		eventType = models.SubscriptionEventDowngraded
	}

	fromPlanID := sub.PlanID
	periodStart := remote.PreviousBillingDate
	periodEnd := remote.NextBillingDate

	if err := s.billingRepo.ApplyPlanChange(
		ctx, sub.ID, newPlan.ID, int(remote.RecurringPreTaxAmount), string(remote.Currency),
		newPlan.BillingInterval, &periodStart, &periodEnd,
	); err != nil {
		return fmt.Errorf("apply plan change: %w", err)
	}

	if newPlan.ID == fromPlanID {
		return nil
	}

	return s.billingRepo.CreateSubscriptionEvent(ctx, &models.SubscriptionEvent{
		SubscriptionID: sub.ID,
		EventType:      eventType,
		FromPlanID:     &fromPlanID,
		ToPlanID:       &newPlan.ID,
	})
}

func (s *BillingService) upsertSubscriptionFromWebhook(ctx context.Context, payload dodoWebhookPayload, status models.SubscriptionStatus) error {
	sub, err := s.billingRepo.GetSubscriptionByDodoSubscriptionID(ctx, payload.Data.SubscriptionID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("lookup subscription: %w", err)
	}

	if sub != nil {
		return s.billingRepo.UpdateSubscriptionStatus(ctx, sub.ID, status, payload.Data.CurrentPeriodStart, payload.Data.CurrentPeriodEnd)
	}

	userID := payload.Data.Metadata["user_id"]
	planIDStr := payload.Data.Metadata["plan_id"]
	if userID == "" || planIDStr == "" {
		return errors.New("webhook missing user_id/plan_id metadata for new subscription")
	}

	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		return fmt.Errorf("parse plan id from metadata: %w", err)
	}

	plan, err := s.billingRepo.GetPlanByID(ctx, planID)
	if err != nil {
		return fmt.Errorf("load plan for new subscription: %w", err)
	}

	dodoSubID := payload.Data.SubscriptionID
	dodoCustID := payload.Data.CustomerID

	newSub := &models.Subscription{
		ID:                        uuid.New(),
		UserID:                    userID,
		PlanID:                    plan.ID,
		Status:                    status,
		DodoSubscriptionID:        nilIfEmpty(dodoSubID),
		DodoCustomerID:            nilIfEmpty(dodoCustID),
		PriceSubunitsAtPurchase:   plan.PriceSubunits,
		CurrencyAtPurchase:        plan.Currency,
		BillingIntervalAtPurchase: plan.BillingInterval,
		CurrentPeriodStart:        payload.Data.CurrentPeriodStart,
		CurrentPeriodEnd:          payload.Data.CurrentPeriodEnd,
	}

	if err := s.billingRepo.CreateSubscription(ctx, newSub); err != nil {
		return fmt.Errorf("create subscription from webhook: %w", err)
	}

	return s.billingRepo.CreateSubscriptionEvent(ctx, &models.SubscriptionEvent{
		SubscriptionID: newSub.ID,
		EventType:      models.SubscriptionEventCreated,
		ToPlanID:       &plan.ID,
	})
}

func (s *BillingService) finalizeSubscriptionStatus(ctx context.Context, dodoSubscriptionID string, status models.SubscriptionStatus) error {
	sub, err := s.billingRepo.GetSubscriptionByDodoSubscriptionID(ctx, dodoSubscriptionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("lookup subscription: %w", err)
	}

	if status == models.SubscriptionStatusCanceled {
		return s.billingRepo.FinalizeSubscriptionCancellation(ctx, sub.ID, time.Now())
	}

	if err := s.billingRepo.UpdateSubscriptionStatus(ctx, sub.ID, status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd); err != nil {
		return err
	}

	if status == models.SubscriptionStatusExpired {
		return s.billingRepo.CreateSubscriptionEvent(ctx, &models.SubscriptionEvent{
			SubscriptionID: sub.ID,
			EventType:      models.SubscriptionEventExpired,
		})
	}

	return nil
}

func (s *BillingService) pauseSubscription(ctx context.Context, dodoSubscriptionID string) error {
	if dodoSubscriptionID == "" {
		return nil
	}

	sub, err := s.billingRepo.GetSubscriptionByDodoSubscriptionID(ctx, dodoSubscriptionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("lookup subscription for pause: %w", err)
	}

	if err := s.billingRepo.UpdateSubscriptionStatus(ctx, sub.ID, models.SubscriptionStatusPaused, sub.CurrentPeriodStart, sub.CurrentPeriodEnd); err != nil {
		return fmt.Errorf("pause subscription: %w", err)
	}

	return s.billingRepo.CreateSubscriptionEvent(ctx, &models.SubscriptionEvent{
		SubscriptionID: sub.ID,
		EventType:      models.SubscriptionEventPaused,
	})
}

func (s *BillingService) unpauseSubscription(ctx context.Context, dodoSubscriptionID string) error {
	if dodoSubscriptionID == "" {
		return nil
	}

	sub, err := s.billingRepo.GetSubscriptionByDodoSubscriptionID(ctx, dodoSubscriptionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("lookup subscription for unpause: %w", err)
	}

	remote, err := s.dodo.GetSubscription(ctx, dodoSubscriptionID)
	if err != nil {
		return fmt.Errorf("fetch subscription from dodo: %w", err)
	}

	periodStart := remote.PreviousBillingDate
	periodEnd := remote.NextBillingDate

	if err := s.billingRepo.UpdateSubscriptionStatus(ctx, sub.ID, models.SubscriptionStatusActive, &periodStart, &periodEnd); err != nil {
		return fmt.Errorf("unpause subscription: %w", err)
	}

	return s.billingRepo.CreateSubscriptionEvent(ctx, &models.SubscriptionEvent{
		SubscriptionID: sub.ID,
		EventType:      models.SubscriptionEventResumed,
	})
}

func (s *BillingService) recordPayment(ctx context.Context, payload dodoWebhookPayload, status models.PaymentStatus) error {
	if payload.Data.PaymentID == "" {
		return nil
	}

	_, err := s.billingRepo.GetPaymentByDodoPaymentID(ctx, payload.Data.PaymentID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("lookup payment: %w", err)
	}

	var sub *models.Subscription
	if payload.Data.SubscriptionID != "" {
		found, subErr := s.billingRepo.GetSubscriptionByDodoSubscriptionID(ctx, payload.Data.SubscriptionID)
		if subErr == nil {
			sub = found
		} else if !errors.Is(subErr, repository.ErrNotFound) {
			return fmt.Errorf("lookup subscription for payment: %w", subErr)
		}
	}

	userID := payload.Data.Metadata["user_id"]
	if userID == "" && sub != nil {
		userID = sub.UserID
	}
	if userID == "" {
		return errors.New("webhook missing user_id for payment")
	}

	currency := payload.Data.Currency
	if currency == "" {
		currency = "USD"
	}

	payment := &models.Payment{
		UserID:                userID,
		DodoPaymentID:         nilIfEmpty(payload.Data.PaymentID),
		DodoCheckoutSessionID: nilIfEmpty(payload.Data.CheckoutSessionID),
		AmountSubunits:        payload.Data.TotalAmount,
		Currency:              currency,
		Status:                status,
	}

	if status == models.PaymentStatusCaptured {
		now := time.Now()
		payment.PaidAt = &now

		if sub != nil {
			planID := sub.PlanID
			inv := &models.Invoice{
				SubscriptionID: sub.ID,
				PlanID:         &planID,
				AmountSubunits: payload.Data.TotalAmount,
				Currency:       currency,
				Status:         models.InvoiceStatusPaid,
				IssuedAt:       now,
			}
			if err := s.billingRepo.CreateInvoice(ctx, inv); err != nil {
				return fmt.Errorf("create invoice for payment: %w", err)
			}
			payment.InvoiceID = &inv.ID
		}
	}

	if err := s.billingRepo.CreatePayment(ctx, payment); err != nil {
		return err
	}

	if sub == nil {
		return nil
	}

	eventType := models.SubscriptionEventPaymentSucceeded
	if status == models.PaymentStatusFailed {
		eventType = models.SubscriptionEventPaymentFailed
	}

	return s.billingRepo.CreateSubscriptionEvent(ctx, &models.SubscriptionEvent{
		SubscriptionID: sub.ID,
		EventType:      eventType,
	})
}

func (s *BillingService) refundPayment(ctx context.Context, payload dodoWebhookPayload) error {
	if payload.Data.PaymentID == "" || payload.Data.RefundID == "" {
		return nil
	}

	amount := parseAmountSubunits(payload.Data.Amount)
	currency := nilIfEmpty(payload.Data.Currency)
	reason := nilIfEmpty(payload.Data.Remarks)

	payment, err := s.billingRepo.ApplyRefund(ctx, payload.Data.PaymentID, payload.Data.RefundID, amount, currency, payload.Data.IsPartialRefund, reason)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("apply refund: %w", err)
	}

	if payment.Status == models.PaymentStatusRefunded && payment.InvoiceID != nil {
		if err := s.billingRepo.UpdateInvoiceStatus(ctx, *payment.InvoiceID, models.InvoiceStatusRefunded); err != nil && !errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("mark invoice refunded: %w", err)
		}
	}

	if payload.Data.SubscriptionID == "" {
		return nil
	}

	sub, err := s.billingRepo.GetSubscriptionByDodoSubscriptionID(ctx, payload.Data.SubscriptionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("lookup subscription for refund: %w", err)
	}

	return s.billingRepo.CreateSubscriptionEvent(ctx, &models.SubscriptionEvent{
		SubscriptionID: sub.ID,
		EventType:      models.SubscriptionEventPaymentRefunded,
	})
}

func (s *BillingService) upsertDispute(ctx context.Context, payload dodoWebhookPayload, status models.DisputeStatus) error {
	if payload.Data.DisputeID == "" || payload.Data.PaymentID == "" {
		return nil
	}

	payment, err := s.billingRepo.GetPaymentByDodoPaymentID(ctx, payload.Data.PaymentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("lookup payment for dispute: %w", err)
	}

	disputeAmount := parseAmountSubunits(payload.Data.Amount)
	disputeCurrency := nilIfEmpty(payload.Data.Currency)
	if disputeCurrency == nil {
		disputeCurrency = nilIfEmpty(payment.Currency)
	}

	existing, err := s.billingRepo.GetDisputeByDodoDisputeID(ctx, payload.Data.DisputeID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("lookup dispute: %w", err)
	}

	if existing == nil {
		dispute := &models.PaymentDispute{
			PaymentID:      payment.ID,
			DodoDisputeID:  payload.Data.DisputeID,
			Status:         status,
			AmountSubunits: disputeAmount,
			Currency:       disputeCurrency,
			Reason:         nilIfEmpty(payload.Data.Remarks),
			Stage:          nilIfEmpty(payload.Data.DisputeStage),
		}
		if err := s.billingRepo.CreateDispute(ctx, dispute); err != nil {
			return fmt.Errorf("create dispute: %w", err)
		}
	} else {
		var resolvedAt *time.Time
		if status.Resolved() {
			now := time.Now()
			resolvedAt = &now
		}
		if err := s.billingRepo.UpdateDisputeStatus(ctx, existing.ID, status, resolvedAt); err != nil {
			return fmt.Errorf("update dispute status: %w", err)
		}
	}

	if status != models.DisputeStatusLost {
		return nil
	}

	isPartial := disputeAmount != nil && *disputeAmount < payment.AmountSubunits

	updated, err := s.billingRepo.ApplyRefund(ctx, payload.Data.PaymentID, "dispute:"+payload.Data.DisputeID, disputeAmount, disputeCurrency, isPartial, nilIfEmpty(payload.Data.Remarks))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("apply refund for lost dispute: %w", err)
	}

	if updated.Status == models.PaymentStatusRefunded && updated.InvoiceID != nil {
		if err := s.billingRepo.UpdateInvoiceStatus(ctx, *updated.InvoiceID, models.InvoiceStatusRefunded); err != nil && !errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("mark invoice refunded after lost dispute: %w", err)
		}
	}

	if payload.Data.SubscriptionID == "" {
		return nil
	}

	sub, err := s.billingRepo.GetSubscriptionByDodoSubscriptionID(ctx, payload.Data.SubscriptionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("lookup subscription for lost dispute: %w", err)
	}

	return s.billingRepo.CreateSubscriptionEvent(ctx, &models.SubscriptionEvent{
		SubscriptionID: sub.ID,
		EventType:      models.SubscriptionEventPaymentRefunded,
	})
}

func parseAmountSubunits(raw json.RawMessage) *int {
	if len(raw) == 0 {
		return nil
	}

	var asInt int
	if err := json.Unmarshal(raw, &asInt); err == nil {
		return &asInt
	}

	var asFloat float64
	if err := json.Unmarshal(raw, &asFloat); err == nil {
		n := int(asFloat)
		return &n
	}

	var asStr string
	if err := json.Unmarshal(raw, &asStr); err == nil {
		if n, err := strconv.Atoi(asStr); err == nil {
			return &n
		}
		if f, err := strconv.ParseFloat(asStr, 64); err == nil {
			n := int(f)
			return &n
		}
	}

	return nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
