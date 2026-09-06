<script>
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { theme } from '../stores/theme.js';

  let { activePage = '' } = $props();
  let drawerOpen = $state(false);
  let userRole = $state('');

  function toggleDrawer() { drawerOpen = !drawerOpen; }
  function closeDrawer() { drawerOpen = false; }

  async function handleLogout() {
    await fetch('/api/auth/logout', { method: 'POST' });
    window.location.href = '/login';
  }

  // Fetch user role
  onMount(async () => {
    theme.init();
    try {
      const res = await fetch('/api/auth/me');
      const data = await res.json();
      if (data.success) userRole = data.data.role || '';
    } catch {}
  });

  // All nav items
  const allNavItems = [
    { href: '/dashboard', page: 'dashboard', icon: 'fa-house', label: 'Beranda' },
    { href: '/pegawai', page: 'pegawai', icon: 'fa-users', label: 'Pegawai' },
    { href: '/scrape', page: 'scrape', icon: 'fa-rotate', label: 'Scrape' },
    { href: '/laporan', page: 'laporan', icon: 'fa-chart-bar', label: 'Laporan' },
    { href: '/settings', page: 'settings', icon: 'fa-gear', label: 'Pengaturan' },
    { href: '/kepatuhan', page: 'kepatuhan', icon: 'fa-scale-balanced', label: 'Kepatuhan' },
    { href: '/superadmin/users', page: 'users', icon: 'fa-user-shield', label: 'User Mgmt', superadminOnly: true },
    { href: '/profil', page: 'profil', icon: 'fa-circle-user', label: 'Profil' }
  ];

  // Filter based on role
  let navItems = $derived(allNavItems.filter(item => !item.superadminOnly || userRole === 'superadmin'));

  function navClass(page) {
    return page === activePage
      ? 'bg-blue-50 text-blue-700 font-semibold dark:bg-blue-500/20 dark:text-blue-300'
      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-slate-400 dark:hover:bg-slate-700 dark:hover:text-slate-100';
  }
</script>

{#if drawerOpen}
  <div class="fixed inset-0 bg-black/50 z-40 lg:hidden" onclick={closeDrawer} onkeydown={(e) => { if (e.key === 'Escape') closeDrawer(); }} role="presentation"></div>
{/if}

<!-- Sidebar Drawer (mobile) -->
<aside class="fixed top-0 left-0 z-50 h-full w-56 bg-white dark:bg-slate-800 border-r border-gray-200 dark:border-slate-700 transform transition-transform duration-200 lg:hidden"
       class:-translate-x-full={!drawerOpen}
       class:translate-x-0={drawerOpen}>
  <div class="flex items-center justify-between px-3 py-2 border-b border-gray-200 dark:border-slate-700">
    <div class="flex items-center gap-1.5">
      <div class="w-7 h-7 bg-gradient-to-br from-blue-600 to-blue-900 dark:from-blue-500 dark:to-blue-950 rounded-md flex items-center justify-center">
        <i class="fa-solid fa-shield-halved text-white text-xs"></i>
      </div>
      <div>
        <div class="text-xs font-bold text-gray-900">Pusaka Monitor</div>
        <div class="text-[9px] text-green-600 font-medium">Read-Only · Hanya Baca</div>
      </div>
    </div>
    <button onclick={closeDrawer} class="p-0.5 rounded hover:bg-gray-100" aria-label="Close menu">
      <i class="fa-solid fa-xmark text-gray-500 text-xs"></i>
    </button>
  </div>
  <nav class="p-2 space-y-0.5">
    {#each navItems as item (item.page)}
      <a href={item.href} onclick={closeDrawer}
         class="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-xs transition-colors {navClass(item.page)}">
        <i class="fa-solid {item.icon} w-4 text-center text-[10px]"></i>
        <span>{item.label}</span>
      </a>
    {/each}
    <div class="pt-1.5 mt-1.5 border-t border-gray-100">
      <button onclick={() => theme.toggle()}
              class="flex items-center gap-2 w-full px-2.5 py-1.5 rounded-md text-xs font-medium text-gray-600 hover:bg-gray-100 dark:text-slate-300 dark:hover:bg-slate-700 transition-colors"
              aria-label="Ganti mode gelap terang">
        <i class="fa-solid {$theme === 'dark' ? 'fa-sun' : 'fa-moon'} w-4 text-center text-[10px]"></i>
        <span>{$theme === 'dark' ? 'Mode Terang' : 'Mode Gelap'}</span>
      </button>
      <button onclick={handleLogout}
              class="flex items-center gap-2 w-full px-2.5 py-1.5 rounded-md text-xs font-medium text-red-600 hover:bg-red-50 transition-colors">
        <i class="fa-solid fa-right-from-bracket w-4 text-center text-[10px]"></i>
        <span>Keluar</span>
      </button>
    </div>
  </nav>
</aside>

<!-- Sidebar Desktop (fixed) -->
<aside class="hidden lg:flex lg:flex-col lg:w-48 lg:fixed lg:inset-y-0 lg:left-0 lg:z-30 bg-white dark:bg-slate-800 border-r border-gray-200 dark:border-slate-700">
  <div class="flex items-center gap-1.5 px-3 py-2 border-b border-gray-200 dark:border-slate-700">
    <div class="w-7 h-7 bg-gradient-to-br from-blue-600 to-blue-900 dark:from-blue-500 dark:to-blue-950 rounded-md flex items-center justify-center">
      <i class="fa-solid fa-shield-halved text-white text-xs"></i>
    </div>
    <div>
      <div class="text-xs font-bold text-gray-900">Pusaka Monitor</div>
      <div class="text-[9px] text-green-600 font-medium">Read-Only · Hanya Baca</div>
    </div>
  </div>
  <nav class="flex-1 p-2 space-y-0.5 overflow-y-auto">
    {#each navItems as item (item.page)}
      <a href={item.href}
         class="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-xs transition-colors {navClass(item.page)}">
        <i class="fa-solid {item.icon} w-4 text-center text-[10px]"></i>
        <span>{item.label}</span>
      </a>
    {/each}
    <div class="pt-1.5 mt-1.5 border-t border-gray-100">
      <button onclick={() => theme.toggle()}
              class="flex items-center gap-2 w-full px-2.5 py-1.5 rounded-md text-xs font-medium text-gray-600 hover:bg-gray-100 dark:text-slate-300 dark:hover:bg-slate-700 transition-colors"
              aria-label="Ganti mode gelap terang">
        <i class="fa-solid {$theme === 'dark' ? 'fa-sun' : 'fa-moon'} w-4 text-center text-[10px]"></i>
        <span>{$theme === 'dark' ? 'Mode Terang' : 'Mode Gelap'}</span>
      </button>
      <button onclick={handleLogout}
              class="flex items-center gap-2 w-full px-2.5 py-1.5 rounded-md text-xs font-medium text-red-600 hover:bg-red-50 transition-colors">
        <i class="fa-solid fa-right-from-bracket w-4 text-center text-[10px]"></i>
        <span>Keluar</span>
      </button>
    </div>
  </nav>
  <div class="px-2 py-1.5 border-t border-gray-200 dark:border-slate-700 text-center text-[9px] text-gray-400">
    v1.0.3 · Hanya membaca Pusaka
  </div>
</aside>
