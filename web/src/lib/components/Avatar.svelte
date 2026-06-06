<script lang="ts">
  import { avatarGradient, avatarInitials } from '../avatar';

  let {
    name,
    size = 36,
    level = 0,
  }: {
    name: string;
    size?: number;
    level?: number;
  } = $props();

  const initials = $derived(avatarInitials(name));
  const gradient = $derived(avatarGradient(name));
  const fontSize = $derived(Math.round(size * 0.4));
  const active = $derived(level > 0.04);
  const glowSize = $derived(2 + level * 7);
  const glowAlpha = $derived(Math.round((0.25 + level * 0.55) * 100));
  const ringStyle = $derived(
    active
      ? `box-shadow: 0 0 0 ${glowSize}px color-mix(in srgb, var(--color-speaking) ${glowAlpha}%, transparent)`
      : '',
  );
</script>

<div
  class="flex shrink-0 items-center justify-center rounded-full font-semibold text-white transition-[box-shadow] duration-75"
  style="width: {size}px; height: {size}px; background: {gradient}; font-size: {fontSize}px; {ringStyle}"
  aria-hidden="true"
>
  {initials}
</div>
