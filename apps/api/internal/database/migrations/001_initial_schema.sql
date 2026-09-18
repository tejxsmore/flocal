create extension if not exists "pgcrypto";
create extension if not exists "pg_trgm";
create extension if not exists "citext";

create type topic_difficulty as enum ('easy', 'medium', 'hard');
create type topic_format as enum ('word', 'quote', 'debate', 'situation', 'story_starter', 'image');
create type session_status as enum ('pending', 'processing', 'completed', 'failed');
create type xp_reason as enum ('session_complete', 'daily_challenge', 'streak_bonus', 'first_session', 'perfect_score', 'badge_earned', 'xp_adjustment', 'xp_reversal');
create type plan_type as enum ('free', 'pro');
create type billing_interval as enum ('monthly', 'annual');
create type subscription_status as enum ('active', 'trialing', 'past_due', 'canceled', 'expired', 'paused');
create type payment_status as enum ('created', 'captured', 'failed', 'refunded');
create type dispute_status as enum ('opened', 'expired', 'accepted', 'cancelled', 'challenged', 'won', 'lost');
create type webhook_processing_status as enum ('received', 'processing', 'processed', 'failed');
create type invoice_status as enum ('open', 'paid', 'void', 'uncollectible', 'refunded');
create type subscription_event_type as enum ('created', 'upgraded', 'downgraded', 'renewed', 'canceled', 'resumed', 'paused', 'trial_started', 'trial_ended', 'payment_succeeded', 'payment_failed', 'payment_refunded', 'expired');
create type onboarding_goal as enum ('confidence', 'conversations', 'public_speaking', 'work_interviews', 'vocabulary', 'habit');
create type focus_area as enum ('work_interviews', 'conversations', 'college_presentations', 'travel', 'everyday_situations', 'public_speaking', 'online_meetings', 'anywhere');
create type daily_time_commitment as enum ('min_5', 'min_10', 'min_15', 'min_30_plus');
create type badge_criteria_type as enum ('streak_days', 'session_count', 'xp_total', 'score_threshold', 'manual');
create type topic_submission_status as enum ('pending', 'approved', 'rejected');
create type notification_type as enum ('daily_reminder', 'streak_risk', 'session_ready', 'badge_earned', 'subscription');
create type notification_channel as enum ('in_app', 'email', 'push');
create type notification_delivery_status as enum ('pending', 'sent', 'delivered', 'failed');
create type leaderboard_period as enum ('all_time', 'weekly', 'monthly');
create type user_role as enum ('user', 'moderator', 'admin');

create or replace function set_updated_at()
returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;

create table "user" (
  id                        text primary key,
  name                      text not null,
  email                     citext not null unique,
  email_verified            boolean not null default false,
  email_verified_at         timestamptz,
  image                     text,
  username                  text unique,
  role                      user_role not null default 'user',
  default_prep_time_seconds int not null default 0 check (default_prep_time_seconds >= 0),
  timezone                  text default 'UTC',
  country_code              char(2) check (country_code is null or country_code ~ '^[A-Z]{2}$'),
  is_active                 boolean not null default true,
  onboarding_completed      boolean not null default false,
  last_login_at             timestamptz,
  deleted_at                timestamptz,
  created_at                timestamptz not null default now(),
  updated_at                timestamptz not null default now()
);

alter table "user"
add constraint chk_email_verified_consistency
check (email_verified = false or email_verified_at is not null);

create trigger trg_user_updated_at
  before update on "user"
  for each row execute function set_updated_at();

create index idx_user_active on "user"(is_active) where deleted_at is null;
create index idx_user_country on "user"(country_code)
  where deleted_at is null and country_code is not null;

create table "session" (
  id            text primary key,
  expires_at    timestamptz not null,
  token         text not null unique,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  ip_address    text,
  user_agent    text,
  user_id       text not null references "user"(id) on delete cascade
);

create trigger trg_session_updated_at
  before update on "session"
  for each row execute function set_updated_at();

create index idx_session_user_id on "session"(user_id);
create index idx_session_expires_at on "session"(expires_at);

create table "account" (
  id                       text primary key,
  account_id               text not null,
  provider_id              text not null,
  user_id                  text not null references "user"(id) on delete cascade,
  access_token             text,
  refresh_token            text,
  id_token                 text,
  access_token_expires_at  timestamptz,
  refresh_token_expires_at timestamptz,
  scope                    text,
  password_hash            text,
  failed_login_attempts    int not null default 0 check (failed_login_attempts >= 0),
  locked_until              timestamptz,
  created_at               timestamptz not null default now(),
  updated_at               timestamptz not null default now(),
  unique (provider_id, account_id)
);

alter table "account"
add constraint chk_credential_account_has_password
check (provider_id <> 'credential' or password_hash is not null);

create trigger trg_account_updated_at
  before update on "account"
  for each row execute function set_updated_at();

create index idx_account_user_id on "account"(user_id);

