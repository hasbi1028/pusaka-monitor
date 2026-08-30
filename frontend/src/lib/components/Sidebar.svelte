<script>
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';

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
    { href: '/superadmin/users', page: 'users', icon: 'fa-user-shield', label: 'User Mgmt', superadminOnly: true },
    { href: '/profil', page: 'profil', icon: 'fa-circle-user', label: 'Profil' }
  ];

  // Filter based on role
  let navItems = $derived(allNavItems.filter(item => !item.superadminOnly || userRole === 'superadmin'));

  function navClass(page) {
    return page === activePage
      ? 'bg-blue-50 text-blue-700 font-semibold'
      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900';
  }
</script>

{#if drawerOpen}
  <div class="fixed inset-0 bg-black/50 z-40 lg:hidden" onclick={closeDrawer}></div>
{/if}

<!-- Sidebar Drawer (mobile) -->
<aside class="fixed top-0 left-0 z-50 h-full w-64 bg-white border-r border-gray-200 transform transition-transform duration-200 lg:hidden"
       class:-translate-x-full={!drawerOpen}
       class:translate-x-0={drawerOpen}>
  <div class="flex items-center justify-between p-4 border-b border-gray-200">
    <div class="flex items-center gap-2">
      <div class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
        <i class="fa-solid fa-shield-halved text-white text-sm"></i>
      </div>
      <div>
        <div class="text-sm font-bold text-gray-900">Pusaka Monitor</div>
        <div class="text-[10px] text-gray-400">Sistem Monitoring</div>
      </div>
    </div>
    <button onclick={closeDrawer} class="p-1 rounded hover:bg-gray-100" aria-label="Close menu">
      <i class="fa-solid fa-xmark text-gray-500"></i>
    </button>
  </div>
  <nav class="p-3 space-y-1">
    {#each navItems as item}
      <a href={item.href} onclick={closeDrawer}
         class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors {navClass(item.page)}">
        <i class="fa-solid {item.icon} w-5 text-center text-xs"></i>
        <span>{item.label}</span>
      </a>
    {/each}
    <div class="pt-2 mt-2 border-t border-gray-100">
      <button onclick={handleLogout}
              class="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm font-medium text-red-600 hover:bg-red-50 transition-colors">
        <i class="fa-solid fa-right-from-bracket w-5 text-center text-xs"></i>
        <span>Keluar</span>
      </button>
    </div>
  </nav>
</aside>

<!-- Sidebar Desktop (fixed) -->
<aside class="hidden lg:flex lg:flex-col lg:w-56 lg:fixed lg:inset-y-0 lg:left-0 lg:z-30 bg-white border-r border-gray-200">
  <div class="flex items-center gap-2 p-4 border-b border-gray-200">
    <div class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
      <i class="fa-solid fa-shield-halved text-white text-sm"></i>
    </div>
    <div>
      <div class="text-sm font-bold text-gray-900">Pusaka Monitor</div>
      <div class="text-[10px] text-gray-400">Monitoring Kehadiran</div>
    </div>
  </div>
  <nav class="flex-1 p-3 space-y-1 overflow-y-auto">
    {#each navItems as item}
      <a href={item.href}
         class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors {navClass(item.page)}">
        <i class="fa-solid {item.icon} w-5 text-center text-xs"></i>
        <span>{item.label}</span>
      </a>
    {/each}
    <div class="pt-2 mt-2 border-t border-gray-100">
      <button onclick={handleLogout}
              class="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm font-medium text-red-600 hover:bg-red-50 transition-colors">
        <i class="fa-solid fa-right-from-bracket w-5 text-center text-xs"></i>
        <span>Keluar</span>
      </button>
    </div>
  </nav>
  <div class="p-3 border-t border-gray-200 text-center text-[10px] text-gray-400">
    v1.0 · Pusaka Kemenag
  </div>
</aside>
