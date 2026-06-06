export const RECONNECT_MAX_ATTEMPTS = 8;
export const RECONNECT_BASE_MS = 1000;
export const RECONNECT_MAX_DELAY_MS = 15000;

export function reconnectDelayMs(attempt: number): number {
  return Math.min(RECONNECT_MAX_DELAY_MS, RECONNECT_BASE_MS * 2 ** Math.max(0, attempt - 1));
}

export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
