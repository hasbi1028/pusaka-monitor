<script>
  import Sidebar from './Sidebar.svelte';
  import Topbar from './Topbar.svelte';
  import BottomNav from './BottomNav.svelte';
  import Toast from './Toast.svelte';

  let { title = '', activePage = '', children } = $props();

  let toggleDrawerFn;
</script>

<Toast />
<Sidebar {activePage} bind:toggleDrawer={toggleDrawerFn} />
<div class="lg:pl-48 min-h-screen pb-16 lg:pb-0 flex flex-col">
  <Topbar {title} onToggleDrawer={() => toggleDrawerFn()} />
  <main class="p-2.5 page-enter flex-1">
    {@render children()}
  </main>
  <footer class="px-3 py-2 text-center text-[9px] text-gray-400 border-t border-gray-100">
    <i class="fa-solid fa-shield-halved mr-0.5"></i>
    Pusaka Monitor <strong class="text-gray-500">hanya membaca</strong> riwayat Pusaka Kemenag — tidak membuat/mengubah absen.
    <a href="/kepatuhan" class="text-blue-500 hover:underline ml-1">Kepatuhan</a>
  </footer>
</div>
<BottomNav {activePage} />
