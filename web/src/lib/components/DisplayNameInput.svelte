<script lang="ts">
  import Icon from './Icon.svelte';
  import { mdiDiceMultiple } from '../icons';
  import { randomDisplayName } from '../random-name';
  import { settings } from '../stores/settings.svelte';
  import { MAX_DISPLAY_NAME_LENGTH } from '../validation';

  let {
    value = $bindable(''),
    placeholder = 'Alex',
    live = false,
    onLiveChange,
  }: {
    value?: string;
    placeholder?: string;
    live?: boolean;
    onLiveChange?: (name: string) => void;
  } = $props();

  function applyName(raw: string) {
    const clean = raw.slice(0, MAX_DISPLAY_NAME_LENGTH);
    value = clean;
    settings.setName(clean.trim());
    if (live) {
      const bounded = clean.trim();
      if (bounded) onLiveChange?.(bounded);
    }
  }

  function onInput(e: Event) {
    applyName((e.target as HTMLInputElement).value);
  }

  function randomize() {
    applyName(randomDisplayName());
  }
</script>

<div class="flex gap-2">
  <input
    type="text"
    {value}
    oninput={onInput}
    maxlength={MAX_DISPLAY_NAME_LENGTH}
    {placeholder}
    class="min-w-0 flex-1 rounded-lg border border-border bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
  />
  <button
    type="button"
    title="Random name"
    aria-label="Random name"
    onclick={randomize}
    class="shrink-0 rounded-lg border border-border bg-surface-2 px-2.5 text-muted transition-colors hover:border-accent hover:text-accent"
  >
    <Icon path={mdiDiceMultiple} size={18} />
  </button>
</div>
