<script lang="ts">
  import { pwaUpdate, applyPendingUpdate } from '../pwa-update.svelte';

  let { inSession = false }: { inSession?: boolean } = $props();
</script>

{#if pwaUpdate.pending && inSession}
  <div
    class="fixed bottom-20 left-1/2 z-40 w-[min(92vw,28rem)] -translate-x-1/2 rounded-xl border border-border bg-surface-1 px-4 py-3 shadow-xl"
    role="status"
  >
    <p class="text-sm font-medium">Update available</p>
    <p class="mt-1 text-xs text-muted">
      A new version is ready. It will apply automatically when you leave the room, or you can
      refresh now.
    </p>
    <div class="mt-3 flex justify-end gap-2">
      <button
        type="button"
        onclick={() => applyPendingUpdate()}
        class="rounded-lg bg-accent px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-accent-hover"
      >
        Refresh now
      </button>
    </div>
  </div>
{/if}
