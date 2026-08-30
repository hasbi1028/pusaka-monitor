<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import { dashboard } from '$lib/api.js';
  import { toasts } from '$lib/stores/toast.js';

  const bulanNames = ['Januari','Februari','Maret','April','Mei','Juni','Juli','Agustus','September','Oktober','November','Desember'];
  let loading = $state(true);

  let mode = $state('harian');
  let today = $state(new Date().toISOString().split('T')[0]);
  let selectedBulan = $state(new Date().getMonth() + 1);
  let selectedTahun = $state(2026);

  // Harian
  let rekap = $state({ total: 0, hadir: 0, terlambat: 0, belum_masuk: 0, belum_pulang: 0, tidak_hadir: 0 });
  let absensi = $state([]);

  // Bulanan
  let bulanMode = $state('rekap');
  let hariKerja = $state(0);
  let rekapPegawai = $state([]);
  let detailHarian = $state([]);

  function statusBadge(s) {
    const m = {
      'Belum Masuk': 'bg-gray-100 text-gray-600',
      'Belum Pulang': 'bg-gray-100 text-gray-600',
      'Terlambat': 'bg-red-100 text-red-700', 'Libur': 'bg-blue-50 text-blue-600', 'Cuti': 'bg-purple-50 text-purple-600',
      'Telat Ringan': 'bg-yellow-100 text-yellow-700',
      'Tepat Waktu': 'bg-green-100 text-green-700',
      'Hadir': 'bg-green-100 text-green-700'
    };
    return m[s] || 'bg-gray-100 text-gray-600';
  }

  async function loadDashboard() {
    const res = await dashboard.get(today);
    if (res.success) {
      rekap = res.data.rekap;
      absensi = res.data.absensi;
    }
    loading = false;
  }

  async function loadBulanPegawai() {
    const res = await dashboard.getBulanPegawai(selectedBulan, selectedTahun);
    if (res.success) {
      hariKerja = res.data.hari_kerja || 0;
      rekapPegawai = res.data.rekap_pegawai || [];
    }
  }

  async function loadBulanDetail() {
    const res = await dashboard.getBulanDetail(selectedBulan, selectedTahun);
    if (res.success) {
      detailHarian = res.data.detail || [];
    }
  }

  function loadBulan() { loadBulanPegawai(); loadBulanDetail(); }

  function setMode(m) {
    mode = m;
    if (m === 'bulanan') loadBulan();
  }

  function setBulanMode(m) {
    bulanMode = m;
    if (m === 'detail') loadBulanDetail();
    else loadBulanPegawai();
  }

  function exportCSV() { dashboard.exportCSV(selectedBulan, selectedTahun); }

  // Init
  loadDashboard();
</script>

