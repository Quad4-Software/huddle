<script lang="ts">
  import Icon from './Icon.svelte';
  import {
    mdiAccountPlusOutline,
    mdiAccountGroup,
    mdiChevronLeft,
    mdiCheck,
    mdiRefresh,
    mdiAccountMultiple,
    mdiQrcode,
  } from '../icons';
  import { buildFullInviteUrl } from '../invite';
  import { session } from '../stores/session.svelte';
  import { connection } from '../stores/connection.svelte';
  import { qualityColor, qualityLabel, iceStateLabel } from '../connection-health';
  import { retryConnection } from '../session-controller';
  import ChatPanel from './ChatPanel.svelte';
  import MemberPanel from './MemberPanel.svelte';
  import VoiceBar from './VoiceBar.svelte';
  import ScreenGrid from './ScreenGrid.svelte';
  import MembersSheet from './MembersSheet.svelte';
  import InviteQrModal from './InviteQrModal.svelte';

  let { onSettings }: { onSettings: () => void } = $props();

  let copied = $state(false);
  let sidebarMinimized = $state(false);
  let membersOpen = $state(false);
  let qrOpen = $state(false);

  const inviteUrl = $derived(
    session.room
      ? buildFullInviteUrl(location.origin, session.room.id, session.invite, session.roomKey)
      : '',
  );

  const peerCount = $derived((session.room?.members.length ?? 1) - 1);

  const needsRetry = $derived(
    connection.status === 'offline' || (!session.connected && connection.status !== 'reconnecting'),
  );

  const connectionTitle = $derived.by(() => {
    const parts = [qualityLabel(session.connectionQuality)];
    if (session.ping !== null) parts.push(`${session.ping} ms`);
    if (session.jitter !== null) parts.push(`jitter ${session.jitter} ms`);
    const ice = iceStateLabel(session.iceState);
    if (ice) parts.push(ice);
    return parts.join(' · ');
  });

  async function copyInvite() {
    if (!inviteUrl) return;
    await navigator.clipboard.writeText(inviteUrl);
    copied = true;
    setTimeout(() => (copied = false), 2000);
  }
</script>

<div class="flex h-full flex-col">
  {#if needsRetry}
    <div
      class="flex items-center justify-between gap-3 border-b border-away/30 bg-away/10 px-4 py-2 text-sm"
    >
      <p class="min-w-0 text-muted">
        {#if connection.status === 'reconnecting'}
          Reconnecting{connection.detail ? `: ${connection.detail}` : '...'}
        {:else if !session.connected}
          Disconnected from the server
        {:else}
          Connection issue
        {/if}
      </p>
      <button
        type="button"
        onclick={() => retryConnection()}
        class="flex shrink-0 items-center gap-1.5 rounded-lg bg-accent px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-accent-hover"
      >
        <Icon path={mdiRefresh} size={16} />
        Retry
      </button>
    </div>
  {/if}

  <header class="flex items-center justify-between border-b border-border bg-surface-1 px-4 py-2.5">
    <div class="flex min-w-0 items-center gap-2.5">
      <div
        class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-accent/15 text-accent"
      >
        <Icon path={mdiAccountGroup} size={18} />
      </div>
      <div class="min-w-0 leading-tight">
        <h1 class="truncate text-sm font-semibold">{session.room?.name ?? 'Room'}</h1>
        <div class="flex flex-wrap items-center gap-x-2 gap-y-0.5 text-xs text-muted">
          <span>{session.sortedMembers.length} in room</span>
          <span class="text-border">·</span>
          <span class="inline-flex items-center gap-1.5" title={connectionTitle}>
            <span
              class="h-1.5 w-1.5 shrink-0 rounded-full {qualityColor(session.connectionQuality)}"
              aria-hidden="true"
            ></span>
            {#if !session.connected}
              disconnected
            {:else if peerCount === 0}
              waiting for others
            {:else if session.meshReady}
              voice connected
            {:else}
              connecting
            {/if}
          </span>
          {#if session.ping !== null}
            <span class="text-border">·</span>
            <span>{session.ping} ms</span>
          {/if}
          {#if session.jitter !== null}
            <span class="text-border">·</span>
            <span>±{session.jitter} ms</span>
          {/if}
          {#if session.sharing}
            <span class="text-border">·</span>
            <span class="text-accent">sharing</span>
          {/if}
        </div>
      </div>
    </div>

    <div class="flex shrink-0 items-center gap-2">
      <button
        type="button"
        onclick={() => (membersOpen = true)}
        class="flex items-center gap-1.5 rounded-lg border border-border bg-surface-2 px-2.5 py-1.5 text-sm font-medium transition-colors hover:bg-surface-3 md:hidden"
        aria-label="Show members"
      >
        <Icon path={mdiAccountMultiple} size={18} />
        <span>{session.sortedMembers.length}</span>
      </button>

      <button
        type="button"
        onclick={() => (qrOpen = true)}
        class="flex items-center gap-2 rounded-lg border border-border bg-surface-2 px-2.5 py-1.5 text-sm font-medium transition-colors hover:bg-surface-3"
        title="Show invite QR code"
        aria-label="Show invite QR code"
      >
        <Icon path={mdiQrcode} size={18} />
        <span class="hidden sm:inline">QR</span>
      </button>

      <button
        type="button"
        onclick={copyInvite}
        class="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {copied
          ? 'bg-success/20 text-success'
          : 'bg-accent text-white hover:bg-accent-hover'}"
      >
        <Icon path={copied ? mdiCheck : mdiAccountPlusOutline} size={18} />
        <span class="hidden sm:inline">{copied ? 'Copied' : 'Invite'}</span>
      </button>
    </div>
  </header>

  <div class="flex flex-1 overflow-hidden">
    <main class="flex flex-1 flex-col overflow-hidden">
      <ScreenGrid />
      {#if !session.focusMode}
        <ChatPanel />
      {/if}
    </main>
    {#if sidebarMinimized}
      <aside class="hidden w-10 shrink-0 border-l border-border bg-surface-1 md:flex md:flex-col">
        <button
          type="button"
          onclick={() => (sidebarMinimized = false)}
          class="flex flex-1 flex-col items-center gap-2 py-3 text-muted transition-colors hover:bg-surface-2 hover:text-foreground"
          title="Show members"
          aria-label="Show members"
        >
          <Icon path={mdiChevronLeft} size={18} />
          <span class="text-[10px] font-medium uppercase tracking-wide [writing-mode:vertical-rl]">
            Members
          </span>
        </button>
      </aside>
    {:else}
      <aside class="hidden w-60 shrink-0 border-l border-border bg-surface-1 md:block">
        <MemberPanel onMinimize={() => (sidebarMinimized = true)} />
      </aside>
    {/if}
  </div>

  <VoiceBar {onSettings} />
</div>

<MembersSheet open={membersOpen} onClose={() => (membersOpen = false)} />

<InviteQrModal
  open={qrOpen}
  url={inviteUrl}
  roomName={session.room?.name ?? 'Room'}
  onClose={() => (qrOpen = false)}
/>
