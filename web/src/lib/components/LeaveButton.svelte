<script lang="ts">
  import Icon from './Icon.svelte';
  import { mdiPhoneHangup } from '../icons';
  import { leaveSession } from '../session-controller';

  const HOLD_MS = 700;

  let holding = $state(false);
  let progress = $state(0);
  let timer: ReturnType<typeof setInterval> | null = null;

  function clearTimer() {
    if (timer) clearInterval(timer);
    timer = null;
  }

  function cancelHold() {
    holding = false;
    progress = 0;
    clearTimer();
  }

  function startHold() {
    cancelHold();
    holding = true;
    const start = Date.now();
    timer = setInterval(() => {
      progress = Math.min(100, ((Date.now() - start) / HOLD_MS) * 100);
      if (progress >= 100) {
        cancelHold();
        leaveSession();
      }
    }, 16);
  }
</script>

<button
  type="button"
  onpointerdown={(e) => {
    e.currentTarget.setPointerCapture(e.pointerId);
    startHold();
  }}
  onpointerup={cancelHold}
  onpointerleave={cancelHold}
  onpointercancel={cancelHold}
  class="relative flex h-11 w-11 items-center justify-center overflow-hidden rounded-xl bg-danger text-white transition-colors hover:bg-danger/90 sm:h-12 sm:w-12"
  title="Hold to leave room"
  aria-label="Hold to leave room"
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
