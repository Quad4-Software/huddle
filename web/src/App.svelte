<script lang="ts">
  import Landing from './lib/components/Landing.svelte';
  import RoomView from './lib/components/RoomView.svelte';
  import SettingsPage from './lib/components/SettingsPage.svelte';
  import { session } from './lib/stores/session.svelte';
  import { joinFromUrl } from './lib/session-controller';
  import type { View } from './lib/types';

  let view = $state<View>('landing');
  let booting = $state(true);

  $effect(() => {
    (async () => {
      const joined = await joinFromUrl().catch(() => false);
      if (joined) {
        view = 'room';
      }
      booting = false;
    })();
  });

  $effect(() => {
    if (session.connected && session.room) {
      view = 'room';
    }
  });

  function openSettings() {
    view = 'settings';
  }

  function backFromSettings() {
    view = session.connected ? 'room' : 'landing';
  }
</script>

{#if booting}
  <div class="flex h-full items-center justify-center text-sm text-muted">Loading...</div>
{:else if view === 'settings'}
  <SettingsPage onBack={backFromSettings} />
{:else if view === 'room' && session.connected}
  <RoomView onSettings={openSettings} />
{:else}
  <Landing onSettings={openSettings} />
{/if}
