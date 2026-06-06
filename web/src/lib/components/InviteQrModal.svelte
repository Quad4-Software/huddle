<script lang="ts">
  import Icon from './Icon.svelte';
  import { mdiClose, mdiCheck, mdiContentCopy } from '../icons';
  import { generateInviteQrDataUrl } from '../invite-qr';

  let {
    open,
    url,
    roomName = 'Room',
    onClose,
  }: {
    open: boolean;
    url: string;
    roomName?: string;
    onClose: () => void;
  } = $props();

  let qrDataUrl = $state('');
  let loading = $state(false);
  let error = $state('');
  let copied = $state(false);

  $effect(() => {
    if (!open || !url) {
      qrDataUrl = '';
      error = '';
      loading = false;
      return;
    }

    let cancelled = false;
    loading = true;
    error = '';
    qrDataUrl = '';

    void generateInviteQrDataUrl(url)
      .then((dataUrl) => {
        if (cancelled) return;
        qrDataUrl = dataUrl;
        loading = false;
      })
      .catch(() => {
        if (cancelled) return;
        error = 'Could not generate QR code';
        loading = false;
      });

    return () => {
      cancelled = true;
    };
  });

  async function copyLink() {
    if (!url) return;
    await navigator.clipboard.writeText(url);
    copied = true;
    setTimeout(() => (copied = false), 2000);
  }

  function closeModal() {
    copied = false;
    onClose();
  }
</script>

{#if open}
  <div
    class="fixed inset-0 z-50 flex items-end justify-center bg-black/60 p-0 sm:items-center sm:p-4"
    role="presentation"
    onclick={closeModal}
  >
    <div
      class="w-full max-w-sm rounded-t-2xl border border-border bg-surface-1 shadow-2xl sm:rounded-xl"
      role="dialog"
      aria-labelledby="invite-qr-title"
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.key === 'Escape' && closeModal()}
      style="padding-bottom: max(0px, env(safe-area-inset-bottom));"
    >
      <div class="flex items-center justify-between border-b border-border px-4 py-4">
        <h2 id="invite-qr-title" class="text-base font-semibold">Invite to {roomName}</h2>
        <button
          type="button"
          onclick={closeModal}
          class="flex h-11 w-11 items-center justify-center rounded-lg text-muted transition-colors hover:bg-surface-2 hover:text-foreground"
          aria-label="Close invite QR"
        >
          <Icon path={mdiClose} size={18} />
        </button>
      </div>

      <div class="flex flex-col items-center px-4 py-6">
        <p class="mb-4 text-center text-sm text-muted">Scan to join on another device</p>

        {#if loading}
          <div
            class="flex h-[280px] w-[280px] items-center justify-center rounded-xl border border-border bg-surface-2"
          >
            <span class="text-sm text-muted">Generating...</span>
          </div>
        {:else if error}
          <div
            class="flex h-[280px] w-[280px] items-center justify-center rounded-xl border border-danger/30 bg-danger/10 px-4 text-center text-sm text-danger"
          >
            {error}
          </div>
        {:else if qrDataUrl}
          <img
            src={qrDataUrl}
            width="280"
            height="280"
            alt="QR code for room invite link"
            class="rounded-xl border border-border"
          />
        {/if}

        <button
          type="button"
          onclick={copyLink}
          disabled={!url}
          class="mt-5 flex w-full items-center justify-center gap-2 rounded-lg border border-border bg-surface-2 py-2.5 text-sm font-medium transition-colors hover:border-accent hover:text-foreground disabled:opacity-50"
        >
          <Icon path={copied ? mdiCheck : mdiContentCopy} size={18} />
          {copied ? 'Link copied' : 'Copy invite link'}
        </button>
      </div>
    </div>
  </div>
{/if}
