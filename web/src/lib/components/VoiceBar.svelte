<script lang="ts">
  import Icon from './Icon.svelte';
  import {
    mdiMicrophone,
    mdiMicrophoneOff,
    mdiHeadphones,
    mdiHeadphonesOff,
    mdiMonitorShare,
    mdiMonitorOff,
    mdiPhoneHangup,
    mdiCog,
  } from '../icons';
  import { session } from '../stores/session.svelte';
  import {
    toggleMute,
    toggleDeafen,
    startShare,
    stopShare,
    leaveSession,
  } from '../session-controller';

  let { onSettings }: { onSettings: () => void } = $props();

  const peerCount = $derived((session.room?.members.length ?? 1) - 1);

  function pingColor(ping: number) {
    if (ping < 80) return 'text-online';
    if (ping < 200) return 'text-away';
    return 'text-danger';
  }

  async function toggleShare() {
    if (session.sharing) {
      stopShare();
    } else {
      try {
        await startShare();
      } catch {}
    }
  }
</script>

<div
  class="grid grid-cols-[1fr_auto_1fr] items-center border-t border-border bg-surface-1 px-4 py-2"
>
  <div class="flex items-center gap-3 text-sm">
    <div class="flex items-center gap-2">
      <div
        class="h-2 w-2 rounded-full {session.connected
          ? session.meshReady || peerCount === 0
            ? 'bg-online'
            : 'bg-away'
          : 'bg-danger'}"
      ></div>
      <span class="text-muted">
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
    </div>

    {#if session.ping !== null}
      <span class="text-xs {pingColor(session.ping)}">{session.ping} ms</span>
    {/if}

    {#if session.sharing}
      <span class="rounded bg-accent/20 px-2 py-0.5 text-xs text-accent">sharing screen</span>
    {/if}
  </div>

  <p class="text-center text-xs text-muted">Ephemeral and encrypted room</p>

  <div class="flex items-center justify-end gap-1">
    <button
      onclick={toggleMute}
      class="rounded-lg p-2.5 transition-colors {session.muted
        ? 'bg-danger/20 text-danger'
        : 'hover:bg-surface-3'}"
      title={session.muted ? 'Unmute' : 'Mute'}
    >
      <Icon path={session.muted ? mdiMicrophoneOff : mdiMicrophone} size={22} />
    </button>

    <button
      onclick={toggleDeafen}
      class="rounded-lg p-2.5 transition-colors {session.deafened
        ? 'bg-danger/20 text-danger'
        : 'hover:bg-surface-3'}"
      title={session.deafened ? 'Undeafen' : 'Deafen'}
    >
      <Icon path={session.deafened ? mdiHeadphonesOff : mdiHeadphones} size={22} />
    </button>

    <button
      onclick={toggleShare}
      class="rounded-lg p-2.5 transition-colors {session.sharing
        ? 'bg-accent/20 text-accent'
        : 'hover:bg-surface-3'}"
      title={session.sharing ? 'Stop sharing' : 'Share screen'}
    >
      <Icon path={session.sharing ? mdiMonitorOff : mdiMonitorShare} size={22} />
    </button>

    <button
      onclick={onSettings}
      class="rounded-lg p-2.5 transition-colors hover:bg-surface-3"
      title="Settings"
    >
      <Icon path={mdiCog} size={22} />
    </button>

    <button
      onclick={leaveSession}
      class="ml-2 rounded-lg bg-danger/20 p-2.5 text-danger transition-colors hover:bg-danger/30"
      title="Leave"
    >
      <Icon path={mdiPhoneHangup} size={22} />
    </button>
  </div>
</div>
