package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
	"github.com/dodopayments/dodopayments-go/shared"
)

var (
	ErrDodoNotFound      = errors.New("dodo: resource not found")
	ErrDodoConflict      = errors.New("dodo: conflicting operation already in progress")
	ErrDodoUnprocessable = errors.New("dodo: request could not be processed for current subscription state")
)

type DodoClient struct {
	sdk           *dodopayments.Client
	webhookSecret string
}

func NewDodoClient(apiKey, webhookSecret string, testMode bool) *DodoClient {
	opts := []option.RequestOption{option.WithBearerToken(apiKey)}
	if testMode {
		opts = append(opts, option.WithEnvironmentTestMode())
	}

	return &DodoClient{
		sdk:           dodopayments.NewClient(opts...),
		webhookSecret: webhookSecret,
	}
}

func classifyDodoError(err error) error {
	if err == nil {
		return nil
	}

	var apiErr *dodopayments.Error
	if !errors.As(err, &apiErr) {
		return err
	}

	switch apiErr.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("%w: %v", ErrDodoNotFound, err)
	case http.StatusConflict:
		return fmt.Errorf("%w: %v", ErrDodoConflict, err)
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("%w: %v", ErrDodoUnprocessable, err)
	default:
		return err
	}
}

type CreateCheckoutSessionInput struct {
	ProductID       string
	CustomerEmail   string
	CustomerName    string
	ReturnURL       string
	BillingCurrency string
	BillingCountry  string
	Metadata        map[string]string
}

type CheckoutSession struct {
	SessionID   string `json:"sessionId"`
	CheckoutURL string `json:"checkoutUrl"`
}

func (d *DodoClient) CreateCheckoutSession(ctx context.Context, in CreateCheckoutSessionInput) (*CheckoutSession, error) {
	metadata := dodopayments.MetadataParam{}
	for k, v := range in.Metadata {
		metadata[k] = shared.UnionString(v)
	}

	req := dodopayments.CheckoutSessionRequestParam{
		ProductCart: dodopayments.F([]dodopayments.ProductItemReqParam{
			{
				ProductID: dodopayments.F(in.ProductID),
				Quantity:  dodopayments.F(int64(1)),
			},
		}),
		Customer: dodopayments.F[dodopayments.CustomerRequestUnionParam](dodopayments.CustomerRequestParam{
			Email: dodopayments.F(in.CustomerEmail),
			Name:  dodopayments.F(in.CustomerName),
		}),
		ReturnURL: dodopayments.F(in.ReturnURL),
		Metadata:  dodopayments.F(metadata),
	}

	if in.BillingCurrency != "" {
		req.BillingCurrency = dodopayments.F(dodopayments.Currency(in.BillingCurrency))
	}

	if in.BillingCountry != "" {
		req.BillingAddress = dodopayments.F(dodopayments.CheckoutSessionBillingAddressParam{
			Country: dodopayments.F(dodopayments.CountryCode(strings.ToUpper(in.BillingCountry))),
		})
	}

	resp, err := d.sdk.CheckoutSessions.New(ctx, dodopayments.CheckoutSessionNewParams{
		CheckoutSessionRequest: req,
	})
	if err != nil {
		return nil, fmt.Errorf("dodo: create checkout session: %w", classifyDodoError(err))
	}

	return &CheckoutSession{
		SessionID:   resp.SessionID,
		CheckoutURL: resp.CheckoutURL,
	}, nil
}

func (d *DodoClient) SetCancelAtNextBillingDate(ctx context.Context, subscriptionID string, cancel bool) (*dodopayments.Subscription, error) {
	params := dodopayments.SubscriptionUpdateParams{
		CancelAtNextBillingDate: dodopayments.F(cancel),
	}
	if cancel {
		params.CancelReason = dodopayments.F(dodopayments.SubscriptionUpdateParamsCancelReasonCancelledByCustomer)
	}

	resp, err := d.sdk.Subscriptions.Update(ctx, subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("dodo: update subscription cancellation: %w", classifyDodoError(err))
	}

	return resp, nil
}

type ChangePlanInput struct {
	SubscriptionID string
	ProductID      string
}

func (d *DodoClient) ChangePlan(ctx context.Context, in ChangePlanInput) error {
	err := d.sdk.Subscriptions.ChangePlan(ctx, in.SubscriptionID, dodopayments.SubscriptionChangePlanParams{
		UpdateSubscriptionPlanReq: dodopayments.UpdateSubscriptionPlanReqParam{
			ProductID:            dodopayments.F(in.ProductID),
			Quantity:             dodopayments.F(int64(1)),
			ProrationBillingMode: dodopayments.F(dodopayments.UpdateSubscriptionPlanReqProrationBillingModeProratedImmediately),
		},
	})
	if err != nil {
		return fmt.Errorf("dodo: change plan: %w", classifyDodoError(err))
	}

	return nil
}

func (d *DodoClient) GetSubscription(ctx context.Context, subscriptionID string) (*dodopayments.Subscription, error) {
	sub, err := d.sdk.Subscriptions.Get(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("dodo: get subscription: %w", classifyDodoError(err))
	}

	return sub, nil
}

type UpdatePaymentMethodResult struct {
	PaymentID    string `json:"paymentId,omitempty"`
	PaymentLink  string `json:"paymentLink,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

func (d *DodoClient) InitiatePaymentMethodUpdate(ctx context.Context, subscriptionID, returnURL string) (*UpdatePaymentMethodResult, error) {
	resp, err := d.sdk.Subscriptions.UpdatePaymentMethod(ctx, subscriptionID, dodopayments.SubscriptionUpdatePaymentMethodParams{
		PaymentMethod: dodopayments.SubscriptionUpdatePaymentMethodParamsPaymentMethodNew{
			Type:      dodopayments.F(dodopayments.SubscriptionUpdatePaymentMethodParamsPaymentMethodNewTypeNew),
			ReturnURL: dodopayments.F(returnURL),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dodo: update payment method: %w", classifyDodoError(err))
	}

	return &UpdatePaymentMethodResult{
		PaymentID:    resp.PaymentID,
		PaymentLink:  resp.PaymentLink,
		ClientSecret: resp.ClientSecret,
	}, nil
}

func (d *DodoClient) VerifyWebhookSignature(id, timestamp, signatureHeader string, body []byte) error {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.New("dodo: invalid webhook timestamp")
	}
	if age := time.Since(time.Unix(ts, 0)); age > 5*time.Minute || age < -5*time.Minute {
		return errors.New("dodo: webhook timestamp outside tolerance window")
	}

	secret := strings.TrimPrefix(d.webhookSecret, "whsec_")
	key, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return fmt.Errorf("dodo: webhook secret is not valid base64: %w", err)
	}

	signedContent := id + "." + timestamp + "." + string(body)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signedContent))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	for _, part := range strings.Fields(signatureHeader) {
		sig := part
		if idx := strings.Index(part, ","); idx >= 0 {
			sig = part[idx+1:]
		}
		if hmac.Equal([]byte(sig), []byte(expected)) {
			return nil
		}
	}

	return errors.New("dodo: webhook signature mismatch")
}
