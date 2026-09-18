package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"flocal/internal/models"
)

type BillingRepository interface {
	ListActivePlans(ctx context.Context) ([]models.SubscriptionPlan, error)
	GetPlanByID(ctx context.Context, id uuid.UUID) (*models.SubscriptionPlan, error)
	GetPlanBySlug(ctx context.Context, slug string) (*models.SubscriptionPlan, error)
	GetPlanByDodoProductID(ctx context.Context, dodoProductID string) (*models.SubscriptionPlan, error)

	CreateSubscription(ctx context.Context, s *models.Subscription) error
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	GetActiveSubscriptionForUser(ctx context.Context, userID string) (*models.Subscription, error)
	GetSubscriptionByDodoSubscriptionID(ctx context.Context, dodoSubscriptionID string) (*models.Subscription, error)
	UpdateSubscriptionStatus(ctx context.Context, id uuid.UUID, status models.SubscriptionStatus, currentPeriodStart, currentPeriodEnd *time.Time) error
	SetSubscriptionCancelAtPeriodEnd(ctx context.Context, id uuid.UUID, cancelAtPeriodEnd bool) error
	FinalizeSubscriptionCancellation(ctx context.Context, id uuid.UUID, canceledAt time.Time) error
	ApplyPlanChange(ctx context.Context, id uuid.UUID, planID uuid.UUID, priceSubunits int, currency string, billingInterval *models.BillingInterval, currentPeriodStart, currentPeriodEnd *time.Time) error

	CreateSubscriptionEvent(ctx context.Context, e *models.SubscriptionEvent) error

	CreateInvoice(ctx context.Context, inv *models.Invoice) error
	GetInvoiceByDodoInvoiceID(ctx context.Context, dodoInvoiceID string) (*models.Invoice, error)
	UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status models.InvoiceStatus) error

	CreatePayment(ctx context.Context, p *models.Payment) error
	GetPaymentByDodoPaymentID(ctx context.Context, dodoPaymentID string) (*models.Payment, error)
	ApplyRefund(ctx context.Context, dodoPaymentID, dodoRefundID string, amountSubunits *int, currency *string, isPartial bool, reason *string) (*models.Payment, error)
	ListPaymentsForUser(ctx context.Context, userID string, limit, offset int) ([]models.Payment, error)

	CreateDispute(ctx context.Context, d *models.PaymentDispute) error
	GetDisputeByDodoDisputeID(ctx context.Context, dodoDisputeID string) (*models.PaymentDispute, error)
	UpdateDisputeStatus(ctx context.Context, id uuid.UUID, status models.DisputeStatus, resolvedAt *time.Time) error

	CreateWebhookEvent(ctx context.Context, e *models.DodoWebhookEvent) error
	GetWebhookEventByDodoEventID(ctx context.Context, dodoEventID string) (*models.DodoWebhookEvent, error)
	UpdateWebhookProcessingStatus(ctx context.Context, id uuid.UUID, status models.WebhookProcessingStatus, processingError *string) error
	MarkWebhookProcessed(ctx context.Context, id uuid.UUID) error

	DeleteAllForUser(ctx context.Context, userID string) error
}

type pgBillingRepository struct {
	pool *pgxpool.Pool
}

func NewBillingRepository(pool *pgxpool.Pool) BillingRepository {
	return &pgBillingRepository{pool: pool}
}

func (r *pgBillingRepository) ListActivePlans(ctx context.Context) ([]models.SubscriptionPlan, error) {
	const q = `
		select id, slug, name, plan_type, billing_interval, price_subunits, currency,
		       daily_session_limit, can_view_analysis, dodo_product_id, is_active, created_at
		from subscription_plans
		where is_active = true
		order by price_subunits asc`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("repository: list active plans: %w", err)
	}
	defer rows.Close()

	plans := make([]models.SubscriptionPlan, 0)
	for rows.Next() {
		var p models.SubscriptionPlan
		if err := rows.Scan(
			&p.ID, &p.Slug, &p.Name, &p.PlanType, &p.BillingInterval, &p.PriceSubunits, &p.Currency,
			&p.DailySessionLimit, &p.CanViewAnalysis, &p.DodoProductID, &p.IsActive, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan plan: %w", err)
		}
		plans = append(plans, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list active plans: %w", err)
	}

	return plans, nil
}

