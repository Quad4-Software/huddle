<script lang="ts">
  import { QUAD4_NAME, QUAD4_URL } from '../version';

  let { class: className = '' }: { class?: string } = $props();
</script>

<p class="text-xs text-muted {className}">
  Created by
  <a
    href={QUAD4_URL}
    target="_blank"
    rel="noopener noreferrer"
    class="text-highlight hover:underline"
  >
    {QUAD4_NAME}
  </a>
</p>
