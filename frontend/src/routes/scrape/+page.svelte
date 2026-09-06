<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import Loading from '$lib/components/Loading.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import { scrape } from '$lib/api.js';
  import { toasts } from '$lib/stores/toast.js';
  import { onMount, onDestroy } from 'svelte';

  let stats = $state({ pending: 0, running: 0, done: 0, failed: 0, cancelled: 0 });
  let jobs = $state([]);
  let searchQ = $state('');
  let statusFilter = $state('');
  let loading = $state(true);
  let busyAction = $state('');
  let live = $state(false);
  let es = null;
  let fallbackTimer = null;

  // Confirm dialog state
  let confirmShow = $state(false);
  let confirmTitle = $state('');
  let confirmMessage = $state('');
  let confirmAction = $state(() => {});
  let confirmVariant = $state('danger');

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
    loading = false;
  }

  let filteredJobs = $derived(
    jobs.filter(j =>
        (!searchQ || (j.employee_name || '').toLowerCase().includes(searchQ.toLowerCase())) &&
        (!statusFilter || j.status === statusFilter)
      )
  );

  // Seksi kategori ala gambar referensi: Berjalan > Menunggu > Selesai > Gagal
  let collapsed = $state({});
  let sections = $derived([
    { key: 'running', label: 'Berjalan', icon: 'fa-solid fa-spinner fa-spin', head: 'bg-cyan-50 text-cyan-700', chip: jobBadge('running'), count: stats.running || 0,
      list: filteredJobs.filter(j => j.status === 'running') },
    { key: 'pending', label: 'Menunggu', icon: 'fa-solid fa-clock', head: 'bg-gray-100 text-gray-700', chip: jobBadge('pending'), count: stats.pending || 0,
      list: filteredJobs.filter(j => j.status === 'pending') },
    { key: 'done', label: 'Selesai', icon: 'fa-solid fa-circle-check', head: 'bg-green-50 text-green-700', chip: jobBadge('done'), count: stats.done || 0,
      list: filteredJobs.filter(j => j.status === 'done') },
    { key: 'fail', label: 'Gagal & Batal', icon: 'fa-solid fa-triangle-exclamation', head: 'bg-red-50 text-red-700', chip: jobBadge('failed'), count: (stats.failed || 0) + (stats.cancelled || 0),
      list: filteredJobs.filter(j => j.status === 'failed' || j.status === 'cancelled') },
  ]);

  // Tanggal + jam: "3/9 14:42:13" (bukan jam saja) agar urutan terbaca
  function fmtDT(iso) {
    if (!iso) return '-';
    const d = new Date(iso);
    return `${d.getDate()}/${d.getMonth() + 1} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}`;
  }

  function openConfirm(title, message, action, variant = 'danger') {
    confirmTitle = title;
    confirmMessage = message;
    confirmAction = action;
    confirmVariant = variant;
    confirmShow = true;
  }

  // Optimistic create — add dummy jobs to pending instantly
  async function createJobs(mode) {
    if (busyAction) return;
    const label = mode === 'belum_masuk' ? 'Belum Masuk' : mode === 'belum_pulang' ? 'Belum Pulang' : 'Semua';
    openConfirm('Buat Job Scrape', 'Buat job scrape: ' + label + '?', async () => {
      busyAction = 'create-' + mode;
      const prevJobs = [...jobs];
      const prevStats = { ...stats };
      // Create 3 optimistic pending jobs (placeholder)
      const dummyJobs = Array.from({ length: 3 }, (_, i) => ({
        id: 'optimistic-' + Date.now() + '-' + i,
        employee_name: 'Memuat...',
        status: 'pending',
        created_at: new Date().toISOString()
      }));
      jobs = [...dummyJobs, ...jobs];
      stats = { ...stats, pending: stats.pending + 3 };
      try {
        const res = await scrape.createJobs(mode);
        busyAction = '';
        // Remove optimistic jobs, fetch real data
        jobs = jobs.filter(j => !j.id.startsWith('optimistic-'));
        stats = { ...stats, pending: Math.max(0, stats.pending - 3) };
        await loadStatus();
        if (res.success) toasts.success('Job scrape ' + label + ' dibuat');
        else toasts.error(res.error || 'Gagal');
      } catch {
        jobs = prevJobs; stats = prevStats; busyAction = '';
        toasts.error('Gagal koneksi');
      }
    });
  }

  // Optimistic retry — mark failed jobs as pending instantly
  async function retryFailed() {
    if (busyAction) return;
    busyAction = 'retry';
    const prevJobs = [...jobs];
    const failedIds = jobs.filter(j => j.status === 'failed').map(j => j.id);
    jobs = jobs.map(j => j.status === 'failed' ? { ...j, status: 'pending', error: '' } : j);
    stats = { ...stats, failed: 0, pending: stats.pending + failedIds.length };
    try {
      const res = await scrape.retryFailed();
      busyAction = '';
      if (res.success) toasts.warning('Job gagal di-retry');
      else { jobs = prevJobs; toasts.error(res.error || 'Gagal'); }
    } catch {
      jobs = prevJobs; busyAction = '';
      toasts.error('Gagal koneksi');
    }
  }

  // Optimistic cancelAll — mark all pending/running as cancelled instantly
  async function cancelAll() {
    if (busyAction) return;
    openConfirm('Batalkan Semua Job', 'Batalkan SEMUA job pending & running?', async () => {
      busyAction = 'cancel-all';
      const prevJobs = [...jobs];
      const prevStats = { ...stats };
      const affected = jobs.filter(j => j.status === 'pending' || j.status === 'running');
      jobs = jobs.map(j => (j.status === 'pending' || j.status === 'running') ? { ...j, status: 'cancelled' } : j);
      stats = { ...stats, pending: 0, running: 0, cancelled: stats.cancelled + affected.length };
      try {
        const res = await scrape.cancelAll();
        busyAction = '';
        if (res.success) toasts.success(res.data.cancelled + ' job dibatalkan');
        else { jobs = prevJobs; stats = prevStats; toasts.error(res.error || 'Gagal'); }
      } catch {
        jobs = prevJobs; stats = prevStats; busyAction = '';
        toasts.error('Gagal koneksi');
      }
    });
  }

  // Optimistic cancelOne — mark single job as cancelled instantly
  async function cancelOne(id) {
    if (busyAction) return;
    openConfirm('Batalkan Job', 'Batalkan job ini?', async () => {
      const prevJobs = [...jobs];
      const prevStats = { ...stats };
      const job = jobs.find(j => j.id === id);
      jobs = jobs.map(j => j.id === id ? { ...j, status: 'cancelled' } : j);
      if (job) {
        if (job.status === 'pending') stats = { ...stats, pending: stats.pending - 1, cancelled: stats.cancelled + 1 };
        else if (job.status === 'running') stats = { ...stats, running: stats.running - 1, cancelled: stats.cancelled + 1 };
      }
      try {
        const res = await scrape.cancelOne(id);
        if (res.success) toasts.success('Job dibatalkan');
        else { jobs = prevJobs; stats = prevStats; toasts.error(res.error || 'Gagal'); }
      } catch {
        jobs = prevJobs; stats = prevStats;
        toasts.error('Gagal koneksi');
      }
    });
  }

  onMount(() => {
    loadStatus(); // snapshot awal via REST (cepat, sekali)
    connectStream();
    document.addEventListener('visibilitychange', onVisibility);
  });

  onDestroy(() => {
    document.removeEventListener('visibilitychange', onVisibility);
    disconnectStream();
  });

  function onVisibility() {
    // Hemat bandwidth + baterai: putus SSE saat tab disembunyikan,
    // sambung ulang + refresh saat tab aktif kembali.
    if (document.hidden) disconnectStream();
    else { loadStatus(); connectStream(); }
  }

  function connectStream() {
    if (es || typeof EventSource === 'undefined') { startFallback(); return; }
    try {
      es = new EventSource('/api/scrape/stream');
      es.addEventListener('stats', (e) => { stats = JSON.parse(e.data); live = true; });
      es.addEventListener('jobs', (e) => {
        const list = JSON.parse(e.data) || [];
        list.sort((a, b) => new Date(b.created_at || 0) - new Date(a.created_at || 0));
        jobs = list;
        live = true;
      });
      es.onerror = () => { disconnectStream(); startFallback(); };
    } catch { startFallback(); }
  }

  function disconnectStream() {
    live = false;
    if (es) { es.close(); es = null; }
    if (fallbackTimer) { clearInterval(fallbackTimer); fallbackTimer = null; }
  }

  // Fallback kalau SSE gagal: polling REST tiap 3 detik (bukan 500ms)
  function startFallback() {
    if (fallbackTimer) return;
    fallbackTimer = setInterval(loadStatus, 3000);
  }
