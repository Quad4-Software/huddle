<script lang="ts">
  import Icon from './Icon.svelte';
  import { mdiAccountPlusOutline, mdiAccountGroup } from '../icons';
  import { session } from '../stores/session.svelte';
  import ChatPanel from './ChatPanel.svelte';
  import MemberPanel from './MemberPanel.svelte';
  import VoiceBar from './VoiceBar.svelte';
  import ScreenGrid from './ScreenGrid.svelte';
  import InviteModal from './InviteModal.svelte';

  let { onSettings }: { onSettings: () => void } = $props();

  let showInvite = $state(false);
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
      onclick={() => (showInvite = true)}
      class="flex items-center gap-2 rounded-lg bg-accent px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-accent-hover"
    >
      <Icon path={mdiAccountPlusOutline} size={18} />
      Invite
    </button>
  </header>

  <div class="flex flex-1 overflow-hidden">
    <main class="flex flex-1 flex-col overflow-hidden">
      <ScreenGrid />
      <ChatPanel />
    </main>
    <aside class="hidden w-60 border-l border-border bg-surface-1 md:block">
      <MemberPanel />
    </aside>
  </div>

  <VoiceBar {onSettings} />
</div>

{#if showInvite}
  <InviteModal onClose={() => (showInvite = false)} />
{/if}
