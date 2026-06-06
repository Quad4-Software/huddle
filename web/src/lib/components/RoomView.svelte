<script lang="ts">
  import Icon from './Icon.svelte';
  import { mdiAccountPlusOutline, mdiAccountGroup, mdiChevronLeft, mdiCheck } from '../icons';
  import { buildFullInviteUrl } from '../invite';
  import { session } from '../stores/session.svelte';
  import ChatPanel from './ChatPanel.svelte';
  import MemberPanel from './MemberPanel.svelte';
  import VoiceBar from './VoiceBar.svelte';
  import ScreenGrid from './ScreenGrid.svelte';

  let { onSettings }: { onSettings: () => void } = $props();

  let copied = $state(false);
  let sidebarMinimized = $state(false);

  const inviteUrl = $derived(
    session.room
      ? buildFullInviteUrl(location.origin, session.room.id, session.invite, session.roomKey)
      : '',
  );

  async function copyInvite() {
    if (!inviteUrl) return;
    await navigator.clipboard.writeText(inviteUrl);
    copied = true;
    setTimeout(() => (copied = false), 2000);
  }
</script>

<div class="flex h-full flex-col">
  <header class="flex items-center justify-between border-b border-border bg-surface-1 px-4 py-2.5">
    <div class="flex items-center gap-2.5">
      <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-accent/15 text-accent">
        <Icon path={mdiAccountGroup} size={18} />
      </div>
      <div class="leading-tight">
        <h1 class="text-sm font-semibold">{session.room?.name ?? 'Room'}</h1>
        <p class="text-xs text-muted">{session.sortedMembers.length} in room</p>
      </div>
    </div>

    <button
      onclick={copyInvite}
      class="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {copied
        ? 'bg-success/20 text-success'
        : 'bg-accent text-white hover:bg-accent-hover'}"
    >
      <Icon path={copied ? mdiCheck : mdiAccountPlusOutline} size={18} />
      {copied ? 'Copied' : 'Invite'}
    </button>
  </header>

  <div class="flex flex-1 overflow-hidden">
    <main class="flex flex-1 flex-col overflow-hidden">
      <ScreenGrid />
      <ChatPanel />
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
