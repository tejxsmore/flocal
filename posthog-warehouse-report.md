# PostHog Data Warehouse — Setup Report

## Summary

3 of 4 detected sources were connected to PostHog automatically. OpenAI requires a special Admin API key and needs to be finished in the browser.

---

## Sources Connected

### 1. PostgreSQL (Supabase) — Connected

- **Source ID:** `01a01372-ee0d-0000-f726-44afe1de4629`
- **Host:** `aws-0-us-east-1.pooler.supabase.com` (Session Pooler)
- **Tables synced:** 50 tables from the `public` schema
- **Sync strategy:** Incremental (`updated_at` or `created_at`) for 39 tables; full refresh for 11 tables

<details>
<summary>Full table list</summary>

| Table | Sync type | Incremental field |
|---|---|---|
| account | incremental | updated_at |
| ai_provider_logs | incremental | created_at |
| audit_log | incremental | created_at |
| badges | incremental | created_at |
| collections | incremental | created_at |
| country_leaderboard | full_refresh | — |
| daily_activity | incremental | updated_at |
| daily_challenge_completions | full_refresh | — |
| daily_challenges | incremental | created_at |
| dodo_webhook_events | incremental | created_at |
| grammar_corrections | incremental | created_at |
| invoices | incremental | created_at |
| leaderboard | full_refresh | — |
| leaderboard_snapshots | incremental | created_at |
| notification_preferences | incremental | updated_at |
| notifications | incremental | created_at |
| payment_disputes | incremental | created_at |
| payment_refunds | incremental | created_at |
| payments | incremental | created_at |
| session | incremental | updated_at |
| session_analysis | incremental | created_at |
| session_report | full_refresh | — |
| session_share_public | full_refresh | — |
| session_summary | full_refresh | — |
| speaking_sessions | incremental | updated_at |
| subscription_events | incremental | created_at |
| subscription_plans | incremental | created_at |
| subscriptions | incremental | updated_at |
| topic_bookmarks | incremental | created_at |
| topic_categories | incremental | created_at |
| topic_collections | incremental | created_at |
| topic_spins | incremental | created_at |
| topic_submissions | incremental | created_at |
| topic_usage_stats | full_refresh | — |
| topics | incremental | updated_at |
| transcripts | incremental | created_at |
| user | incremental | updated_at |
| user_badges | full_refresh | — |
| user_category_preferences | incremental | created_at |
| user_current_plan | full_refresh | — |
| user_focus_areas | incremental | created_at |
| user_preferences | incremental | updated_at |
| user_skill_stats | incremental | updated_at |
| user_stats | incremental | updated_at |
| user_topic_stats | full_refresh | — |
| user_vocabulary | full_refresh | — |
| verification | incremental | updated_at |
| vocabulary_suggestions | incremental | created_at |
| vocabulary_words | incremental | created_at |
| xp_transactions | incremental | created_at |

</details>

---

### 2. Resend — Connected

- **Source ID:** `01a01377-9ec3-0000-425b-0241d386163a`
- **Table prefix:** `resend_`
- **Tables synced:** 5 tables (full refresh)
  - `audiences`, `broadcasts`, `domains`, `emails`, `contacts`

---

### 3. Deepgram — Connected

- **Source ID:** `01a0137f-387b-0000-6141-eca004400964`
- **Table prefix:** `deepgram_`
- **Tables synced:** 6 tables
  - `projects`, `members`, `keys`, `balances`, `invites` — full refresh
  - `requests` — incremental (`created` field)

---

## Sources Requiring Browser Setup

### OpenAI — Manual setup needed

The OpenAI source requires an **Admin API key** (`sk-admin-...`). This is different from the project key in your `.env` — only organization owners can create one.

**Steps:**
1. Go to [platform.openai.com/settings/organization/admin-keys](https://platform.openai.com/settings/organization/admin-keys)
2. Create a new Admin API key
3. Open the link below and paste it in

[Connect OpenAI in PostHog →](https://us.posthog.com/project/563478/data-warehouse/new-source?kind=OpenAI&utm_source=wizard&utm_campaign=warehouse-source)

---

## Files Modified or Created

- `posthog-warehouse-report.md` (this file) — created/updated as a setup summary

No application source files were modified.
