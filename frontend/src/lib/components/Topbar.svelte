<script>
  import { auth } from '../stores/auth.js';

  let { title = '', onToggleDrawer = () => {} } = $props();
</script>

<header class="sticky top-0 z-30 bg-white border-b border-gray-200 px-4 py-2 flex items-center justify-between lg:pl-56">
  <div class="flex items-center gap-3">
    <button onclick={onToggleDrawer} class="lg:hidden p-2 rounded-lg hover:bg-gray-100 text-gray-600">
      <i class="fa-solid fa-bars text-lg"></i>
    </button>
    <h1 class="text-base font-semibold text-gray-900">{title}</h1>
  </div>
  <div class="flex items-center gap-2">
    {#if $auth.user}
      <span class="text-xs text-gray-500 hidden sm:inline">{$auth.user.username}</span>
    {/if}
    <div class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center">
      <i class="fa-solid fa-user text-blue-600 text-xs"></i>
    </div>
  </div>
</header>
