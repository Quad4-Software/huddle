<script lang="ts">
  import Landing from './lib/components/Landing.svelte';
  import RoomView from './lib/components/RoomView.svelte';
  import SettingsPage from './lib/components/SettingsPage.svelte';
  import LoadingScreen from './lib/components/LoadingScreen.svelte';
  import { session } from './lib/stores/session.svelte';
  import { loading } from './lib/stores/loading.svelte';
  import { joinFromUrl } from './lib/session-controller';
  import { SITE_NAME, canonicalUrl, pageDescription, pageTitle, robotsDirective } from './lib/seo';
  import type { View } from './lib/types';

  function hasInviteUrl(): boolean {
    const roomId = location.pathname.match(/^\/r\/([^/]+)/)?.[1];
    const invite = new URLSearchParams(location.search).get('t');
    const key = location.hash.match(/key=([^&]+)/)?.[1];
    return !!(roomId && invite && key);
  }

  const inviteBoot = hasInviteUrl();
  if (inviteBoot) loading.start('joining');

  let view = $state<View>('landing');
  let booting = $state(inviteBoot);

  const seoView = $derived(
    view === 'settings' ? 'settings' : view === 'room' && session.connected ? 'room' : 'landing',
  );
  const seoRoomName = $derived(session.room?.name);
  const seoTitle = $derived(pageTitle(seoView, seoRoomName));
  const seoDescription = $derived(pageDescription(seoView, seoRoomName));
  const seoRobots = $derived(robotsDirective(location.pathname));
  const seoCanonical = $derived(canonicalUrl(location.pathname, location.search));

  $effect(() => {
    void (async () => {
      try {
        const joined = await joinFromUrl().catch(() => false);
        if (joined) {
          view = 'room';
        }
      } finally {
        booting = false;
      }
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

<svelte:head>
  <title>{seoTitle}</title>
  <meta name="description" content={seoDescription} />
  <meta name="robots" content={seoRobots} />
  <link rel="canonical" href={seoCanonical} />
  <meta property="og:type" content="website" />
  <meta property="og:site_name" content={SITE_NAME} />
  <meta property="og:locale" content="en_US" />
  <meta property="og:title" content={seoTitle} />
  <meta property="og:description" content={seoDescription} />
  <meta property="og:url" content={seoCanonical} />
  <meta name="twitter:card" content="summary" />
  <meta name="twitter:title" content={seoTitle} />
  <meta name="twitter:description" content={seoDescription} />
</svelte:head>

{#if booting || loading.active}
  <LoadingScreen />
{:else if view === 'settings'}
  <SettingsPage onBack={backFromSettings} />
{:else if view === 'room' && session.connected}
  <RoomView onSettings={openSettings} />
{:else}
  <Landing onSettings={openSettings} />
{/if}
