import { RECONNECT_MAX_ATTEMPTS } from '../reconnect';

export type ConnectionStatus = 'online' | 'reconnecting' | 'offline';

class ConnectionStore {
  status = $state<ConnectionStatus>('online');
  attempt = $state(0);
  detail = $state('');
  readonly maxAttempts = RECONNECT_MAX_ATTEMPTS;

  startReconnect() {
    this.status = 'reconnecting';
  }

  setAttempt(attempt: number) {
    this.attempt = attempt;
  }

  setDetail(detail: string) {
    this.detail = detail;
  }

  setOnline() {
    this.status = 'online';
    this.attempt = 0;
    this.detail = '';
  }

  setOffline() {
    this.status = 'offline';
    this.detail = '';
  }

  reset() {
    this.status = 'online';
    this.attempt = 0;
    this.detail = '';
  }
}

export const connection = new ConnectionStore();