create unique index idx_account_one_credential_per_user
  on "account"(user_id)
  where provider_id = 'credential';

create table "verification" (
  id          text primary key,
  identifier  text not null,
  value_hash  text not null,
  used_at     timestamptz,
  expires_at  timestamptz not null,
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now()
);

create trigger trg_verification_updated_at
  before update on "verification"
  for each row execute function set_updated_at();

create index idx_verification_identifier on "verification"(identifier);

create table user_preferences (
  user_id               text primary key references "user"(id) on delete cascade,
  primary_goal           onboarding_goal,
  daily_time_commitment  daily_time_commitment,
  completed_at            timestamptz,
  created_at             timestamptz not null default now(),
  updated_at             timestamptz not null default now()
);

create trigger trg_user_preferences_updated_at
  before update on user_preferences
  for each row execute function set_updated_at();

create table user_focus_areas (
  user_id       text not null references "user"(id) on delete cascade,
  focus_area    focus_area not null,
  created_at    timestamptz not null default now(),
  primary key (user_id, focus_area)
);

create table subscription_plans (
  id                    uuid primary key default gen_random_uuid(),
  slug                  text not null unique,
  name                  text not null,
  plan_type             plan_type not null,
  billing_interval      billing_interval,
  price_subunits        int not null default 0 check (price_subunits >= 0),
  currency              text not null default 'USD' check (currency ~ '^[A-Z]{3}$'),
  daily_session_limit   int check (daily_session_limit is null or daily_session_limit >= 0),
  can_view_analysis     boolean not null default true,
  dodo_product_id       text unique,
  is_active             boolean not null default true,
  created_at            timestamptz not null default now()
);

alter table subscription_plans
add constraint chk_subscription_plan_pricing
check (
  (
    plan_type = 'free'
    and price_subunits = 0
    and billing_interval is null
  )
  or
  (
    plan_type = 'pro'
    and price_subunits > 0
    and billing_interval is not null
  )
);

create table subscriptions (
  id                           uuid primary key default gen_random_uuid(),
  user_id                      text not null references "user"(id) on delete cascade,
  plan_id                      uuid not null references subscription_plans(id),
  status                       subscription_status not null default 'active',
  dodo_subscription_id         text unique,
  dodo_customer_id             text,
  price_subunits_at_purchase   int not null check (price_subunits_at_purchase >= 0),
  currency_at_purchase         text not null check (currency_at_purchase ~ '^[A-Z]{3}$'),
  billing_interval_at_purchase billing_interval,
  current_period_start        timestamptz,
  current_period_end          timestamptz,
  cancel_at_period_end        boolean not null default false,
  canceled_at                 timestamptz,
  created_at                   timestamptz not null default now(),
  updated_at                   timestamptz not null default now()
);

alter table subscriptions
add constraint chk_subscription_period_order
check (
  current_period_start is null
  or current_period_end is null
  or current_period_start <= current_period_end
);

alter table subscriptions
add constraint chk_subscription_cancel_has_period_end
check (cancel_at_period_end = false or current_period_end is not null);

alter table subscriptions
add constraint chk_subscription_active_has_period_end
check (
  status not in ('active', 'trialing', 'past_due')
  or current_period_end is not null
);

alter table subscriptions
add constraint chk_subscription_canceled_has_timestamp
check (status <> 'canceled' or canceled_at is not null);

create trigger trg_subscriptions_updated_at
  before update on subscriptions
  for each row execute function set_updated_at();

create index idx_subscriptions_user_id on subscriptions(user_id);
create index idx_subscriptions_status on subscriptions(status);
create index idx_subscriptions_dodo_customer_id on subscriptions(dodo_customer_id);

create unique index idx_one_current_subscription_per_user
  on subscriptions(user_id)
  where status in ('active', 'trialing', 'past_due');

create table subscription_events (
  id              uuid primary key default gen_random_uuid(),
  subscription_id uuid not null references subscriptions(id) on delete cascade,
  event_type      subscription_event_type not null,
  from_plan_id    uuid references subscription_plans(id),
  to_plan_id      uuid references subscription_plans(id),
  metadata        jsonb not null default '{}' check (jsonb_typeof(metadata) = 'object'),
  created_at      timestamptz not null default now()
);

create index idx_subscription_events_subscription
  on subscription_events(subscription_id, created_at desc);

create table invoices (
  id              uuid primary key default gen_random_uuid(),
  subscription_id uuid not null references subscriptions(id) on delete cascade,
  plan_id         uuid references subscription_plans(id),
  dodo_invoice_id text unique,
  amount_subunits int not null check (amount_subunits >= 0),
  currency        text not null default 'USD' check (currency ~ '^[A-Z]{3}$'),
  status          invoice_status not null default 'open',
  issued_at       timestamptz not null default now(),
  created_at      timestamptz not null default now()
);

create index idx_invoices_subscription
  on invoices(subscription_id, issued_at desc);

