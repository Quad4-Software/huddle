<script lang="ts">
  import Icon from './Icon.svelte';
  import Avatar from './Avatar.svelte';
  import {
    mdiMonitorShare,
    mdiClose,
    mdiPlay,
    mdiPause,
    mdiEyeOutline,
    mdiFullscreen,
    mdiPictureInPictureBottomRight,
    mdiVolumeHigh,
    mdiVolumeOff,
  } from '../icons';
  import { session } from '../stores/session.svelte';
  import { settings } from '../stores/settings.svelte';
  import { setSharePaused, toggleShareAudioMuted } from '../session-controller';
  import { setOutputDevice } from '../webrtc/audio';
  import { bindScreenVideo, hasScreenAudio } from '../screen-video';

  let mainVideo: HTMLVideoElement | undefined = $state();

  function audioAction(el: HTMLAudioElement, stream: MediaStream) {
    el.srcObject = stream;
    el.play().catch(() => {});
    if (settings.outputDeviceId) {
      setOutputDevice(el, settings.outputDeviceId).catch(() => {});
    }
    return {
      update(s: MediaStream) {
        el.srcObject = s;
        el.play().catch(() => {});
      },
      destroy() {
        el.srcObject = null;
      },
    };
  }

  const allShares = $derived.by(() => {
    const shares = [...session.screenShares];
    if (session.localScreen && session.sharing) {
      shares.unshift({ peerId: session.peerId, stream: session.localScreen });
    }
    return shares;
  });

  const focused = $derived(
    allShares.find((s) => s.peerId === session.focusedShare) ?? allShares[0] ?? null,
  );

  const focusedWatchers = $derived(
    focused ? (session.watchers[focused.peerId] ?? []).filter((p) => p !== focused.peerId) : [],
  );

  const focusedPaused = $derived(
    focused ? focused.peerId !== session.peerId && isPaused(focused.peerId) : false,
  );

  const focusedHasAudio = $derived(focused ? hasScreenAudio(focused.stream) : false);

  $effect(() => {
    if (!mainVideo || !focused) return;
    if (focusedPaused) {
      mainVideo.pause();
    } else {
      void mainVideo.play().catch(() => {});
    }
  });

  function isPaused(peerId: string) {
    return session.pausedShares[peerId] === true;
  }

  function selectShare(peerId: string) {
    session.focusedShare = peerId;
  }

  function clearFocus() {
    session.focusedShare = null;
  }

  function togglePause(peerId: string) {
    setSharePaused(peerId, !isPaused(peerId));
  }

  async function enterFullscreen() {
    if (!mainVideo) return;
    if (document.fullscreenElement) {
      await document.exitFullscreen();
      return;
    }
    await mainVideo.requestFullscreen();
  }

  async function enterPictureInPicture() {
    if (!mainVideo || !document.pictureInPictureEnabled) return;
    if (document.pictureInPictureElement === mainVideo) {
      await document.exitPictureInPicture();
      return;
    }
    await mainVideo.requestPictureInPicture();
  }
</script>

