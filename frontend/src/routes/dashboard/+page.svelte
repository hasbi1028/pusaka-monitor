<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import Loading from '$lib/components/Loading.svelte';
  import ErrorState from '$lib/components/ErrorState.svelte';
  import { dashboard, rekap as rekapApi, invalidateCache, prefetch } from '$lib/api.js';
  import { toasts } from '$lib/stores/toast.js';

  const bulanNames = ['Januari','Februari','Maret','April','Mei','Juni','Juli','Agustus','September','Oktober','November','Desember'];
  let loading = $state(true);
  let loadError = $state(false);
  let bulanLoading = $state(false);
  let bulanError = $state(false);
  let kirimLoading = $state(false);
  let kirimVia = $state('');

  let mode = $state('harian');
  let today = $state(new Date().toLocaleDateString('en-CA'));
  let selectedBulan = $state(new Date().getMonth() + 1);
  let selectedTahun = $state(new Date().getFullYear());

  // Dynamic year range — current year ± 1
  const thisYear = new Date().getFullYear();
  const yearRange = [thisYear - 1, thisYear, thisYear + 1];

  // Harian
  let rekap = $state({ total: 0, hadir: 0, terlambat: 0, belum_masuk: 0, belum_pulang: 0, tidak_hadir: 0 });
  let absensi = $state([]);

  // Bulanan
  let bulanMode = $state('rekap');
  let hariKerja = $state(0);
  let rekapPegawai = $state([]);
  let detailHarian = $state([]);

  // Virtual scroll for detail harian
  let vsScrollTop = $state(0);
  const VS_ROW_H = 28;
  const VS_BUFFER = 5;
  let vsStart = $derived(Math.max(0, Math.floor(vsScrollTop / VS_ROW_H) - VS_BUFFER));
  let vsVisible = $derived(Math.min(detailHarian.length, vsStart + Math.ceil(500 / VS_ROW_H) + VS_BUFFER * 2));
  let vsItems = $derived(detailHarian.slice(vsStart, vsVisible));
  let vsTotalH = $derived(detailHarian.length * VS_ROW_H);
  let vsOffsetY = $derived(vsStart * VS_ROW_H);

  function statusBadge(s) {
    const m = {
      'Belum Masuk': 'bg-gray-100 text-gray-600',
      'Belum Pulang': 'bg-gray-100 text-gray-600',
      'Terlambat': 'bg-red-100 text-red-700', 'Libur': 'bg-blue-50 text-blue-600', 'Cuti': 'bg-purple-50 text-purple-600',
      'Telat Ringan': 'bg-yellow-100 text-yellow-700', 'Tepat Waktu': 'bg-green-100 text-green-700',
      'Hadir': 'bg-green-100 text-green-700'
    };
    return m[s] || 'bg-gray-100 text-gray-600';
  }

  async function loadDashboard() {
    loadError = false;
    try {
      const res = await dashboard.get(today);
      if (res.success) {
        rekap = res.data.rekap;
        absensi = res.data.absensi;
      } else {
        loadError = true;
      }
    } catch {
      loadError = true;
    }
    loading = false;
  }

  async function loadBulanPegawai() {
    bulanError = false;
    try {
      const res = await dashboard.getBulanPegawai(selectedBulan, selectedTahun);
      if (res.success) {
        hariKerja = res.data.hari_kerja || 0;
        rekapPegawai = res.data.rekap_pegawai || [];
      } else {
        bulanError = true;
      }
    } catch {
      bulanError = true;
    }
  }

  async function loadBulanDetail() {
    try {
      const res = await dashboard.getBulanDetail(selectedBulan, selectedTahun);
      if (res.success) {
        detailHarian = res.data.detail || [];
      }
    } catch {
      // silently fail for detail
    }
  }

  async function loadBulan() {
    bulanLoading = true;
    await Promise.all([loadBulanPegawai(), loadBulanDetail()]);
    bulanLoading = false;
  }

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

  async function kirimRekap(via) {
    kirimLoading = true;
    kirimVia = via;
    try {
      const fn = via === 'wa' ? rekapApi.kirimWA : rekapApi.kirimTelegram;
      const res = await fn(today);
      if (res.success) {
        toasts.success('Terkirim via ' + (via === 'wa' ? 'WhatsApp' : 'Telegram') + '!');
      } else {
        toasts.error(res.error || 'Gagal mengirim');
      }
    } catch (e) {
      toasts.error('Error: ' + e.message);
    }
    kirimLoading = false;
    kirimVia = '';
  }

  // Init
  loadDashboard();

  // Prefetch likely next-navigate data during idle time
  prefetch(['/api/scrape/status', '/api/settings']);
