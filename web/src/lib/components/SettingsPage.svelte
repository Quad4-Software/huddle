<script lang="ts">
  import Icon from './Icon.svelte';
  import DisplayNameInput from './DisplayNameInput.svelte';
  import { mdiArrowLeft } from '../icons';
  import { settings } from '../stores/settings.svelte';
  import { listAudioDevices } from '../webrtc/audio';
  import { refreshMic, changeName } from '../session-controller';

  let { onBack }: { onBack: () => void } = $props();

  let inputs = $state<MediaDeviceInfo[]>([]);
  let outputs = $state<MediaDeviceInfo[]>([]);

  $effect(() => {
    (async () => {
      await navigator.mediaDevices
        .getUserMedia({ audio: true })
        .then((s) => {
          s.getTracks().forEach((t) => t.stop());
        })
        .catch(() => {});
      const devs = await listAudioDevices();
      inputs = devs.inputs;
      outputs = devs.outputs;
    })();
  });

  async function onInputChange(e: Event) {
    const id = (e.target as HTMLSelectElement).value;
    settings.setInput(id);
    if (settings.displayName) await refreshMic().catch(() => {});
  }

  function onOutputChange(e: Event) {
    settings.setOutput((e.target as HTMLSelectElement).value);
  }
</script>

<div class="mx-auto max-w-lg p-6">
  <button
    onclick={onBack}
    class="mb-6 flex items-center gap-2 text-sm text-muted transition-colors hover:text-text"
  >
    <Icon path={mdiArrowLeft} size={18} />
    Back
  </button>

  <h1 class="mb-6 text-xl font-semibold">Settings</h1>

  <div class="space-y-5">
    <label class="block">
      <span class="mb-1.5 block text-xs font-medium text-muted">Display name</span>
      <DisplayNameInput
        bind:value={settings.displayName}
        placeholder="Your name"
        live
        onLiveChange={changeName}
      />
      <p class="mt-1.5 text-xs text-muted">Changes update live for everyone in the room.</p>
    </label>

    <label class="block">
      <span class="mb-1.5 block text-xs font-medium text-muted">Microphone</span>
      <select
        value={settings.inputDeviceId}
        onchange={onInputChange}
        class="w-full rounded-lg border border-border bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
      >
        <option value="">System default</option>
        {#each inputs as dev (dev.deviceId)}
          <option value={dev.deviceId}>{dev.label || 'Microphone'}</option>
        {/each}
      </select>
    </label>

    <label class="block">
      <span class="mb-1.5 block text-xs font-medium text-muted">Speaker</span>
      <select
        value={settings.outputDeviceId}
        onchange={onOutputChange}
        class="w-full rounded-lg border border-border bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
      >
        <option value="">System default</option>
        {#each outputs as dev (dev.deviceId)}
          <option value={dev.deviceId}>{dev.label || 'Speaker'}</option>
        {/each}
      </select>
      <p class="mt-1.5 text-xs text-muted">Applies to everyone you hear.</p>
    </label>
  </div>
</div>
