<script lang="ts">
  import Icon from './Icon.svelte';
  import DisplayNameInput from './DisplayNameInput.svelte';
  import { mdiClose, mdiRefresh, mdiMicrophone, mdiMicrophoneOff } from '../icons';
  import { settings } from '../stores/settings.svelte';
  import { listAudioDevices } from '../webrtc/audio';
  import { micPreview } from '../webrtc/mic-preview';
  import { applyAudioSettings, changeName, refreshMic } from '../session-controller';
  import { formatKeyCode } from '../keybind';
  import { APP_NAME, APP_VERSION } from '../version';

  let { onClose }: { onClose: () => void } = $props();

  let inputs = $state<MediaDeviceInfo[]>([]);
  let outputs = $state<MediaDeviceInfo[]>([]);
  let capturingKey = $state(false);
  let micListening = $state(false);
  let micPreviewError = $state('');

  $effect(() => {
    return () => {
      void micPreview.stop();
    };
  });

  async function syncMicPreview() {
    if (!micListening) return;
    try {
      await micPreview.restart(
        settings.inputDeviceId,
        settings.inputVolume,
        settings.outputDeviceId,
      );
      micPreviewError = '';
    } catch {
      micPreviewError = 'Could not listen to microphone';
      micListening = false;
      await micPreview.stop();
    }
  }

  async function toggleMicListen() {
    micPreviewError = '';
    if (micListening) {
      await micPreview.stop();
      micListening = false;
      return;
    }
    try {
      await micPreview.start(
        settings.inputDeviceId,
        settings.inputVolume,
        settings.outputDeviceId,
      );
      micListening = true;
    } catch {
      micPreviewError = 'Could not listen to microphone';
      micListening = false;
    }
  }

  function closeModal() {
    void micPreview.stop();
    micListening = false;
    onClose();
  }

  $effect(() => {
    void (async () => {
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
    settings.setInput((e.target as HTMLSelectElement).value);
    if (settings.displayName) await refreshMic().catch(() => {});
    await syncMicPreview();
  }

  async function onOutputChange(e: Event) {
    settings.setOutput((e.target as HTMLSelectElement).value);
    await syncMicPreview();
  }

  function onInputModeChange(e: Event) {
    const value = (e.target as HTMLSelectElement).value;
    settings.setInputMode(value === 'pushToTalk' ? 'pushToTalk' : 'voiceActivation');
    applyAudioSettings();
  }

  function onKeyCapture(e: KeyboardEvent) {
    if (!capturingKey) return;
    e.preventDefault();
    e.stopPropagation();
    if (e.code === 'Escape') {
      capturingKey = false;
      return;
    }
    settings.setPushToTalkKey(e.code);
    capturingKey = false;
    applyAudioSettings();
  }

  async function resetSettings() {
    await micPreview.stop();
    micListening = false;
    settings.reset();
    applyAudioSettings();
  }
</script>

<svelte:window onkeydown={onKeyCapture} />

<div
  class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
  role="presentation"
  onclick={closeModal}
>
  <div
    class="flex max-h-[min(90vh,720px)] w-full max-w-lg flex-col rounded-xl border border-border bg-surface-1 shadow-2xl"
    role="dialog"
    aria-labelledby="settings-title"
    tabindex="-1"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => e.key === 'Escape' && closeModal()}
  >
    <div class="flex items-center justify-between border-b border-border px-5 py-4">
      <h2 id="settings-title" class="text-base font-semibold">Settings</h2>
      <button
        type="button"
        onclick={closeModal}
        class="rounded-lg p-1 text-muted transition-colors hover:bg-surface-2 hover:text-text"
        aria-label="Close settings"
      >
        <Icon path={mdiClose} size={18} />
      </button>
    </div>

    <div class="flex-1 space-y-6 overflow-y-auto px-5 py-5">
      <section>
        <h3 class="mb-3 text-xs font-semibold uppercase tracking-wide text-muted">Profile</h3>
        <label class="block">
          <span class="mb-1.5 block text-xs font-medium text-muted">Display name</span>
          <DisplayNameInput
            bind:value={settings.displayName}
            placeholder="Your name"
            live
            onLiveChange={changeName}
          />
        </label>
      </section>

      <section>
        <h3 class="mb-3 text-xs font-semibold uppercase tracking-wide text-muted">Devices</h3>
        <div class="space-y-4">
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
            <button
              type="button"
              onclick={toggleMicListen}
              class="mt-2 flex w-full items-center justify-center gap-2 rounded-lg border px-3 py-2 text-sm font-medium transition-colors {micListening
                ? 'border-accent bg-accent/15 text-accent'
                : 'border-border bg-surface-2 text-muted hover:border-accent hover:text-text'}"
            >
              <Icon path={micListening ? mdiMicrophoneOff : mdiMicrophone} size={18} />
              {micListening ? 'Stop listening' : 'Listen to microphone'}
            </button>
            {#if micPreviewError}
              <p class="mt-1.5 text-xs text-danger">{micPreviewError}</p>
            {/if}
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
          </label>
        </div>
      </section>

      <section>
        <h3 class="mb-3 text-xs font-semibold uppercase tracking-wide text-muted">Voice</h3>
        <div class="space-y-4">
          <label class="block">
            <span class="mb-1.5 block text-xs font-medium text-muted">Input mode</span>
            <select
              value={settings.inputMode}
              onchange={onInputModeChange}
              class="w-full rounded-lg border border-border bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
            >
              <option value="voiceActivation">Voice activation</option>
              <option value="pushToTalk">Push to talk</option>
            </select>
          </label>

          {#if settings.inputMode === 'voiceActivation'}
            <label class="block">
              <span class="mb-1.5 flex justify-between text-xs font-medium text-muted">
                <span>Activation sensitivity</span>
                <span>{settings.voiceActivationThreshold}%</span>
              </span>
              <input
                type="range"
                min="1"
                max="100"
                value={settings.voiceActivationThreshold}
                oninput={(e) => {
                  settings.setVoiceActivationThreshold(
                    Number((e.target as HTMLInputElement).value),
                  );
                }}
                class="w-full accent-accent"
              />
            </label>
          {:else}
            <label class="block">
              <span class="mb-1.5 block text-xs font-medium text-muted">Push to talk key</span>
              <button
                type="button"
                onclick={() => (capturingKey = true)}
                class="w-full rounded-lg border border-border bg-surface-2 px-3 py-2 text-left text-sm transition-colors hover:border-accent {capturingKey
                  ? 'border-accent text-accent'
                  : ''}"
              >
                {capturingKey ? 'Press a key...' : formatKeyCode(settings.pushToTalkKey)}
              </button>
            </label>
          {/if}

          <label class="block">
            <span class="mb-1.5 flex justify-between text-xs font-medium text-muted">
              <span>Input volume</span>
              <span>{settings.inputVolume}%</span>
            </span>
            <input
              type="range"
              min="0"
              max="200"
              value={settings.inputVolume}
              oninput={(e) => {
                settings.setInputVolume(Number((e.target as HTMLInputElement).value));
                applyAudioSettings();
                void syncMicPreview();
              }}
              class="w-full accent-accent"
            />
          </label>

          <label class="block">
            <span class="mb-1.5 flex justify-between text-xs font-medium text-muted">
              <span>Output volume</span>
              <span>{settings.outputVolume}%</span>
            </span>
            <input
              type="range"
              min="0"
              max="200"
              value={settings.outputVolume}
              oninput={(e) => {
                settings.setOutputVolume(Number((e.target as HTMLInputElement).value));
                void syncMicPreview();
              }}
              class="w-full accent-accent"
            />
          </label>
        </div>
      </section>

      <section>
        <h3 class="mb-3 text-xs font-semibold uppercase tracking-wide text-muted">About</h3>
        <div class="rounded-lg border border-border bg-surface-2 px-4 py-3 text-sm">
          <p class="font-medium">{APP_NAME}</p>
          <p class="mt-1 text-xs text-muted">Version {APP_VERSION}</p>
          <p class="mt-2 text-xs leading-relaxed text-muted">
            E2EE and P2P voice, text, and screen sharing for small groups.
          </p>
        </div>
      </section>
    </div>

    <div class="border-t border-border px-5 py-4">
      <button
        type="button"
        onclick={resetSettings}
        class="flex w-full items-center justify-center gap-2 rounded-lg border border-border bg-surface-2 py-2.5 text-sm font-medium text-muted transition-colors hover:border-accent hover:text-text"
      >
        <Icon path={mdiRefresh} size={18} />
        Reset settings to default
      </button>
    </div>
  </div>
</div>
