// Astrisk service worker: shows push alerts and opens the right page on tap.
// It deliberately does not cache API responses: member data stays fresh and
// is never stored offline on shared devices.
self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', event => event.waitUntil(self.clients.claim()));

self.addEventListener('push', event => {
  let data = { title: 'Astrisk', body: 'You have new activity.', url: '/#matrimony/interests' };
  try { data = { ...data, ...event.data.json() }; } catch { /* plain text or empty */ }
  event.waitUntil(self.registration.showNotification(data.title, {
    body: data.body, icon: '/icon-192.png', badge: '/icon-192.png', data: { url: data.url }, tag: 'astrisk-activity', renotify: true,
  }));
});

self.addEventListener('notificationclick', event => {
  event.notification.close();
  const target = new URL(event.notification.data?.url || '/', self.location.origin);
  if (target.origin !== self.location.origin) return;
  event.waitUntil((async () => {
    const wins = await self.clients.matchAll({ type: 'window', includeUncontrolled: true });
    for (const w of wins) { if ('focus' in w) { await w.navigate(target.href).catch(() => {}); return w.focus(); } }
    return self.clients.openWindow(target.href);
  })());
});

// A fetch handler (network only) makes the app installable.
self.addEventListener('fetch', () => {});
