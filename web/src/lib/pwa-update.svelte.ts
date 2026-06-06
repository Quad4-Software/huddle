import { registerSW } from 'virtual:pwa-register';

const CHECK_INTERVAL_MS = 15 * 60 * 1000;

class PwaUpdateStore {
  pending = $state(false);
  offlineReady = $state(false);
}

export const pwaUpdate = new PwaUpdateStore();

let applyUpdate: ((reloadPage?: boolean) => Promise<void>) | null = null;
let registration: ServiceWorkerRegistration | undefined;

function isInVoiceSession(): boolean {
  return sessionStorage.getItem('huddle:pwa:in-session') === '1';
}

export function setPwaInSession(active: boolean) {
  if (active) {
    sessionStorage.setItem('huddle:pwa:in-session', '1');
  } else {
    sessionStorage.removeItem('huddle:pwa:in-session');
  }
}

export function applyPendingUpdate() {
  if (!pwaUpdate.pending || !applyUpdate) return;
  pwaUpdate.pending = false;
  void applyUpdate(true);
}

export function initPwa() {
  if (!import.meta.env.PROD) return;

  applyUpdate = registerSW({
    immediate: true,
    onOfflineReady() {
      pwaUpdate.offlineReady = true;
    },
    onNeedRefresh() {
      pwaUpdate.pending = true;
      if (!isInVoiceSession()) {
        applyPendingUpdate();
      }
    },
    onRegistered(reg) {
      registration = reg;
      if (!reg) return;

      const check = () => reg.update().catch(() => {});
      check();
      window.setInterval(check, CHECK_INTERVAL_MS);
      document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible') check();
      });
    },
    onRegisterError() {},
  });
}

export async function checkForAppUpdate() {
  await registration?.update().catch(() => {});
}