create table payments (
  id                       uuid primary key default gen_random_uuid(),
  user_id                  text not null references "user"(id) on delete cascade,
  invoice_id               uuid references invoices(id) on delete set null,
  dodo_payment_id          text unique,
  dodo_checkout_session_id text,
  amount_subunits          int not null check (amount_subunits >= 0),
  refunded_amount_subunits int not null default 0 check (refunded_amount_subunits >= 0),
  currency                 text not null default 'USD' check (currency ~ '^[A-Z]{3}$'),
  payment_method           text,
  status                   payment_status not null default 'created',
  idempotency_key          text unique,
  paid_at                  timestamptz,
  created_at               timestamptz not null default now()
);

alter table payments
add constraint chk_payment_refunded_not_exceeding
check (refunded_amount_subunits <= amount_subunits);

create index idx_payments_user_id
  on payments(user_id, created_at desc);

create index idx_payments_invoice_id
  on payments(invoice_id);

create unique index idx_payments_checkout_session
  on payments(dodo_checkout_session_id)
  where dodo_checkout_session_id is not null;

create table payment_refunds (
  id               uuid primary key default gen_random_uuid(),
  payment_id       uuid not null references payments(id) on delete cascade,
  dodo_refund_id   text not null unique,
  amount_subunits  int check (amount_subunits is null or amount_subunits >= 0),
  currency         text check (currency is null or currency ~ '^[A-Z]{3}$'),
  is_partial       boolean not null default false,
  reason           text,
  created_at       timestamptz not null default now()
);

create index idx_payment_refunds_payment
  on payment_refunds(payment_id);

create table payment_disputes (
  id              uuid primary key default gen_random_uuid(),
  payment_id      uuid not null references payments(id) on delete cascade,
  dodo_dispute_id text not null unique,
  status          dispute_status not null default 'opened',
  amount_subunits int check (amount_subunits is null or amount_subunits >= 0),
  currency        text check (currency is null or currency ~ '^[A-Z]{3}$'),
  reason          text,
  stage           text,
  opened_at       timestamptz not null default now(),
  resolved_at     timestamptz,
  created_at      timestamptz not null default now()
);

create index idx_payment_disputes_payment
  on payment_disputes(payment_id);

create index idx_payment_disputes_status
  on payment_disputes(status)
  where status in ('opened', 'challenged');

create table dodo_webhook_events (
  id                 uuid primary key default gen_random_uuid(),
  dodo_event_id      text not null unique,
  event_type         text not null,
  payload            jsonb not null,
  processing_status  webhook_processing_status not null default 'received',
  processing_error   text,
  processed_at       timestamptz,
  created_at         timestamptz not null default now()
);

create index idx_dodo_webhook_processing
  on dodo_webhook_events(processing_status, created_at)
  where processing_status in ('received', 'failed');

create view user_current_plan with (security_invoker = true) as
select
  u.id as user_id,
  coalesce(cp.slug, fp.slug) as plan_slug,
  coalesce(cp.daily_session_limit, fp.daily_session_limit) as daily_session_limit,
  coalesce(cp.can_view_analysis, fp.can_view_analysis) as can_view_analysis,
  cp.current_period_end as plan_renews_at,
  cp.status as subscription_status
from "user" u
left join lateral (
  select
    sp.slug,
    sp.daily_session_limit,
    sp.can_view_analysis,
    s.current_period_end,
    s.status
  from subscriptions s
  join subscription_plans sp on sp.id = s.plan_id
  where s.user_id = u.id
    and s.status in ('active', 'trialing', 'past_due')
  order by s.current_period_end desc nulls last
  limit 1
) cp on true
cross join lateral (
  select
    slug,
    daily_session_limit,
    can_view_analysis
  from subscription_plans
  where slug = 'free'
    and is_active = true
  limit 1
) fp;

create table topic_categories (
  id         uuid primary key default gen_random_uuid(),
  slug       text not null unique,
  name       text not null,
  icon       text,
  is_active  boolean not null default true,
  sort_order int not null default 0,
  created_at timestamptz not null default now()
);

create table topics (
  id                       uuid primary key default gen_random_uuid(),
  category_id              uuid not null references topic_categories(id) on delete restrict,
  title                    text not null,
  format                   topic_format not null,
  difficulty               topic_difficulty not null default 'medium',
  recommended_prep_seconds int not null default 0 check (recommended_prep_seconds >= 0),
  tags                     text[] not null default '{}',
  metadata                 jsonb not null default '{}' check (jsonb_typeof(metadata) = 'object'),
  is_active                boolean not null default true,
  is_premium               boolean not null default false,
  source                   text not null default 'manual'
    check (source in ('manual', 'ai_generated', 'community')),
  language_code            text not null default 'en',
  submitted_by             text references "user"(id) on delete set null,
  deleted_at               timestamptz,
  created_at               timestamptz not null default now(),
  updated_at               timestamptz not null default now()
);

create trigger trg_topics_updated_at
  before update on topics
  for each row execute function set_updated_at();