func (r *pgBillingRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*models.SubscriptionPlan, error) {
	const q = `
		select id, slug, name, plan_type, billing_interval, price_subunits, currency,
		       daily_session_limit, can_view_analysis, dodo_product_id, is_active, created_at
		from subscription_plans
		where id = $1`

	var p models.SubscriptionPlan
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.Slug, &p.Name, &p.PlanType, &p.BillingInterval, &p.PriceSubunits, &p.Currency,
		&p.DailySessionLimit, &p.CanViewAnalysis, &p.DodoProductID, &p.IsActive, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get plan by id: %w", err)
	}

	return &p, nil
}

func (r *pgBillingRepository) GetPlanBySlug(ctx context.Context, slug string) (*models.SubscriptionPlan, error) {
	const q = `
		select id, slug, name, plan_type, billing_interval, price_subunits, currency,
		       daily_session_limit, can_view_analysis, dodo_product_id, is_active, created_at
		from subscription_plans
		where slug = $1`

	var p models.SubscriptionPlan
	err := r.pool.QueryRow(ctx, q, slug).Scan(
		&p.ID, &p.Slug, &p.Name, &p.PlanType, &p.BillingInterval, &p.PriceSubunits, &p.Currency,
		&p.DailySessionLimit, &p.CanViewAnalysis, &p.DodoProductID, &p.IsActive, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get plan by slug: %w", err)
	}

	return &p, nil
}

func (r *pgBillingRepository) GetPlanByDodoProductID(ctx context.Context, dodoProductID string) (*models.SubscriptionPlan, error) {
	const q = `
		select id, slug, name, plan_type, billing_interval, price_subunits, currency,
		       daily_session_limit, can_view_analysis, dodo_product_id, is_active, created_at
		from subscription_plans
		where dodo_product_id = $1`

	var p models.SubscriptionPlan
	err := r.pool.QueryRow(ctx, q, dodoProductID).Scan(
		&p.ID, &p.Slug, &p.Name, &p.PlanType, &p.BillingInterval, &p.PriceSubunits, &p.Currency,
		&p.DailySessionLimit, &p.CanViewAnalysis, &p.DodoProductID, &p.IsActive, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get plan by dodo product id: %w", err)
	}

	return &p, nil
}

func (r *pgBillingRepository) CreateSubscription(ctx context.Context, s *models.Subscription) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	const q = `
		insert into subscriptions (
			id, user_id, plan_id, status, dodo_subscription_id, dodo_customer_id,
			price_subunits_at_purchase, currency_at_purchase, billing_interval_at_purchase,
			current_period_start, current_period_end, cancel_at_period_end, canceled_at,
			created_at, updated_at
		)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,now(),now())
		returning created_at, updated_at`

	err := r.pool.QueryRow(ctx, q,
		s.ID, s.UserID, s.PlanID, s.Status, s.DodoSubscriptionID, s.DodoCustomerID,
		s.PriceSubunitsAtPurchase, s.CurrencyAtPurchase, s.BillingIntervalAtPurchase,
		s.CurrentPeriodStart, s.CurrentPeriodEnd, s.CancelAtPeriodEnd, s.CanceledAt,
	).Scan(&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository: create subscription: %w", err)
	}

	return nil
}

func (r *pgBillingRepository) scanSubscription(row pgx.Row) (*models.Subscription, error) {
	var s models.Subscription
	err := row.Scan(
		&s.ID, &s.UserID, &s.PlanID, &s.Status, &s.DodoSubscriptionID, &s.DodoCustomerID,
		&s.PriceSubunitsAtPurchase, &s.CurrencyAtPurchase, &s.BillingIntervalAtPurchase,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CancelAtPeriodEnd, &s.CanceledAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &s, nil
}

const subscriptionColumns = `
	id, user_id, plan_id, status, dodo_subscription_id, dodo_customer_id,
	price_subunits_at_purchase, currency_at_purchase, billing_interval_at_purchase,
	current_period_start, current_period_end, cancel_at_period_end, canceled_at,
	created_at, updated_at`

func (r *pgBillingRepository) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	q := fmt.Sprintf(`select %s from subscriptions where id = $1`, subscriptionColumns)

	s, err := r.scanSubscription(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get subscription by id: %w", err)
	}

	return s, nil
}

