<script lang="ts">
  import Icon from './Icon.svelte';
  import Avatar from './Avatar.svelte';
  import { mdiMicrophoneOff, mdiHeadphonesOff, mdiAccountRemove, mdiCrown } from '../icons';
  import { session } from '../stores/session.svelte';
  import { memberStatus } from '../members';
  import { kickMember } from '../session-controller';

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
</script>

<div class="flex h-full flex-col">
  <div class="border-b border-border px-4 py-3">
    <span class="text-xs font-semibold uppercase tracking-wide text-muted">
      Members {session.sortedMembers.length}
    </span>
  </div>
  <div class="flex-1 overflow-y-auto p-2">
    {#each session.sortedMembers as member (member.id)}
      {@const online = isOnline(member.id)}
      {@const connecting = isConnecting(member.id)}
      {@const status = memberStatus(member, online, connecting)}
      {@const active = member.speaking && online && !member.muted}
      {@const host = isRoomHost(member.id)}
      <div
        class="group flex items-center gap-2.5 rounded-lg px-2 py-2 transition-colors {active
          ? 'bg-speaking/10'
          : 'hover:bg-surface-2/60'}"
      >
        <div class="relative shrink-0">
          <Avatar name={member.name} size={36} ring={active} />
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
          {/if}
          {#if active}
            <div class="voice-bars" aria-hidden="true">
              <span></span>
              <span></span>
              <span></span>
            </div>
          {/if}
          {#if session.isHost && member.id !== session.peerId}
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
