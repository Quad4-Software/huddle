<script lang="ts">
  import Icon from './Icon.svelte';
  import { mdiAccountGroup } from '../icons';
  import { createRoom } from '../session-controller';
  import { session } from '../stores/session.svelte';
  import { settings } from '../stores/settings.svelte';

  let { onSettings }: { onSettings: () => void } = $props();

  let displayName = $state(settings.displayName);
  let name = $state('');
  let password = $state('');
  let loading = $state(false);
  let error = $state('');

  $effect(() => {
    if (session.error) {
      error = session.error;
      session.error = '';
    }
  });

  async function handleCreate() {
    if (!displayName.trim()) {
      error = 'Enter your name';
      return;
    }
    if (!name.trim()) {
      error = 'Enter a room name';
      return;
    }
    loading = true;
    error = '';
    settings.setName(displayName.trim());
    try {
      await createRoom(name.trim(), password);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Could not create room';
      loading = false;
    }
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
        <input
          type="text"
          bind:value={displayName}
          placeholder="Alex"
          class="w-full rounded-lg border border-border bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
        />
      </label>

      <label class="mb-4 block">
        <span class="mb-1.5 block text-xs font-medium text-muted">Room name</span>
        <input
          type="text"
          bind:value={name}
          placeholder="Team sync"
          class="w-full rounded-lg border border-border bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
        />
      </label>

      <label class="mb-6 block">
        <span class="mb-1.5 block text-xs font-medium text-muted">Password (optional)</span>
        <input
          type="password"
          bind:value={password}
          placeholder="Leave empty for invite only"
          class="w-full rounded-lg border border-border bg-surface-2 px-3 py-2 text-sm outline-none focus:border-accent"
        />
      </label>

      {#if error}
        <p class="mb-4 text-sm text-danger">{error}</p>
      {/if}

      <button
        onclick={handleCreate}
        disabled={loading}
        class="w-full rounded-lg bg-accent py-2.5 text-sm font-medium text-white transition-colors hover:bg-accent-hover disabled:opacity-50"
      >
        {loading ? 'Creating' : 'Create room'}
      </button>
    </div>

    <p class="mt-6 text-center text-xs text-muted">
      End to end encrypted, peer to peer, no accounts.
      <button onclick={onSettings} class="ml-1 text-highlight hover:underline">Settings</button>
    </p>
  </div>
</div>
