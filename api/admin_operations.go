package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Aggregate operational counts only; never conversation content or birth details.
func (s *Server) adminOperations(w http.ResponseWriter, r *http.Request, _ string) {
	var raw []byte
	err := s.membersDB().QueryRow(r.Context(), `SELECT json_build_object(
 'as_of',now(),
 'registrations_7d',(SELECT count(*) FROM member_accounts WHERE created_at>=now()-interval '7 days'),
 'registrations_30d',(SELECT count(*) FROM member_accounts WHERE created_at>=now()-interval '30 days'),
 'verified_emails',(SELECT count(*) FROM member_accounts WHERE email_verified_at IS NOT NULL),
 'accounts_with_active_sessions',(SELECT count(DISTINCT account_id) FROM member_sessions WHERE expires_at>now()),
 'active_matrimony_profiles',(SELECT count(*) FROM matrimony_profiles WHERE active AND NOT hidden),
 'incomplete_birthplaces',(SELECT count(*) FROM matrimony_profiles p JOIN member_accounts a ON a.id=p.account_id WHERE p.active AND p.horoscope_visible AND a.birth_place IS NULL),
 'interests_created_7d',(SELECT count(*) FROM matrimony_interests WHERE created_at>=now()-interval '7 days'),
 'those_interests_accepted',(SELECT count(*) FROM matrimony_interests WHERE created_at>=now()-interval '7 days' AND status='accepted'),
 'pending_photo_reviews',(SELECT count(*) FROM matrimony_verifications WHERE status='pending'),
 'oldest_photo_review',(SELECT min(created_at) FROM matrimony_verifications WHERE status='pending'),
 'open_reports_over_24h',((SELECT count(*) FROM community_reports WHERE status='open' AND created_at<now()-interval '24 hours')+(SELECT count(*) FROM matrimony_reports WHERE status='open' AND created_at<now()-interval '24 hours')),
 'admins_with_authenticator',(SELECT count(*) FROM member_totp t JOIN member_accounts a ON a.id=t.account_id WHERE t.enabled AND a.role='superadmin'))`).Scan(&raw)
	if err != nil {
		problem(w, 503, fmt.Errorf("operations summary unavailable"))
		return
	}
	_, keyErr := totpCipher()
	writeJSON(w, 200, map[string]any{"metrics": json.RawMessage(raw), "email_delivery_configured": s.mailer != nil, "phone_verification_configured": phoneReady(), "admin_authenticator_available": keyErr == nil})
}
