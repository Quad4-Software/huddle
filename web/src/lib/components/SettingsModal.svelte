<script lang="ts">
  import Icon from './Icon.svelte';
  import DisplayNameInput from './DisplayNameInput.svelte';
  import KeybindRecorder from './KeybindRecorder.svelte';
  import MicActivationMeter from './MicActivationMeter.svelte';
  import { mdiClose, mdiRefresh, mdiMicrophone, mdiMicrophoneOff } from '../icons';
  import { settings, voiceActivationLevelThreshold } from '../stores/settings.svelte';
  import { session } from '../stores/session.svelte';
  import { audioLevels } from '../stores/audio-levels.svelte';
  import { localMicLevel } from '../stores/local-mic-level.svelte';
  import { listAudioDevices } from '../webrtc/audio';
  import { micPreview } from '../webrtc/mic-preview';
  import { settingsMicSampler } from '../webrtc/settings-mic-level';
  import { applyAudioSettings, changeName, refreshMic } from '../session-controller';
  import { APP_NAME, APP_VERSION } from '../version';

  let { onClose }: { onClose: () => void } = $props();

  let inputs = $state<MediaDeviceInfo[]>([]);
  let outputs = $state<MediaDeviceInfo[]>([]);
  let micListening = $state(false);
  let micPreviewError = $state('');
  let tab = $state<'audio' | 'room'>('audio');

  const micLevel = $derived(
    session.connected ? audioLevels.level(session.peerId) : localMicLevel.level,
  );
  const activationThreshold = $derived(voiceActivationLevelThreshold());

  $effect(() => {
    return () => {
      void micPreview.stop();
      void settingsMicSampler.stop();
    };
  });

  $effect(() => {
    const mode = settings.inputMode;
    const deviceId = settings.inputDeviceId;
    const volume = settings.inputVolume;
    const inRoom = session.connected;

    if (mode !== 'voiceActivation' || inRoom || micListening) {
      void settingsMicSampler.stop();
      return;
    }

    void settingsMicSampler.restart(deviceId, volume);
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
      await micPreview.start(settings.inputDeviceId, settings.inputVolume, settings.outputDeviceId);
      micListening = true;
    } catch {
      micPreviewError = 'Could not listen to microphone';
      micListening = false;
    }
  }

  function closeModal() {
    void micPreview.stop();
    void settingsMicSampler.stop();
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

  async function onNoiseSuppressionChange(e: Event) {
    settings.setNoiseSuppression((e.target as HTMLInputElement).checked);
    if (session.connected) await refreshMic().catch(() => {});
    await syncMicPreview();
  }

  async function resetSettings() {
    await micPreview.stop();
    await settingsMicSampler.stop();
    micListening = false;
    settings.reset();
    applyAudioSettings();
  }
</script>

<div
  class="fixed inset-0 z-50 flex items-end justify-center bg-black/60 p-0 sm:items-center sm:p-4"
  role="presentation"
  onclick={closeModal}
>
  <div
    class="flex max-h-[min(92dvh,820px)] w-full max-w-xl flex-col rounded-t-2xl border border-border bg-surface-1 shadow-2xl sm:max-h-[min(90vh,820px)] sm:max-w-2xl sm:rounded-xl"
    role="dialog"
    aria-labelledby="settings-title"
    tabindex="-1"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => e.key === 'Escape' && closeModal()}
    style="padding-bottom: max(0px, env(safe-area-inset-bottom));"
  >
    <div class="flex items-center justify-between border-b border-border px-4 py-4 sm:px-6">
      <h2 id="settings-title" class="text-base font-semibold">Settings</h2>
      <button
        type="button"
        onclick={closeModal}
        class="rounded-lg p-2 text-muted transition-colors hover:bg-surface-2 hover:text-text min-h-11 min-w-11 flex items-center justify-center"
        aria-label="Close settings"
      >
        <Icon path={mdiClose} size={18} />
      </button>
    </div>

    <div class="flex gap-1 border-b border-border px-4 sm:px-6">
      <button
        type="button"
        onclick={() => (tab = 'audio')}
        class="border-b-2 px-3 py-2.5 text-sm font-medium transition-colors {tab === 'audio'
          ? 'border-accent text-foreground'
          : 'border-transparent text-muted hover:text-foreground'}"
      >
        Audio
      </button>
      <button
        type="button"
        onclick={() => (tab = 'room')}
        class="border-b-2 px-3 py-2.5 text-sm font-medium transition-colors {tab === 'room'
          ? 'border-accent text-foreground'
          : 'border-transparent text-muted hover:text-foreground'}"
      >
        Room
      </button>
    </div>

    <div class="flex-1 space-y-6 overflow-y-auto overscroll-contain px-4 py-5 sm:px-6">
      {#if tab === 'audio'}
        <section>
          <h3 class="mb-3 text-xs font-semibold uppercase tracking-wide text-muted">Devices</h3>
          <div class="space-y-4">
            <label class="block">
              <span class="mb-1.5 block text-xs font-medium text-muted">Microphone</span>
              <select
                value={settings.inputDeviceId}
                onchange={onInputChange}
                class="min-h-11 w-full rounded-lg border border-border bg-surface-2 px-3 py-2.5 text-sm outline-none focus:border-accent"
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
                class="min-h-11 w-full rounded-lg border border-border bg-surface-2 px-3 py-2.5 text-sm outline-none focus:border-accent"
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
            <label class="flex items-center gap-3">
              <input
                type="checkbox"
                checked={settings.noiseSuppression}
                onchange={onNoiseSuppressionChange}
                class="h-4 w-4 rounded border-border accent-accent"
              />
              <span class="text-sm">Noise suppression</span>
            </label>
            <p class="text-xs leading-relaxed text-muted">
              Reduces background noise. Turn off if your voice sounds clipped or unnatural.
            </p>

            <label class="block">
              <span class="mb-1.5 block text-xs font-medium text-muted">Input mode</span>
              <select
                value={settings.inputMode}
                onchange={onInputModeChange}
                class="min-h-11 w-full rounded-lg border border-border bg-surface-2 px-3 py-2.5 text-sm outline-none focus:border-accent"
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
                  class="h-2 w-full accent-accent"
                />
              </label>
              <MicActivationMeter
                level={micLevel}
                threshold={activationThreshold}
                sensitivity={settings.voiceActivationThreshold}
              />
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
                class="h-2 w-full accent-accent"
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
                class="h-2 w-full accent-accent"
              />
            </label>
          </div>
        </section>
      {:else}
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
          <h3 class="mb-1 text-xs font-semibold uppercase tracking-wide text-muted">Keybinds</h3>
          <p class="mb-3 text-xs leading-relaxed text-muted">
            Record a key for each action. Deafen also mutes your mic and restores the previous mute
            state when you undeafen.
          </p>
          <div class="space-y-3">
            <KeybindRecorder
              label="Push to talk"
              hint="Hold while speaking when push-to-talk mode is enabled"
              value={settings.pushToTalkKey}
              defaultValue="Space"
              required
              onchange={(code) => settings.setPushToTalkKey(code)}
            />
            <KeybindRecorder
              label="Toggle mute"
              hint="Optional shortcut to mute and unmute your microphone"
              value={settings.toggleMuteKey}
              onchange={(code) => settings.setToggleMuteKey(code)}
            />
            <KeybindRecorder
              label="Toggle deafen"
              hint="Deafens incoming audio and mutes your mic until undeafened"
              value={settings.toggleDeafenKey}
              onchange={(code) => settings.setToggleDeafenKey(code)}
            />
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
      {/if}
    </div>

    <div class="border-t border-border px-4 py-4 sm:px-6">
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
