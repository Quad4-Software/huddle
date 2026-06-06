<script lang="ts">
  import Icon from './Icon.svelte';
  import MemberPanel from './MemberPanel.svelte';
  import { mdiClose } from '../icons';

  let { open, onClose }: { open: boolean; onClose: () => void } = $props();
</script>

{#if open}
  <div class="fixed inset-0 z-40 bg-black/50 md:hidden" role="presentation" onclick={onClose}></div>
  <div
    class="fixed inset-x-0 bottom-0 z-50 flex max-h-[min(85dvh,640px)] flex-col rounded-t-2xl border border-border bg-surface-1 shadow-2xl md:hidden"
    role="dialog"
    aria-label="Members"
    style="padding-bottom: max(0px, env(safe-area-inset-bottom));"
  >
    <div class="flex items-center justify-between border-b border-border px-4 py-3">
      <h2 class="text-sm font-semibold">Members</h2>
      <button
        type="button"
        onclick={onClose}
        class="flex h-10 w-10 items-center justify-center rounded-lg text-muted transition-colors hover:bg-surface-2 hover:text-foreground"
        aria-label="Close members"
      >
        <Icon path={mdiClose} size={18} />
      </button>
    </div>
    <div class="min-h-0 flex-1 overflow-hidden">
      <MemberPanel />
    </div>
  </div>
{/if}
