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
<div class="lg:pl-56 min-h-screen pb-20 lg:pb-0">
  <Topbar {title} onToggleDrawer={() => toggleDrawerFn()} />
  <main class="p-4 page-enter">
    {@render children()}
  </main>
</div>
<BottomNav {activePage} />
