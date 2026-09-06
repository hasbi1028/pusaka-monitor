<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import Loading from '$lib/components/Loading.svelte';
  import { dashboard } from '$lib/api.js';
  import { toasts } from '$lib/stores/toast.js';
  import { onMount } from 'svelte';

  const bulanNames = ['Januari','Februari','Maret','April','Mei','Juni','Juli','Agustus','September','Oktober','November','Desember'];

  let selectedBulan = $state(new Date().getMonth() + 1);
  let selectedTahun = $state(new Date().getFullYear());
  let loading = $state(false);
  let filterStatus = $state('all');
  let page = $state(1);
  let perPage = $state(15);

  let hariKerja = $state(0);
  let totalPegawai = $state(0);
  let rekapPegawai = $state([]);
  let detailHarian = $state([]);
  let settings = $state({ jam_masuk: '07:30', jam_pulang: '16:00', toleransi: 15 });
  let summary = $state({ avgHadir: 0, avgTerlambat: 0, avgTidakHadir: 0, avgPersen: 0, rataJamMasuk: '-' });

  // Donut chart
  let donutTotal = $derived(summary.avgHadir + summary.avgTerlambat + summary.avgTidakHadir || 1);
  let donutPctHadir = $derived(summary.avgHadir / donutTotal * 100);
  let donutPctTelat = $derived(summary.avgTerlambat / donutTotal * 100);

  // Filter + paginate (NO heavy compute here)
  let filteredPegawai = $derived.by(() => {
    if (filterStatus === 'terlambat') return rekapPegawai.filter(r => r.terlambat > 0);
    if (filterStatus === 'hadir') return rekapPegawai.filter(r => (parseFloat(r.persen) || 0) >= 80);
    return rekapPegawai;
  });
  let totalPages = $derived(Math.max(1, Math.ceil(filteredPegawai.length / perPage)));
  let pagedPegawai = $derived(filteredPegawai.slice((page - 1) * perPage, page * perPage));

  // Compute avg masuk & durasi for ONE employee (called per visible row only)
  function computeEmployeeAvg(nama) {
    const pegawaiDetail = detailHarian.filter(d => d.nama === nama);
    const jmList = pegawaiDetail
      .filter(d => d.jam_masuk && d.jam_masuk !== '-' && d.jam_masuk.includes(':'))
      .map(d => { const [h,m] = d.jam_masuk.split(':').map(Number); return h*60+m; })
      .filter(v => !isNaN(v) && v > 0);
    const durList = pegawaiDetail
      .filter(d => d.jam_masuk && d.jam_masuk !== '-' && d.jam_pulang && d.jam_pulang !== '-'
                   && d.jam_masuk.includes(':') && d.jam_pulang.includes(':'))
      .map(d => {
        const [h1,m1] = d.jam_masuk.split(':').map(Number);
        const [h2,m2] = d.jam_pulang.split(':').map(Number);
        return (h2*60+m2) - (h1*60+m1);
      })
      .filter(v => !isNaN(v) && v > 0);
    const avgJm = jmList.length > 0
      ? (() => { const a = Math.round(jmList.reduce((s,v)=>s+v,0)/jmList.length); return String(Math.floor(a/60)).padStart(2,'0') + ':' + String(a%60).padStart(2,'0'); })()
      : '-';
    const avgDur = durList.length > 0
      ? (() => { const a = Math.round(durList.reduce((s,v)=>s+v,0)/durList.length); return Math.floor(a/60) + 'j ' + (a%60) + 'm'; })()
      : '-';
    return { avgMasuk: avgJm, avgDurasi: avgDur };
  }

  // Compute rata-rata jam masuk global
  function computeRataJamMasuk(details) {
    const times = details
      .filter(d => d.jam_masuk && d.jam_masuk !== '-' && d.jam_masuk.includes(':'))
      .map(d => { const [h,m] = d.jam_masuk.split(':').map(Number); return h*60+m; })
      .filter(v => !isNaN(v) && v > 0);
    if (times.length === 0) return '-';
    const avg = Math.round(times.reduce((s,v) => s+v, 0) / times.length);
    return String(Math.floor(avg/60)).padStart(2,'0') + ':' + String(avg%60).padStart(2,'0');
  }

  async function loadData() {
    loading = true;
    try {
      const [pegawaiRes, detailRes, settingsRes] = await Promise.all([
        dashboard.getBulanPegawai(selectedBulan, selectedTahun),
        dashboard.getBulanDetail(selectedBulan, selectedTahun),
        fetch('/api/instansi/settings').then(r => r.json())
      ]);

      if (settingsRes.success && settingsRes.data) settings = settingsRes.data;
      hariKerja = pegawaiRes.data?.hari_kerja || 0;
      totalPegawai = pegawaiRes.data?.total_pegawai || 0;
      rekapPegawai = pegawaiRes.data?.rekap_pegawai || [];
      detailHarian = detailRes.data?.detail || [];

      // Pre-compute summary only (lightweight)
      const total = rekapPegawai.length;
      if (total > 0) {
        const tHadir = rekapPegawai.reduce((s, r) => s + (r.hadir || 0), 0);
        const tTelat = rekapPegawai.reduce((s, r) => s + (r.terlambat || 0), 0);
        const tTidak = rekapPegawai.reduce((s, r) => s + (r.tidak_hadir || 0), 0);
        const tPersen = rekapPegawai.reduce((s, r) => s + (parseFloat(r.persen) || 0), 0);
        summary = {
          avgHadir: Math.round(tHadir / total), avgTerlambat: Math.round(tTelat / total),
          avgTidakHadir: Math.round(tTidak / total), avgPersen: tPersen / total,
          rataJamMasuk: computeRataJamMasuk(detailHarian)
        };
      } else {
        summary = { avgHadir: 0, avgTerlambat: 0, avgTidakHadir: 0, avgPersen: 0, rataJamMasuk: '-' };
      }
      page = 1;
    } catch { toasts.error('Gagal memuat data laporan'); }
    finally { loading = false; }
  }

  function exportCSV() { dashboard.exportCSV(selectedBulan, selectedTahun); }
  function statusBadge(s) {
    return { 'Belum Masuk': 'bg-gray-100 text-gray-600', 'Belum Pulang': 'bg-gray-100 text-gray-600',
      'Terlambat': 'bg-red-100 text-red-700', 'Libur': 'bg-blue-50 text-blue-600', 'Cuti': 'bg-purple-50 text-purple-600', 'Telat Ringan': 'bg-yellow-100 text-yellow-700',
      'Tepat Waktu': 'bg-green-100 text-green-700' }[s] || 'bg-gray-100 text-gray-600';
  }
  function persenColor(p) { if (p>=90) return 'bg-green-500'; if (p>=80) return 'bg-blue-500'; if (p>=60) return 'bg-yellow-500'; return 'bg-red-500'; }
  function persenTextColor(p) { if (p>=90) return 'text-green-600'; if (p>=80) return 'text-blue-600'; if (p>=60) return 'text-yellow-600'; return 'text-red-600'; }

  onMount(() => { loadData(); });
