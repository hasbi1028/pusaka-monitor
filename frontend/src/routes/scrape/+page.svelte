<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import { scrape } from '$lib/api.js';
  import { toasts } from '$lib/stores/toast.js';
  import { onMount, onDestroy } from 'svelte';

  let stats = $state({ pending: 0, running: 0, done: 0, failed: 0 });
  let jobs = $state([]);
  let searchQ = $state('');
  let statusFilter = $state('');
  let interval;

  function jobBadge(status) {
    const m = {done:'bg-green-100 text-green-700',failed:'bg-red-100 text-red-700',running:'bg-cyan-100 text-cyan-700',pending:'bg-gray-100 text-gray-600',cancelled:'bg-gray-100 text-gray-500'};
    return m[status] || 'bg-gray-100 text-gray-600';
  }

  async function loadStatus() {
    const res = await scrape.status();
    if (res.success) {
      stats = res.data.stats;
      jobs = res.data.jobs;
    }
  }

  let filteredJobs = $derived(jobs.filter(j =>
    (!searchQ || (j.employee_name || '').toLowerCase().includes(searchQ.toLowerCase())) &&
    (!statusFilter || j.status === statusFilter)
  ));

  async function createJobs(mode) {
    const label = mode === 'belum_masuk' ? 'Belum Masuk' : mode === 'belum_pulang' ? 'Belum Pulang' : 'Semua';
    if (!confirm('Buat job scrape: ' + label + '?')) return;
    await scrape.createJobs(mode);
    loadStatus();
    toasts.success('Job scrape ' + label + ' dibuat');
  }

  async function retryFailed() {
    await scrape.retryFailed();
    loadStatus();
    toasts.warning('Job gagal di-retry');
  }

  async function cancelAll() {
    if (!confirm('Batalkan SEMUA job pending & running?')) return;
    const res = await scrape.cancelAll();
    loadStatus();
    if (res.success) toasts.success(res.data.cancelled + ' job dibatalkan');
    else toasts.error(res.error || 'Gagal');
  }

  async function cancelOne(id) {
    if (!confirm('Batalkan job ini?')) return;
    const res = await scrape.cancelOne(id);
    loadStatus();
    if (res.success) toasts.success('Job dibatalkan');
    else toasts.error(res.error || 'Gagal');
  }

  onMount(() => { loadStatus(); interval = setInterval(loadStatus, 5000); });
  onDestroy(() => { if (interval) clearInterval(interval); });
</script>

