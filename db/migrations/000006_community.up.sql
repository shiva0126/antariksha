ALTER TABLE member_accounts ADD COLUMN recovery_hash TEXT;
CREATE TABLE member_settings (
 account_id TEXT PRIMARY KEY REFERENCES member_accounts(id) ON DELETE CASCADE,
 community BOOLEAN NOT NULL DEFAULT false,
 birth_date DATE,
 avatar TEXT NOT NULL DEFAULT 'sun',
 accent TEXT NOT NULL DEFAULT '#d6b467',
 public_bio TEXT NOT NULL DEFAULT '',
 interests TEXT[] NOT NULL DEFAULT '{}',
 links JSONB NOT NULL DEFAULT '[]',
 preferences JSONB NOT NULL DEFAULT '{}'
);
CREATE TABLE member_blocks (
 actor TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 target TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 PRIMARY KEY(actor,target), CHECK(actor<>target)
);
CREATE TABLE member_follows (
 follower TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 target TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 accepted BOOLEAN NOT NULL DEFAULT false,
 PRIMARY KEY(follower,target), CHECK(follower<>target)
);
CREATE TABLE family_groups (
 id TEXT PRIMARY KEY, owner TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 name TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE family_members (
 group_id TEXT REFERENCES family_groups(id) ON DELETE CASCADE,
 account_id TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 accepted BOOLEAN NOT NULL DEFAULT false,
 role TEXT NOT NULL DEFAULT 'viewer' CHECK(role IN ('viewer','editor')),
 PRIMARY KEY(group_id,account_id)
);
CREATE TABLE family_people (
 id TEXT PRIMARY KEY, group_id TEXT NOT NULL REFERENCES family_groups(id) ON DELETE CASCADE,
 name TEXT NOT NULL, note TEXT NOT NULL DEFAULT '', created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE family_edges (
 group_id TEXT NOT NULL REFERENCES family_groups(id) ON DELETE CASCADE,
 source TEXT REFERENCES family_people(id) ON DELETE CASCADE,
 target TEXT REFERENCES family_people(id) ON DELETE CASCADE,
 relation TEXT NOT NULL CHECK(relation IN ('parent','adoptive_parent','step_parent','guardian','partner','sibling')),
 PRIMARY KEY(source,target,relation), CHECK(source<>target)
);
CREATE TABLE community_posts (
 id BIGSERIAL PRIMARY KEY, owner TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 caption TEXT NOT NULL, audience TEXT NOT NULL DEFAULT 'private' CHECK(audience IN ('private','followers','family','community')),
 family_id TEXT REFERENCES family_groups(id) ON DELETE CASCADE,
 hidden BOOLEAN NOT NULL DEFAULT false,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 CHECK(audience<>'family' OR family_id IS NOT NULL)
);
CREATE INDEX community_posts_feed ON community_posts(id DESC);
CREATE TABLE community_media (
 id TEXT PRIMARY KEY, owner TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 post_id BIGINT REFERENCES community_posts(id) ON DELETE CASCADE,
 data BYTEA NOT NULL, mime TEXT NOT NULL, alt TEXT NOT NULL DEFAULT '',
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX community_media_post ON community_media(post_id);
CREATE TABLE community_reactions (
 account_id TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 post_id BIGINT REFERENCES community_posts(id) ON DELETE CASCADE,
 kind TEXT CHECK(kind IN ('like','bookmark')),
 PRIMARY KEY(account_id,post_id,kind)
);
CREATE TABLE community_comments (
 id BIGSERIAL PRIMARY KEY, owner TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 post_id BIGINT NOT NULL REFERENCES community_posts(id) ON DELETE CASCADE,
 body TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE community_reports (
 id BIGSERIAL PRIMARY KEY, reporter TEXT REFERENCES member_accounts(id) ON DELETE SET NULL,
 post_id BIGINT REFERENCES community_posts(id) ON DELETE CASCADE,
 reason TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open','dismissed','hidden')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE member_audit (
 id BIGSERIAL PRIMARY KEY, actor TEXT REFERENCES member_accounts(id) ON DELETE SET NULL,
 action TEXT NOT NULL, subject TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE matrimony_profiles (
 account_id TEXT PRIMARY KEY REFERENCES member_accounts(id) ON DELETE CASCADE,
 active BOOLEAN NOT NULL DEFAULT false, details JSONB NOT NULL DEFAULT '{}',
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE matrimony_interests (
 sender TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 recipient TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','accepted','declined')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 PRIMARY KEY(sender,recipient), CHECK(sender<>recipient)
);
CREATE TABLE member_messages (
 id BIGSERIAL PRIMARY KEY, sender TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 recipient TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 body TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), CHECK(sender<>recipient)
);
CREATE TABLE member_phone (
 account_id TEXT PRIMARY KEY REFERENCES member_accounts(id) ON DELETE CASCADE,
 phone_hash TEXT UNIQUE NOT NULL, ciphertext BYTEA NOT NULL, verified BOOLEAN NOT NULL DEFAULT false,
 attempts INTEGER NOT NULL DEFAULT 0, sent_at TIMESTAMPTZ
);
CREATE TABLE chat_owners (
 session_id TEXT PRIMARY KEY REFERENCES chat_sessions(id) ON DELETE CASCADE,
 account_id TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE
);

CREATE FUNCTION member_blocked(a TEXT,b TEXT) RETURNS BOOLEAN LANGUAGE SQL STABLE AS $$
 SELECT EXISTS(SELECT 1 FROM member_blocks WHERE (actor=a AND target=b) OR (actor=b AND target=a))
$$;
CREATE FUNCTION family_access(g TEXT,u TEXT) RETURNS BOOLEAN LANGUAGE SQL STABLE AS $$
 SELECT EXISTS(SELECT 1 FROM family_groups f WHERE f.id=g AND (f.owner=u OR
 (NOT member_blocked(f.owner,u) AND EXISTS(SELECT 1 FROM family_members m WHERE m.group_id=g AND m.account_id=u AND m.accepted))))
$$;
CREATE FUNCTION post_access(p BIGINT,u TEXT) RETURNS BOOLEAN LANGUAGE SQL STABLE AS $$
 SELECT EXISTS(SELECT 1 FROM community_posts c WHERE c.id=p AND
 (c.owner=u OR (NOT c.hidden AND NOT member_blocked(c.owner,u) AND
 EXISTS(SELECT 1 FROM member_settings s WHERE s.account_id=c.owner AND s.community) AND
 EXISTS(SELECT 1 FROM member_settings s WHERE s.account_id=u AND s.community) AND (
 c.audience='community' OR
 (c.audience='followers' AND EXISTS(SELECT 1 FROM member_follows f WHERE f.follower=u AND f.target=c.owner AND f.accepted)) OR
 (c.audience='family' AND family_access(c.family_id,u))))))
$$;
