<script lang="ts">
  import Icon from './Icon.svelte';
  import Avatar from './Avatar.svelte';
  import EmojiPicker from './EmojiPicker.svelte';
  import MessageAttachment from './MessageAttachment.svelte';
  import { mdiSend, mdiPaperclip, mdiEmoticonOutline } from '../icons';
  import { session } from '../stores/session.svelte';
  import { sendMessage, sendFile, toggleReaction } from '../session-controller';
  import { MAX_CHAT_MESSAGE_LENGTH, MAX_FILE_SIZE } from '../validation';

  let text = $state('');
  let fileInput: HTMLInputElement;
  let chatEl: HTMLDivElement;
  let picker = $state<string | null>(null);

  const messages = $derived(session.messagesForChannel(session.activeChannel));
  const peerCount = $derived((session.room?.members.length ?? 1) - 1);

  $effect(() => {
    const count = messages.length;
    if (count >= 0 && chatEl) chatEl.scrollTop = chatEl.scrollHeight;
  });

  async function submit() {
    const value = text.slice(0, MAX_CHAT_MESSAGE_LENGTH).trim();
    if (!value) return;
    text = '';
    session.error = '';
    await sendMessage(value);
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      submit();
    }
  }

  async function onFile(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (file && file.size > MAX_FILE_SIZE) {
      session.error = 'File is too large';
      input.value = '';
      return;
    }
    if (file) await sendFile(file);
    input.value = '';
  }

  function formatTime(ts: number) {
    return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  function onPick(emoji: string) {
    if (picker === 'composer') {
      text += emoji;
    } else if (picker) {
      toggleReaction(picker, emoji);
    }
    picker = null;
  }

  function react(messageId: string, emoji: string) {
    toggleReaction(messageId, emoji);
  }
</script>

<div class="relative flex h-full flex-col">
  <div class="flex items-center justify-between border-b border-border px-4 py-3">
    <span class="text-sm font-semibold">{session.room?.name ?? 'Room'}</span>
    <span class="text-xs {session.meshReady || peerCount === 0 ? 'text-online' : 'text-away'}">
      {peerCount === 0 ? 'just you' : session.meshReady ? 'connected' : 'connecting'}
    </span>
  </div>

  {#if session.error}
    <p class="border-b border-danger/30 bg-danger/10 px-4 py-2 text-xs text-danger">
      {session.error}
    </p>
  {/if}

  <div bind:this={chatEl} class="flex-1 space-y-4 overflow-y-auto p-4">
    {#each messages as msg (msg.id)}
      {@const mine = msg.authorId === session.peerId}
      {@const reactions = session.reactions[msg.id] ?? []}
      {@const attachBlob = msg.attachment ? session.attachments[msg.attachment.id] : undefined}
      {@const hasBody = msg.text || msg.attachment}
      <div class="group flex gap-3 {mine ? 'flex-row-reverse' : ''}">
        <Avatar name={msg.authorName} size={32} />
        <div
          class="flex min-w-0 max-w-[78%] flex-col {mine ? 'items-end text-right' : 'items-start'}"
        >
          <div class="mb-1 flex items-baseline gap-2 {mine ? 'flex-row-reverse' : ''}">
            <span class="text-sm font-medium text-highlight">{mine ? 'You' : msg.authorName}</span>
            <span class="text-xs text-muted">{formatTime(msg.timestamp)}</span>
          </div>

          {#if hasBody}
            <div
              class="inline-block max-w-full overflow-hidden rounded-xl text-sm leading-relaxed {mine
                ? 'border border-bubble-self-border bg-bubble-self text-text'
                : 'bg-surface-2'}"
            >
              {#if msg.text}
                <p class="whitespace-pre-wrap break-words px-3 py-2">{msg.text}</p>
              {/if}
              {#if msg.attachment}
                {#if attachBlob}
                  <MessageAttachment
                    blob={attachBlob}
                    name={msg.attachment.name}
                    mime={msg.attachment.mime}
                  />
                {:else}
                  <p class="px-3 py-2 text-xs text-muted">Receiving {msg.attachment.name}</p>
                {/if}
              {/if}
            </div>
          {/if}

          <div
            class="mt-1.5 flex max-w-full flex-wrap items-center gap-1 {mine ? 'justify-end' : ''}"
          >
            {#each reactions as r (r.emoji)}
              <button
                onclick={() => react(msg.id, r.emoji)}
                class="flex items-center gap-1 rounded-full border px-1.5 py-0.5 text-xs transition-colors {r.peerIds.includes(
                  session.peerId,
                )
                  ? 'border-accent bg-accent/15 text-accent'
                  : 'border-border bg-surface-2 text-muted hover:bg-surface-3'}"
              >
                <span>{r.emoji}</span>
                <span>{r.peerIds.length}</span>
              </button>
            {/each}
            <button
              onclick={(e) => {
                e.stopPropagation();
                picker = picker === msg.id ? null : msg.id;
              }}
              class="rounded-full p-1 text-muted opacity-0 transition-opacity hover:bg-surface-3 hover:text-text group-hover:opacity-100"
              title="Add reaction"
            >
              <Icon path={mdiEmoticonOutline} size={16} />
            </button>
          </div>
        </div>
      </div>
    {:else}
      <p class="mt-8 text-center text-sm text-muted">
        Messages are end to end encrypted and peer to peer. Say hello.
      </p>
    {/each}
  </div>

  {#if picker}
    <div class="absolute bottom-20 right-4 z-30">
      <EmojiPicker {onPick} onClose={() => (picker = null)} />
    </div>
  {/if}

  <div class="border-t border-border p-3">
    <div class="flex items-center gap-2">
      <div class="relative min-w-0 flex-1">
        <input
          type="text"
          bind:value={text}
          maxlength={MAX_CHAT_MESSAGE_LENGTH}
          onkeydown={onKey}
          placeholder="Send a message"
          class="w-full rounded-lg border border-border bg-surface-2 py-2 pl-3 pr-[4.25rem] text-sm outline-none focus:border-accent"
        />
        <div class="absolute top-1/2 right-1 flex -translate-y-1/2 items-center">
          <button
            onclick={() => fileInput.click()}
            class="rounded-md p-1.5 text-muted transition-colors hover:bg-surface-3 hover:text-text"
            title="Attach file, image, or video"
          >
            <Icon path={mdiPaperclip} size={18} />
          </button>
          <input
            type="file"
            accept="image/*,video/*,*/*"
            bind:this={fileInput}
            class="hidden"
            onchange={onFile}
          />
          <button
            onclick={(e) => {
              e.stopPropagation();
              picker = picker === 'composer' ? null : 'composer';
            }}
            class="rounded-md p-1.5 text-muted transition-colors hover:bg-surface-3 hover:text-text"
            title="Emoji"
          >
            <Icon path={mdiEmoticonOutline} size={18} />
          </button>
        </div>
      </div>
      <button
        onclick={submit}
        class="shrink-0 rounded-lg bg-accent p-2 text-white transition-colors hover:bg-accent-hover"
        title="Send"
      >
        <Icon path={mdiSend} size={20} />
      </button>
    </div>
  </div>
</div>
