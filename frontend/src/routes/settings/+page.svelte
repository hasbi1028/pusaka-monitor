<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import Loading from '$lib/components/Loading.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import { superadmin, instansi, waGroups, libur as liburApi } from '$lib/api.js';
  import { toasts } from '$lib/stores/toast.js';
  import { onMount } from 'svelte';

  let concInput = $state(8);
  let fileInput;
  let username = $state('');
  let role = $state('');

  // Jam Kerja state
  let jamMasuk = $state('08:00');
  let jamPulang = $state('16:00');
  let toleransi = $state(15);
  let jamMasukRam = $state('06:00');
  let jamPulangRam = $state('14:00');
  let toleransiRam = $state(15);
  let modeRamadan = $state(false);
  let jamKerjaSaving = $state(false);

  // Notifikasi WA state
  let waGroup = $state('');
  let waSaving = $state(false);
  let waTesting = $state('');

  // Rekap harian state (semua role: superadmin + admin instansi)
  let recapAktif = $state(true);
  let recapSaving = $state(false);
  let instansiList = $state([]);

  // Schedule state
  let schedules = $state([]);
  let newJam = $state(7);
  let newMenit = $state(0);
  let newLabel = $state('');
  let newMode = $state('all');
  let newTelegramEnabled = $state(false);
  let newWAEnabled = $state(true);
  let newWAGroup = $state('');
  let showAddForm = $state(false);
  let editingId = $state(null);

  async function loadWAGroups() {
    waGroupLoading = true;
    try {
      const res = await waGroups.list();
      if (res.success && Array.isArray(res.data)) {
        waGroupList = res.data;
      }
    } catch {}
    waGroupLoading = false;
  }

  // Confirm dialog state
  let confirmShow = $state(false);
  let confirmTitle = $state('');
  let confirmMessage = $state('');
  let confirmAction = $state(() => {});

  // WA Groups state
  let waGroupList = $state([]);
  let waGroupLoading = $state(false);

  async function loadData() {
    // User info
    try {
      const res = await fetch('/api/auth/me');
      const data = await res.json();
      if (data.success) { username = data.data.username || ''; role = data.data.role || ''; }
    } catch {}

    // Concurrency
    try {
      const res = await superadmin.getConcurrency();
      if (res.success) concInput = res.data.concurrency;
    } catch {}

    // Jam kerja
    try {
      const res = await instansi.getSettings();
      if (res.success && res.data) {
        jamMasuk = res.data.jam_masuk || '07:30';
        jamPulang = res.data.jam_pulang || '16:00';
        toleransi = res.data.toleransi || 15;
        jamMasukRam = res.data.jam_masuk_ram || '07:00';
        jamPulangRam = res.data.jam_pulang_ram || '15:30';
        toleransiRam = res.data.toleransi_ram || 15;
        modeRamadan = res.data.mode_ramadan || false;
        waGroup = res.data.wa_group || '';
        recapAktif = res.data.recap_aktif !== false;
      }
    } catch {}

    // Daftar instansi (superadmin: kelola rekap per instansi)
    if (role === 'superadmin') {
      try {
        const res = await superadmin.listInstansi();
        if (res.success) instansiList = res.data || [];
      } catch {}
    }

    // Schedules
    try {
      const res = await fetch('/api/schedules');
      const data = await res.json();
      if (data.success) schedules = data.data || [];
    } catch {}
  }

  // Hari libur state
  const namaHari = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'];
  let liburMingguan = $state([0]);
  let liburTanggal = $state([]);
  let liburSaving = $state(false);
  let newLiburTgl = $state('');
  let newLiburKet = $state('');

  async function loadLibur() {
    try {
      const res = await liburApi.get();
      if (res.success && res.data) {
        liburMingguan = res.data.mingguan || [0];
        liburTanggal = res.data.tanggal || [];
      }
    } catch {}
  }

  function toggleHariLibur(d) {
    liburMingguan = liburMingguan.includes(d) ? liburMingguan.filter(x => x !== d) : [...liburMingguan, d];
  }

  async function saveLiburMingguan() {
    liburSaving = true;
    try {
      const res = await liburApi.setMingguan(liburMingguan);
      if (res.success) toasts.success('Hari libur mingguan disimpan');
      else toasts.error(res.error || 'Gagal');
    } catch { toasts.error('Gagal menyimpan'); }
    liburSaving = false;
  }

  async function addLiburTanggal() {
    if (!newLiburTgl) { toasts.warning('Pilih tanggal dulu'); return; }
    try {
      const res = await liburApi.addTanggal(newLiburTgl, newLiburKet);
      if (res.success) {
        toasts.success('Tanggal libur ditambahkan');
        newLiburTgl = ''; newLiburKet = '';
        loadLibur();
      } else toasts.error(res.error || 'Gagal');
    } catch { toasts.error('Gagal menyimpan'); }
  }

  async function delLiburTanggal(id, tgl) {
    confirmTitle = 'Hapus hari libur?';
    confirmMessage = 'Hapus ' + tgl + ' dari daftar libur?';
    confirmAction = async () => {
      confirmShow = false;
      const res = await liburApi.delTanggal(id);
      if (res.success) { toasts.success('Dihapus'); loadLibur(); }
      else toasts.error(res.error || 'Gagal');
    };
    confirmShow = true;
  }

  onMount(() => { loadData(); loadWAGroups(); loadLibur(); });

  async function saveConc() {
    const n = parseInt(concInput) || 1;
    const res = await superadmin.setConcurrency(n);
    if (res.success) toasts.success('Worker diatur ke ' + n);
    else toasts.error(res.error || 'Gagal');
  }

  async function saveJamKerja() {
    jamKerjaSaving = true;
    try {
      const res = await instansi.updateSettings({
        jam_masuk: jamMasuk, jam_pulang: jamPulang, toleransi,
        jam_masuk_ram: jamMasukRam, jam_pulang_ram: jamPulangRam, toleransi_ram: toleransiRam,
        mode_ramadan: modeRamadan
      });
      if (res.success) toasts.success('Jam kerja diperbarui');
      else toasts.error(res.error || 'Gagal');
    } catch { toasts.error('Gagal menyimpan'); }
    jamKerjaSaving = false;
  }

  async function saveWaGroup() {
    waSaving = true;
    try {
      const res = await instansi.updateSettings({ wa_group: waGroup.trim() });
      if (res.success) { waGroup = res.data?.wa_group || ''; toasts.success('Grup WA diperbarui'); }
      else toasts.error(res.error || 'Gagal');
    } catch { toasts.error('Gagal menyimpan'); }
    waSaving = false;
  }

  async function testWaSend() {
    if (waTesting) return;
    waTesting = 'sending';
    try {
      const res = await fetch('/api/recap/test', { method: 'POST' }).then(r => r.json());
      if (res.success) { waTesting = ''; toasts.success('Terkirim ke ' + (res.data?.target || 'grup')); }
      else { waTesting = ''; toasts.error(res.error || 'Gagal kirim'); }
    } catch { waTesting = ''; toasts.error('Koneksi gagal'); }
  }

  async function saveRecap() {
    recapSaving = true;
    try {
      const res = await instansi.updateSettings({ recap_aktif: recapAktif });
      if (res.success) { recapAktif = res.data?.recap_aktif !== false; toasts.success(recapAktif ? 'Rekap harian diaktifkan' : 'Rekap harian dinonaktifkan'); }
      else toasts.error(res.error || 'Gagal');
    } catch { toasts.error('Gagal menyimpan'); }
    recapSaving = false;
  }

  async function toggleInstansiRecap(ins) {
    const prev = ins.recap_aktif;
    instansiList = instansiList.map(i => i.id === ins.id ? { ...i, recap_aktif: !prev } : i);
    const res = await superadmin.updateInstansi(ins.id, { recap_aktif: !prev });
    if (res.success) toasts.success(ins.nama + (prev ? ' dinonaktifkan' : ' diaktifkan'));
    else {
      instansiList = instansiList.map(i => i.id === ins.id ? { ...i, recap_aktif: prev } : i);
      toasts.error(res.error || 'Gagal');
    }
  }

  async function importPegawai() {
    if (!fileInput?.files?.length) { toasts.warning('Pilih file JSON dulu'); return; }
    const res = await superadmin.importPegawai(fileInput.files[0]);
    if (res.success) toasts.success('Import: ' + res.data.imported + ' pegawai, ' + res.data.skipped + ' skip');
    else toasts.error(res.error || 'Gagal');
  }

  function startEdit(s) {
    editingId = s.id;
    newJam = s.jam; newMenit = s.menit; newLabel = s.label; newMode = s.mode;
    newTelegramEnabled = s.telegram_enabled || false;
    newWAEnabled = s.wa_enabled !== false;
    newWAGroup = s.wa_group || '';
    showAddForm = true;
  }

  function cancelEdit() {
    editingId = null; showAddForm = false;
    newJam = 7; newMenit = 0; newLabel = ''; newMode = 'all';
    newTelegramEnabled = false; newWAEnabled = true; newWAGroup = '';
  }

  async function saveSchedule() {
    if (!newLabel) { toasts.warning('Isi label jadwal'); return; }
    const body = {
      jam: newJam, menit: newMenit, label: newLabel, mode: newMode,
      telegram_enabled: newTelegramEnabled, wa_enabled: newWAEnabled, wa_group: newWAGroup
    };
    const url = editingId ? `/api/schedules/${editingId}` : '/api/schedules';
    const method = editingId ? 'PUT' : 'POST';
    const res = await fetch(url, { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
    const data = await res.json();
    if (data.success) { toasts.success(editingId ? 'Jadwal diperbarui' : 'Jadwal ditambahkan'); cancelEdit(); await loadData(); }
    else toasts.error(data.error || 'Gagal');
  }

  // Optimistic toggle — update UI instantly, revert on error
  async function toggleSchedule(id) {
    const idx = schedules.findIndex(s => s.id === id);
    if (idx === -1) return;
    const prev = schedules[idx].aktif;
    // Optimistic update
    schedules = schedules.map(s => s.id === id ? { ...s, aktif: !s.aktif } : s);
    try {
      const res = await fetch(`/api/schedules/${id}/toggle`, { method: 'POST' });
      if (!res.ok) {
        // Revert on error
        schedules = schedules.map(s => s.id === id ? { ...s, aktif: prev } : s);
        toasts.error('Gagal toggle jadwal');
      }
    } catch {
      schedules = schedules.map(s => s.id === id ? { ...s, aktif: prev } : s);
      toasts.error('Gagal koneksi');
    }
  }

  // Optimistic delete — remove from list instantly, revert on error
  async function deleteSchedule(id) {
    confirmTitle = 'Hapus Jadwal';
    confirmMessage = 'Hapus jadwal ini?';
    confirmAction = async () => {
      const prev = [...schedules];
      schedules = schedules.filter(s => s.id !== id);
      try {
        const res = await fetch(`/api/schedules/${id}`, { method: 'DELETE' });
        if (res.ok) toasts.success('Jadwal dihapus');
        else { schedules = prev; toasts.error('Gagal menghapus jadwal'); }
      } catch {
        schedules = prev;
        toasts.error('Gagal koneksi');
      }
    };
    confirmShow = true;
  }

  function formatTime(jam, menit) { return String(jam).padStart(2, '0') + ':' + String(menit).padStart(2, '0'); }
  function modeLabel(mode) { return { 'all': 'Semua', 'belum_masuk': 'Belum Masuk', 'belum_pulang': 'Belum Pulang' }[mode] || mode; }
</script>

<Layout title="Pengaturan" activePage="settings">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Pengaturan' }]} />
  <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
    <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700">
      <i class="fa-solid fa-id-badge mr-0.5"></i> Info Akun
    </div>
    <div class="p-2.5 space-y-1">
      <div class="flex justify-between text-xs">
        <span class="text-gray-500">Username</span>
        <span class="font-semibold text-gray-900">{username || '-'}</span>
      </div>
      <div class="flex justify-between text-xs">
        <span class="text-gray-500">Role</span>
        <span class="inline-flex items-center px-1.5 py-0.5 rounded-full text-[10px] font-medium bg-blue-100 text-blue-700">{role || '-'}</span>
      </div>
    </div>
  </div>

  {#if role === 'superadmin' || role === 'admin'}
  <!-- Rekap Harian (instansi sendiri) -->
  <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
    <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700">
      <i class="fa-solid fa-image mr-0.5"></i> Rekap Harian (Gambar WA)
    </div>
    <div class="p-2.5">
      <div class="flex items-center justify-between mb-2">
        <div><span class="text-xs font-medium text-gray-700">Kirim rekap instansi ini</span><p class="text-[9px] text-gray-400">Matikan jika instansi ini tidak punya grup WA</p></div>
        <button onclick={() => recapAktif = !recapAktif}
                class="w-9 h-5 rounded-full transition-colors {recapAktif ? 'bg-green-500' : 'bg-gray-300'} relative" aria-label="Toggle rekap">
          <span class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full shadow transition-transform {recapAktif ? 'translate-x-4' : ''}"></span>
        </button>
      </div>
      <button onclick={saveRecap} disabled={recapSaving}
              class="w-full px-3 py-1.5 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white rounded-md text-xs font-medium transition-colors">
        {recapSaving ? 'Menyimpan...' : 'Simpan'}
      </button>
    </div>
  </div>
  {/if}
  
  {#if role === 'superadmin'}
    <!-- Rekap per Instansi -->
    <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
      <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700">
        <i class="fa-solid fa-building mr-0.5"></i> Rekap per Instansi
      </div>
      <div class="p-2.5 space-y-1.5">
        {#if instansiList.length === 0}
          <p class="text-xs text-gray-400 text-center py-2">Belum ada instansi</p>
        {:else}
          {#each instansiList as ins (ins.id)}
            <div class="flex items-center justify-between p-2 rounded-md border {ins.recap_aktif ? 'border-gray-200' : 'border-gray-100 opacity-60'}">
              <div><div class="text-xs font-medium text-gray-900">{ins.nama}</div><div class="text-[10px] text-gray-500">{ins.wa_group || 'grup default'}</div></div>
              <button onclick={() => toggleInstansiRecap(ins)} class="w-8 h-5 rounded-full transition-colors {ins.recap_aktif ? 'bg-green-500' : 'bg-gray-300'} relative" aria-label="Toggle rekap {ins.nama}">
                <span class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full shadow transition-transform {ins.recap_aktif ? 'translate-x-3' : ''}"></span></button>
            </div>
          {/each}
        {/if}
      </div>
    </div>

    <!-- Jam Kerja -->
    <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
      <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700">
        <i class="fa-solid fa-clock mr-0.5"></i> Jam Kerja
      </div>
      <div class="p-2.5">
        <div class="grid md:grid-cols-2 gap-2.5 mb-2.5">
          <div class="bg-gray-50 rounded-md p-2">
            <div class="text-[10px] font-semibold text-gray-600 mb-1.5 flex items-center gap-1">
              <i class="fa-solid fa-briefcase"></i> Jam Normal
            </div>
            <div class="grid grid-cols-2 gap-1.5">
              <div><label for="jm" class="block text-[10px] text-gray-500 mb-0.5">Masuk</label>
                <input id="jm" type="time" bind:value={jamMasuk} class="w-full px-1.5 py-1 border border-gray-300 rounded-md text-xs" /></div>
              <div><label for="jp" class="block text-[10px] text-gray-500 mb-0.5">Pulang</label>
                <input id="jp" type="time" bind:value={jamPulang} class="w-full px-1.5 py-1 border border-gray-300 rounded-md text-xs" /></div>
            </div>
            <div class="mt-1.5"><label for="tol" class="block text-[10px] text-gray-500 mb-0.5">Toleransi Telat</label>
              <div class="flex items-center gap-1"><input id="tol" type="number" min="0" max="60" bind:value={toleransi} class="w-16 px-1.5 py-1 border border-gray-300 rounded-md text-xs" /><span class="text-[10px] text-gray-500">mnt</span></div></div>
          </div>
          <div class="bg-gray-50 rounded-md p-2">
            <div class="text-[10px] font-semibold text-gray-600 mb-1.5 flex items-center gap-1">
              <i class="fa-solid fa-mosque"></i> Jam Ramadan
            </div>
            <div class="grid grid-cols-2 gap-1.5">
              <div><label for="jmr" class="block text-[10px] text-gray-500 mb-0.5">Masuk</label>
                <input id="jmr" type="time" bind:value={jamMasukRam} class="w-full px-1.5 py-1 border border-gray-300 rounded-md text-xs" /></div>
              <div><label for="jpr" class="block text-[10px] text-gray-500 mb-0.5">Pulang</label>
                <input id="jpr" type="time" bind:value={jamPulangRam} class="w-full px-1.5 py-1 border border-gray-300 rounded-md text-xs" /></div>
            </div>
            <div class="mt-1.5"><label for="tolr" class="block text-[10px] text-gray-500 mb-0.5">Toleransi Telat</label>
              <div class="flex items-center gap-1"><input id="tolr" type="number" min="0" max="60" bind:value={toleransiRam} class="w-16 px-1.5 py-1 border border-gray-300 rounded-md text-xs" /><span class="text-[10px] text-gray-500">mnt</span></div></div>
          </div>
        </div>
        <div class="flex items-center justify-between mb-2">
          <div><span class="text-xs font-medium text-gray-700">Mode Ramadan</span><p class="text-[9px] text-gray-400">Aktifkan jika sedang masa Ramadan</p></div>
          <button onclick={() => modeRamadan = !modeRamadan}
                  aria-label="Toggle Mode Ramadan"
                  class="w-9 h-5 rounded-full transition-colors {modeRamadan ? 'bg-green-500' : 'bg-gray-300'} relative">
            <span class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full shadow transition-transform {modeRamadan ? 'translate-x-4' : ''}"></span>
          </button>
        </div>
        <button onclick={saveJamKerja} disabled={jamKerjaSaving}
                class="w-full px-3 py-1.5 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white rounded-md text-xs font-medium transition-colors">
          {jamKerjaSaving ? 'Menyimpan...' : 'Simpan Jam Kerja'}
        </button>
      </div>
    </div>

    <!-- Notifikasi WA -->
    <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
      <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700">
        <i class="fa-brands fa-whatsapp mr-0.5"></i> Notifikasi WA (Rekap Gambar)
      </div>
      <div class="p-2.5">
        <label for="wag" class="block text-xs font-medium text-gray-700 mb-0.5">Grup WA Instansi</label>
        <input id="wag" type="text" bind:value={waGroup} placeholder="1203...@g.us atau 628xx (kosong = grup default)"
               class="w-full px-2 py-1.5 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-green-500 outline-none mb-1.5" />
        <p class="text-[9px] text-gray-400 mb-2">Rekap gambar harian dikirim ke grup ini tiap batch auto-scrape selesai (pagi + sore). Kosongkan untuk pakai grup default server.</p>
        <div class="grid grid-cols-2 gap-1.5">
          <button onclick={saveWaGroup} disabled={waSaving}
                  class="px-3 py-1.5 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white rounded-md text-xs font-medium transition-colors">
            {waSaving ? 'Menyimpan...' : 'Simpan Grup'}
          </button>
          <button onclick={testWaSend} disabled={waTesting === 'sending'}
                  class="px-3 py-1.5 border border-gray-300 hover:bg-gray-50 disabled:opacity-50 rounded-md text-xs font-medium text-gray-700 transition-colors">
            {#if waTesting === 'sending'}
              <Loading variant="button" label="Mengirim..." />
            {:else}
              <i class="fa-solid fa-paper-plane mr-0.5"></i>Kirim Tes
            {/if}
          </button>
        </div>
      </div>
    </div>

    <!-- Jadwal Auto Scrape -->
    <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
      <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700 flex items-center justify-between">
        <span><i class="fa-solid fa-calendar-days mr-0.5"></i> Jadwal Auto Scrape</span>
        <button onclick={() => { cancelEdit(); showAddForm = true; }} class="text-[10px] px-1.5 py-0.5 bg-blue-50 text-blue-600 rounded-md hover:bg-blue-100"><i class="fa-solid fa-plus mr-0.5"></i> Tambah</button>
      </div>
      <div class="p-2.5">
        {#if showAddForm}
          <div class="bg-gray-50 rounded-md p-2.5 mb-2 space-y-2">
            <div class="grid grid-cols-2 gap-1.5">
              <div><label for="sj" class="block text-[10px] text-gray-500 mb-0.5">Jam</label><input id="sj" type="number" min="0" max="23" bind:value={newJam} class="w-full px-1.5 py-1 border border-gray-300 rounded-md text-xs" /></div>
              <div><label for="sm" class="block text-[10px] text-gray-500 mb-0.5">Menit</label><input id="sm" type="number" min="0" max="59" bind:value={newMenit} class="w-full px-1.5 py-1 border border-gray-300 rounded-md text-xs" /></div>
            </div>
            <div><label for="sl" class="block text-[10px] text-gray-500 mb-0.5">Label</label><input id="sl" type="text" bind:value={newLabel} placeholder="Pagi, Sore, Rekap" class="w-full px-1.5 py-1 border border-gray-300 rounded-md text-xs" /></div>
            <div><label for="smd" class="block text-[10px] text-gray-500 mb-0.5">Mode Scrape</label>
              <select id="smd" bind:value={newMode} class="w-full px-1.5 py-1 border border-gray-300 rounded-md text-xs">
                <option value="all">Semua Pegawai</option><option value="belum_masuk">Belum Masuk</option><option value="belum_pulang">Belum Pulang</option>
              </select></div>
            
            <!-- Kirim Notifikasi -->
            <div class="border-t border-gray-200 pt-2">
              <div class="text-[10px] font-semibold text-gray-600 mb-1.5"><i class="fa-solid fa-paper-plane mr-0.5"></i> Kirim Notifikasi</div>
              
              <!-- WhatsApp -->
              <div class="flex items-center justify-between mb-1.5">
                <div class="flex items-center gap-1.5">
                  <i class="fa-brands fa-whatsapp text-green-600 text-xs"></i>
                  <span class="text-[10px] font-medium text-gray-700">WhatsApp</span>
                </div>
                <button onclick={() => newWAEnabled = !newWAEnabled}
                        class="w-8 h-4.5 rounded-full transition-colors {newWAEnabled ? 'bg-green-500' : 'bg-gray-300'} relative" aria-label="Toggle WA">
                  <span class="absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow transition-transform {newWAEnabled ? 'translate-x-3.5' : ''}"></span>
                </button>
              </div>
              {#if newWAEnabled}
                <select bind:value={newWAGroup} class="w-full px-1.5 py-1 border border-gray-300 rounded-md text-xs mb-1.5">
                  <option value="">Grup default instansi</option>
                  {#each waGroupList as g (g.id)}
                    <option value={g.id}>{g.name || g.id}</option>
                  {/each}
                </select>
                {#if waGroupLoading}
                  <p class="text-[9px] text-gray-400">Memuat daftar grup WA...</p>
                {:else if waGroupList.length === 0}
                  <p class="text-[9px] text-gray-400">Tidak ada grup WA ditemukan. Cek config GoWA.</p>
                {/if}
              {/if}

              <!-- Telegram -->
              <div class="flex items-center justify-between mt-1.5">
                <div class="flex items-center gap-1.5">
                  <i class="fa-brands fa-telegram text-blue-500 text-xs"></i>
                  <span class="text-[10px] font-medium text-gray-700">Telegram</span>
                </div>
                <button onclick={() => newTelegramEnabled = !newTelegramEnabled}
                        class="w-8 h-4.5 rounded-full transition-colors {newTelegramEnabled ? 'bg-green-500' : 'bg-gray-300'} relative" aria-label="Toggle Telegram">
                  <span class="absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow transition-transform {newTelegramEnabled ? 'translate-x-3.5' : ''}"></span>
                </button>
              </div>
            </div>

            <div class="flex gap-1.5">
              <button onclick={saveSchedule} class="flex-1 px-2.5 py-1.5 bg-blue-600 text-white rounded-md text-xs font-medium hover:bg-blue-700">{editingId ? 'Update' : 'Simpan'}</button>
              <button onclick={cancelEdit} class="px-2.5 py-1.5 border border-gray-300 rounded-md text-xs hover:bg-gray-50">Batal</button>
            </div>
          </div>
        {/if}
        {#if schedules.length === 0}
          <p class="text-xs text-gray-400 text-center py-3">Belum ada jadwal</p>
        {:else}
          <div class="space-y-1.5">
            {#each schedules as s (s.id)}
              <div class="p-2 rounded-md border {s.aktif ? 'border-gray-200' : 'border-gray-100 opacity-60'}">
                <!-- Row 1: Time + Label + Actions -->
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class="w-8 h-8 rounded-md {s.aktif ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-400'} flex items-center justify-center font-bold text-[10px]">{formatTime(s.jam, s.menit)}</div>
                    <div><div class="text-xs font-medium text-gray-900">{s.label}</div><div class="text-[10px] text-gray-500">{modeLabel(s.mode)}</div></div>
                  </div>
                  <div class="flex items-center gap-0.5">
                    <button onclick={() => startEdit(s)} class="p-1 text-gray-400 hover:text-blue-500 rounded-md hover:bg-blue-50" aria-label="Edit"><i class="fa-solid fa-pen text-[10px]"></i></button>
                    <button onclick={() => toggleSchedule(s.id)} class="w-8 h-5 rounded-full transition-colors {s.aktif ? 'bg-green-500' : 'bg-gray-300'} relative" aria-label="Toggle">
                      <span class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full shadow transition-transform {s.aktif ? 'translate-x-3' : ''}"></span></button>
                    <button onclick={() => deleteSchedule(s.id)} class="p-1 text-gray-400 hover:text-red-500 rounded-md hover:bg-red-50" aria-label="Hapus"><i class="fa-solid fa-trash text-[10px]"></i></button>
                  </div>
                </div>
                <!-- Row 2: Notification badges -->
                <div class="flex items-center gap-1.5 mt-1.5 ml-10">
                  {#if s.wa_enabled !== false}
                    <span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-[8px] font-medium bg-green-50 text-green-700">
                      <i class="fa-brands fa-whatsapp"></i> WA
                      {#if s.wa_group}<span class="text-green-500 ml-0.5">({s.wa_group.substring(0, 12)}...)</span>{/if}
                    </span>
                  {:else}
                    <span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-[8px] font-medium bg-gray-50 text-gray-400">
                      <i class="fa-brands fa-whatsapp"></i> WA off
                    </span>
                  {/if}
                  {#if s.telegram_enabled}
                    <span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-[8px] font-medium bg-blue-50 text-blue-700">
                      <i class="fa-brands fa-telegram"></i> TG
                    </span>
                  {:else}
                    <span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded-full text-[8px] font-medium bg-gray-50 text-gray-400">
                      <i class="fa-brands fa-telegram"></i> TG off
                    </span>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <!-- Hari Libur -->
    <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
      <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700"><i class="fa-solid fa-umbrella-beach mr-0.5"></i> Hari Libur</div>
      <div class="p-2.5">
        <div class="text-[10px] font-semibold text-gray-600 mb-1.5">Libur mingguan (auto-scrape dilewati)</div>
        <div class="flex gap-1.5 flex-wrap mb-2">
          {#each [0,1,2,3,4,5,6] as d (d)}
            <button onclick={() => toggleHariLibur(d)}
                    class="px-2.5 py-1 rounded-md text-xs font-medium border transition-colors {liburMingguan.includes(d) ? 'bg-blue-600 text-white border-blue-600' : 'bg-white text-gray-600 border-gray-200'}">
              {namaHari[d]}
            </button>
          {/each}
        </div>
        <button onclick={saveLiburMingguan} disabled={liburSaving}
                class="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white rounded-md text-xs font-medium mb-3">
          {liburSaving ? 'Menyimpan...' : 'Simpan Hari Libur'}
        </button>
        <div class="text-[10px] font-semibold text-gray-600 mb-1.5 border-t border-gray-100 pt-2">Tanggal merah / libur khusus</div>
        <div class="flex gap-1.5 mb-2">
          <input type="date" bind:value={newLiburTgl} aria-label="Tanggal libur" class="flex-1 px-2 py-1 border border-gray-300 rounded-md text-xs outline-none focus:ring-2 focus:ring-blue-500" />
          <input type="text" bind:value={newLiburKet} placeholder="Keterangan" class="flex-1 px-2 py-1 border border-gray-300 rounded-md text-xs outline-none focus:ring-2 focus:ring-blue-500" />
          <button onclick={addLiburTanggal} class="px-2.5 py-1 bg-green-600 hover:bg-green-700 text-white rounded-md text-xs" aria-label="Tambah libur"><i class="fa-solid fa-plus"></i></button>
        </div>
        {#if liburTanggal.length === 0}
          <p class="text-xs text-gray-400 text-center py-2">Belum ada tanggal libur khusus</p>
        {:else}
          <div class="space-y-1">
            {#each liburTanggal as l (l.id)}
              <div class="flex items-center justify-between px-2 py-1 rounded-md border border-gray-100 text-xs">
                <span class="font-medium text-gray-800">{l.tanggal}{#if l.keterangan}<span class="text-gray-400 font-normal"> — {l.keterangan}</span>{/if}</span>
                <button onclick={() => delLiburTanggal(l.id, l.tanggal)} class="p-1 text-gray-400 hover:text-red-500" aria-label="Hapus"><i class="fa-solid fa-trash text-[10px]"></i></button>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <!-- Worker -->
    <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
      <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700"><i class="fa-solid fa-microchip mr-0.5"></i> Worker Scrape</div>
      <div class="p-2.5">
        <label for="wc" class="block text-xs font-medium text-gray-700 mb-0.5">Jumlah Worker Paralel</label>
        <div class="flex gap-1.5">
          <input id="wc" type="number" min="1" max="100" bind:value={concInput} class="flex-1 px-2 py-1.5 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-blue-500 outline-none" />
          <button onclick={saveConc} class="px-3 py-1.5 border border-gray-300 rounded-md text-xs font-medium hover:bg-gray-50 transition-colors">Simpan</button>
        </div>
        <p class="text-[9px] text-gray-400 mt-1">Max 60 req/jam/akun rate-limit Pusaka.</p>
      </div>
    </div>

    <!-- Import -->
    <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
      <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700"><i class="fa-solid fa-file-import mr-0.5"></i> Import Pegawai</div>
      <div class="p-2.5">
        <input type="file" accept=".json" bind:this={fileInput} class="block w-full text-[10px] text-gray-500 file:mr-3 file:py-1 file:px-3 file:rounded-md file:border-0 file:text-[10px] file:font-medium file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 mb-2" />
        <button onclick={importPegawai} class="w-full px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-xs font-medium transition-colors"><i class="fa-solid fa-cloud-arrow-up mr-0.5"></i> Import dari JSON</button>
      </div>
    </div>
  {/if}
</Layout>

<ConfirmDialog show={confirmShow} title={confirmTitle} message={confirmMessage}
               confirmLabel="Hapus" variant="danger"
               onConfirm={confirmAction} onCancel={() => confirmShow = false} />
