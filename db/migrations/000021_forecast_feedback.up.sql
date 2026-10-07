-- Opt-in answers to "did this period feel as described?", used only to
-- measure and tune forecast accuracy. Only the period, area, tone shown and
-- answer are kept; rows go with the account.
CREATE TABLE forecast_feedback (
 account_id TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 period_from DATE NOT NULL,
 period_to DATE NOT NULL CHECK (period_to > period_from),
 area TEXT NOT NULL CHECK (area IN ('career','money','partnership','home','learning','effort','wellbeing','change','travel')),
 tone TEXT NOT NULL CHECK (tone IN ('supportive','mixed','challenging')),
 response TEXT NOT NULL CHECK (response IN ('yes','partly','no')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 PRIMARY KEY (account_id, period_from, area)
);