create index idx_topics_category_id on topics(category_id);
create index idx_topics_active on topics(is_active) where is_active = true;
create index idx_topics_title_trgm on topics using gin(title gin_trgm_ops);
create index idx_topics_difficulty on topics(difficulty);
create index idx_topics_format on topics(format);
create index idx_topics_category_difficulty on topics(category_id, difficulty);

create index idx_topics_selectable
  on topics(category_id, format, difficulty, is_premium)
  where is_active = true and deleted_at is null;

create table topic_submissions (
  id              uuid primary key default gen_random_uuid(),
  submitted_by    text not null references "user"(id) on delete cascade,
  category_id     uuid references topic_categories(id) on delete set null,
  title           text not null,
  format          topic_format not null,
  difficulty      topic_difficulty not null default 'medium',
  tags            text[] not null default '{}',
  status          topic_submission_status not null default 'pending',
  review_note     text,
  reviewed_by     text references "user"(id) on delete set null,
  reviewed_at     timestamptz,
  topic_id        uuid references topics(id) on delete set null,
  created_at      timestamptz not null default now()
);

create index idx_topic_submissions_status
  on topic_submissions(status, created_at);

create table collections (
  id          uuid primary key default gen_random_uuid(),
  slug        text not null unique,
  name        text not null,
  description text,
  is_active   boolean not null default true,
  created_at  timestamptz not null default now()
);

create table topic_collections (
  topic_id      uuid not null references topics(id) on delete restrict,
  collection_id uuid not null references collections(id) on delete cascade,
  created_at    timestamptz not null default now(),
  primary key (topic_id, collection_id)
);

create index idx_topic_collections_collection
  on topic_collections(collection_id);

