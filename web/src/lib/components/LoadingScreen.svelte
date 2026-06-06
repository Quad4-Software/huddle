<script lang="ts">
  import Icon from './Icon.svelte';
  import {
    mdiShieldLock,
    mdiLogin,
    mdiPlus,
    mdiAccessPointNetwork,
    mdiAccountGroup,
  } from '../icons';
  import { loading } from '../stores/loading.svelte';
  import Quad4Credit from './Quad4Credit.svelte';

  const phaseConfig = {
    connecting: {
      icon: mdiAccessPointNetwork,
      title: 'Connecting',
      subtitle: 'Establishing secure connection',
    },
    pow: {
      icon: mdiShieldLock,
      title: 'Solving PoW',
      subtitle: 'Anti-Bot Challenge',
    },
    creating: {
      icon: mdiPlus,
      title: 'Creating room',
      subtitle: 'Setting up your space',
    },
    joining: {
      icon: mdiLogin,
      title: 'Joining room',
      subtitle: 'Entering voice channel',
    },
  } as const;

  const config = $derived(phaseConfig[loading.phase]);
  const subtitle = $derived(loading.detail || config.subtitle);
</script>

<div class="flex h-full items-center justify-center p-6">
  <div class="w-full max-w-sm text-center">
    <div
      class="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-accent/15 text-accent"
    >
      <Icon path={config.icon} size={32} />
    </div>
    <h1 class="text-xl font-semibold tracking-tight">{config.title}</h1>
    <p class="mt-2 text-sm text-muted">{subtitle}</p>

    <div class="mt-8 h-1.5 overflow-hidden rounded-full bg-surface-3">
      <div class="h-full rounded-full bg-accent" style="width: {loading.progress}%"></div>
    </div>

    <p class="mt-2 text-xs tabular-nums text-muted">{Math.round(loading.progress)}%</p>

    <div class="mx-auto mt-10 flex flex-col items-center justify-center gap-2">
      <div class="flex items-center justify-center gap-2 text-xs text-muted">
        <Icon path={mdiAccountGroup} size={14} />
        <span>Huddle</span>
      </div>
      <Quad4Credit />
    </div>
  </div>
</div>
