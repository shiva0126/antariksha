# Antariksha: profiles, community, family and matrimony

Status: approved product direction; implementation roadmap, not a claim that
these features are deployed. Updated 2026-09-29.

## Product

One account supports a private personal profile, an optional community presence,
a private family space and an explicitly enabled matrimony profile. Panchang,
Kundali and chart interpretation remain available. Joining the community does
not enroll someone in matchmaking; joining a family does not publish a profile.

Navigation: Home feed · Explore · Create post · Family · Matches · Astrology · Me.
Use a responsive mobile bottom navigation with secondary destinations in Me.
The first release is a mobile-friendly web app; native apps are a later decision.
Continue native Go/PostgreSQL deployment without Docker.

## Existing application and migration boundary

- Saved birth profiles currently live in browser localStorage, not authenticated
  user records (`web/src/ui/common/profiles.ts`). Offer a preview-and-confirm import
  into private account storage after sign-in. Never automatically publish them.
- Chat sessions currently use session IDs and birth data. Add authenticated
  ownership before associating chat history with social or matrimony accounts.
  Knowledge of a session ID must not grant access. Do not claim legacy sessions
  merely from a matching birth date, phone number or name.
- The existing chart comparison is an astrology tool, not a matchmaking service.
- Public launch requires an authorization audit of existing history, deletion,
  export and sharing routes as well as new endpoints.

## Account and profile

Phone OTP sign-in with international E.164 normalization, expiry, one-time use,
attempt/resend limits and delivery-cost limits. Never expose OTPs through API
responses or production logs. Development fake delivery must be explicit and
disabled in production. Add recovery email or passkeys, session/device management,
logout-all, phone-change reauthentication and recycled-number recovery controls.

Use opaque account IDs. Encrypt contact data at rest; keyed lookup digests support
phone uniqueness without logging cleartext. Session cookies are HttpOnly and
SameSite, Secure in HTTPS deployments. Enforce CSRF and exact allowed origins on
state-changing requests. Replace permissive cross-origin behavior for account APIs.

Progressive fields:

- Display name, handle, bio, city, languages and profile photo.
- Private date of birth and optional birth time/place with accuracy indicator.
- Education, occupation, optional employer and income band with field visibility.
- Interests, hobbies, skills, learning goals and preferred shared activities.
- Optional values, lifestyle, communication preferences and relationship intentions.
- Matrimony-only preferences: marriage timeline, relocation, children, family
  involvement and partner criteria. Allow skips and later editing.

Every field has an owner, source, visibility and update date. Distinguish
self-declared, connected-account-confirmed, photo-verified and ID-verified data.
Phone verification confirms number access only. Collect no identity document
unless required for an explicitly chosen verification flow.

## Personal posts and feed

MVP: text and photo carousel posts (up to 10 images), captions, alt text, interest
tags, optional coarse location, drafts, edit/delete, chronological feed, likes,
comments, bookmarks and follow requests. Private accounts approve followers.
Likes are unique per account/post; bookmarks are private. Cursor pagination and
idempotent write requests prevent duplicates and unstable scrolling.

Post audience choices:

| Audience | Viewers |
| --- | --- |
| Only me | Author only; draft and new-post default |
| Followers | Accepted followers, excluding blocked accounts |
| Family | Members of explicitly selected family groups |
| Community | Signed-in eligible community members |

Posting does not automatically share birth details, phone numbers, family trees,
or matrimony status. Preview the audience before publishing. Do not permit
reposting private content to wider audiences. Re-check visibility for comments,
notifications, search, recommendations, media and shared links, not just feed rows.
Removing a follower, leaving a group or blocking an account revokes access.

Media pipeline: authenticated upload → size/type/dimension validation → decode
and re-encode → remove EXIF/GPS → moderation → thumbnail variants → publish.
Reject unsupported formats, oversized/decompression-bomb images and executable
content. Store media outside the public web root; authorize delivery or issue
short-lived signed URLs. Delete originals/derivatives on expiry or deletion,
with documented backup retention. Apply storage and upload quotas.

