<script lang="ts">
  import Icon from './Icon.svelte';
  import Avatar from './Avatar.svelte';
  import VoiceLevelBars from './VoiceLevelBars.svelte';
  import {
    mdiMicrophoneOff,
    mdiHeadphonesOff,
    mdiAccountRemove,
    mdiCrown,
    mdiChevronRight,
    mdiChevronDown,
    mdiChevronUp,
    mdiMonitorShare,
    mdiEyeOutline,
  } from '../icons';
  import { session } from '../stores/session.svelte';
  import { audioLevels } from '../stores/audio-levels.svelte';
  import { memberStatus } from '../members';
  import { kickMember, moderateMember } from '../session-controller';

  function isOnline(peerId: string) {
    if (peerId === session.peerId) return session.connected;
    return session.peerOnline[peerId] === true;
  }

  function isConnecting(peerId: string) {
    if (peerId === session.peerId) return false;
    return session.peerOnline[peerId] !== true;
  }

  function statusColor(status: string) {
    if (status === 'offline') return 'bg-offline';
    if (status === 'connecting') return 'bg-away';
    if (status === 'speaking') return 'bg-speaking';
    if (status === 'muted' || status === 'deafened') return 'bg-away';
    return 'bg-online';
  }

  function statusText(status: string) {
    if (status === 'offline') return 'text-offline';
    if (status === 'speaking') return 'text-speaking';
    if (status === 'online') return 'text-online';
    return 'text-away';
  }

  function isRoomHost(peerId: string) {
    return session.room?.hostId === peerId;
  }

  let { onMinimize }: { onMinimize?: () => void } = $props();

  let shareMenuOpen = $state(false);

  function viewShare(peerId: string) {
    session.showScreenPanel(peerId);
    shareMenuOpen = false;
  }
</script>

<div class="flex h-full flex-col">
  <div class="flex items-center justify-between border-b border-border px-4 py-3">
    <span class="text-xs font-semibold uppercase tracking-wide text-muted">
      Members {session.sortedMembers.length}
    </span>
    {#if onMinimize}
      <button
        type="button"
        onclick={onMinimize}
        class="rounded p-1 text-muted transition-colors hover:bg-surface-3 hover:text-foreground"
        title="Hide members"
        aria-label="Hide members"
      >
        <Icon path={mdiChevronRight} size={18} />
      </button>
    {/if}
  </div>

  {#if session.allActiveShares.length > 0}
    <div class="border-b border-border">
      <button
        type="button"
        onclick={() => (shareMenuOpen = !shareMenuOpen)}
        class="flex w-full items-center gap-2 px-4 py-2.5 text-left transition-colors hover:bg-surface-2"
        aria-expanded={shareMenuOpen}
      >
        <Icon path={mdiMonitorShare} size={16} class="shrink-0 text-accent" />
        <span class="min-w-0 flex-1 truncate text-xs">
          {#if session.allActiveShares.length === 1}
            {session.memberName(session.allActiveShares[0].peerId)} is sharing
          {:else}
            {session.allActiveShares.length} screens live
          {/if}
        </span>
        <Icon
          path={shareMenuOpen ? mdiChevronUp : mdiChevronDown}
          size={16}
          class="shrink-0 text-muted"
        />
      </button>
      {#if shareMenuOpen}
        <div class="space-y-0.5 border-t border-border px-2 py-1.5">
          {#each session.allActiveShares as share (share.peerId)}
            <button
              type="button"
              onclick={() => viewShare(share.peerId)}
              class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-xs transition-colors hover:bg-surface-2"
            >
              <Icon path={mdiEyeOutline} size={14} class="shrink-0 text-muted" />
              <span class="truncate">View {session.memberName(share.peerId)}</span>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  <div class="flex-1 overflow-y-auto p-2">
    {#each session.sortedMembers as member (member.id)}
      {@const online = isOnline(member.id)}
      {@const connecting = isConnecting(member.id)}
      {@const status = memberStatus(member, online, connecting)}
      {@const level =
        online && !member.muted && !member.deafened ? audioLevels.level(member.id) : 0}
      {@const host = isRoomHost(member.id)}
      <div
        class="group flex items-center gap-2.5 rounded-lg px-2 py-2 transition-colors {level > 0.04
          ? 'bg-speaking/10'
          : 'hover:bg-surface-2/60'}"
      >
        <div class="relative shrink-0">
          <Avatar name={member.name} size={36} {level} />
          <span
            class="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-surface-1 {statusColor(
              status,
            )}"
            title={status}
          ></span>
        </div>
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm font-medium">
            {member.name}
            {#if member.id === session.peerId}
              <span class="font-normal text-muted">you</span>
            {/if}
            {#if host}
              <Icon path={mdiCrown} size={14} class="ml-1 inline text-accent" title="Host" />
            {/if}
          </p>
          <p class="text-xs capitalize {statusText(status)}">{status}</p>
        </div>
        <div class="flex items-center gap-1.5 text-muted">
          {#if member.deafened}
            <Icon path={mdiHeadphonesOff} size={16} class="text-danger" />
          {:else if member.muted}
            <Icon path={mdiMicrophoneOff} size={16} class="text-danger" />
          {:else if online}
            <VoiceLevelBars {level} />
          {/if}
          {#if session.isHost && member.id !== session.peerId}
            <button
              onclick={() => moderateMember(member.id, !member.muted, member.deafened)}
              class="rounded-md p-1 text-muted opacity-0 transition-all hover:bg-surface-3 hover:text-text group-hover:opacity-100"
              title={member.muted ? 'Unmute member' : 'Mute member'}
              aria-label={member.muted ? 'Unmute member' : 'Mute member'}
            >
              <Icon path={mdiMicrophoneOff} size={16} />
            </button>
            <button
              onclick={() =>
                moderateMember(member.id, member.deafened ? false : true, !member.deafened)}
              class="rounded-md p-1 text-muted opacity-0 transition-all hover:bg-surface-3 hover:text-text group-hover:opacity-100"
              title={member.deafened ? 'Undeafen member' : 'Deafen member'}
              aria-label={member.deafened ? 'Undeafen member' : 'Deafen member'}
            >
              <Icon path={mdiHeadphonesOff} size={16} />
            </button>
            <button
              onclick={() => kickMember(member.id)}
              class="rounded-md p-1 text-muted opacity-0 transition-all hover:bg-danger/10 hover:text-danger group-hover:opacity-100"
              title="Remove from room"
            >
              <Icon path={mdiAccountRemove} size={16} />
            </button>
          {/if}
        </div>
      </div>
    {/each}
  </div>
</div>
