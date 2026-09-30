import { useEffect, useState } from 'react';
import { Avatar, Feedback, Field, memberAPI, useAction, type Family } from './shared';
type Post = { id: number; caption: string; audience: string; family_id?: string; mine: boolean; hidden: boolean; handle: string; author_id: string; avatar: string; accent: string; created_at: string; media: { id: string; alt: string }[]; likes: number; liked: boolean; bookmarked: boolean };
type Comment = { id: number; body: string; handle: string; mine: boolean };
function PostCard({ post, refresh }: { post: Post; refresh: () => Promise<void> }) {
  const action = useAction(); const [comments, setComments] = useState<Comment[] | null>(null); const [body, setBody] = useState('');
  const loadComments = async () => setComments(await memberAPI(`/api/community/posts/${post.id}/comments`));
  return <article className="card social-post">
    <header className="community-row"><Avatar kind={post.avatar} accent={post.accent} /><div><strong>@{post.handle}</strong><p className="small muted">{new Date(post.created_at).toLocaleString()} · {post.audience}{post.hidden ? ' · Hidden by moderation' : ''}</p></div></header>
    <p className="post-caption">{post.caption}</p>
    {post.media.length > 0 && <div className="photo-carousel">{post.media.map(m => <img key={m.id} loading="lazy" src={`/api/community/media/${m.id}`} alt={m.alt || 'Photo shared by ' + post.handle} />)}</div>}
    <div className="community-actions">
      <button disabled={action.busy} aria-pressed={post.liked} onClick={() => void action.run(async () => { await memberAPI(`/api/community/posts/${post.id}/react`, 'POST', { kind: 'like', active: !post.liked }); await refresh(); })}>{post.liked ? '♥' : '♡'} {post.likes}</button>
      <button disabled={action.busy} aria-pressed={post.bookmarked} onClick={() => void action.run(async () => { await memberAPI(`/api/community/posts/${post.id}/react`, 'POST', { kind: 'bookmark', active: !post.bookmarked }); await refresh(); })}>{post.bookmarked ? 'Saved' : 'Save'}</button>
      <button onClick={() => void action.run(loadComments)}>Comments</button>
      {post.mine ? <><button onClick={() => { const caption = prompt('Edit caption', post.caption); if (caption !== null) void action.run(async () => { await memberAPI(`/api/community/posts/${post.id}`, 'PUT', { caption, audience: post.audience, family_id: post.family_id || "", media: post.media.map(m => m.id) }); await refresh(); }); }}>Edit caption</button><button onClick={() => { if (confirm('Delete this post and its photos?')) void action.run(async () => { await memberAPI(`/api/community/posts/${post.id}`, 'DELETE'); await refresh(); }); }}>Delete</button></> : <><button onClick={() => { const reason = prompt('Why are you reporting this post?'); if (reason) void action.run(async () => { await memberAPI(`/api/community/posts/${post.id}/report`, 'POST', { reason }); action.setMessage('Report submitted for review.'); }); }}>Report</button><button onClick={() => { if (confirm(`Block @${post.handle}? Connections and message access will be removed.`)) void action.run(async () => { await memberAPI('/api/community/blocks', 'POST', { target: post.author_id, block: true }); await refresh(); }); }}>Block</button></>}
    </div>
    {comments && <div className="post-comments">{comments.length === 0 && <p>No comments yet.</p>}{comments.map(c => <p key={c.id}><strong>@{c.handle}</strong> {c.body} {(c.mine || post.mine) && <button className="link-button" onClick={() => void action.run(async () => { await memberAPI(`/api/community/comments/${c.id}`, 'DELETE'); await loadComments(); })}>Remove</button>}</p>)}
      <form onSubmit={e => { e.preventDefault(); void action.run(async () => { await memberAPI(`/api/community/posts/${post.id}/comments`, 'POST', { body }); setBody(''); await loadComments(); }); }}><Field label="Your comment"><input required maxLength={1000} value={body} onChange={e => setBody(e.target.value)} /></Field><button disabled={action.busy}>Comment</button></form>
    </div>}<Feedback {...action} />
  </article>;
}
export function Feed() {
  const action = useAction(); const [posts, setPosts] = useState<Post[]>([]), [mode, setMode] = useState('all');
  const [caption, setCaption] = useState(''), [audience, setAudience] = useState('private'), [familyId, setFamilyId] = useState('');
  const [families, setFamilies] = useState<Family[]>([]), [photos, setPhotos] = useState<File[]>([]), [alt, setAlt] = useState('');
  const [uploadKey, setUploadKey] = useState(0), [more, setMore] = useState(false);
  const refresh = async () => { const data = await memberAPI<Post[]>(`/api/community/posts?mode=${mode}`); setPosts(data); setMore(data.length === 20); };
  useEffect(() => { void action.run(refresh); }, [mode]);
  useEffect(() => { memberAPI<Family[]>('/api/families').then(f => setFamilies(f.filter(x => x.mine || x.accepted))).catch(() => {}); }, []);
  return <div className="social-layout"><aside className="card post-composer"><p className="kicker">Your everyday universe</p><h2>Share a moment</h2><p>A thought, a hobby, a little piece of your day.</p>
    <form onSubmit={e => { e.preventDefault(); void action.run(async () => { const media: string[] = []; for (const photo of photos) { const form = new FormData(); form.append('photo', photo); form.append('alt', alt); const result = await memberAPI<{ id: string }>('/api/community/media', 'POST', form); media.push(result.id); } await memberAPI('/api/community/posts', 'POST', { caption, audience, family_id: audience === 'family' ? familyId : '', media }); setCaption(''); setPhotos([]); setUploadKey(k => k + 1); action.setMessage(audience === 'private' ? 'Saved privately. Find it under My posts.' : 'Your post has been shared.'); await refresh(); }); }}>
      <Field label="Caption"><textarea rows={5} maxLength={2200} value={caption} onChange={e => setCaption(e.target.value)} placeholder="What has been inspiring you?" /></Field>
      <Field label="Photos (up to 10)"><input key={uploadKey} type="file" multiple accept="image/jpeg,image/png" onChange={e => { const files = Array.from(e.target.files || []); if (files.length > 10) { action.setMessage('Choose at most 10 photos.'); e.target.value = ''; return; } setPhotos(files); }} /></Field>
      {photos.length > 0 && <><p>{photos.length} photo(s) selected · JPEG/PNG, up to 8 MB each</p><Field label="Photo description"><input maxLength={300} value={alt} onChange={e => setAlt(e.target.value)} /></Field></>}
      <Field label="Who can see this?"><select value={audience} onChange={e => setAudience(e.target.value)}><option value="private">Only me</option><option value="followers">Approved followers</option><option value="family">A family group</option><option value="community">Signed-in community</option></select></Field>
      {audience === 'family' && <Field label="Family group"><select required value={familyId} onChange={e => setFamilyId(e.target.value)}><option value="">Choose a group</option>{families.map(f => <option key={f.id} value={f.id}>{f.name}</option>)}</select></Field>}
      <button className="primary" disabled={action.busy || (!caption.trim() && !photos.length)}> {audience === 'private' ? 'Save privately' : 'Share post'}</button>
    </form><Feedback {...action} /></aside><div className="social-feed"><div className="segmented">{[['all', 'Community'], ['following', 'Following'], ['mine', 'My posts'], ['saved', 'Saved']].map(([id, label]) => <button key={id} className={mode === id ? 'active' : ''} onClick={() => setMode(id)}>{label}</button>)}</div>
    {!posts.length && <div className="card"><h3>A little space for your story</h3><p>No visible posts yet. Enable community in Profile & interests, then share your first moment or follow someone.</p></div>}
    {posts.map(p => <PostCard key={p.id} post={p} refresh={refresh} />)}
    {more && <button disabled={action.busy} onClick={() => void action.run(async () => { const data = await memberAPI<Post[]>(`/api/community/posts?mode=${mode}&before=${posts[posts.length - 1].id}`); setPosts([...posts, ...data]); setMore(data.length === 20); })}>Load older posts</button>}
  </div></div>;
}
