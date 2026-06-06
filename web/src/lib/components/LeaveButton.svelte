<script lang="ts">
  import Icon from './Icon.svelte';
  import { mdiPhoneHangup } from '../icons';
  import { leaveSession } from '../session-controller';

  const HOLD_MS = 700;
  const DOUBLE_CLICK_MS = 400;
  const HINT_VISIBLE_MS = 3000;

  let holding = $state(false);
  let progress = $state(0);
  let hintVisible = $state(false);
  let timer: ReturnType<typeof setInterval> | null = null;
  let hintTimer: ReturnType<typeof setTimeout> | null = null;
  let lastClickAt = 0;
  let skipNextClick = false;

  function clearHoldTimer() {
    if (timer) clearInterval(timer);
    timer = null;
  }

  function clearHintTimer() {
    if (hintTimer) clearTimeout(hintTimer);
    hintTimer = null;
  }

  function hideHint() {
    clearHintTimer();
    hintVisible = false;
  }

  function showHint() {
    clearHintTimer();
    hintVisible = true;
    hintTimer = setTimeout(hideHint, HINT_VISIBLE_MS);
  }

  function cancelHold() {
    holding = false;
    progress = 0;
    clearHoldTimer();
  }

  function startHold() {
    cancelHold();
    holding = true;
    const start = Date.now();
    timer = setInterval(() => {
      progress = Math.min(100, ((Date.now() - start) / HOLD_MS) * 100);
      if (progress >= 100) {
        cancelHold();
        hideHint();
        skipNextClick = true;
        leaveSession();
      }
    }, 16);
  }

  function handleClick() {
    if (skipNextClick) {
      skipNextClick = false;
      return;
    }

    const now = Date.now();
    if (now - lastClickAt <= DOUBLE_CLICK_MS) {
      lastClickAt = 0;
      cancelHold();
      hideHint();
      leaveSession();
      return;
    }

    lastClickAt = now;
    showHint();
  }
</script>

<div class="relative">
  {#if hintVisible}
    <div
      role="tooltip"
      class="pointer-events-none absolute bottom-[calc(100%+0.5rem)] left-1/2 z-10 w-max max-w-[min(16rem,calc(100vw-2rem))] -translate-x-1/2 rounded-lg border border-border bg-surface-2 px-3 py-2 text-center text-xs leading-snug text-muted shadow-lg"
    >
      Hold to leave
      <span class="text-border"> · </span>
      double-click to leave now
    </div>
  {/if}

  <button
    type="button"
    onclick={handleClick}
    onpointerdown={(e) => {
      e.currentTarget.setPointerCapture(e.pointerId);
      startHold();
    }}
    onpointerup={cancelHold}
    onpointerleave={cancelHold}
    onpointercancel={cancelHold}
    class="relative flex h-11 w-11 items-center justify-center overflow-hidden rounded-xl bg-danger text-white transition-colors hover:bg-danger/90 sm:h-12 sm:w-12"
    title="Hold to leave · double-click to leave now"
    aria-label="Leave room. Hold to confirm, or double-click to leave immediately."
  >
    {#if holding}
      <span
        class="pointer-events-none absolute inset-0 bg-white/25"
        style="clip-path: inset({100 - progress}% 0 0 0);"
        aria-hidden="true"
      ></span>
    {/if}
    <Icon path={mdiPhoneHangup} size={22} />
  </button>
</div>