MVP safety: report post/comment/account, block, mute, comment controls, spam
limits, moderator queue, appeals and audit records. Deleted or hidden posts
must disappear from feeds and notifications. Private content never enters a
public search index. Public follower counts are optional; do not rank people
by popularity for matrimony.

Later: stories, short videos, mentions with approval, hobby groups and events.
Video requires transcoding, storage budgets and stronger moderation; defer it
until photo-post operations are stable. Disable precise live location by default.

## Personality and personal character

Offer a customizable avatar, interest badges, personal goals, voice introduction
and editable profile prompts. An avatar is creative self-expression and must
not be labeled a verified likeness. Keep verification photos separate.

Personality assessment uses an established IPIP scale with versioned items,
scoring and optional answers. Translated or shortened questionnaires need their
own validation; do not claim accuracy from an unvalidated adaptation. Results
remain private unless shared. Do not issue moral character, honesty, fidelity,
mental-health diagnosis or marriage-success scores.

AI can draft a bio or suggest hobbies and conversation topics from chosen inputs.
Show the basis, permit corrections and require approval before saving or publishing.
Do not infer sensitive traits from photos, social activity or family members.
Astrology is a separate optional interpretation and never overrides self-reported
interests or verifies personality.

## Family tree

Represent people separately from login accounts. Relationships support biological,
adoptive, step, guardian and partner links. Record provenance and disputed/pending
states. One account can manage a private tree without creating accounts for relatives.
No automatic merging by name, phone, birth details or suspected relationship.

Support invite/accept/decline, profile claiming with verification, group viewer/editor
roles, change history, correction requests, revocation and scoped exports. Obtain
permission before sharing identifiable living relatives; minimal private placeholders
can be added without publishing them. Do not collect contact lists automatically.
Children's family entries stay private and cannot join social discovery or matrimony.

Family-assisted matrimony uses revocable grants: view, shortlist, suggest edits,
or request an introduction. The adult profile owner approves publication and contact
sharing. Parent access must not silently expose private conversations.

## Matrimony journey and recommendations

Explicit enrollment → eligible private profile → preview publication → discovery
→ interest → mutual acceptance → conversation → optional family introduction.
Pause/hide/close enrollment without deleting the personal account.

Use reciprocal preferences and eligibility first, then explainable compatibility
on values, lifestyle, location, goals and shared interests. Missing answers are
unknown, not negative. Give specific agreements and discussion points; no invented
precision or guaranteed success. Optional Kundali comparison has its own screen
and never becomes a hidden exclusion rule.

Do not infer caste, religion, sexuality or health from names, images or social
profiles. Avoid attractiveness, wealth-status and engagement-popularity rankings.
Show verification type and limitations. Private contact exchange needs explicit
permission even after a match. Include block/report/unmatch, spam limits,
moderation, safety guidance and appeals from the first matchmaking release.

## Social accounts

Users add their own profile URLs or connect supported providers with minimal OAuth
scopes. Label manual links unverified. Connected account control is not identity,
employment, education or character verification. Imports have a preview, per-field
selection and revocation/deletion controls. Encrypt tokens and never request passwords.

No automatic phone-number reverse lookup, hidden account discovery, face search,
address-book scraping or dossiers about prospective partners. Do not promise
personal Instagram imports where the provider API does not support them.

## Backend and storage

Keep Go as a modular monolith and PostgreSQL as the primary database. Add a worker
for SMS delivery, media processing, notifications, exports and deletion. Use a
transactional outbox with retries and idempotency. Native private disk storage works
for development; choose private object storage before a public multi-instance launch.

Proposed table groups (versioned migrations, not yet applied):

- Identity: accounts, account_contacts, verification_challenges, sessions,
  consent_events, privacy_settings, identity_verifications.
- Profiles: member_profiles, profile_fields, interests, member_interests,
  assessment_versions, assessment_answers, avatar_settings, social_connections.
- Community: posts, post_media, media_assets, follows, post_likes, comments,
  bookmarks, blocks, reports, moderation_actions.
- Family: people, family_groups, memberships, relationships, invitations,
  profile_claims, delegated_permissions, relationship_history.
- Matrimony: matrimony_profiles, partner_preferences, interests_sent, matches,
  conversations, conversation_members, messages, contact_share_grants.
