<script>
  let { items = [] } = $props();
</script>

<nav data-testid="breadcrumb" class="flex items-center gap-1 text-[10px] text-gray-400 mb-2">
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
