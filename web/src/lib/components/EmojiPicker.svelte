<script lang="ts">
  import { onMount } from 'svelte';
  import 'emoji-picker-element';
  import emojiDataUrl from 'emoji-picker-element-data/en/emojibase/data.json?url';

  let { onPick, onClose }: { onPick: (emoji: string) => void; onClose: () => void } = $props();

  let hostEl = $state<HTMLDivElement | null>(null);

  function attach(node: HTMLElement) {
    const handler = (event: Event) => {
      const detail = (event as CustomEvent<{ unicode?: string }>).detail;
      if (detail?.unicode) onPick(detail.unicode);
    };
    node.addEventListener('emoji-click', handler);
    return {
      destroy() {
        node.removeEventListener('emoji-click', handler);
      },
    };
  }

  onMount(() => {
    const onDocClick = (e: MouseEvent) => {
      if (hostEl && !hostEl.contains(e.target as Node)) onClose();
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    const id = window.setTimeout(() => {
      document.addEventListener('click', onDocClick);
    }, 0);
    document.addEventListener('keydown', onKey);
    return () => {
      window.clearTimeout(id);
      document.removeEventListener('click', onDocClick);
      document.removeEventListener('keydown', onKey);
    };
  });
</script>

<div
  bind:this={hostEl}
  class="emoji-host overflow-hidden rounded-xl border border-border shadow-xl"
  role="dialog"
  aria-label="Emoji picker"
  tabindex="-1"
>
  <emoji-picker data-source={emojiDataUrl} use:attach></emoji-picker>
</div>

<style>
  .emoji-host :global(emoji-picker) {
    --background: var(--color-surface-1);
    --border-color: var(--color-border);
    --input-border-color: var(--color-border);
    --input-background: var(--color-surface-2);
    --category-emoji-size: 1.1rem;
    --num-columns: 8;
    height: 320px;
    width: 320px;
  }
</style>