<section class="shrink-0 border-b border-border bg-surface-0">
  <div class="flex items-center justify-between px-4 py-2">
    <div class="flex items-center gap-2 text-sm font-medium">
      <Icon path={mdiMonitorShare} size={18} class="text-accent" />
      Screen
    </div>
    {#if allShares.length > 0}
      <span class="text-xs text-speaking">{allShares.length} live</span>
    {:else}
      <span class="text-xs text-muted">No active shares</span>
    {/if}
  </div>

  {#if allShares.length === 0}
    <div class="px-4 pb-3 text-xs leading-relaxed text-muted">
      Share your screen from the bar below and it appears here for everyone.
    </div>
  {:else if focused}
    {@const mine = focused.peerId === session.peerId}
    {@const paused = focusedPaused}
    <div class="bg-black/40 px-3 pb-3">
      <div
        class="relative mx-auto w-full max-w-full overflow-hidden rounded-lg border border-border bg-black"
      >
        <video
          bind:this={mainVideo}
          use:bindScreenVideo={focused.stream}
          class="block max-h-80 w-full object-contain {paused ? 'invisible' : ''}"
          autoplay
          playsinline
          muted
        ></video>

        {#if paused}
          <div
            class="absolute inset-0 flex max-h-80 flex-col items-center justify-center gap-3 bg-surface-0 text-muted"
          >
            <Icon path={mdiPause} size={32} />
            <span class="text-sm">Paused to save resources</span>
            <button
              onclick={() => togglePause(focused.peerId)}
              class="flex items-center gap-1.5 rounded-lg bg-accent px-3 py-1.5 text-sm text-white hover:bg-accent-hover"
            >
              <Icon path={mdiPlay} size={16} />
              Resume
            </button>
          </div>
        {/if}

        <div class="absolute left-3 top-3 flex items-center gap-2">
          <span class="rounded bg-speaking/90 px-2 py-0.5 text-xs font-medium text-surface-0">
            LIVE
          </span>
          <span class="rounded bg-black/60 px-2 py-0.5 text-xs">
            {session.memberName(focused.peerId)}
          </span>
        </div>

        <div class="absolute right-3 top-3 flex items-center gap-1.5">
          {#if !paused}
            <button
              onclick={enterPictureInPicture}
              class="rounded bg-black/60 p-1.5 text-muted hover:text-text"
              title="Picture in picture"
            >
              <Icon path={mdiPictureInPictureBottomRight} size={16} />
            </button>
            <button
              onclick={enterFullscreen}
              class="rounded bg-black/60 p-1.5 text-muted hover:text-text"
              title="Fullscreen"
            >
              <Icon path={mdiFullscreen} size={16} />
            </button>
          {/if}
          {#if focusedHasAudio && !mine}
            <button
              onclick={() => toggleShareAudioMuted(focused.peerId)}
              class="rounded bg-black/60 p-1.5 text-muted hover:text-text"
              title={session.isShareAudioMuted(focused.peerId)
                ? 'Unmute share audio'
                : 'Mute share audio'}
            >
              <Icon
                path={session.isShareAudioMuted(focused.peerId) ? mdiVolumeOff : mdiVolumeHigh}
                size={16}
              />
            </button>
          {/if}
          {#if !mine}
            <button
              onclick={() => togglePause(focused.peerId)}
              class="rounded bg-black/60 p-1.5 text-muted hover:text-text"
              title={paused ? 'Resume' : 'Pause to save resources'}
            >
              <Icon path={paused ? mdiPlay : mdiPause} size={16} />
            </button>
          {/if}
          {#if session.focusedShare}
            <button
              onclick={clearFocus}
              class="rounded bg-black/60 p-1.5 text-muted hover:text-text"
              title="Back to first share"
            >
              <Icon path={mdiClose} size={16} />
            </button>
          {/if}
        </div>

        {#if focusedWatchers.length > 0}
          <div
            class="absolute bottom-3 right-3 flex items-center gap-1.5 rounded-full bg-black/60 px-2 py-1"
          >
            <Icon path={mdiEyeOutline} size={14} class="text-muted" />
            <div class="flex -space-x-2">
              {#each focusedWatchers.slice(0, 5) as watcher (watcher)}
                <div class="rounded-full ring-2 ring-black/60" title={session.memberName(watcher)}>
                  <Avatar name={session.memberName(watcher)} size={20} />
                </div>
              {/each}
            </div>
            {#if focusedWatchers.length > 5}
              <span class="text-xs text-muted">+{focusedWatchers.length - 5}</span>
            {/if}
          </div>
        {/if}
      </div>

      {#if allShares.length > 1}
        <div class="mt-2 flex gap-2 overflow-x-auto">
          {#each allShares as share (share.peerId)}
            <button
              onclick={() => selectShare(share.peerId)}
              class="relative shrink-0 overflow-hidden rounded-md border-2 transition-colors {focused.peerId ===
              share.peerId
                ? 'border-accent'
                : 'border-transparent'}"
            >
              <video
                use:bindScreenVideo={share.stream}
                class="h-16 w-28 object-cover {isPaused(share.peerId) &&
                share.peerId !== session.peerId
                  ? 'invisible'
                  : ''}"
                autoplay
                playsinline
                muted
              ></video>
              {#if isPaused(share.peerId) && share.peerId !== session.peerId}
                <div
                  class="absolute inset-0 flex items-center justify-center bg-surface-2 text-muted"
                >
                  <Icon path={mdiPause} size={18} />
                </div>
              {/if}
              <span class="absolute bottom-0 left-0 right-0 truncate bg-black/60 px-1 text-[10px]">
                {session.memberName(share.peerId)}
              </span>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</section>

{#each Object.entries(session.remoteVoiceStreams) as [peerId, stream] (peerId)}
  {#if stream.getAudioTracks().length > 0}
    <audio use:audioAction={stream} class="hidden" autoplay></audio>
  {/if}
{/each}

{#if focused && focusedHasAudio && !focusedPaused && !session.isShareAudioMuted(focused.peerId)}
  <audio use:audioAction={focused.stream} class="hidden" autoplay></audio>
{/if}
