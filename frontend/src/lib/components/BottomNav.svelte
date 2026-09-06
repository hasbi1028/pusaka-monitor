<script>
  import Icon from '$lib/components/Icon.svelte';
  let { activePage = '' } = $props();
  
  const items = [
    { href: '/dashboard', page: 'dashboard', icon: 'house', label: 'Beranda' },
    { href: '/pegawai', page: 'pegawai', icon: 'users', label: 'Pegawai' },
    { href: '/scrape', page: 'scrape', icon: 'rotate', label: 'Scrape' },
    { href: '/laporan', page: 'laporan', icon: 'chart-bar', label: 'Laporan' },
    { href: '/settings', page: 'settings', icon: 'gear', label: 'Lainnya' }
  ];
</script>

<nav class="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 lg:hidden z-30">
  <div class="flex justify-around py-0.5">
    {#each items as item}
      <a href={item.href}
         class="flex flex-col items-center gap-0 px-2 py-1 text-[9px] font-medium transition-colors"
         class:text-blue-600={item.page === activePage}
         class:text-gray-500={item.page !== activePage}>
        <Icon name={item.icon} class="text-sm" />
        <span>{item.label}</span>
      </a>
    {/each}
  </div>
</nav>
