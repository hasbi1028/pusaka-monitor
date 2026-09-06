<script>
  import { auth } from '../stores/auth.js';
  import Icon from '$lib/components/Icon.svelte';

  let { title = '', onToggleDrawer = () => {} } = $props();
</script>

<header class="sticky top-0 z-30 bg-white border-b border-gray-200 px-3 py-1.5 flex items-center justify-between lg:pl-48">
  <div class="flex items-center gap-2">
    <button onclick={onToggleDrawer} class="lg:hidden p-1 rounded-md hover:bg-gray-100 text-gray-600" aria-label="Buka menu">
      <Icon name="bars" class="text-sm" />
    </button>
    <h1 class="text-sm font-semibold text-gray-900">{title}</h1>
    <a href="/kepatuhan" title="Aplikasi ini hanya membaca data — tidak membuat absen"
       class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full bg-green-50 border border-green-200 text-green-700 text-[9px] font-bold tracking-wide">
      <Icon name="eye" class="text-[8px]" />READ-ONLY
    </a>
  </div>
  <div class="flex items-center gap-1.5">
    {#if $auth.user}
      <span class="text-[10px] text-gray-500 hidden sm:inline">{$auth.user.username}</span>
    {/if}
    <div class="w-6 h-6 rounded-full bg-blue-100 flex items-center justify-center">
      <Icon name="user" class="text-blue-600 text-[10px]" />
    </div>
  </div>
</header>
