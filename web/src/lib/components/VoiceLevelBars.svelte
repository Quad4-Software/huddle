<script lang="ts">
  let { level = 0, bars = 4 }: { level?: number; bars?: number } = $props();

  const scales = [0.55, 0.85, 1, 0.7, 0.6];
  const heights = $derived(
    Array.from({ length: bars }, (_, i) =>
      Math.max(0.1, Math.min(1, level * (scales[i] ?? 0.8) * 1.2)),
    ),
  );
</script>

<div class="voice-level-bars" aria-hidden="true">
  {#each heights as height, index (index)}
    <span style="transform: scaleY({height})"></span>
  {/each}
</div>