- Operations: notifications, outbox_jobs, audit_events, export_jobs, deletion_jobs.

Foreign keys, unique constraints and transactions enforce ownership and lifecycle
rules. PostgreSQL adjacency tables are sufficient for family relationships initially;
no graph database is required. Private user data is never mixed into the shared
classical-text pgvector corpus. Any future user-content embeddings need separate
authorization, retention and deletion paths.

API groups: /api/auth/*, /api/me/*, /api/profiles/*, /api/posts/*, /api/feed,
/api/media/*, /api/follows/*, /api/families/*, /api/matrimony/*,
/api/conversations/*, /api/reports and restricted /api/admin/moderation/*.
Access rules must be centralized and reused by every read, mutation and worker.

## Delivery sequence and acceptance gates

| Phase | Scope | Required evidence before moving on |
| --- | --- | --- |
| M0 | Authentication, consent, account ownership, private profiles | OTP replay/expiry/limits; account A cannot access B; logout revokes session; protected legacy chat routes |
| M1 | Interests, avatar, profile visibility and optional assessment | Users edit/delete/export their data; private fields absent from other users' API responses |
| M2 | Photo posts, feed, follows, likes/comments/bookmarks, moderation | Audience matrix tests; revoked followers lose access; blocks apply to media too; upload validation and pagination tests |
| M3 | Private family tree, invitations and delegates | Nonmembers denied; child entries excluded from discovery; revoked delegates lose access; no unauthorized profile claiming |
| M4 | Matrimony enrollment, reciprocal discovery, interests and chat | Only eligible consenting adults discoverable; mutual acceptance enforced; block/unmatch/contact permissions verified |
| M5 | Social connections, optional identity provider, notification delivery | OAuth ownership/CSRF checks; minimal scopes; revocation and token deletion; provider failure recovery |
| M6 | Explainable recommendations, calls, stories/video and events | Evaluated matching explanations, moderation capacity, storage/cost limits and operational monitoring |

Each phase includes migrations, API contract, UI, authorization tests and native
deployment checks. Existing Panchang/chart routes must retain regression coverage.
No phase is marked complete based on mock delivery, placeholder verification,
synthetic identities or an untested UI alone.

## External dependencies and launch readiness

Production SMS needs a funded provider, approved sender/templates as applicable
and delivery monitoring. OAuth needs provider registrations and callbacks. Email,
identity checks, calls and media storage each need explicitly configured providers.
The existing exhausted embedding credits do not block account or social features.

Community and matrimony launch requires an age/eligibility policy reviewed for the
launch jurisdictions, named moderation operators, report response procedures,
privacy notices, retention periods and account recovery. Use an adult-only community
MVP; family records do not create child social accounts. Plan against India's DPDP
requirements and phased commencement with a launch-specific legal review.

Preserve the current free-for-users product direction. Budget ongoing SMS, storage,
media processing, LLM and moderation costs; no paid tier is assumed by this plan.

Measure profile completion, successful mutual conversations, helpful introductions,
report response time and privacy failures. Follower growth and time spent scrolling
are not proxies for successful matches.

## Research references

- Shaadi contact privacy: https://support.shaadi.com/support/solutions/articles/48000755468-i-want-to-change-or-hide-my-phone-number
- Bumble safety features: https://safety.bumble.com/en_US
- IPIP measures and usage: https://www.ipip.ori.org/
- LinkedIn OIDC limitations: https://learn.microsoft.com/en-us/linkedin/consumer/integrations/self-serve/sign-in-with-linkedin-v2
- Meta Instagram API limitations: https://www.postman.com/meta/instagram/folder/u4g5a2a/instagram-api-with-facebook-login
- FamilySearch living-relative privacy: https://www.familysearch.org/en/help/helpcenter/article/who-can-see-my-living-relatives-in-family-tree
- India DPDP Rules notification: https://www.meity.gov.in/static/uploads/2025/11/53450e6e5dc0bfa85ebd78686cadad39.pdf

References informed the research on 2026-09-29. Recheck provider capabilities and
legal commencement dates before implementing integrations or launching publicly.