</script>

<Layout title="Dashboard" activePage="dashboard">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Dashboard' }]} />

  {#if loading}
    <Loading variant="skeleton" rows={5} cols={6} label="" />
  {:else if loadError}
    <ErrorState onRetry={loadDashboard} />
  {:else}
    <!-- Marquee Banner -->
    <div class="bg-green-50 border border-green-200 rounded-md px-2.5 py-1.5 mb-2 text-[10px] text-green-800">
      <i class="fa-solid fa-shield-halved mr-0.5"></i>
      <strong>READ-ONLY:</strong> hanya membaca riwayat Pusaka Kemenag — tidak membuat/mengubah absen.
      <a href="/kepatuhan" class="underline ml-1">Kepatuhan</a>
    </div>
    
    <!-- Stat Cards -->
    <div class="grid grid-cols-3 sm:grid-cols-3 lg:grid-cols-6 gap-2 mb-2">
      <div class="bg-white rounded-lg border border-gray-200 p-2 shadow-sm border-l-3 border-l-gray-400">
        <div class="text-base font-bold text-gray-900">{rekap.total}</div>
        <div class="text-[9px] text-gray-500 uppercase tracking-wide">Total</div>
      </div>
      <div class="bg-white rounded-lg border border-gray-200 p-2 shadow-sm border-l-3 border-l-green-500">
        <div class="text-base font-bold text-green-600">{rekap.hadir}</div>
        <div class="text-[9px] text-gray-500 uppercase tracking-wide">Hadir</div>
      </div>
      <div class="bg-white rounded-lg border border-gray-200 p-2 shadow-sm border-l-3 border-l-yellow-500">
        <div class="text-base font-bold text-yellow-600">{rekap.terlambat}</div>
        <div class="text-[9px] text-gray-500 uppercase tracking-wide">Telat</div>
      </div>
      <div class="bg-white rounded-lg border border-gray-200 p-2 shadow-sm border-l-3 border-l-orange-500">
        <div class="text-base font-bold text-orange-600">{rekap.belum_masuk}</div>
        <div class="text-[9px] text-gray-500 uppercase tracking-wide">Blm Masuk</div>
      </div>
      <div class="bg-white rounded-lg border border-gray-200 p-2 shadow-sm border-l-3 border-l-cyan-500">
        <div class="text-base font-bold text-cyan-600">{rekap.belum_pulang}</div>
        <div class="text-[9px] text-gray-500 uppercase tracking-wide">Blm Pulang</div>
      </div>
      <div class="bg-white rounded-lg border border-gray-200 p-2 shadow-sm border-l-3 border-l-red-500">
        <div class="text-base font-bold text-red-600">{rekap.tidak_hadir}</div>
        <div class="text-[9px] text-gray-500 uppercase tracking-wide">Alfa</div>
      </div>
    </div>
    
    <!-- Tab Switcher -->
    <div class="flex gap-1.5 mb-2">
      <button onclick={() => setMode('harian')}
              class="px-3 py-1 rounded-md text-xs font-medium transition-colors"
              class:bg-blue-600={mode === 'harian'} class:text-white={mode === 'harian'}
              class:bg-white={mode !== 'harian'} class:text-gray-600={mode !== 'harian'}
              class:border={mode !== 'harian'} class:border-gray-200={mode !== 'harian'}>
        <i class="fa-solid fa-calendar-day mr-0.5"></i>Harian
      </button>
      <button onclick={() => setMode('bulanan')}
              class="px-3 py-1 rounded-md text-xs font-medium transition-colors"
              class:bg-blue-600={mode === 'bulanan'} class:text-white={mode === 'bulanan'}
              class:bg-white={mode !== 'bulanan'} class:text-gray-600={mode !== 'bulanan'}
              class:border={mode !== 'bulanan'} class:border-gray-200={mode !== 'bulanan'}>
        <i class="fa-solid fa-calendar mr-0.5"></i>Bulanan
      </button>
    </div>
    
    {#if mode === 'harian'}
      <!-- Harian View -->
      <div class="bg-white rounded-lg border border-gray-200 shadow-sm p-2.5 mb-2">
        <div class="flex items-center gap-1.5 flex-wrap">
          <label for="dt" class="text-xs font-medium text-gray-700">Tanggal:</label>
          <input id="dt" type="date" bind:value={today} onchange={loadDashboard}
                 class="px-2 py-1 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-blue-500 outline-none" />
          <button onclick={loadDashboard} aria-label="Cari"
                  class="px-2 py-1 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-xs transition-colors">
            <i class="fa-solid fa-magnifying-glass text-[10px]"></i>
          </button>
        </div>
      </div>
      <!-- Kirim Rekap Buttons -->
      <div class="bg-white rounded-lg border border-gray-200 shadow-sm p-2.5 mb-2">
        <div class="flex items-center gap-2 flex-wrap">
          <span class="text-xs font-medium text-gray-700"><i class="fa-solid fa-paper-plane mr-0.5"></i> Kirim Rekap:</span>
          <button onclick={() => kirimRekap('wa')}
                  disabled={kirimLoading}
                  class="px-3 py-1.5 bg-green-600 hover:bg-green-700 disabled:bg-green-300 text-white rounded-md text-xs font-medium transition-colors flex items-center gap-1">
            {#if kirimLoading && kirimVia === 'wa'}
              <div class="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
              Mengirim...
            {:else}
              <i class="fa-brands fa-whatsapp text-sm"></i> WhatsApp
            {/if}
          </button>
          <button onclick={() => kirimRekap('telegram')}
                  disabled={kirimLoading}
                  class="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white rounded-md text-xs font-medium transition-colors flex items-center gap-1">
            {#if kirimLoading && kirimVia === 'telegram'}
              <div class="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
              Mengirim...
            {:else}
              <i class="fa-brands fa-telegram text-sm"></i> Telegram
            {/if}
          </button>
        </div>
      </div>
      <div class="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
        <div class="px-3 py-2 border-b border-gray-200 flex items-center justify-between">
          <span class="text-xs font-semibold text-gray-700"><i class="fa-solid fa-table mr-0.5"></i> Presensi</span>
          <span class="bg-gray-100 text-gray-600 text-[10px] font-medium px-1.5 py-0.5 rounded-full">{absensi.length}</span>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-xs">
            <thead class="bg-gray-50 text-gray-500 text-[10px] uppercase">
              <tr><th class="px-3 py-1.5 text-left">Nama</th><th class="px-3 py-1.5 text-center">Masuk</th><th class="px-3 py-1.5 text-center">Pulang</th><th class="px-3 py-1.5 text-center">Status</th></tr>
            </thead>
            <tbody class="divide-y divide-gray-100">
              {#if absensi.length === 0}
                <tr><td colspan="4" class="px-3 py-4 text-center text-gray-400 text-xs">Belum ada data</td></tr>
              {:else}
                {#each absensi as a (a.id)}
                  <tr class="hover:bg-gray-50">
                    <td class="px-3 py-1.5 font-medium text-gray-900">{a.nama}</td>
                    <td class="px-3 py-1.5 text-center text-gray-600">{a.jam_masuk || '-'}</td>
                    <td class="px-3 py-1.5 text-center text-gray-600">{a.jam_pulang || '-'}</td>
                    <td class="px-3 py-1.5 text-center">
                      <span class="inline-flex items-center px-1.5 py-0.5 rounded-full text-[10px] font-medium {statusBadge(a.status)}">
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
      <div class="relative">
        {#if bulanLoading}
          <Loading variant="skeleton" rows={5} cols={3} label="" />
        {:else if bulanError}
          <ErrorState message="Gagal memuat data bulanan" onRetry={loadBulan} />
        {:else}
          <div class="bg-white rounded-lg border border-gray-200 shadow-sm p-2.5 mb-2">
            <div class="flex items-center gap-1.5 flex-wrap">
              <label for="dbulan" class="text-xs font-medium text-gray-700">Bulan:</label>
              <select id="dbulan" bind:value={selectedBulan} onchange={loadBulan}
                      class="px-2 py-1 border border-gray-300 rounded-md text-xs bg-white focus:ring-2 focus:ring-blue-500 outline-none">
                {#each Array.from({length: 12}, (_, i) => i + 1) as b (b)}
                  <option value={b}>{bulanNames[b - 1]}</option>
                {/each}
              </select>
              <label for="dtahun" class="text-xs font-medium text-gray-700">Tahun:</label>
              <select id="dtahun" bind:value={selectedTahun} onchange={loadBulan}
                      class="px-2 py-1 border border-gray-300 rounded-md text-xs bg-white focus:ring-2 focus:ring-blue-500 outline-none">
                {#each yearRange as y (y)}
                  <option value={y}>{y}</option>
                {/each}
              </select>
              <button onclick={exportCSV}
                      class="px-2 py-1 border border-gray-300 rounded-md text-xs hover:bg-gray-50 transition-colors">
                <i class="fa-solid fa-download mr-0.5"></i>CSV
              </button>
            </div>
          </div>
          
          <!-- Sub-tabs -->
          <div class="flex gap-1.5 mb-2">
            <button onclick={() => setBulanMode('rekap')}
                    class="px-3 py-1 rounded-md text-xs font-medium transition-colors"
                    class:bg-blue-600={bulanMode === 'rekap'} class:text-white={bulanMode === 'rekap'}
                    class:bg-white={bulanMode !== 'rekap'} class:text-gray-600={bulanMode !== 'rekap'}
                    class:border={bulanMode !== 'rekap'} class:border-gray-200={bulanMode !== 'rekap'}>
              Rekap Pegawai
            </button>
            <button onclick={() => setBulanMode('detail')}
                    class="px-3 py-1 rounded-md text-xs font-medium transition-colors"
                    class:bg-blue-600={bulanMode === 'detail'} class:text-white={bulanMode === 'detail'}
                    class:bg-white={bulanMode !== 'detail'} class:text-gray-600={bulanMode !== 'detail'}
                    class:border={bulanMode !== 'detail'} class:border-gray-200={bulanMode !== 'detail'}>
              Detail Harian
            </button>
          </div>
          
          {#if bulanMode === 'rekap'}
            <div class="grid grid-cols-3 gap-2 mb-2">
              <div class="bg-white rounded-lg border border-gray-200 shadow-sm p-2 text-center">
                <div class="text-sm font-bold">{hariKerja}</div>
                <div class="text-[9px] text-gray-500 uppercase">Hari Kerja</div>
              </div>
              <div class="bg-white rounded-lg border border-gray-200 shadow-sm p-2 text-center">
                <div class="text-sm font-bold text-green-600">
                  {#if rekapPegawai.length > 0}
                    {@const totalHadir = rekapPegawai.reduce((s, r) => s + r.hadir, 0)}
                    {@const totalHK = hariKerja * rekapPegawai.length}
                    {totalHK > 0 ? Math.round(totalHadir / totalHK * 100) : 0}%
                  {:else}0%{/if}
                </div>
                <div class="text-[9px] text-gray-500 uppercase">% Hadir</div>
              </div>
              <div class="bg-white rounded-lg border border-gray-200 shadow-sm p-2 text-center">
                <div class="text-sm font-bold">{rekapPegawai.length}</div>
                <div class="text-[9px] text-gray-500 uppercase">Pegawai</div>
              </div>
            </div>
            <div class="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
              <div class="px-3 py-2 border-b border-gray-200 text-xs font-semibold text-gray-700">
                <i class="fa-solid fa-users mr-0.5"></i> Rekap per Pegawai
              </div>
              <div class="overflow-x-auto max-h-[50vh] overflow-y-auto">
                <table class="w-full text-xs">
                  <thead class="bg-gray-50 text-gray-500 text-[10px] uppercase sticky top-0">
                    <tr><th class="px-3 py-1.5 text-left">Nama</th><th class="px-3 py-1.5 text-center">Hadir</th><th class="px-3 py-1.5 text-center">Telat</th><th class="px-3 py-1.5 text-center">Alfa</th><th class="px-3 py-1.5 text-center">%</th></tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100">
                    {#each rekapPegawai as r (r.nama)}
                      <tr class="hover:bg-gray-50">
                        <td class="px-3 py-1.5 font-medium text-gray-900">{r.nama}</td>
                        <td class="px-3 py-1.5 text-center text-green-600 font-medium">{r.hadir}</td>
                        <td class="px-3 py-1.5 text-center text-yellow-600">{r.terlambat}</td>
                        <td class="px-3 py-1.5 text-center text-red-600">{r.tidak_hadir}</td>
                        <td class="px-3 py-1.5 text-center font-bold" class:text-green-600={r.persen >= 80} class:text-red-600={r.persen < 80}>
                          {r.persen.toFixed(0)}%
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>
          {:else}
            <div class="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
              <div class="px-3 py-2 border-b border-gray-200 text-xs font-semibold text-gray-700 flex items-center justify-between">
                <span><i class="fa-solid fa-list-ul mr-0.5"></i> Detail Harian</span>
                <span class="text-[10px] font-normal text-gray-400">{detailHarian.length} baris</span>
              </div>
              {#if detailHarian.length === 0}
                <div class="px-3 py-4 text-center text-gray-400 text-xs">Tidak ada data</div>
              {:else}
                <div class="overflow-y-auto" style="height: min(55vh, 500px)" onscroll={(e) => vsScrollTop = e.target.scrollTop}>
                  <div style="height: {vsTotalH}px; position: relative;">
                    <div style="transform: translateY({vsOffsetY}px);">
                      {#each vsItems as a (a.id || (a.tanggal + a.nama))}
                        <div class="flex items-center border-b border-gray-50 hover:bg-gray-50 px-3 h-[28px]">
                          <span class="w-[18%] text-gray-600 shrink-0 text-[10px]">{a.tanggal}</span>
                          <span class="w-[28%] font-medium text-gray-900 shrink-0 truncate text-[10px]">{a.nama}</span>
                          <span class="w-[16%] text-center text-gray-600 shrink-0 text-[10px]">{a.jam_masuk || '-'}</span>
                          <span class="w-[16%] text-center text-gray-600 shrink-0 text-[10px]">{a.jam_pulang || '-'}</span>
                          <span class="w-[22%] text-center shrink-0">
                            <span class="inline-flex items-center px-1.5 py-0.5 rounded-full text-[10px] font-medium {statusBadge(a.status)}">
                              {a.status || '-'}
                            </span>
                          </span>
                        </div>
                      {/each}
                    </div>
                  </div>
                </div>
                <div class="px-3 py-1 text-center text-[10px] text-gray-400 bg-gray-50 border-t border-gray-100">
                  Menampilkan {vsStart + 1}–{Math.min(vsVisible, detailHarian.length)} dari {detailHarian.length}
                </div>
              {/if}
            </div>
          {/if}
        {/if}
      </div>
    {/if}
  {/if}
</Layout>
