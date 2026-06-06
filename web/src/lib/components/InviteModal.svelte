<script lang="ts">
  import Icon from './Icon.svelte';
  import { mdiContentCopy, mdiCheck, mdiLinkVariant, mdiClose } from '../icons';
  import { buildInviteUrl } from '../session-controller';
  import { session } from '../stores/session.svelte';

  let { onClose }: { onClose: () => void } = $props();

  let copied = $state(false);

  const url = $derived(
    session.room
      ? `${location.origin}${buildInviteUrl(session.room.id, session.invite, session.roomKey)}`
      : '',
  );

  async function copy() {
    await navigator.clipboard.writeText(url);
    copied = true;
    setTimeout(() => (copied = false), 2000);
  }
</script>

<div
  class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
  role="presentation"
  onclick={onClose}
>
  <div
    class="w-full max-w-md rounded-xl border border-border bg-surface-1 p-5 shadow-2xl"
    role="dialog"
    tabindex="-1"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => e.key === 'Escape' && onClose()}
  >
    <div class="mb-4 flex items-start justify-between">
      <div class="flex items-center gap-2.5">
        <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-accent/15 text-accent">
          <Icon path={mdiLinkVariant} size={20} />
        </div>
        <div>
          <h2 class="text-base font-semibold">Invite to {session.room?.name ?? 'room'}</h2>
          <p class="text-xs text-muted">The decryption key never leaves this link.</p>
        </div>
      </div>
      <button onclick={onClose} class="rounded-lg p-1 text-muted hover:text-text">
        <Icon path={mdiClose} size={18} />
      </button>
    </div>

    <div class="flex gap-2">
      <input
        readonly
        value={url}
        class="flex-1 rounded-lg border border-border bg-surface-2 px-3 py-2 text-xs outline-none"
      />
      <button
        onclick={copy}
        class="flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-medium transition-colors {copied
          ? 'bg-success/20 text-success'
          : 'bg-accent text-white hover:bg-accent-hover'}"
      >
        <Icon path={copied ? mdiCheck : mdiContentCopy} size={18} />
        {copied ? 'Copied' : 'Copy'}
      </button>
    </div>
  </div>
</div>