</script>

<Layout title="Laporan Kehadiran" activePage="laporan">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Laporan' }]} />
  <!-- Filter -->
  <div class="bg-white rounded-lg border border-gray-200 shadow-sm p-2.5 mb-2 flex flex-wrap gap-2 items-center">
    <div class="flex items-center gap-1.5">
      <label for="bln" class="text-xs text-gray-600">Bulan</label>
      <select id="bln" bind:value={selectedBulan} onchange={() => loadData()} class="px-2 py-1 border border-gray-300 rounded-md text-xs">
        {#each bulanNames as name, i (i)}<option value={i + 1}>{name}</option>{/each}
      </select>
    </div>
    <div class="flex items-center gap-1.5">
      <label for="thn" class="text-xs text-gray-600">Tahun</label>
      <select id="thn" bind:value={selectedTahun} onchange={() => loadData()} class="px-2 py-1 border border-gray-300 rounded-md text-xs">
        {#each [2025, 2026, 2027] as y (y)}<option value={y}>{y}</option>{/each}
      </select>
    </div>
    <div class="flex gap-0.5 bg-gray-100 rounded-md p-0.5">
      {#each [['all','Semua'],['hadir','Hadir ≥80%'],['terlambat','Terlambat']] as [val, label] (val)}
        <button onclick={() => { filterStatus = val; page = 1; }}
                class="px-2 py-0.5 text-[10px] font-medium rounded transition-colors {filterStatus === val ? 'bg-white shadow text-gray-900' : 'text-gray-500 hover:text-gray-700'}">{label}</button>
      {/each}
    </div>
    <div class="flex-1"></div>
    <button onclick={exportCSV} class="px-2 py-1 bg-green-600 text-white text-[10px] font-medium rounded-md hover:bg-green-700">
      <i class="fa-solid fa-download mr-0.5"></i> Export CSV
    </button>
  </div>

  {#if loading}
    <Loading variant="spinner" label="Mengambil data dari server..." />
  {:else}
    <!-- Summary Cards -->
    <div class="grid grid-cols-2 md:grid-cols-5 gap-2 mb-2">
      <div class="bg-white rounded-lg border p-2 text-center"><div class="text-sm font-bold text-blue-600">{hariKerja}</div><div class="text-[9px] text-gray-500">Hari Kerja</div></div>
      <div class="bg-white rounded-lg border p-2 text-center"><div class="text-sm font-bold text-gray-900">{totalPegawai}</div><div class="text-[9px] text-gray-500">Total Pegawai</div></div>
      <div class="bg-white rounded-lg border p-2 text-center"><div class="text-sm font-bold text-green-600">{summary.avgHadir}</div><div class="text-[9px] text-gray-500">Rata-rata Hadir</div></div>
      <div class="bg-white rounded-lg border p-2 text-center"><div class="text-sm font-bold text-yellow-600">{summary.avgTerlambat}</div><div class="text-[9px] text-gray-500">Rata-rata Telat</div></div>
      <div class="bg-white rounded-lg border p-2 text-center"><div class="text-sm font-bold text-red-600">{summary.avgTidakHadir}</div><div class="text-[9px] text-gray-500">Rata-rata Alfa</div></div>
    </div>

    <!-- Donut + Jam Kerja -->
    <div class="grid md:grid-cols-3 gap-2 mb-2">
      <div class="bg-white rounded-lg border p-3">
        <h3 class="text-xs font-semibold text-gray-700 mb-2"><i class="fa-solid fa-chart-pie mr-0.5"></i> Ringkasan Status</h3>
        <div class="flex items-center gap-4">
          <div class="relative w-24 h-24">
            <svg viewBox="0 0 36 36" class="w-24 h-24 -rotate-90">
              <circle cx="18" cy="18" r="15.9" fill="none" stroke="#e5e7eb" stroke-width="3"/>
              <circle cx="18" cy="18" r="15.9" fill="none" stroke="#22c55e" stroke-width="3" stroke-dasharray="{donutPctHadir} {100-donutPctHadir}" stroke-dashoffset="0"/>
              <circle cx="18" cy="18" r="15.9" fill="none" stroke="#eab308" stroke-width="3" stroke-dasharray="{donutPctTelat} {100-donutPctTelat}" stroke-dashoffset="-{donutPctHadir}"/>
            </svg>
            <div class="absolute inset-0 flex items-center justify-center">
              <div class="text-center"><div class="text-sm font-bold text-gray-900">{(summary.avgPersen || 0).toFixed(0)}%</div><div class="text-[8px] text-gray-400 uppercase">Rata-rata</div></div>
            </div>
          </div>
          <div class="space-y-0.5 text-[10px]">
            <div class="flex items-center gap-1.5"><span class="w-2 h-2 rounded-full bg-green-500"></span> Hadir: {summary.avgHadir}</div>
            <div class="flex items-center gap-1.5"><span class="w-2 h-2 rounded-full bg-yellow-500"></span> Terlambat: {summary.avgTerlambat}</div>
            <div class="flex items-center gap-1.5"><span class="w-2 h-2 rounded-full bg-gray-300"></span> Tidak Hadir: {summary.avgTidakHadir}</div>
          </div>
        </div>
      </div>
      <div class="bg-white rounded-lg border p-3">
        <h3 class="text-xs font-semibold text-gray-700 mb-2"><i class="fa-solid fa-clock mr-0.5"></i> Jam Kerja</h3>
        <div class="space-y-1 text-xs">
          <div class="flex justify-between"><span class="text-gray-500">Jam Masuk</span><span class="font-semibold">{settings.jam_masuk || '-'}</span></div>
          <div class="flex justify-between"><span class="text-gray-500">Jam Pulang</span><span class="font-semibold">{settings.jam_pulang || '-'}</span></div>
          <div class="flex justify-between"><span class="text-gray-500">Toleransi</span><span class="font-semibold">{settings.toleransi || 0} mnt</span></div>
          <div class="flex justify-between"><span class="text-gray-500">Mode</span><span class="font-semibold">{settings.mode_ramadan ? '🌙 Ramadan' : '☀️ Normal'}</span></div>
        </div>
      </div>
    </div>

    <!-- Tabel Rekap -->
    <div class="bg-white rounded-lg border shadow-sm">
      <div class="px-3 py-2 border-b border-gray-200 flex items-center justify-between">
        <h3 class="text-xs font-semibold text-gray-700"><i class="fa-solid fa-table mr-0.5"></i> Rekap Per Pegawai <span class="text-gray-400 font-normal">({filteredPegawai.length} pegawai)</span></h3>
        <div class="flex items-center gap-1.5 text-[10px] text-gray-500">
          <span>Baris:</span>
          <select bind:value={perPage} onchange={() => page = 1} class="px-1.5 py-0.5 border rounded text-[10px]">
            <option value={10}>10</option><option value={15}>15</option><option value={25}>25</option><option value={50}>50</option>
          </select>
        </div>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-xs">
          <thead class="bg-gray-50 text-[10px] text-gray-500 uppercase">
            <tr><th class="px-3 py-1.5 text-center w-8">#</th><th class="px-3 py-1.5 text-left">Nama</th><th class="px-3 py-1.5 text-center">Hadir</th><th class="px-3 py-1.5 text-center">Terlambat</th><th class="px-3 py-1.5 text-center">Tidak Hadir</th><th class="px-3 py-1.5 text-center">Persentase</th><th class="px-3 py-1.5 text-center">Rata-rata Masuk</th><th class="px-3 py-1.5 text-center">Rata-rata Durasi</th></tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            {#each pagedPegawai as r, i (r.nip || r.nama)}
              {@const empAvg = computeEmployeeAvg(r.nama)}
              {@const empPersen = parseFloat(r.persen) || 0}
              <tr class="hover:bg-gray-50 transition-colors">
                <td class="px-3 py-1.5 text-gray-400 text-[10px] text-center">{(page - 1) * perPage + i + 1}</td>
                <td class="px-3 py-1.5"><div class="font-medium text-gray-900">{r.nama}</div>{#if r.nip}<div class="text-[10px] text-gray-400">{r.nip}</div>{/if}</td>
                <td class="px-3 py-1.5 text-center"><span class="inline-flex items-center justify-center w-6 h-6 rounded-md bg-green-50 text-green-700 font-semibold text-[10px]">{r.hadir || 0}</span></td>
                <td class="px-3 py-1.5 text-center"><span class="inline-flex items-center justify-center w-6 h-6 rounded-md {r.terlambat > 0 ? 'bg-yellow-50 text-yellow-700' : 'bg-gray-50 text-gray-400'} font-semibold text-[10px]">{r.terlambat || 0}</span></td>
                <td class="px-3 py-1.5 text-center"><span class="inline-flex items-center justify-center w-6 h-6 rounded-md {(r.tidak_hadir || 0) > 0 ? 'bg-red-50 text-red-700' : 'bg-gray-50 text-gray-400'} font-semibold text-[10px]">{r.tidak_hadir || 0}</span></td>
                <td class="px-3 py-1.5 text-center">
                  
                  <div class="flex items-center justify-center gap-1.5">
                    <div class="w-12 h-1.5 bg-gray-100 rounded-full overflow-hidden"><div class="h-full rounded-full {persenColor(empPersen)}" style="width: {Math.min(empPersen, 100)}%"></div></div>
                    <span class="text-[10px] font-bold {persenTextColor(empPersen)} min-w-[30px]">{empPersen.toFixed(0)}%</span>
                  </div>
                </td>
                <td class="px-3 py-1.5 text-center text-gray-600 font-medium">{empAvg.avgMasuk}</td>
                <td class="px-3 py-1.5 text-center text-gray-600 font-medium">{empAvg.avgDurasi}</td>
              </tr>
            {/each}
            {#if pagedPegawai.length === 0}
              <tr><td colspan="8" class="px-3 py-6 text-center text-gray-400">Tidak ada data</td></tr>
            {/if}
          </tbody>
        </table>
      </div>
      {#if totalPages > 1}
        <div class="px-3 py-2 border-t border-gray-200 flex items-center justify-between text-[10px] text-gray-500">
          <span>Halaman {page} dari {totalPages} ({filteredPegawai.length} data)</span>
          <div class="flex gap-0.5">
            <button onclick={() => page = Math.max(1, page - 1)} disabled={page <= 1} class="px-2 py-0.5 border rounded-md hover:bg-gray-50 disabled:opacity-40">‹ Prev</button>
            {#each Array.from({length: Math.min(totalPages, 7)}, (_, i) => { let p; if (totalPages <= 7) p = i + 1; else if (page <= 4) p = i + 1; else if (page >= totalPages - 3) p = totalPages - 6 + i; else p = page - 3 + i; return p; }) as pn (pn)}
              <button onclick={() => page = pn} class="px-2 py-0.5 border rounded-md {pn === page ? 'bg-blue-600 text-white border-blue-600' : 'hover:bg-gray-50'}">{pn}</button>
            {/each}
            <button onclick={() => page = Math.min(totalPages, page + 1)} disabled={page >= totalPages} class="px-2 py-0.5 border rounded-md hover:bg-gray-50 disabled:opacity-40">Next ›</button>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</Layout>
