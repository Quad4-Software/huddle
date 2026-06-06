<script lang="ts">
  import Icon from './Icon.svelte';
  import DisplayNameInput from './DisplayNameInput.svelte';
  import { mdiAccountGroup, mdiDiceMultiple } from '../icons';
  import { createRoom } from '../session-controller';
  import { randomRoomName } from '../random-name';
  import { session } from '../stores/session.svelte';
  import { loading } from '../stores/loading.svelte';
  import { settings } from '../stores/settings.svelte';
  import { MAX_PASSWORD_LENGTH, MAX_ROOM_NAME_LENGTH } from '../validation';

  let { onSettings }: { onSettings: () => void } = $props();

  let name = $state(randomRoomName());
  let password = $state('');
  let error = $state('');

  $effect(() => {
    if (session.error) {
      error = session.error;
      session.error = '';
    }
  });

  async function handleCreate() {
    const nextDisplayName = settings.displayName.trim();
    const nextName = name.trim().slice(0, MAX_ROOM_NAME_LENGTH);
    const nextPassword = password.slice(0, MAX_PASSWORD_LENGTH);

    if (!nextDisplayName) {
      error = 'Enter your name';
      return;
    }
    if (!nextName) {
      error = 'Enter a room name';
      return;
    }
    error = '';
    try {
      await createRoom(nextName, nextPassword);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Could not create room';
    }
  }

  function randomizeRoomName() {
    name = randomRoomName();
  }
</script>

<div class="flex min-h-full items-center justify-center p-6">
  <div class="w-full max-w-md">
    <div class="mb-10 text-center">
      <div
        class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-accent/15 text-accent"
      >
        <Icon path={mdiAccountGroup} size={28} />
      </div>
      <h1 class="text-2xl font-semibold tracking-tight">Huddle</h1>
      <p class="mt-2 text-sm text-muted">Private voice, text, and screen sharing</p>
    </div>

    <div class="rounded-xl border border-border bg-surface-1 p-6">
      <label class="mb-4 block">
        <span class="mb-1.5 block text-xs font-medium text-muted">Your name</span>
        <DisplayNameInput bind:value={settings.displayName} placeholder="Alex" />
      </label>

      <label class="mb-4 block">
        <span class="mb-1.5 block text-xs font-medium text-muted">Room name</span>
        <div class="flex gap-2">
          <input
            type="text"
            bind:value={name}
            maxlength={MAX_ROOM_NAME_LENGTH}
            placeholder="Team sync"
            class="min-w-0 flex-1 rounded-lg border border-border bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
          />
          <button
            type="button"
            title="Random room name"
            aria-label="Random room name"
            onclick={randomizeRoomName}
            class="shrink-0 rounded-lg border border-border bg-surface-2 px-2.5 text-muted transition-colors hover:border-accent hover:text-accent"
          >
            <Icon path={mdiDiceMultiple} size={18} />
          </button>
        </div>
      </label>

      <label class="mb-6 block">
        <span class="mb-1.5 block text-xs font-medium text-muted">Password (optional)</span>
        <input
          type="password"
          bind:value={password}
          maxlength={MAX_PASSWORD_LENGTH}
          placeholder="Leave empty for invite only"
          class="w-full rounded-lg border border-border bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
        />
      </label>

      {#if error}
        <p class="mb-4 text-sm text-danger">{error}</p>
      {/if}

      <button
        onclick={handleCreate}
        disabled={loading.active}
        class="w-full rounded-lg bg-accent py-2.5 text-sm font-medium text-white transition-colors hover:bg-accent-hover disabled:opacity-50"
      >
        Create room
      </button>
    </div>

    <p class="mt-6 text-center text-xs text-muted">
      End to end encrypted, peer to peer, no accounts.
      <button onclick={onSettings} class="ml-1 text-highlight hover:underline">Settings</button>
    </p>
  </div>
</div>
