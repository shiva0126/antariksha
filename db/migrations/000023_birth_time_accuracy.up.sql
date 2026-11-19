-- A recorded clock time is not necessarily an exact birth time.  Keep this
-- separate from the value itself so features that depend on lagna can refuse
-- false precision while Moon/date-only features remain available.
ALTER TABLE member_accounts
  ADD COLUMN birth_time_accuracy TEXT NOT NULL DEFAULT 'exact'
  CHECK (birth_time_accuracy IN ('exact', 'approximate', 'unknown'));