func (r *pgBillingRepository) GetActiveSubscriptionForUser(ctx context.Context, userID string) (*models.Subscription, error) {
	q := fmt.Sprintf(`
		select %s
		from subscriptions
		where user_id = $1 and status in ('active', 'trialing', 'past_due')
		order by current_period_end desc nulls last
		limit 1`, subscriptionColumns)

	s, err := r.scanSubscription(r.pool.QueryRow(ctx, q, userID))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get active subscription for user: %w", err)
	}

	return s, nil
}

func (r *pgBillingRepository) GetSubscriptionByDodoSubscriptionID(ctx context.Context, dodoSubscriptionID string) (*models.Subscription, error) {
	q := fmt.Sprintf(`select %s from subscriptions where dodo_subscription_id = $1`, subscriptionColumns)

	s, err := r.scanSubscription(r.pool.QueryRow(ctx, q, dodoSubscriptionID))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get subscription by dodo subscription id: %w", err)
	}

	return s, nil
}

func (r *pgBillingRepository) UpdateSubscriptionStatus(ctx context.Context, id uuid.UUID, status models.SubscriptionStatus, currentPeriodStart, currentPeriodEnd *time.Time) error {
	const q = `
		update subscriptions
		set status = $2, current_period_start = $3, current_period_end = $4, updated_at = now()
		where id = $1`

	result, err := r.pool.Exec(ctx, q, id, status, currentPeriodStart, currentPeriodEnd)
	if err != nil {
		return fmt.Errorf("repository: update subscription status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgBillingRepository) SetSubscriptionCancelAtPeriodEnd(ctx context.Context, id uuid.UUID, cancelAtPeriodEnd bool) error {
	const q = `
		update subscriptions
		set cancel_at_period_end = $2, updated_at = now()
		where id = $1`

	result, err := r.pool.Exec(ctx, q, id, cancelAtPeriodEnd)
	if err != nil {
		return fmt.Errorf("repository: set subscription cancel at period end: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgBillingRepository) FinalizeSubscriptionCancellation(ctx context.Context, id uuid.UUID, canceledAt time.Time) error {
	const q = `
		update subscriptions
		set status = 'canceled', canceled_at = $2, updated_at = now()
		where id = $1`

	result, err := r.pool.Exec(ctx, q, id, canceledAt)
	if err != nil {
		return fmt.Errorf("repository: finalize subscription cancellation: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgBillingRepository) ApplyPlanChange(ctx context.Context, id uuid.UUID, planID uuid.UUID, priceSubunits int, currency string, billingInterval *models.BillingInterval, currentPeriodStart, currentPeriodEnd *time.Time) error {
	const q = `
		update subscriptions
		set plan_id = $2, price_subunits_at_purchase = $3, currency_at_purchase = $4,
		    billing_interval_at_purchase = $5, current_period_start = $6, current_period_end = $7,
		    updated_at = now()
		where id = $1`

	result, err := r.pool.Exec(ctx, q, id, planID, priceSubunits, currency, billingInterval, currentPeriodStart, currentPeriodEnd)
	if err != nil {
		return fmt.Errorf("repository: apply plan change: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgBillingRepository) CreateSubscriptionEvent(ctx context.Context, e *models.SubscriptionEvent) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.Metadata.IsNull() {
		e.Metadata = models.JSONB("{}")
	}

	const q = `
		insert into subscription_events (
			id, subscription_id, event_type, from_plan_id, to_plan_id, metadata, created_at
		)
		values ($1,$2,$3,$4,$5,$6,now())
		returning created_at`

	err := r.pool.QueryRow(ctx, q,
		e.ID, e.SubscriptionID, e.EventType, e.FromPlanID, e.ToPlanID, e.Metadata,
	).Scan(&e.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: create subscription event: %w", err)
	}

	return nil
}

func (r *pgBillingRepository) CreateInvoice(ctx context.Context, inv *models.Invoice) error {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}

	const q = `
		insert into invoices (
			id, subscription_id, plan_id, dodo_invoice_id, amount_subunits, currency, status, issued_at, created_at
		)
		values ($1,$2,$3,$4,$5,$6,$7,coalesce($8, now()),now())
		returning created_at`

	err := r.pool.QueryRow(ctx, q,
		inv.ID, inv.SubscriptionID, inv.PlanID, inv.DodoInvoiceID, inv.AmountSubunits, inv.Currency, inv.Status, inv.IssuedAt,
	).Scan(&inv.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: create invoice: %w", err)
	}

	return nil
}

func (r *pgBillingRepository) GetInvoiceByDodoInvoiceID(ctx context.Context, dodoInvoiceID string) (*models.Invoice, error) {
	const q = `
		select id, subscription_id, plan_id, dodo_invoice_id, amount_subunits, currency, status, issued_at, created_at
		from invoices
		where dodo_invoice_id = $1`

	var inv models.Invoice
	err := r.pool.QueryRow(ctx, q, dodoInvoiceID).Scan(
		&inv.ID, &inv.SubscriptionID, &inv.PlanID, &inv.DodoInvoiceID, &inv.AmountSubunits, &inv.Currency, &inv.Status, &inv.IssuedAt, &inv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get invoice by dodo invoice id: %w", err)
	}

	return &inv, nil
}

func (r *pgBillingRepository) UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status models.InvoiceStatus) error {
	const q = `update invoices set status = $2 where id = $1`

	result, err := r.pool.Exec(ctx, q, id, status)
	if err != nil {
		return fmt.Errorf("repository: update invoice status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgBillingRepository) CreatePayment(ctx context.Context, p *models.Payment) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	const q = `
		insert into payments (
			id, user_id, invoice_id, dodo_payment_id, dodo_checkout_session_id,
			amount_subunits, currency, payment_method, status, idempotency_key, paid_at, created_at
		)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now())
		returning created_at`

	err := r.pool.QueryRow(ctx, q,
		p.ID, p.UserID, p.InvoiceID, p.DodoPaymentID, p.DodoCheckoutSessionID,
		p.AmountSubunits, p.Currency, p.PaymentMethod, p.Status, p.IdempotencyKey, p.PaidAt,
	).Scan(&p.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: create payment: %w", err)
	}

	return nil
}

func (r *pgBillingRepository) GetPaymentByDodoPaymentID(ctx context.Context, dodoPaymentID string) (*models.Payment, error) {
	const q = `
		select id, user_id, invoice_id, dodo_payment_id, dodo_checkout_session_id,
		       amount_subunits, refunded_amount_subunits, currency, payment_method, status, idempotency_key, paid_at, created_at
		from payments
		where dodo_payment_id = $1`

	var p models.Payment
	err := r.pool.QueryRow(ctx, q, dodoPaymentID).Scan(
		&p.ID, &p.UserID, &p.InvoiceID, &p.DodoPaymentID, &p.DodoCheckoutSessionID,
		&p.AmountSubunits, &p.RefundedAmountSubunits, &p.Currency, &p.PaymentMethod, &p.Status, &p.IdempotencyKey, &p.PaidAt, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get payment by dodo payment id: %w", err)
	}

	return &p, nil
}

func (r *pgBillingRepository) ApplyRefund(ctx context.Context, dodoPaymentID, dodoRefundID string, amountSubunits *int, currency *string, isPartial bool, reason *string) (*models.Payment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: apply refund: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var paymentID uuid.UUID
	var totalAmount, refundedAmount int
	err = tx.QueryRow(ctx, `
		select id, amount_subunits, refunded_amount_subunits
		from payments
		where dodo_payment_id = $1
		for update`, dodoPaymentID).Scan(&paymentID, &totalAmount, &refundedAmount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: apply refund: lock payment: %w", err)
	}

	var alreadyApplied bool
	err = tx.QueryRow(ctx, `select exists(select 1 from payment_refunds where dodo_refund_id = $1)`, dodoRefundID).Scan(&alreadyApplied)
	if err != nil {
		return nil, fmt.Errorf("repository: apply refund: check existing refund: %w", err)
	}

	if !alreadyApplied {
		delta := totalAmount - refundedAmount
		if amountSubunits != nil {
			delta = *amountSubunits
		}
		if delta < 0 {
			delta = 0
		}

		newRefunded := refundedAmount + delta
		if newRefunded > totalAmount {
			newRefunded = totalAmount
		}

		if _, err := tx.Exec(ctx, `
			insert into payment_refunds (id, payment_id, dodo_refund_id, amount_subunits, currency, is_partial, reason, created_at)
			values ($1,$2,$3,$4,$5,$6,$7,now())`,
			uuid.New(), paymentID, dodoRefundID, amountSubunits, currency, isPartial, reason,
		); err != nil {
			return nil, fmt.Errorf("repository: apply refund: insert refund record: %w", err)
		}

		newStatus := models.PaymentStatusCaptured
		if newRefunded >= totalAmount {
			newStatus = models.PaymentStatusRefunded
		}

		if _, err := tx.Exec(ctx, `
			update payments
			set refunded_amount_subunits = $2, status = $3
			where id = $1`,
			paymentID, newRefunded, newStatus,
		); err != nil {
			return nil, fmt.Errorf("repository: apply refund: update payment: %w", err)
		}
	}

	const q = `
		select id, user_id, invoice_id, dodo_payment_id, dodo_checkout_session_id,
		       amount_subunits, refunded_amount_subunits, currency, payment_method, status, idempotency_key, paid_at, created_at
		from payments
		where id = $1`

	var p models.Payment
	if err := tx.QueryRow(ctx, q, paymentID).Scan(
		&p.ID, &p.UserID, &p.InvoiceID, &p.DodoPaymentID, &p.DodoCheckoutSessionID,
		&p.AmountSubunits, &p.RefundedAmountSubunits, &p.Currency, &p.PaymentMethod, &p.Status, &p.IdempotencyKey, &p.PaidAt, &p.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("repository: apply refund: reload payment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("repository: apply refund: commit: %w", err)
	}

	return &p, nil
}

func (r *pgBillingRepository) ListPaymentsForUser(ctx context.Context, userID string, limit, offset int) ([]models.Payment, error) {
	const q = `
		select id, user_id, invoice_id, dodo_payment_id, dodo_checkout_session_id,
		       amount_subunits, refunded_amount_subunits, currency, payment_method, status, idempotency_key, paid_at, created_at
		from payments
		where user_id = $1
		order by created_at desc
		limit $2 offset $3`

	rows, err := r.pool.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list payments: %w", err)
	}
	defer rows.Close()

	payments := make([]models.Payment, 0)
	for rows.Next() {
		var p models.Payment
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.InvoiceID, &p.DodoPaymentID, &p.DodoCheckoutSessionID,
			&p.AmountSubunits, &p.RefundedAmountSubunits, &p.Currency, &p.PaymentMethod, &p.Status, &p.IdempotencyKey, &p.PaidAt, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan payment: %w", err)
		}
		payments = append(payments, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: list payments: %w", err)
	}

	return payments, nil
}

func (r *pgBillingRepository) CreateDispute(ctx context.Context, d *models.PaymentDispute) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	const q = `
		insert into payment_disputes (
			id, payment_id, dodo_dispute_id, status, amount_subunits, currency, reason, stage, opened_at, created_at
		)
		values ($1,$2,$3,$4,$5,$6,$7,$8,now(),now())
		returning opened_at, created_at`

	err := r.pool.QueryRow(ctx, q,
		d.ID, d.PaymentID, d.DodoDisputeID, d.Status, d.AmountSubunits, d.Currency, d.Reason, d.Stage,
	).Scan(&d.OpenedAt, &d.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: create dispute: %w", err)
	}

	return nil
}

func (r *pgBillingRepository) GetDisputeByDodoDisputeID(ctx context.Context, dodoDisputeID string) (*models.PaymentDispute, error) {
	const q = `
		select id, payment_id, dodo_dispute_id, status, amount_subunits, currency, reason, stage, opened_at, resolved_at, created_at
		from payment_disputes
		where dodo_dispute_id = $1`

	var d models.PaymentDispute
	err := r.pool.QueryRow(ctx, q, dodoDisputeID).Scan(
		&d.ID, &d.PaymentID, &d.DodoDisputeID, &d.Status, &d.AmountSubunits, &d.Currency, &d.Reason, &d.Stage, &d.OpenedAt, &d.ResolvedAt, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get dispute by dodo dispute id: %w", err)
	}

	return &d, nil
}

func (r *pgBillingRepository) UpdateDisputeStatus(ctx context.Context, id uuid.UUID, status models.DisputeStatus, resolvedAt *time.Time) error {
	const q = `
		update payment_disputes
		set status = $2, resolved_at = coalesce($3, resolved_at)
		where id = $1`

	result, err := r.pool.Exec(ctx, q, id, status, resolvedAt)
	if err != nil {
		return fmt.Errorf("repository: update dispute status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgBillingRepository) CreateWebhookEvent(ctx context.Context, e *models.DodoWebhookEvent) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}

	const q = `
		insert into dodo_webhook_events (
			id, dodo_event_id, event_type, payload, processing_status, processing_error, processed_at, created_at
		)
		values ($1,$2,$3,$4,$5,$6,$7,now())
		returning created_at`

	err := r.pool.QueryRow(ctx, q,
		e.ID, e.DodoEventID, e.EventType, e.Payload, e.ProcessingStatus, e.ProcessingError, e.ProcessedAt,
	).Scan(&e.CreatedAt)
	if err != nil {
		return fmt.Errorf("repository: create webhook event: %w", err)
	}

	return nil
}

func (r *pgBillingRepository) GetWebhookEventByDodoEventID(ctx context.Context, dodoEventID string) (*models.DodoWebhookEvent, error) {
	const q = `
		select id, dodo_event_id, event_type, payload, processing_status, processing_error, processed_at, created_at
		from dodo_webhook_events
		where dodo_event_id = $1`

	var e models.DodoWebhookEvent
	err := r.pool.QueryRow(ctx, q, dodoEventID).Scan(
		&e.ID, &e.DodoEventID, &e.EventType, &e.Payload, &e.ProcessingStatus, &e.ProcessingError, &e.ProcessedAt, &e.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get webhook event: %w", err)
	}

	return &e, nil
}

func (r *pgBillingRepository) UpdateWebhookProcessingStatus(ctx context.Context, id uuid.UUID, status models.WebhookProcessingStatus, processingError *string) error {
	const q = `
		update dodo_webhook_events
		set processing_status = $2, processing_error = $3
		where id = $1`

	result, err := r.pool.Exec(ctx, q, id, status, processingError)
	if err != nil {
		return fmt.Errorf("repository: update webhook processing status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgBillingRepository) MarkWebhookProcessed(ctx context.Context, id uuid.UUID) error {
	const q = `
		update dodo_webhook_events
		set processing_status = 'processed', processed_at = now(), processing_error = null
		where id = $1`

	result, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("repository: mark webhook processed: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgBillingRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: delete all for user: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `delete from payments where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete payments for user: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		delete from subscription_events
		where subscription_id in (select id from subscriptions where user_id = $1)`, userID); err != nil {
		return fmt.Errorf("repository: delete subscription events for user: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		delete from invoices
		where subscription_id in (select id from subscriptions where user_id = $1)`, userID); err != nil {
		return fmt.Errorf("repository: delete invoices for user: %w", err)
	}
	if _, err := tx.Exec(ctx, `delete from subscriptions where user_id = $1`, userID); err != nil {
		return fmt.Errorf("repository: delete subscriptions for user: %w", err)
	}

	return tx.Commit(ctx)
}
