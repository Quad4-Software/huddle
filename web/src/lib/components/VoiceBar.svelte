<script lang="ts">
  import Icon from './Icon.svelte';
  import {
    mdiMicrophone,
    mdiMicrophoneOff,
    mdiHeadphones,
    mdiHeadphonesOff,
    mdiMonitorShare,
    mdiMonitorOff,
    mdiCog,
    mdiGestureTapHold,
  } from '../icons';
  import { session } from '../stores/session.svelte';
  import { settings } from '../stores/settings.svelte';
  import { audioLevels } from '../stores/audio-levels.svelte';
  import { formatKeyCode } from '../keybind';
  import {
    toggleMute,
    toggleDeafen,
    startShare,
    stopShare,
    setPttActive,
    retryMic,
  } from '../session-controller';
  import VoiceLevelBars from './VoiceLevelBars.svelte';
  import LeaveButton from './LeaveButton.svelte';

  let { onSettings }: { onSettings: () => void } = $props();

  const localLevel = $derived(
    session.muted || session.deafened ? 0 : audioLevels.level(session.peerId),
  );
  const isPtt = $derived(settings.inputMode === 'pushToTalk');
  const inputModeLabel = $derived(isPtt ? 'PTT' : 'Voice');

  function actionTooltip(label: string, key: string): string {
    return key ? `${label} (${formatKeyCode(key)})` : label;
  }

  const muteTooltip = $derived(
    session.micError
      ? session.micError
      : actionTooltip(session.muted ? 'Unmute' : 'Mute', settings.toggleMuteKey),
  );

  const deafenTooltip = $derived(
    actionTooltip(
      session.deafened ? 'Undeafen' : 'Deafen (also mutes mic)',
      settings.toggleDeafenKey,
    ),
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

  function startPttHold(e: PointerEvent) {
    e.currentTarget.setPointerCapture(e.pointerId);
    setPttActive(true);
  }

  function endPttHold() {
    setPttActive(false);
  }
</script>

{#if session.micError}
  <div
    class="flex items-center justify-between gap-3 border-t border-danger/30 bg-danger/10 px-4 py-2 text-sm"
  >
    <p class="min-w-0 text-danger">{session.micError}</p>
    <div class="flex shrink-0 items-center gap-2">
      <button
        type="button"
        onclick={() => retryMic()}
        class="rounded-lg bg-surface-2 px-3 py-1.5 text-xs font-medium transition-colors hover:bg-surface-3"
      >
        Retry mic
      </button>
      <button
        type="button"
        onclick={onSettings}
        class="rounded-lg bg-surface-2 px-3 py-1.5 text-xs font-medium transition-colors hover:bg-surface-3"
      >
        Pick device
      </button>
    </div>
  </div>
{/if}

<footer
  class="border-t border-border bg-surface-1 px-3 pt-3"
  style="padding-bottom: max(0.75rem, env(safe-area-inset-bottom));"
>
  <div class="mx-auto flex max-w-lg items-center justify-center gap-2 sm:max-w-xl sm:gap-3">
    <div
      class="flex items-center gap-0.5 rounded-2xl bg-surface-2 p-1 sm:gap-1"
      role="toolbar"
      aria-label="Voice controls"
    >
      {#if isPtt}
        <button
          type="button"
          onpointerdown={startPttHold}
          onpointerup={endPttHold}
          onpointerleave={endPttHold}
          onpointercancel={endPttHold}
          class="flex h-11 min-w-11 items-center justify-center gap-1 rounded-xl px-2 text-xs font-semibold uppercase tracking-wide text-accent transition-colors hover:bg-surface-3 sm:h-12"
          title="Hold to talk"
          aria-label="Hold to talk"
        >
          <Icon path={mdiGestureTapHold} size={18} />
          <span class="hidden sm:inline">Talk</span>
        </button>
      {:else if session.connected && !session.muted && !session.deafened}
        <div class="hidden px-1 sm:block" aria-hidden="true">
          <VoiceLevelBars level={localLevel} bars={3} />
        </div>
      {/if}

      <button
        type="button"
        onclick={toggleMute}
        class="relative flex h-11 w-11 items-center justify-center rounded-xl transition-colors sm:h-12 sm:w-12 {session.micError
          ? 'bg-danger/25 text-danger ring-2 ring-danger/40'
          : session.muted
            ? 'bg-danger/25 text-danger'
            : 'text-foreground hover:bg-surface-3'}"
        title={muteTooltip}
        aria-label={session.muted ? 'Unmute microphone' : 'Mute microphone'}
        aria-pressed={session.muted}
      >
        <Icon
          path={session.muted || session.micError ? mdiMicrophoneOff : mdiMicrophone}
          size={22}
        />
        <span
          class="pointer-events-none absolute -right-0.5 -top-0.5 rounded bg-surface-3 px-1 text-[9px] font-semibold uppercase leading-tight text-muted"
        >
          {inputModeLabel}
        </span>
      </button>

      <button
        type="button"
        onclick={toggleDeafen}
        class="flex h-11 w-11 items-center justify-center rounded-xl transition-colors sm:h-12 sm:w-12 {session.deafened
          ? 'bg-danger/25 text-danger'
          : 'text-foreground hover:bg-surface-3'}"
        title={deafenTooltip}
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

    <LeaveButton />
  </div>
</footer>
