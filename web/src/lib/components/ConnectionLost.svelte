<script lang="ts">
  import Icon from './Icon.svelte';
  import { mdiCloudOffOutline, mdiAccountGroup, mdiRefresh } from '../icons';
  import { connection } from '../stores/connection.svelte';
  import { session } from '../stores/session.svelte';
  import { leaveSession, retryConnection } from '../session-controller';

  const reconnecting = $derived(connection.status === 'reconnecting');
  const title = $derived(reconnecting ? 'Reconnecting' : 'Connection lost');
  const subtitle = $derived(
    connection.detail ||
      (reconnecting
        ? 'Trying to restore your session with the server.'
        : 'Unable to reach the server. Your room is still saved locally.'),
  );
  const progress = $derived(
    reconnecting ? Math.min(100, (connection.attempt / connection.maxAttempts) * 100) : 0,
  );
</script>

<div class="flex h-full items-center justify-center p-6">
  <div class="w-full max-w-sm text-center">
    <div
      class="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-danger/15 text-danger"
    >
      <Icon path={mdiCloudOffOutline} size={32} />
    </div>
    <h1 class="text-xl font-semibold tracking-tight">{title}</h1>
    <p class="mt-2 text-sm text-muted">{subtitle}</p>
    {#if session.room?.name}
      <p class="mt-1 text-xs text-muted">{session.room.name}</p>
    {/if}

    {#if reconnecting}
      <div class="mt-8 h-1.5 overflow-hidden rounded-full bg-surface-3">
        <div class="h-full rounded-full bg-accent" style="width: {progress}%"></div>
      </div>
      <p class="mt-2 text-xs tabular-nums text-muted">
        Attempt {connection.attempt} of {connection.maxAttempts}
      </p>
    {/if}

    <div class="mt-8 flex flex-col gap-2">
      {#if !reconnecting}
        <button
          type="button"
          onclick={() => retryConnection()}
          class="flex w-full items-center justify-center gap-2 rounded-lg bg-accent py-2.5 text-sm font-medium text-white transition-colors hover:bg-accent-hover"
        >
          <Icon path={mdiRefresh} size={18} />
          Retry connection
        </button>
      {/if}
      <button
        type="button"
        onclick={() => leaveSession()}
        class="w-full rounded-lg border border-border bg-surface-2 py-2.5 text-sm font-medium text-muted transition-colors hover:border-danger hover:text-danger"
      >
        Leave room
      </button>
    </div>

    <div class="mx-auto mt-10 flex items-center justify-center gap-2 text-xs text-muted">
      <Icon path={mdiAccountGroup} size={14} />
      <span>Huddle</span>
    </div>
  </div>
</div>