</script>

<Layout title="Scrape" activePage="scrape">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Scrape' }]} />
  <!-- Action Buttons -->
  <div class="bg-white rounded-lg border border-gray-200 shadow-sm p-2.5 mb-2">
    <button onclick={() => createJobs('all')} disabled={busyAction}
            class="w-full px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-xs font-medium transition-colors mb-2 disabled:opacity-60">
      {#if busyAction === 'create-all'}
        <Loading variant="button" label="Membuat job..." />
      {:else}
        <i class="fa-solid fa-bolt mr-0.5"></i> Scrape Semua
      {/if}
    </button>
    <div class="grid grid-cols-2 gap-1.5">
      <button onclick={() => createJobs('belum_masuk')} disabled={busyAction}
              class="px-2 py-1.5 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-md text-xs font-medium transition-colors disabled:opacity-60">
        {#if busyAction === 'create-belum_masuk'}
          <Loading variant="button" size="sm" label="..." />
        {:else}
          <i class="fa-solid fa-clock mr-0.5"></i> Belum Masuk
        {/if}
      </button>
      <button onclick={() => createJobs('belum_pulang')} disabled={busyAction}
              class="px-2 py-1.5 border border-gray-200 hover:bg-gray-50 text-gray-700 rounded-md text-xs font-medium transition-colors disabled:opacity-60">
        {#if busyAction === 'create-belum_pulang'}
          <Loading variant="button" size="sm" label="..." />
        {:else}
          <i class="fa-solid fa-clock mr-0.5"></i> Belum Pulang
        {/if}
      </button>
    </div>
  </div>
  
  <!-- Stats -->
  <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 mb-2">
    <div class="bg-white rounded-lg border border-gray-200 p-2 shadow-sm border-l-3 border-l-gray-400">
      <div class="text-sm font-bold">{stats.pending}</div>
      <div class="text-[9px] text-gray-500 uppercase">Pending</div>
    </div>
    <div class="bg-white rounded-lg border border-gray-200 p-2 shadow-sm border-l-3 border-l-cyan-500">
      <div class="text-sm font-bold text-cyan-600">{stats.running}</div>
      <div class="text-[9px] text-gray-500 uppercase">Running</div>
    </div>
    <div class="bg-white rounded-lg border border-gray-200 p-2 shadow-sm border-l-3 border-l-green-500">
      <div class="text-sm font-bold text-green-600">{stats.done}</div>
      <div class="text-[9px] text-gray-500 uppercase">Selesai</div>
    </div>
    <div class="bg-white rounded-lg border border-gray-200 p-2 shadow-sm border-l-3 border-l-red-500">
      <div class="text-sm font-bold text-red-600">{stats.failed}</div>
      <div class="text-[9px] text-gray-500 uppercase">Gagal</div>
    </div>
  </div>
  
  <!-- Retry + Cancel -->
  <div class="grid grid-cols-2 gap-1.5 mb-2">
    <button onclick={retryFailed} disabled={busyAction}
            class="px-2 py-1.5 border border-gray-200 hover:bg-gray-50 rounded-md text-xs font-medium text-gray-700 transition-colors disabled:opacity-50">
      {#if busyAction === 'retry'}
        <Loading variant="button" label="..." />
      {:else}
        <i class="fa-solid fa-rotate mr-0.5"></i> Retry Gagal
      {/if}
    </button>
    <button onclick={cancelAll} disabled={busyAction}
            class="px-2 py-1.5 bg-red-50 hover:bg-red-100 text-red-700 rounded-md text-xs font-medium transition-colors disabled:opacity-50">
      {#if busyAction === 'cancel-all'}
        <Loading variant="button" label="Membatalkan..." />
      {:else}
        <i class="fa-solid fa-circle-xmark mr-0.5"></i> Cancel All
      {/if}
    </button>
  </div>
  
  <!-- Jobs Table -->
  <div class="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
    <div class="px-3 py-1.5 border-b border-gray-200 flex items-center justify-between">
      <span class="text-xs font-semibold text-gray-700"><i class="fa-solid fa-list-check mr-0.5"></i> Antrian Job</span>
      <span class="flex items-center gap-1.5">
        {#if live}
          <span class="inline-flex items-center gap-1 text-[10px] font-medium text-green-700"><span class="w-1.5 h-1.5 rounded-full bg-green-500 animate-pulse"></span>Live</span>
        {:else}
          <span class="inline-flex items-center gap-1 text-[10px] font-medium text-gray-400"><span class="w-1.5 h-1.5 rounded-full bg-gray-300"></span>Poll</span>
        {/if}
        <span class="bg-gray-100 text-gray-600 text-[10px] font-medium px-1.5 py-0.5 rounded-full">{jobs.length}</span>
      </span>
    </div>
    <div class="px-3 py-1.5 border-b border-gray-100 flex gap-1.5">
      <input type="text" bind:value={searchQ} placeholder="Cari nama..."
             class="flex-1 px-2 py-1 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-blue-500 outline-none" />
      <select bind:value={statusFilter}
              class="px-2 py-1 border border-gray-300 rounded-md text-xs bg-white focus:ring-2 focus:ring-blue-500 outline-none">
        <option value="">Semua</option>
        <option value="pending">Pending</option>
        <option value="running">Running</option>
        <option value="done">Selesai</option>
        <option value="failed">Gagal</option>
        <option value="cancelled">Dibatalkan</option>
      </select>
    </div>
    <div class="max-h-[55vh] overflow-y-auto">
      {#if loading}
        <Loading variant="spinner" size="sm" label="Memuat antrian job..." />
      {:else}
        {#each sections as sec (sec.key)}
          <button onclick={() => collapsed[sec.key] = !collapsed[sec.key]}
                  class="w-full flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold {sec.head} border-b border-gray-200 sticky top-0 z-10">
            <i class="{sec.icon} text-[10px]"></i>
            {sec.label}
            <span class="bg-white/70 rounded-full px-1.5 text-[10px]">{sec.count}</span>
            <span class="flex-1"></span>
            <i class="fa-solid {collapsed[sec.key] ? 'fa-chevron-down' : 'fa-chevron-up'} text-[10px] opacity-60"></i>
          </button>
          {#if !collapsed[sec.key]}
            {#each sec.list as j (j.id)}
              <div class="flex items-center gap-2 px-3 py-1.5 border-b border-gray-100 hover:bg-gray-50 text-xs">
                <span class="inline-flex items-center px-1.5 py-0.5 rounded-full text-[10px] font-medium {sec.chip} shrink-0">{j.status}</span>
                <div class="flex-1 min-w-0">
                  <div class="font-medium text-gray-900 truncate">{j.employee_name}</div>
                  {#if j.status === 'running' || j.status === 'pending'}
                    {@const total = j.total_steps || 10}
                    {@const pct = Math.min(100, Math.round((j.progress || 0) / total * 100))}
                    <div class="flex items-center gap-1.5 mt-0.5">
                      <div class="w-16 h-1.5 bg-gray-100 rounded-full overflow-hidden"><div class="h-full rounded-full bg-cyan-500" style="width: {pct}%"></div></div>
                      <span class="text-[9px] text-gray-500">{j.progress || 0}/{total}{#if j.step_label} · {j.step_label}{/if}</span>
                    </div>
                  {:else if j.jam_masuk && j.jam_masuk !== '-'}
                    <div class="text-[10px] text-gray-500">Masuk {j.jam_masuk}{#if j.jam_pulang && j.jam_pulang !== '-'} · Pulang {j.jam_pulang}{/if}</div>
                  {:else if j.error}
                    <div class="text-[10px] text-red-500 truncate">⚠ {j.error}</div>
                  {/if}
                </div>
                <div class="text-right shrink-0">
                  <div class="text-[10px] text-gray-600">{fmtDT(j.created_at)}</div>
                  {#if j.completed_at}
                    <div class="text-[10px] text-green-600">✓ {fmtDT(j.completed_at)}</div>
                  {/if}
                  {#if j.status === 'pending' || j.status === 'running'}
                    <button onclick={() => cancelOne(j.id)} class="mt-0.5 p-0.5 rounded hover:bg-red-50 text-red-500" title="Batalkan">
                      <i class="fa-solid fa-xmark text-[10px]"></i>
                    </button>
                  {/if}
                </div>
              </div>
            {:else}
              <div class="px-3 py-2 text-center text-gray-300 text-[10px]">— kosong —</div>
            {/each}
          {/if}
        {/each}
        <div class="px-3 py-1.5 text-center text-[10px] text-gray-400 bg-gray-50">100 job terbaru · total antrean lihat badge di atas</div>
      {/if}
    </div>
  </div>
</Layout>

<ConfirmDialog show={confirmShow} title={confirmTitle} message={confirmMessage}
               confirmLabel="Ya, Lanjut" variant={confirmVariant}
               onConfirm={confirmAction} onCancel={() => confirmShow = false} />
