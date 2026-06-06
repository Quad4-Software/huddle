<script lang="ts">
  let {
    level = 0,
    threshold = 0,
    sensitivity = 40,
  }: {
    level?: number;
    threshold?: number;
    sensitivity?: number;
  } = $props();

  const levelPct = $derived(Math.round(level * 100));
  const thresholdPct = $derived(Math.round(threshold * 100));
  const active = $derived(level >= threshold);
  const thresholdLeft = $derived(`${Math.min(100, Math.max(0, threshold * 100))}%`);
  const fillWidth = $derived(`${Math.min(100, Math.max(0, level * 100))}%`);
</script>

<div class="rounded-lg border border-border bg-surface-2/60 p-3">
  <div class="mb-2 flex items-center justify-between gap-2 text-xs">
    <span class="font-medium text-muted">Microphone input</span>
    <span class="tabular-nums {active ? 'text-speaking' : 'text-muted'}">
      {levelPct}% · {active ? 'Would activate' : 'Below threshold'}
    </span>
  </div>
  <div
    class="relative h-3 overflow-hidden rounded-full bg-surface-0"
    role="meter"
    aria-valuemin={0}
    aria-valuemax={100}
    aria-valuenow={levelPct}
    aria-label="Microphone input level"
  >
    <div
      class="absolute inset-y-0 left-0 rounded-full transition-[width] duration-75 {active
        ? 'bg-speaking'
        : 'bg-accent/70'}"
      style="width: {fillWidth}"
    ></div>
    <div
      class="absolute inset-y-0 w-0.5 -translate-x-1/2 bg-highlight"
      style="left: {thresholdLeft}"
      title="Activation threshold at {thresholdPct}%"
    ></div>
  </div>
  <div class="mt-2 flex justify-between text-[11px] text-muted">
    <span>Background pickup</span>
    <span>Sensitivity {sensitivity}% · threshold {thresholdPct}%</span>
  </div>
</div>
