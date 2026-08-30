<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import { superadmin, instansi } from '$lib/api.js';
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

  // Schedule state
  let schedules = $state([]);
  let newJam = $state(7);
  let newMenit = $state(0);
  let newLabel = $state('');
  let newMode = $state('all');
  let showAddForm = $state(false);
  let editingId = $state(null);

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
      }
    } catch {}

    // Schedules
    try {
      const res = await fetch('/api/schedules');
      const data = await res.json();
      if (data.success) schedules = data.data || [];
    } catch {}
  }

  onMount(() => { loadData(); });

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

  async function importPegawai() {
    if (!fileInput?.files?.length) { toasts.warning('Pilih file JSON dulu'); return; }
    const res = await superadmin.importPegawai(fileInput.files[0]);
    if (res.success) toasts.success('Import: ' + res.data.imported + ' pegawai, ' + res.data.skipped + ' skip');
    else toasts.error(res.error || 'Gagal');
  }

  function startEdit(s) {
    editingId = s.id;
    newJam = s.jam; newMenit = s.menit; newLabel = s.label; newMode = s.mode;
    showAddForm = true;
  }

  function cancelEdit() {
    editingId = null; showAddForm = false;
    newJam = 7; newMenit = 0; newLabel = ''; newMode = 'all';
  }

  async function saveSchedule() {
    if (!newLabel) { toasts.warning('Isi label jadwal'); return; }
    const body = { jam: newJam, menit: newMenit, label: newLabel, mode: newMode };
    const url = editingId ? `/api/schedules/${editingId}` : '/api/schedules';
    const method = editingId ? 'PUT' : 'POST';
    const res = await fetch(url, { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
    const data = await res.json();
    if (data.success) { toasts.success(editingId ? 'Jadwal diperbarui' : 'Jadwal ditambahkan'); cancelEdit(); await loadData(); }
    else toasts.error(data.error || 'Gagal');
  }

  async function toggleSchedule(id) {
    await fetch(`/api/schedules/${id}/toggle`, { method: 'POST' });
    await loadData();
  }

  async function deleteSchedule(id) {
    if (!confirm('Hapus jadwal ini?')) return;
    await fetch(`/api/schedules/${id}`, { method: 'DELETE' });
    toasts.success('Jadwal dihapus');
    await loadData();
  }

  function formatTime(jam, menit) { return String(jam).padStart(2, '0') + ':' + String(menit).padStart(2, '0'); }
  function modeLabel(mode) { return { 'all': 'Semua', 'belum_masuk': 'Belum Masuk', 'belum_pulang': 'Belum Pulang' }[mode] || mode; }
</script>

<Layout title="Pengaturan" activePage="settings">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Pengaturan' }]} />
  <div class="bg-white rounded-xl border border-gray-200 shadow-sm mb-4">
    <div class="px-4 py-2.5 border-b border-gray-200 text-sm font-semibold text-gray-700">
      <i class="fa-solid fa-id-badge mr-1"></i> Info Akun
    </div>
    <div class="p-4 space-y-2">
      <div class="flex justify-between text-sm">
        <span class="text-gray-500">Username</span>
        <span class="font-semibold text-gray-900">{username || '-'}</span>
      </div>
      <div class="flex justify-between text-sm">
        <span class="text-gray-500">Role</span>
        <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-700">{role || '-'}</span>
      </div>
    </div>
  </div>
  
  {#if role === 'superadmin'}
    <!-- Jam Kerja -->
    <div class="bg-white rounded-xl border border-gray-200 shadow-sm mb-4">
      <div class="px-4 py-2.5 border-b border-gray-200 text-sm font-semibold text-gray-700">
        <i class="fa-solid fa-clock mr-1"></i> Jam Kerja
      </div>
      <div class="p-4">
        <div class="grid md:grid-cols-2 gap-4 mb-4">
          <div class="bg-gray-50 rounded-lg p-3">
            <div class="text-xs font-semibold text-gray-600 mb-2 flex items-center gap-1">
              <i class="fa-solid fa-briefcase"></i> Jam Normal
            </div>
            <div class="grid grid-cols-2 gap-2">
              <div><label for="jm" class="block text-xs text-gray-500 mb-1">Masuk</label>
                <input id="jm" type="time" bind:value={jamMasuk} class="w-full px-2 py-1.5 border border-gray-300 rounded-lg text-sm" /></div>
              <div><label for="jp" class="block text-xs text-gray-500 mb-1">Pulang</label>
                <input id="jp" type="time" bind:value={jamPulang} class="w-full px-2 py-1.5 border border-gray-300 rounded-lg text-sm" /></div>
            </div>
            <div class="mt-2"><label for="tol" class="block text-xs text-gray-500 mb-1">Toleransi Telat</label>
              <div class="flex items-center gap-1"><input id="tol" type="number" min="0" max="60" bind:value={toleransi} class="w-20 px-2 py-1.5 border border-gray-300 rounded-lg text-sm" /><span class="text-xs text-gray-500">menit</span></div></div>
          </div>
          <div class="bg-gray-50 rounded-lg p-3">
            <div class="text-xs font-semibold text-gray-600 mb-2 flex items-center gap-1">
              <i class="fa-solid fa-mosque"></i> Jam Ramadan
            </div>
            <div class="grid grid-cols-2 gap-2">
              <div><label for="jmr" class="block text-xs text-gray-500 mb-1">Masuk</label>
                <input id="jmr" type="time" bind:value={jamMasukRam} class="w-full px-2 py-1.5 border border-gray-300 rounded-lg text-sm" /></div>
              <div><label for="jpr" class="block text-xs text-gray-500 mb-1">Pulang</label>
                <input id="jpr" type="time" bind:value={jamPulangRam} class="w-full px-2 py-1.5 border border-gray-300 rounded-lg text-sm" /></div>
            </div>
            <div class="mt-2"><label for="tolr" class="block text-xs text-gray-500 mb-1">Toleransi Telat</label>
              <div class="flex items-center gap-1"><input id="tolr" type="number" min="0" max="60" bind:value={toleransiRam} class="w-20 px-2 py-1.5 border border-gray-300 rounded-lg text-sm" /><span class="text-xs text-gray-500">menit</span></div></div>
          </div>
        </div>
        <div class="flex items-center justify-between mb-3">
          <div><span class="text-sm font-medium text-gray-700">Mode Ramadan</span><p class="text-xs text-gray-400">Aktifkan jika sedang masa Ramadan</p></div>
          <button onclick={() => modeRamadan = !modeRamadan}
                  class="w-11 h-6 rounded-full transition-colors {modeRamadan ? 'bg-green-500' : 'bg-gray-300'} relative">
            <span class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform {modeRamadan ? 'translate-x-5' : ''}"></span>
          </button>
        </div>
        <button onclick={saveJamKerja} disabled={jamKerjaSaving}
                class="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors">
          {jamKerjaSaving ? 'Menyimpan...' : 'Simpan Jam Kerja'}
        </button>
      </div>
    </div>

    <!-- Jadwal Auto Scrape -->
    <div class="bg-white rounded-xl border border-gray-200 shadow-sm mb-4">
      <div class="px-4 py-2.5 border-b border-gray-200 text-sm font-semibold text-gray-700 flex items-center justify-between">
        <span><i class="fa-solid fa-calendar-days mr-1"></i> Jadwal Auto Scrape</span>
        <button onclick={() => { cancelEdit(); showAddForm = true; }} class="text-xs px-2 py-1 bg-blue-50 text-blue-600 rounded-lg hover:bg-blue-100"><i class="fa-solid fa-plus mr-1"></i> Tambah</button>
      </div>
      <div class="p-4">
        {#if showAddForm}
          <div class="bg-gray-50 rounded-lg p-3 mb-3 space-y-2">
            <div class="grid grid-cols-2 gap-2">
              <div><label for="sj" class="block text-xs text-gray-500 mb-1">Jam</label><input id="sj" type="number" min="0" max="23" bind:value={newJam} class="w-full px-2 py-1.5 border border-gray-300 rounded-lg text-sm" /></div>
              <div><label for="sm" class="block text-xs text-gray-500 mb-1">Menit</label><input id="sm" type="number" min="0" max="59" bind:value={newMenit} class="w-full px-2 py-1.5 border border-gray-300 rounded-lg text-sm" /></div>
            </div>
            <div><label for="sl" class="block text-xs text-gray-500 mb-1">Label</label><input id="sl" type="text" bind:value={newLabel} placeholder="Pagi, Sore" class="w-full px-2 py-1.5 border border-gray-300 rounded-lg text-sm" /></div>
            <div><label for="smd" class="block text-xs text-gray-500 mb-1">Mode</label>
              <select id="smd" bind:value={newMode} class="w-full px-2 py-1.5 border border-gray-300 rounded-lg text-sm">
                <option value="all">Semua</option><option value="belum_masuk">Belum Masuk</option><option value="belum_pulang">Belum Pulang</option>
              </select></div>
            <div class="flex gap-2">
              <button onclick={saveSchedule} class="flex-1 px-3 py-1.5 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700">{editingId ? 'Update' : 'Simpan'}</button>
              <button onclick={cancelEdit} class="px-3 py-1.5 border border-gray-300 rounded-lg text-sm hover:bg-gray-50">Batal</button>
            </div>
          </div>
        {/if}
        {#if schedules.length === 0}
          <p class="text-sm text-gray-400 text-center py-4">Belum ada jadwal</p>
        {:else}
          <div class="space-y-2">
            {#each schedules as s}
              <div class="flex items-center justify-between p-3 rounded-lg border {s.aktif ? 'border-gray-200' : 'border-gray-100 opacity-60'}">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-lg {s.aktif ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-400'} flex items-center justify-center font-bold text-sm">{formatTime(s.jam, s.menit)}</div>
                  <div><div class="text-sm font-medium text-gray-900">{s.label}</div><div class="text-xs text-gray-500">{modeLabel(s.mode)}</div></div>
                </div>
                <div class="flex items-center gap-1">
                  <button onclick={() => startEdit(s)} class="p-1.5 text-gray-400 hover:text-blue-500 rounded-lg hover:bg-blue-50" aria-label="Edit"><i class="fa-solid fa-pen text-xs"></i></button>
                  <button onclick={() => toggleSchedule(s.id)} class="w-10 h-6 rounded-full transition-colors {s.aktif ? 'bg-green-500' : 'bg-gray-300'} relative" aria-label="Toggle">
                    <span class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform {s.aktif ? 'translate-x-4' : ''}"></span></button>
                  <button onclick={() => deleteSchedule(s.id)} class="p-1.5 text-gray-400 hover:text-red-500 rounded-lg hover:bg-red-50" aria-label="Hapus"><i class="fa-solid fa-trash text-xs"></i></button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <!-- Worker -->
    <div class="bg-white rounded-xl border border-gray-200 shadow-sm mb-4">
      <div class="px-4 py-2.5 border-b border-gray-200 text-sm font-semibold text-gray-700"><i class="fa-solid fa-microchip mr-1"></i> Worker Scrape</div>
      <div class="p-4">
        <label for="wc" class="block text-sm font-medium text-gray-700 mb-1">Jumlah Worker Paralel</label>
        <div class="flex gap-2">
          <input id="wc" type="number" min="1" max="100" bind:value={concInput} class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
          <button onclick={saveConc} class="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium hover:bg-gray-50 transition-colors">Simpan</button>
        </div>
        <p class="text-xs text-gray-400 mt-2">Max 60 req/jam/akun rate-limit Pusaka.</p>
      </div>
    </div>

    <!-- Import -->
    <div class="bg-white rounded-xl border border-gray-200 shadow-sm mb-4">
      <div class="px-4 py-2.5 border-b border-gray-200 text-sm font-semibold text-gray-700"><i class="fa-solid fa-file-import mr-1"></i> Import Pegawai</div>
      <div class="p-4">
        <input type="file" accept=".json" bind:this={fileInput} class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-medium file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 mb-3" />
        <button onclick={importPegawai} class="w-full px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors"><i class="fa-solid fa-cloud-arrow-up mr-1"></i> Import dari JSON</button>
      </div>
    </div>
  {/if}
</Layout>