create table topic_bookmarks (
  user_id    text not null references "user"(id) on delete cascade,
  topic_id   uuid not null references topics(id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (user_id, topic_id)
);

create index idx_topic_bookmarks_user
  on topic_bookmarks(user_id, created_at desc);

create table user_category_preferences (
  user_id     text not null references "user"(id) on delete cascade,
  category_id uuid not null references topic_categories(id) on delete cascade,
  weight      int not null default 1 check (weight >= 0),
  created_at  timestamptz not null default now(),
  primary key (user_id, category_id)
);

create table topic_spins (
  id                 uuid primary key default gen_random_uuid(),
  user_id            text not null references "user"(id) on delete cascade,
  topic_id           uuid not null references topics(id) on delete restrict,
  category_filter_id uuid references topic_categories(id),
  format_filter      topic_format,
  session_id         uuid,
  created_at         timestamptz not null default now()
);

create index idx_topic_spins_user_id_created
  on topic_spins(user_id, created_at desc);

create index idx_topic_spins_topic_id
  on topic_spins(topic_id);

create view topic_usage_stats with (security_invoker = true) as
select
  topic_id,
  count(*) as spin_count
from topic_spins
group by topic_id;

create table user_topic_stats (
  user_id           text not null references "user"(id) on delete cascade,
  topic_id          uuid not null references topics(id) on delete cascade,
  times_spun        int not null default 0 check (times_spun >= 0),
  times_completed   int not null default 0 check (times_completed >= 0),
  average_score     numeric(5,2)
    check (average_score is null or average_score between 0 and 100),
  last_completed_at timestamptz,
  primary key (user_id, topic_id)
);

create table daily_challenges (
  id                 uuid primary key default gen_random_uuid(),
  topic_id           uuid not null references topics(id) on delete restrict,
  challenge_date     date not null,
  language_code      text not null default 'en',
  difficulty         topic_difficulty not null default 'medium',
  speak_time_seconds int not null default 120 check (speak_time_seconds > 0),
  xp_reward          int not null default 50 check (xp_reward >= 0),
  created_at         timestamptz not null default now(),
  unique (challenge_date, language_code, difficulty)
);

create table speaking_sessions (
  id                     uuid primary key default gen_random_uuid(),
  user_id                text not null references "user"(id) on delete cascade,
  topic_id               uuid not null references topics(id) on delete restrict,
  daily_challenge_id     uuid references daily_challenges(id) on delete set null,
  client_upload_token    text unique,
  debate_stance          text check (debate_stance is null or debate_stance in ('for', 'against')),
  prep_time_seconds      int not null default 0 check (prep_time_seconds >= 0),
  speak_time_seconds     int not null default 60 check (speak_time_seconds > 0),
  status                 session_status not null default 'pending',
  processing_step        text,
  processing_attempt_id  uuid,
  processing_started_at  timestamptz,
  processing_worker_id   text,
  failure_reason         text,
  retry_count            int not null default 0 check (retry_count >= 0),
  last_error             text,
  last_retry_at          timestamptz,
  next_retry_at          timestamptz,
  audio_s3_key           text,
  audio_duration_seconds numeric(6,2)
    check (audio_duration_seconds is null or audio_duration_seconds >= 0),
  audio_format           text,
  audio_size_bytes       bigint
    check (audio_size_bytes is null or audio_size_bytes >= 0),
  audio_expires_at       timestamptz,
  audio_deleted_at       timestamptz,
  share_token_hash       text unique,
  share_enabled          boolean not null default false,
  share_expires_at       timestamptz,
  started_at             timestamptz,
  submitted_at           timestamptz,
  completed_at           timestamptz,
  created_at             timestamptz not null default now(),
  updated_at             timestamptz not null default now()
);

alter table speaking_sessions
add constraint chk_completed_session_timestamp
check (status <> 'completed' or completed_at is not null);

alter table speaking_sessions
add constraint chk_session_timestamp_order
check (
  (started_at is null or submitted_at is null or started_at <= submitted_at)
  and
  (submitted_at is null or completed_at is null or submitted_at <= completed_at)
);

alter table speaking_sessions
add constraint chk_processing_session_has_start
check (status <> 'processing' or processing_started_at is not null);

alter table speaking_sessions
add constraint chk_failed_session_has_reason
check (status <> 'failed' or failure_reason is not null);

alter table speaking_sessions
add constraint chk_audio_deleted_after_expiry
check (
  audio_deleted_at is null
  or audio_expires_at is null
  or audio_deleted_at >= audio_expires_at
);

alter table speaking_sessions
add constraint chk_share_requires_token
check (share_enabled = false or share_token_hash is not null);

create trigger trg_speaking_sessions_updated_at
  before update on speaking_sessions
  for each row execute function set_updated_at();

create index idx_speaking_sessions_user_id_created
  on speaking_sessions(user_id, created_at desc);

create index idx_speaking_sessions_topic_id
  on speaking_sessions(topic_id);

create index idx_speaking_sessions_status
  on speaking_sessions(status);

create index idx_speaking_sessions_next_retry
  on speaking_sessions(next_retry_at)
  where status = 'failed';

create index idx_speaking_sessions_processing_queue
  on speaking_sessions(created_at)
  where status = 'pending';

create index idx_speaking_sessions_audio_expiry
  on speaking_sessions(audio_expires_at)
  where audio_deleted_at is null;

create index idx_speaking_sessions_daily_challenge
  on speaking_sessions(daily_challenge_id)
  where daily_challenge_id is not null;

alter table topic_spins
add constraint fk_topic_spins_session
foreign key (session_id)
references speaking_sessions(id)
on delete set null;

create table daily_challenge_completions (
  id                 uuid primary key default gen_random_uuid(),
  user_id            text not null references "user"(id) on delete cascade,
  daily_challenge_id uuid not null references daily_challenges(id) on delete cascade,
  session_id         uuid references speaking_sessions(id) on delete set null,
  completed_at       timestamptz not null default now(),
  unique (user_id, daily_challenge_id)
);

create table transcripts (
  id             uuid primary key default gen_random_uuid(),
  session_id     uuid not null unique references speaking_sessions(id) on delete cascade,
  raw_text       text not null,
  word_count     int not null default 0 check (word_count >= 0),
  language       text default 'en',
  avg_confidence numeric(4,3)
    check (avg_confidence is null or avg_confidence between 0 and 1),
  word_timings   jsonb,
  created_at     timestamptz not null default now()
);

create index idx_transcripts_text_trgm
  on transcripts using gin(raw_text gin_trgm_ops);

create table session_analysis (
  id                         uuid primary key default gen_random_uuid(),
  session_id                 uuid not null unique references speaking_sessions(id) on delete cascade,
  words_per_minute           numeric(6,2)
    check (words_per_minute is null or words_per_minute >= 0),
  speaking_duration_seconds  numeric(6,2)
    check (speaking_duration_seconds is null or speaking_duration_seconds >= 0),
  filler_word_count          int not null default 0 check (filler_word_count >= 0),
  filler_word_breakdown      jsonb not null default '{}'
    check (jsonb_typeof(filler_word_breakdown) = 'object'),
  filler_rate                numeric(5,4)
    check (filler_rate is null or filler_rate >= 0),
  pause_count                int not null default 0 check (pause_count >= 0),
  longest_pause_seconds      numeric(6,2)
    check (longest_pause_seconds is null or longest_pause_seconds >= 0),
  total_pause_seconds        numeric(6,2)
    check (total_pause_seconds is null or total_pause_seconds >= 0),
  pause_rate                 numeric(5,4)
    check (pause_rate is null or pause_rate >= 0),
  unique_word_count          int
    check (unique_word_count is null or unique_word_count >= 0),
  lexical_diversity          numeric(5,4)
    check (lexical_diversity is null or lexical_diversity between 0 and 1),
  sentence_count             int
    check (sentence_count is null or sentence_count >= 0),
  average_sentence_length    numeric(6,2)
    check (average_sentence_length is null or average_sentence_length >= 0),
  overall_score              numeric(5,2) not null
    check (overall_score between 0 and 100),
  clarity_score              numeric(5,2)
    check (clarity_score between 0 and 100),
  delivery_score             numeric(5,2)
    check (delivery_score between 0 and 100),
  content_score              numeric(5,2)
    check (content_score between 0 and 100),
  vocabulary_score           numeric(5,2)
    check (vocabulary_score between 0 and 100),
  grammar_score              numeric(5,2)
    check (grammar_score between 0 and 100),
  strengths                  jsonb not null default '[]'
    check (jsonb_typeof(strengths) = 'array'),
  improvements              jsonb not null default '[]'
    check (jsonb_typeof(improvements) = 'array'),
  coach_message              text,
  ai_provider                text not null default 'openai',
  ai_model                   text not null default 'gpt-5-mini',
  ai_prompt_version          text not null default 'v1',
  prompt_template_id         text,
  analysis_version           text not null default 'v1',
  ai_temperature             numeric(3,2),
  ai_seed                    text,
  request_id                 text,
  created_at                 timestamptz not null default now()
);

create index idx_session_analysis_overall_score
  on session_analysis(overall_score);

create table ai_provider_logs (
  id                    uuid primary key default gen_random_uuid(),
  session_id            uuid not null references speaking_sessions(id) on delete cascade,
  processing_attempt_id uuid,
  source                text not null check (source in ('deepgram', 'openai')),
  operation             text,
  input_tokens          bigint check (input_tokens is null or input_tokens >= 0),
  output_tokens         bigint check (output_tokens is null or output_tokens >= 0),
  cached_tokens         bigint check (cached_tokens is null or cached_tokens >= 0),
  latency_ms            int check (latency_ms is null or latency_ms >= 0),
  cost_subunits         bigint check (cost_subunits is null or cost_subunits >= 0),
  raw_response          jsonb,
  created_at            timestamptz not null default now(),
  purge_after           timestamptz not null default (now() + interval '30 days')
);

create index idx_ai_provider_logs_purge
  on ai_provider_logs(purge_after);

create index idx_ai_provider_logs_session
  on ai_provider_logs(session_id);

create table grammar_corrections (
  id             uuid primary key default gen_random_uuid(),
  session_id     uuid not null references speaking_sessions(id) on delete cascade,
  transcript_id  uuid not null references transcripts(id) on delete cascade,
  original_text  text not null,
  corrected_text text not null,
  explanation    text,
  error_type     text,
  start_char     int check (start_char is null or start_char >= 0),
  end_char       int check (end_char is null or end_char >= start_char),
  created_at     timestamptz not null default now()
);

create index idx_grammar_corrections_session
  on grammar_corrections(session_id);

create index idx_grammar_corrections_error_type
  on grammar_corrections(error_type);

create table vocabulary_suggestions (
  id              uuid primary key default gen_random_uuid(),
  session_id      uuid not null references speaking_sessions(id) on delete cascade,
  transcript_id   uuid not null references transcripts(id) on delete cascade,
  original_word   text not null,
  suggested_words text[] not null default '{}',
  reason          text,
  context_snippet text,
  start_char      int check (start_char is null or start_char >= 0),
  end_char        int check (end_char is null or end_char >= start_char),
  created_at      timestamptz not null default now()
);

create index idx_vocab_suggestions_session
  on vocabulary_suggestions(session_id);

create table vocabulary_words (
  id         uuid primary key default gen_random_uuid(),
  word       citext not null unique,
  definition text,
  difficulty topic_difficulty,
  created_at timestamptz not null default now()
);

create table user_vocabulary (
  id              uuid primary key default gen_random_uuid(),
  user_id         text not null references "user"(id) on delete cascade,
  word_id         uuid not null references vocabulary_words(id) on delete cascade,
  times_seen      int not null default 1 check (times_seen >= 0),
  times_suggested int not null default 0 check (times_suggested >= 0),
  mastered        boolean not null default false,
  first_seen_at   timestamptz not null default now(),
  last_seen_at    timestamptz not null default now(),
  unique (user_id, word_id)
);

create index idx_user_vocabulary_user
  on user_vocabulary(user_id, last_seen_at desc);

create table user_skill_stats (
  user_id          text primary key references "user"(id) on delete cascade,
  avg_clarity      numeric(5,2)
    check (avg_clarity is null or avg_clarity between 0 and 100),
  avg_delivery     numeric(5,2)
    check (avg_delivery is null or avg_delivery between 0 and 100),
  avg_content      numeric(5,2)
    check (avg_content is null or avg_content between 0 and 100),
  avg_vocabulary   numeric(5,2)
    check (avg_vocabulary is null or avg_vocabulary between 0 and 100),
  avg_grammar      numeric(5,2)
    check (avg_grammar is null or avg_grammar between 0 and 100),
  sessions_counted int not null default 0 check (sessions_counted >= 0),
  updated_at       timestamptz not null default now()
);

create trigger trg_user_skill_stats_updated_at
  before update on user_skill_stats
  for each row execute function set_updated_at();

create table xp_transactions (
  id              uuid primary key default gen_random_uuid(),
  user_id         text not null references "user"(id) on delete cascade,
  session_id      uuid references speaking_sessions(id) on delete set null,
  amount          int not null,
  reason          xp_reason not null,
  idempotency_key text not null unique,
  created_at      timestamptz not null default now()
);

create index idx_xp_transactions_user_id
  on xp_transactions(user_id, created_at desc);

create table daily_activity (
  id             uuid primary key default gen_random_uuid(),
  user_id        text not null references "user"(id) on delete cascade,
  activity_date  date not null,
  sessions_count int not null default 0 check (sessions_count >= 0),
  xp_earned      int not null default 0 check (xp_earned >= 0),
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now(),
  unique (user_id, activity_date)
);

create trigger trg_daily_activity_updated_at
  before update on daily_activity
  for each row execute function set_updated_at();

create index idx_daily_activity_user_date
  on daily_activity(user_id, activity_date desc);

create table user_stats (
  user_id             text primary key references "user"(id) on delete cascade,
  total_sessions      int not null default 0 check (total_sessions >= 0),
  scored_sessions     int not null default 0 check (scored_sessions >= 0),
  average_score       numeric(5,2) not null default 0
    check (average_score between 0 and 100),
  best_score          numeric(5,2) not null default 0
    check (best_score between 0 and 100),
  current_streak_days int not null default 0 check (current_streak_days >= 0),
  longest_streak_days int not null default 0 check (longest_streak_days >= 0),
  last_session_date   date,
  total_xp            bigint not null default 0 check (total_xp >= 0),
  updated_at          timestamptz not null default now(),

  constraint chk_user_stats_scored_sessions
    check (scored_sessions <= total_sessions)
);

create trigger trg_user_stats_updated_at
  before update on user_stats
  for each row execute function set_updated_at();

create index idx_user_stats_leaderboard
  on user_stats(total_xp desc, current_streak_days desc);

create table badges (
  id              uuid primary key default gen_random_uuid(),
  slug            text not null unique,
  name            text not null,
  description     text,
  icon            text,
  criteria_type   badge_criteria_type not null,
  criteria_value  int check (criteria_value is null or criteria_value >= 0),
  is_active       boolean not null default true,
  sort_order      int not null default 0,
  created_at      timestamptz not null default now()
);

alter table badges
add constraint chk_badge_criteria_value
check (criteria_type = 'manual' or criteria_value is not null);

create table user_badges (
  id         uuid primary key default gen_random_uuid(),
  user_id    text not null references "user"(id) on delete cascade,
  badge_id   uuid not null references badges(id) on delete cascade,
  session_id uuid references speaking_sessions(id) on delete set null,
  earned_at  timestamptz not null default now(),
  unique (user_id, badge_id)
);

create index idx_user_badges_user
  on user_badges(user_id, earned_at desc);

create view leaderboard with (security_invoker = true) as
select
  u.id as user_id,
  u.username,
  u.image,
  u.country_code,
  us.total_xp,
  us.current_streak_days,
  us.total_sessions,
  row_number() over (
    order by us.total_xp desc, us.current_streak_days desc
  ) as rank
from user_stats us
join "user" u on u.id = us.user_id
where u.is_active = true
  and u.deleted_at is null;

create view country_leaderboard with (security_invoker = true) as
select
  u.id as user_id,
  u.username,
  u.image,
  u.country_code,
  us.total_xp,
  us.current_streak_days,
  us.total_sessions,
  row_number() over (
    partition by u.country_code
    order by us.total_xp desc, us.current_streak_days desc
  ) as rank
from user_stats us
join "user" u on u.id = us.user_id
where u.is_active = true
  and u.deleted_at is null
  and u.country_code is not null;

create table leaderboard_snapshots (
  id                  uuid primary key default gen_random_uuid(),
  period              leaderboard_period not null,
  period_start        date not null,
  period_end          date not null,
  country_code        char(2)
    check (country_code is null or country_code ~ '^[A-Z]{2}$'),
  user_id             text not null references "user"(id) on delete cascade,
  rank                int not null check (rank > 0),
  total_xp            bigint not null check (total_xp >= 0),
  current_streak_days int not null default 0 check (current_streak_days >= 0),
  total_sessions      int not null default 0 check (total_sessions >= 0),
  created_at          timestamptz not null default now(),
  unique (period, period_start, period_end, country_code, user_id)
);

create index idx_leaderboard_snapshots_global
  on leaderboard_snapshots(period, period_start, rank)
  where country_code is null;

create index idx_leaderboard_snapshots_country
  on leaderboard_snapshots(period, period_start, country_code, rank)
  where country_code is not null;

create table notification_preferences (
  user_id                  text primary key references "user"(id) on delete cascade,
  daily_reminder_enabled   boolean not null default true,
  streak_risk_enabled      boolean not null default true,
  session_ready_enabled    boolean not null default true,
  reminder_time             time default '19:00',
  updated_at                timestamptz not null default now()
);

create trigger trg_notification_preferences_updated_at
  before update on notification_preferences
  for each row execute function set_updated_at();

create table notifications (
  id                   uuid primary key default gen_random_uuid(),
  user_id              text not null references "user"(id) on delete cascade,
  type                 notification_type not null,
  channel              notification_channel not null default 'in_app',
  delivery_status      notification_delivery_status not null default 'pending',
  title                text not null,
  body                 text,
  data                 jsonb not null default '{}'
    check (jsonb_typeof(data) = 'object'),
  provider_message_id  text,
  read_at              timestamptz,
  sent_at              timestamptz,
  failed_at            timestamptz,
  created_at           timestamptz not null default now()
);

create index idx_notifications_user_unread
  on notifications(user_id, created_at desc)
  where read_at is null;

create table audit_log (
  id            uuid primary key default gen_random_uuid(),
  actor_user_id text references "user"(id) on delete set null,
  action        text not null,
  entity_type   text not null,
  entity_id     text not null,
  old_value     jsonb,
  new_value     jsonb,
  source        text,
  created_at    timestamptz not null default now()
);

create index idx_audit_log_entity
  on audit_log(entity_type, entity_id, created_at desc);

create index idx_audit_log_actor
  on audit_log(actor_user_id, created_at desc);

create or replace view session_report with (security_invoker = true) as
select
  s.id                       as session_id,
  s.user_id,
  s.topic_id,
  t.title                    as topic_title,
  t.format                   as topic_format,
  tc.slug                    as topic_category,
  s.status,
  s.debate_stance,
  s.prep_time_seconds,
  s.speak_time_seconds,
  s.audio_s3_key,
  s.audio_expires_at,
  s.share_token_hash,
  s.share_enabled,
  s.share_expires_at,
  tr.raw_text                as transcript,
  sa.words_per_minute,
  sa.filler_word_count,
  sa.filler_rate,
  sa.pause_rate,
  sa.longest_pause_seconds,
  sa.unique_word_count,
  sa.lexical_diversity,
  sa.sentence_count,
  sa.average_sentence_length,
  sa.overall_score,
  sa.clarity_score,
  sa.delivery_score,
  sa.content_score,
  sa.vocabulary_score,
  sa.grammar_score,
  sa.strengths,
  sa.improvements,
  sa.coach_message,
  s.created_at,
  s.completed_at
from speaking_sessions s
join topics t on t.id = s.topic_id
join topic_categories tc on tc.id = t.category_id
left join transcripts tr on tr.session_id = s.id
left join session_analysis sa on sa.session_id = s.id;

create view session_summary with (security_invoker = true) as
select
  s.id                       as session_id,
  s.user_id,
  s.topic_id,
  t.title                    as topic_title,
  t.format                   as topic_format,
  tc.slug                    as topic_category,
  s.status,
  s.debate_stance,
  s.prep_time_seconds,
  s.speak_time_seconds,
  tr.raw_text                as transcript,
  sa.words_per_minute,
  sa.filler_word_count,
  sa.overall_score,
  sa.clarity_score,
  sa.delivery_score,
  sa.content_score,
  sa.vocabulary_score,
  sa.grammar_score,
  sa.strengths,
  sa.improvements,
  sa.coach_message,
  s.created_at,
  s.completed_at
from speaking_sessions s
join topics t on t.id = s.topic_id
join topic_categories tc on tc.id = t.category_id
left join transcripts tr on tr.session_id = s.id
left join session_analysis sa on sa.session_id = s.id;

create view session_share_public with (security_invoker = true) as
select
  s.id              as session_id,
  t.title           as topic_title,
  t.format          as topic_format,
  tc.slug           as topic_category,
  tr.raw_text       as transcript,
  sa.overall_score,
  sa.clarity_score,
  sa.delivery_score,
  sa.content_score,
  sa.vocabulary_score,
  sa.grammar_score,
  sa.words_per_minute,
  s.completed_at
from speaking_sessions s
join topics t on t.id = s.topic_id
join topic_categories tc on tc.id = t.category_id
left join transcripts tr on tr.session_id = s.id
left join session_analysis sa on sa.session_id = s.id
where
  s.share_enabled = true
  and (s.share_expires_at is null or s.share_expires_at > now());

insert into topic_categories (slug, name, icon, sort_order) values
  ('general',      'General',      '💬', 1),
  ('business',     'Business',     '💼', 2),
  ('education',    'Education',    '🎓', 3),
  ('culture',      'Culture',      '🎨', 4),
  ('sports',       'Sports',       '⚽', 5),
  ('philosophy',   'Philosophy',   '🧠', 6),
  ('politics',     'Politics',     '🏛️', 7),
  ('health',       'Health',       '🩺', 8),
  ('environment',  'Environment',  '🌱', 9),
  ('technology',   'Technology',   '💻', 10)
on conflict (slug) do nothing;

alter table "user" enable row level security;