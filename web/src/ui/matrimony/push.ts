import { memberAPI } from '../community/shared';

// Web push: the service worker (public/sw.js) shows alerts sent by the server.
const b64ToBytes = (b64: string) => {
  const s = atob((b64 + '='.repeat((4 - b64.length % 4) % 4)).replace(/-/g, '+').replace(/_/g, '/'));
  return Uint8Array.from(s, c => c.charCodeAt(0));
};

export const pushSupported = () => 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window;

export async function pushEnabled(): Promise<boolean> {
  if (!pushSupported()) return false;
  const reg = await navigator.serviceWorker.getRegistration();
  return !!(await reg?.pushManager.getSubscription());
}

export async function enablePush(): Promise<void> {
  if (!pushSupported()) throw new Error('This browser does not support notifications. On iPhone, add Astrisk to your Home Screen first.');
  if (await Notification.requestPermission() !== 'granted') throw new Error('Notifications were not allowed. You can allow them in the browser settings.');
  const reg = await navigator.serviceWorker.register('/sw.js');
  await navigator.serviceWorker.ready;
  const { key } = await memberAPI<{ key: string }>('/api/push/key');
  const sub = await reg.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: b64ToBytes(key) });
  await memberAPI('/api/push/subscribe', 'POST', sub.toJSON());
}

export async function disablePush(): Promise<void> {
  const sub = await (await navigator.serviceWorker.getRegistration())?.pushManager.getSubscription();
  if (!sub) return;
  await memberAPI('/api/push/subscribe', 'DELETE', { endpoint: sub.endpoint }).catch(() => {});
  await sub.unsubscribe();
}
