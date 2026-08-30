<script>
  let { items = [] } = $props();
</script>

<nav class="flex items-center gap-1.5 text-xs text-gray-400 mb-4">
  {#each items as item, i}
    {#if i > 0}
      <span class="text-gray-300">/</span>
    {/if}
    {#if item.href}
      <a href={item.href} class="hover:text-blue-600 transition-colors">{item.label}</a>
    {:else}
      <span class="text-gray-700 font-medium">{item.label}</span>
    {/if}
  {/each}
</nav>
