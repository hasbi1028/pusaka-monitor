<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import { pegawai, scrape } from '$lib/api.js';
  import { toasts } from '$lib/stores/toast.js';

  let pegawaiList = $state([]);
  let showModal = $state(false);
  let editingId = $state('');
  let form = $state({ nip: '', nama: '', password_pusaka: '' });

  async function loadPegawai() {
    const res = await pegawai.list();
    if (res.success) pegawaiList = res.data;
  }

  function showAdd() { editingId = ''; form = { nip: '', nama: '', password_pusaka: '' }; showModal = true; }
  function showEdit(p) { editingId = p.id; form = { nip: p.nip, nama: p.nama, password_pusaka: '' }; showModal = true; }
  function closeModal() { showModal = false; }

  async function save() {
    const body = { ...form };
    const res = editingId ? await pegawai.update(editingId, body) : await pegawai.create(body);
    if (res.success) { closeModal(); loadPegawai(); toasts.success('Pegawai berhasil disimpan'); }
  }

  async function del(id) {
    if (!confirm('Nonaktifkan pegawai ini?')) return;
    await pegawai.delete(id);
    loadPegawai();
    toasts.warning('Pegawai dinonaktifkan');
  }

  async function doScrape(nip) {
    const res = await scrape.scrapeOne(nip);
    if (res.success) toasts.success(res.message || 'Job scrape dibuat');
    else toasts.error(res.error || 'Gagal');
  }

  loadPegawai();
</script>

<Layout title="Pegawai" activePage="pegawai">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Pegawai' }]} />
  <div class="flex items-center justify-between mb-4">
    <h2 class="text-sm font-semibold text-gray-700"><i class="fa-solid fa-users mr-1"></i> Daftar Pegawai</h2>
    <button onclick={showAdd}
            class="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors">
      <i class="fa-solid fa-plus mr-1"></i>Tambah
    </button>
  </div>
  
  <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 text-gray-500 text-xs uppercase">
          <tr><th class="px-4 py-2 text-left">NIP</th><th class="px-4 py-2 text-left">Nama</th><th class="px-4 py-2 text-center">Scrape</th><th class="px-4 py-2 text-center">Aksi</th></tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          {#each pegawaiList as p}
            <tr class="hover:bg-gray-50">
              <td class="px-4 py-2 text-gray-500 text-xs">{p.nip}</td>
              <td class="px-4 py-2 font-medium text-gray-900">{p.nama}</td>
              <td class="px-4 py-2 text-center">
                <button onclick={() => doScrape(p.nip)}
                        class="p-1.5 rounded-lg border border-gray-200 hover:bg-yellow-50 text-yellow-600 transition-colors" title="Scrape">
                  <i class="fa-solid fa-bolt text-sm"></i>
                </button>
              </td>
              <td class="px-4 py-2 text-center">
                <div class="flex items-center justify-center gap-1">
                  <button onclick={() => showEdit(p)}
                          class="p-1.5 rounded-lg border border-gray-200 hover:bg-blue-50 text-blue-600 transition-colors">
                    <i class="fa-solid fa-pen text-sm"></i>
                  </button>
                  <button onclick={() => del(p.id)}
                          class="p-1.5 rounded-lg border border-gray-200 hover:bg-red-50 text-red-600 transition-colors">
                    <i class="fa-solid fa-trash text-sm"></i>
                  </button>
                </div>
              </td>
            </tr>
          {:else}
            <tr><td colspan="4" class="px-4 py-6 text-center text-gray-400 text-sm">Belum ada pegawai</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</Layout>

{#if showModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div class="absolute inset-0 bg-black/50" onclick={closeModal}></div>
    <div class="bg-white rounded-2xl shadow-xl w-full max-w-sm p-6 relative z-10">
      <button onclick={closeModal} class="absolute top-3 right-3 p-1 rounded-lg hover:bg-gray-100 text-gray-400">
        <i class="fa-solid fa-xmark"></i>
      </button>
      <h3 class="text-lg font-semibold text-gray-900 mb-1">{editingId ? 'Edit' : 'Tambah'} Pegawai</h3>
      <p class="text-sm text-gray-500 mb-4">Isi data pegawai dan password Pusaka.</p>
      <div class="space-y-3">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">NIP <span class="text-red-500">*</span></label>
          <input type="text" bind:value={form.nip} class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" placeholder="Nomor Induk Pegawai" required />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Nama Lengkap <span class="text-red-500">*</span></label>
          <input type="text" bind:value={form.nama} class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" required />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Password Pusaka</label>
          <input type="password" bind:value={form.password_pusaka} class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" placeholder="Password login Pusaka Kemenag" />
        </div>
      </div>
      <div class="flex gap-2 mt-6">
        <button onclick={closeModal} class="flex-1 px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors">Batal</button>
        <button onclick={save} class="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors">Simpan</button>
      </div>
    </div>
  </div>
{/if}