<Layout title="Scrape" activePage="scrape">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Scrape' }]} />
  <!-- Action Buttons -->
  <div class="bg-white rounded-xl border border-gray-200 shadow-sm p-4 mb-4">
    <button onclick={() => createJobs('all')}
            class="w-full px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors mb-3">
      <i class="fa-solid fa-bolt mr-1"></i> Scrape Semua
    </button>
    <div class="grid grid-cols-2 gap-2">
      <button onclick={() => createJobs('belum_masuk')}
              class="px-3 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg text-sm font-medium transition-colors">
        <i class="fa-solid fa-clock mr-1"></i> Belum Masuk
      </button>
      <button onclick={() => createJobs('belum_pulang')}
              class="px-3 py-2 border border-gray-200 hover:bg-gray-50 text-gray-700 rounded-lg text-sm font-medium transition-colors">
        <i class="fa-solid fa-clock mr-1"></i> Belum Pulang
      </button>
    </div>
  </div>
  
  <!-- Stats -->
  <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
    <div class="bg-white rounded-xl border border-gray-200 p-3 shadow-sm border-l-4 border-l-gray-400">
      <div class="text-xl font-bold">{stats.pending}</div>
      <div class="text-[10px] text-gray-500 uppercase">Pending</div>
    </div>
    <div class="bg-white rounded-xl border border-gray-200 p-3 shadow-sm border-l-4 border-l-cyan-500">
      <div class="text-xl font-bold text-cyan-600">{stats.running}</div>
      <div class="text-[10px] text-gray-500 uppercase">Running</div>
    </div>
    <div class="bg-white rounded-xl border border-gray-200 p-3 shadow-sm border-l-4 border-l-green-500">
      <div class="text-xl font-bold text-green-600">{stats.done}</div>
      <div class="text-[10px] text-gray-500 uppercase">Selesai</div>
    </div>
    <div class="bg-white rounded-xl border border-gray-200 p-3 shadow-sm border-l-4 border-l-red-500">
      <div class="text-xl font-bold text-red-600">{stats.failed}</div>
      <div class="text-[10px] text-gray-500 uppercase">Gagal</div>
    </div>
  </div>
  
  <!-- Retry + Cancel -->
  <div class="grid grid-cols-2 gap-2 mb-4">
    <button onclick={retryFailed}
            class="px-3 py-2 border border-gray-200 hover:bg-gray-50 rounded-lg text-sm font-medium text-gray-700 transition-colors">
      <i class="fa-solid fa-rotate mr-1"></i> Retry Gagal
    </button>
    <button onclick={cancelAll}
            class="px-3 py-2 bg-red-50 hover:bg-red-100 text-red-700 rounded-lg text-sm font-medium transition-colors">
      <i class="fa-solid fa-circle-xmark mr-1"></i> Cancel All
    </button>
  </div>
  
  <!-- Jobs Table -->
  <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
    <div class="px-4 py-2.5 border-b border-gray-200 flex items-center justify-between">
      <span class="text-sm font-semibold text-gray-700"><i class="fa-solid fa-list-check mr-1"></i> Antrian Job</span>
      <span class="bg-gray-100 text-gray-600 text-xs font-medium px-2 py-0.5 rounded-full">{jobs.length}</span>
    </div>
    <div class="px-4 py-2 border-b border-gray-100 flex gap-2">
      <input type="text" bind:value={searchQ} placeholder="Cari nama..."
             class="flex-1 px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
      <select bind:value={statusFilter}
              class="px-3 py-1.5 border border-gray-300 rounded-lg text-sm bg-white focus:ring-2 focus:ring-blue-500 outline-none">
        <option value="">Semua</option>
        <option value="pending">Pending</option>
        <option value="running">Running</option>
        <option value="done">Selesai</option>
        <option value="failed">Gagal</option>
        <option value="cancelled">Dibatalkan</option>
      </select>
    </div>
    <div class="overflow-x-auto max-h-[50vh] overflow-y-auto">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 text-gray-500 text-xs uppercase sticky top-0">
          <tr><th class="px-4 py-2 text-left">Nama</th><th class="px-4 py-2 text-center">Status</th><th class="px-4 py-2 text-center">Waktu</th><th class="px-4 py-2 text-left">Error</th><th class="px-4 py-2 text-center"></th></tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          {#each filteredJobs as j}
            <tr class="hover:bg-gray-50">
              <td class="px-4 py-2 font-medium text-gray-900">{j.employee_name}</td>
              <td class="px-4 py-2 text-center">
                <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium {jobBadge(j.status)}">{j.status}</span>
              </td>
              <td class="px-4 py-2 text-center text-xs">
                {j.created_at ? new Date(j.created_at).toLocaleTimeString('id-ID', {hour:'2-digit',minute:'2-digit',second:'2-digit'}) : '-'}
                {#if j.completed_at}
                  <br/><span class="text-green-600">✓ {new Date(j.completed_at).toLocaleTimeString('id-ID', {hour:'2-digit',minute:'2-digit',second:'2-digit'})}</span>
                {/if}
              </td>
              <td class="px-4 py-2 text-red-500 text-xs">{j.error || ''}</td>
              <td class="px-4 py-2 text-center">
                {#if j.status === 'pending' || j.status === 'running'}
                  <button onclick={() => cancelOne(j.id)} class="p-1 rounded hover:bg-red-50 text-red-500">
                    <i class="fa-solid fa-xmark"></i>
                  </button>
                {/if}
              </td>
            </tr>
          {:else}
            <tr><td colspan="5" class="px-4 py-6 text-center text-gray-400 text-sm">Tidak ada job</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</Layout>
