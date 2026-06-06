<script lang="ts">
  import {
    captureKeyFromEvent,
    claimKeyRecorder,
    formatKeyCode,
    releaseKeyRecorder,
  } from '../keybind';

  let {
    label,
    hint = '',
    value = $bindable(''),
    defaultValue = '',
    required = false,
    onchange,
  }: {
    label: string;
    hint?: string;
    value?: string;
    defaultValue?: string;
    required?: boolean;
    onchange?: (code: string) => void;
  } = $props();

  function updateValue(code: string) {
    value = code;
    onchange?.(code);
  }

  let recording = $state(false);

  function stopRecording() {
    recording = false;
    releaseKeyRecorder(stopRecording);
  }

  function startRecording() {
    claimKeyRecorder(stopRecording);
    recording = true;
  }

  function resetToDefault() {
    updateValue(defaultValue);
    stopRecording();
  }

  function clearBinding() {
    if (required) return;
    updateValue('');
    stopRecording();
  }

  function onKeyCapture(e: KeyboardEvent) {
    if (!recording) return;
    e.preventDefault();
    e.stopPropagation();
    const result = captureKeyFromEvent(e);
    if (result === null) return;
    if (result === 'cancel') {
      stopRecording();
      return;
    }
    if (result === 'clear') {
      clearBinding();
      return;
    }
    updateValue(result.code);
    stopRecording();
  }

  const display = $derived(recording ? 'Press a key...' : formatKeyCode(value));
</script>

<svelte:window onkeydown={onKeyCapture} />

<div class="rounded-lg border border-border bg-surface-2/50 p-3">
  <div class="mb-2">
    <p class="text-sm font-medium">{label}</p>
    {#if hint}
      <p class="mt-0.5 text-xs text-muted">{hint}</p>
    {/if}
  </div>
  <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
    <div
      class="min-h-11 flex flex-1 items-center rounded-lg border px-3 py-2 text-sm tabular-nums {recording
        ? 'border-accent bg-accent/10 text-accent'
        : 'border-border bg-surface-1 text-text'}"
    >
      {display}
    </div>
    <div class="flex gap-2">
      <button
        type="button"
        onclick={recording ? stopRecording : startRecording}
        class="min-h-11 flex-1 rounded-lg px-3 py-2 text-sm font-medium transition-colors sm:flex-none {recording
          ? 'bg-accent text-white'
          : 'border border-border bg-surface-1 text-text hover:border-accent'}"
      >
        {recording ? 'Cancel' : 'Record'}
      </button>
      {#if !required}
        <button
          type="button"
          onclick={clearBinding}
          disabled={!value}
          class="min-h-11 rounded-lg border border-border bg-surface-1 px-3 py-2 text-sm text-muted transition-colors hover:border-accent hover:text-text disabled:opacity-40"
        >
          Clear
        </button>
      {/if}
      <button
        type="button"
        onclick={resetToDefault}
        class="min-h-11 rounded-lg border border-border bg-surface-1 px-3 py-2 text-sm text-muted transition-colors hover:border-accent hover:text-text"
      >
        Reset
      </button>
    </div>
  </div>
</div>