<Layout title="Dashboard" activePage="dashboard">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Dashboard' }]} />

  {#if loading}
    <!-- Loading Skeleton -->
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-4">
      {#each Array(6) as _}
        <div class="bg-white rounded-xl border border-gray-200 p-3 shadow-sm animate-pulse">
          <div class="h-6 bg-gray-200 rounded w-12 mb-2"></div>
          <div class="h-3 bg-gray-100 rounded w-16"></div>
        </div>
      {/each}
    </div>
    <div class="bg-white rounded-xl border border-gray-200 shadow-sm p-4 animate-pulse">
      <div class="h-4 bg-gray-200 rounded w-32 mb-4"></div>
      <div class="space-y-3">
        {#each Array(5) as _}
          <div class="flex gap-4">
            <div class="h-4 bg-gray-100 rounded flex-1"></div>
            <div class="h-4 bg-gray-100 rounded w-16"></div>
            <div class="h-4 bg-gray-100 rounded w-16"></div>
            <div class="h-4 bg-gray-100 rounded w-20"></div>
          </div>
        {/each}
      </div>
    </div>
  {:else}
  <!-- Marquee Banner -->
  <div class="bg-blue-50 border border-blue-200 rounded-lg px-4 py-2 mb-4 text-xs text-blue-700">
    <i class="fa-solid fa-circle-info mr-1"></i>
    <strong>HANYA membaca</strong> riwayat presensi dari Pusaka Kemenag — tidak mengubah atau membuat data.
  </div>
  
  <!-- Stat Cards -->
  <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-4">
    <div class="bg-white rounded-xl border border-gray-200 p-3 shadow-sm border-l-4 border-l-gray-400">
      <div class="text-xl font-bold text-gray-900">{rekap.total}</div>
      <div class="text-[10px] text-gray-500 uppercase tracking-wide">Total</div>
    </div>
    <div class="bg-white rounded-xl border border-gray-200 p-3 shadow-sm border-l-4 border-l-green-500">
      <div class="text-xl font-bold text-green-600">{rekap.hadir}</div>
      <div class="text-[10px] text-gray-500 uppercase tracking-wide">Hadir</div>
    </div>
    <div class="bg-white rounded-xl border border-gray-200 p-3 shadow-sm border-l-4 border-l-yellow-500">
      <div class="text-xl font-bold text-yellow-600">{rekap.terlambat}</div>
      <div class="text-[10px] text-gray-500 uppercase tracking-wide">Telat</div>
    </div>
    <div class="bg-white rounded-xl border border-gray-200 p-3 shadow-sm border-l-4 border-l-orange-500">
      <div class="text-xl font-bold text-orange-600">{rekap.belum_masuk}</div>
      <div class="text-[10px] text-gray-500 uppercase tracking-wide">Blm Masuk</div>
    </div>
    <div class="bg-white rounded-xl border border-gray-200 p-3 shadow-sm border-l-4 border-l-cyan-500">
      <div class="text-xl font-bold text-cyan-600">{rekap.belum_pulang}</div>
      <div class="text-[10px] text-gray-500 uppercase tracking-wide">Blm Pulang</div>
    </div>
    <div class="bg-white rounded-xl border border-gray-200 p-3 shadow-sm border-l-4 border-l-red-500">
      <div class="text-xl font-bold text-red-600">{rekap.tidak_hadir}</div>
      <div class="text-[10px] text-gray-500 uppercase tracking-wide">Alfa</div>
    </div>
  </div>
  
  <!-- Tab Switcher -->
  <div class="flex gap-2 mb-4">
    <button onclick={() => setMode('harian')}
            class="px-4 py-1.5 rounded-lg text-sm font-medium transition-colors"
            class:bg-blue-600={mode === 'harian'} class:text-white={mode === 'harian'}
            class:bg-white={mode !== 'harian'} class:text-gray-600={mode !== 'harian'}
            class:border={mode !== 'harian'} class:border-gray-200={mode !== 'harian'}>
      <i class="fa-solid fa-calendar-day mr-1"></i>Harian
    </button>
    <button onclick={() => setMode('bulanan')}
            class="px-4 py-1.5 rounded-lg text-sm font-medium transition-colors"
            class:bg-blue-600={mode === 'bulanan'} class:text-white={mode === 'bulanan'}
            class:bg-white={mode !== 'bulanan'} class:text-gray-600={mode !== 'bulanan'}
            class:border={mode !== 'bulanan'} class:border-gray-200={mode !== 'bulanan'}>
      <i class="fa-solid fa-calendar mr-1"></i>Bulanan
    </button>
  </div>
  
  {#if mode === 'harian'}
    <!-- Harian View -->
    <div class="bg-white rounded-xl border border-gray-200 shadow-sm p-4 mb-4">
      <div class="flex items-center gap-2 flex-wrap">
        <label class="text-sm font-medium text-gray-700">Tanggal:</label>
        <input type="date" bind:value={today} onchange={loadDashboard}
               class="px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
        <button onclick={loadDashboard}
                class="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm transition-colors">
          <i class="fa-solid fa-magnifying-glass"></i>
        </button>
      </div>
    </div>
    <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
      <div class="px-4 py-2.5 border-b border-gray-200 flex items-center justify-between">
        <span class="text-sm font-semibold text-gray-700"><i class="fa-solid fa-table mr-1"></i> Presensi</span>
        <span class="bg-gray-100 text-gray-600 text-xs font-medium px-2 py-0.5 rounded-full">{absensi.length}</span>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-gray-50 text-gray-500 text-xs uppercase">
            <tr><th class="px-4 py-2 text-left">Nama</th><th class="px-4 py-2 text-center">Masuk</th><th class="px-4 py-2 text-center">Pulang</th><th class="px-4 py-2 text-center">Status</th></tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            {#if absensi.length === 0}
              <tr><td colspan="4" class="px-4 py-6 text-center text-gray-400 text-sm">Belum ada data</td></tr>
            {:else}
              {#each absensi as a}
                <tr class="hover:bg-gray-50">
                  <td class="px-4 py-2 font-medium text-gray-900">{a.nama}</td>
                  <td class="px-4 py-2 text-center text-gray-600">{a.jam_masuk || '-'}</td>
                  <td class="px-4 py-2 text-center text-gray-600">{a.jam_pulang || '-'}</td>
                  <td class="px-4 py-2 text-center">
                    <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium {statusBadge(a.status)}">
                      {a.status || (a.jam_masuk ? 'Hadir' : 'Alfa')}
                    </span>
                  </td>
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      </div>
    </div>
  {:else}
    <!-- Bulanan View -->
    <div class="bg-white rounded-xl border border-gray-200 shadow-sm p-4 mb-4">
      <div class="flex items-center gap-2 flex-wrap">
        <label class="text-sm font-medium text-gray-700">Bulan:</label>
        <select bind:value={selectedBulan} onchange={loadBulan}
                class="px-3 py-1.5 border border-gray-300 rounded-lg text-sm bg-white focus:ring-2 focus:ring-blue-500 outline-none">
          {#each Array.from({length: 12}, (_, i) => i + 1) as b}
            <option value={b}>{bulanNames[b - 1]}</option>
          {/each}
        </select>
        <label class="text-sm font-medium text-gray-700">Tahun:</label>
        <select bind:value={selectedTahun} onchange={loadBulan}
                class="px-3 py-1.5 border border-gray-300 rounded-lg text-sm bg-white focus:ring-2 focus:ring-blue-500 outline-none">
          <option value={2026}>2026</option>
          <option value={2025}>2025</option>
        </select>
        <button onclick={exportCSV}
                class="px-3 py-1.5 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 transition-colors">
          <i class="fa-solid fa-download mr-1"></i>CSV
        </button>
      </div>
    </div>
    
    <!-- Sub-tabs -->
    <div class="flex gap-2 mb-4">
      <button onclick={() => setBulanMode('rekap')}
              class="px-4 py-1.5 rounded-lg text-sm font-medium transition-colors"
              class:bg-blue-600={bulanMode === 'rekap'} class:text-white={bulanMode === 'rekap'}
              class:bg-white={bulanMode !== 'rekap'} class:text-gray-600={bulanMode !== 'rekap'}
              class:border={bulanMode !== 'rekap'} class:border-gray-200={bulanMode !== 'rekap'}>
        Rekap Pegawai
      </button>
      <button onclick={() => setBulanMode('detail')}
              class="px-4 py-1.5 rounded-lg text-sm font-medium transition-colors"
              class:bg-blue-600={bulanMode === 'detail'} class:text-white={bulanMode === 'detail'}
              class:bg-white={bulanMode !== 'detail'} class:text-gray-600={bulanMode !== 'detail'}
              class:border={bulanMode !== 'detail'} class:border-gray-200={bulanMode !== 'detail'}>
        Detail Harian
      </button>
    </div>
    
    {#if bulanMode === 'rekap'}
      <div class="grid grid-cols-3 gap-3 mb-4">
        <div class="bg-white rounded-xl border border-gray-200 shadow-sm p-3 text-center">
          <div class="text-lg font-bold">{hariKerja}</div>
          <div class="text-[10px] text-gray-500 uppercase">Hari Kerja</div>
        </div>
        <div class="bg-white rounded-xl border border-gray-200 shadow-sm p-3 text-center">
          <div class="text-lg font-bold text-green-600">
            {#if rekapPegawai.length > 0}
              {@const totalHadir = rekapPegawai.reduce((s, r) => s + r.hadir, 0)}
              {@const totalHK = hariKerja * rekapPegawai.length}
              {totalHK > 0 ? Math.round(totalHadir / totalHK * 100) : 0}%
            {:else}0%{/if}
          </div>
          <div class="text-[10px] text-gray-500 uppercase">% Hadir</div>
        </div>
        <div class="bg-white rounded-xl border border-gray-200 shadow-sm p-3 text-center">
          <div class="text-lg font-bold">{rekapPegawai.length}</div>
          <div class="text-[10px] text-gray-500 uppercase">Pegawai</div>
        </div>
      </div>
      <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
        <div class="px-4 py-2.5 border-b border-gray-200 text-sm font-semibold text-gray-700">
          <i class="fa-solid fa-users mr-1"></i> Rekap per Pegawai
        </div>
        <div class="overflow-x-auto max-h-[50vh] overflow-y-auto">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 text-gray-500 text-xs uppercase sticky top-0">
              <tr><th class="px-4 py-2 text-left">Nama</th><th class="px-4 py-2 text-center">Hadir</th><th class="px-4 py-2 text-center">Telat</th><th class="px-4 py-2 text-center">Alfa</th><th class="px-4 py-2 text-center">%</th></tr>
            </thead>
            <tbody class="divide-y divide-gray-100">
              {#each rekapPegawai as r}
                <tr class="hover:bg-gray-50">
                  <td class="px-4 py-2 font-medium text-gray-900">{r.nama}</td>
                  <td class="px-4 py-2 text-center text-green-600 font-medium">{r.hadir}</td>
                  <td class="px-4 py-2 text-center text-yellow-600">{r.terlambat}</td>
                  <td class="px-4 py-2 text-center text-red-600">{r.tidak_hadir}</td>
                  <td class="px-4 py-2 text-center font-bold" class:text-green-600={r.persen >= 80} class:text-red-600={r.persen < 80}>
                    {r.persen.toFixed(0)}%
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    {:else}
      <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
        <div class="px-4 py-2.5 border-b border-gray-200 text-sm font-semibold text-gray-700">
          <i class="fa-solid fa-list-ul mr-1"></i> Detail Harian
        </div>
        <div class="overflow-x-auto max-h-[55vh] overflow-y-auto">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 text-gray-500 text-xs uppercase sticky top-0">
              <tr><th class="px-4 py-2 text-left">Tanggal</th><th class="px-4 py-2 text-left">Nama</th><th class="px-4 py-2 text-center">Masuk</th><th class="px-4 py-2 text-center">Pulang</th><th class="px-4 py-2 text-center">Status</th></tr>
            </thead>
            <tbody class="divide-y divide-gray-100">
              {#each detailHarian as a}
                <tr class="hover:bg-gray-50">
                  <td class="px-4 py-2 text-gray-600">{a.tanggal}</td>
                  <td class="px-4 py-2 font-medium text-gray-900">{a.nama}</td>
                  <td class="px-4 py-2 text-center text-gray-600">{a.jam_masuk || '-'}</td>
                  <td class="px-4 py-2 text-center text-gray-600">{a.jam_pulang || '-'}</td>
                  <td class="px-4 py-2 text-center">
                    <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium {statusBadge(a.status)}">
                      {a.status || '-'}
                    </span>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    {/if}
  {/if}
  {/if}
</Layout>
