<script lang="ts">
  import Icon from './Icon.svelte';
  import { mdiPaperclip } from '../icons';
  import { isGifMime, isImageMime, isVideoMime } from '../attachments';

  let {
    blob,
    name,
    mime,
  }: {
    blob: Blob;
    name: string;
    mime: string;
  } = $props();

  let objectUrl = $state('');

  $effect(() => {
    const url = URL.createObjectURL(blob);
    objectUrl = url;
    return () => URL.revokeObjectURL(url);
  });

  const showImage = $derived(isImageMime(mime) || isGifMime(mime, name));
  const showVideo = $derived(isVideoMime(mime));
</script>

{#if showImage}
  <a href={objectUrl} target="_blank" rel="noreferrer" class="block">
    <img src={objectUrl} alt={name} class="max-h-72 max-w-full object-contain" />
  </a>
{:else if showVideo}
  <video src={objectUrl} controls playsinline preload="metadata" class="max-h-72 max-w-full">
    <track kind="captions" />
  </video>
{:else}
  <a
    href={objectUrl}
    download={name}
    class="m-2 inline-flex items-center gap-1.5 rounded-md bg-surface-3 px-2.5 py-1.5 text-xs text-highlight hover:bg-surface-2"
  >
    <Icon path={mdiPaperclip} size={14} />
    {name}
  </a>
{/if}
