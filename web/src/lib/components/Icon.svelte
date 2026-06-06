<script lang="ts">
  let {
    path,
    size = 20,
    class: className = '',
  }: { path: string; size?: number; class?: string } = $props();
</script>

<svg
  viewBox="0 0 24 24"
  width={size}
  height={size}
  class={className}
  fill="currentColor"
  aria-hidden="true"
>
  <path d={path} />
</svg>
