<script lang="ts">
  import { avatarGradient, avatarInitials } from '../avatar';

  let { name, size = 36, ring = false }: { name: string; size?: number; ring?: boolean } = $props();

  const initials = $derived(avatarInitials(name));
  const gradient = $derived(avatarGradient(name));
  const fontSize = $derived(Math.round(size * 0.4));
</script>

<div
  class="flex shrink-0 items-center justify-center rounded-full font-semibold text-white {ring
    ? 'speak-ring'
    : ''}"
  style="width: {size}px; height: {size}px; background: {gradient}; font-size: {fontSize}px;"
  aria-hidden="true"
>
  {initials}
</div>
