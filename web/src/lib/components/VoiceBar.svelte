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
  import { audioLevels } from '../stores/audio-levels.svelte';
  import {
    toggleMute,
    toggleDeafen,
    startShare,
    stopShare,
    leaveSession,
  } from '../session-controller';
  import VoiceLevelBars from './VoiceLevelBars.svelte';

  let { onSettings }: { onSettings: () => void } = $props();

  const localLevel = $derived(
    session.muted || session.deafened ? 0 : audioLevels.level(session.peerId),
  );

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

<footer
  class="border-t border-border bg-surface-1 px-3 pt-3"
  style="padding-bottom: max(0.75rem, env(safe-area-inset-bottom));"
>
  <div class="mx-auto flex max-w-md items-center justify-center gap-2 sm:max-w-lg sm:gap-3">
    <div
      class="flex items-center gap-0.5 rounded-2xl bg-surface-2 p-1 sm:gap-1"
      role="toolbar"
      aria-label="Voice controls"
    >
      {#if session.connected && !session.muted && !session.deafened}
        <div class="hidden px-1 sm:block" aria-hidden="true">
          <VoiceLevelBars level={localLevel} bars={3} />
        </div>
      {/if}

      <button
        type="button"
        onclick={toggleMute}
        class="flex h-11 w-11 items-center justify-center rounded-xl transition-colors sm:h-12 sm:w-12 {session.muted
          ? 'bg-danger/25 text-danger'
          : 'text-foreground hover:bg-surface-3'}"
        title={session.muted ? 'Unmute' : 'Mute'}
        aria-label={session.muted ? 'Unmute microphone' : 'Mute microphone'}
        aria-pressed={session.muted}
      >
        <Icon path={session.muted ? mdiMicrophoneOff : mdiMicrophone} size={22} />
      </button>

      <button
        type="button"
        onclick={toggleDeafen}
        class="flex h-11 w-11 items-center justify-center rounded-xl transition-colors sm:h-12 sm:w-12 {session.deafened
          ? 'bg-danger/25 text-danger'
          : 'text-foreground hover:bg-surface-3'}"
        title={session.deafened ? 'Undeafen' : 'Deafen'}
        aria-label={session.deafened ? 'Undeafen' : 'Deafen'}
        aria-pressed={session.deafened}
      >
        <Icon path={session.deafened ? mdiHeadphonesOff : mdiHeadphones} size={22} />
      </button>

      <button
        type="button"
        onclick={toggleShare}
        class="flex h-11 w-11 items-center justify-center rounded-xl transition-colors sm:h-12 sm:w-12 {session.sharing
          ? 'bg-accent/25 text-accent'
          : 'text-foreground hover:bg-surface-3'}"
        title={session.sharing ? 'Stop sharing' : 'Share screen'}
        aria-label={session.sharing ? 'Stop sharing screen' : 'Share screen'}
        aria-pressed={session.sharing}
      >
        <Icon path={session.sharing ? mdiMonitorOff : mdiMonitorShare} size={22} />
      </button>
    </div>

    <div class="h-8 w-px shrink-0 bg-border" aria-hidden="true"></div>

    <button
      type="button"
      onclick={onSettings}
      class="flex h-11 w-11 items-center justify-center rounded-xl text-foreground transition-colors hover:bg-surface-2 sm:h-12 sm:w-12"
      title="Settings"
      aria-label="Settings"
    >
      <Icon path={mdiCog} size={22} />
    </button>

    <button
      type="button"
      onclick={leaveSession}
      class="flex h-11 w-11 items-center justify-center rounded-xl bg-danger text-white transition-colors hover:bg-danger/90 sm:h-12 sm:w-12"
      title="Leave room"
      aria-label="Leave room"
    >
      <Icon path={mdiPhoneHangup} size={22} />
    </button>
  </div>
</footer>
