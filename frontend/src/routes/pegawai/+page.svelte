<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import Loading from '$lib/components/Loading.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import { pegawai, scrape } from '$lib/api.js';
  import { toasts } from '$lib/stores/toast.js';
  import Icon from '$lib/components/Icon.svelte';

  let pegawaiList = $state([]);
  let showModal = $state(false);
  let editingId = $state('');
  let loading = $state(true);
  let saving = $state(false);
  let scrapingNip = $state('');
  let busy = $state(false);
  let form = $state({ nip: '', nama: '', password_pusaka: '' });

  // Confirm dialog state
  let confirmShow = $state(false);
  let confirmTitle = $state('');
  let confirmMessage = $state('');
  let confirmAction = $state(() => {});

  async function loadPegawai() {
    loading = true;
    const res = await pegawai.list();
    if (res.success) pegawaiList = res.data;
    loading = false;
  }

  function showAdd() { editingId = ''; form = { nip: '', nama: '', password_pusaka: '' }; showModal = true; }
  function showEdit(p) { editingId = p.id; form = { nip: p.nip, nama: p.nama, password_pusaka: '' }; showModal = true; }
  function closeModal() { showModal = false; }

  async function save() {
    if (busy) return;
    busy = true; saving = true;
    const body = { ...form };
    const res = editingId ? await pegawai.update(editingId, body) : await pegawai.create(body);
    if (res.success) { closeModal(); await loadPegawai(); toasts.success('Pegawai berhasil disimpan'); }
    saving = false; busy = false;
  }

  // Optimistic delete — remove from list instantly, revert on error
  function del(id) {
    confirmTitle = 'Nonaktifkan Pegawai';
    confirmMessage = 'Nonaktifkan pegawai ini?';
    confirmAction = async () => {
      const prev = [...pegawaiList];
      const target = pegawaiList.find(p => p.id === id);
      pegawaiList = pegawaiList.filter(p => p.id !== id);
      try {
        const res = await pegawai.delete(id);
        if (res.success) toasts.warning('Pegawai dinonaktifkan');
        else { pegawaiList = prev; toasts.error(res.error || 'Gagal'); }
      } catch {
        pegawaiList = prev;
        toasts.error('Gagal koneksi');
      }
    };
    confirmShow = true;
  }

  async function doScrape(nip) {
    if (!nip || scrapingNip) return;
    scrapingNip = nip;
    const res = await scrape.scrapeOne(nip);
    if (res.success) toasts.success(res.message || 'Job scrape dibuat');
    else toasts.error(res.error || 'Gagal');
    scrapingNip = '';
  }

  loadPegawai();
</script>

<Layout title="Pegawai" activePage="pegawai">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Pegawai' }]} />
  <div class="flex items-center justify-between mb-2">
    <h2 class="text-xs font-semibold text-gray-700"><Icon name="users" class="mr-0.5" /> Daftar Pegawai</h2>
    <button onclick={showAdd}
            class="px-2.5 py-1 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-xs font-medium transition-colors">
      <Icon name="plus" class="mr-0.5" />Tambah
    </button>
  </div>
  
  <div class="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
    {#if loading}
      <Loading variant="spinner" size="sm" label="Memuat daftar pegawai..." />
    {:else}
    <div class="overflow-x-auto">
      <table class="w-full text-xs">
        <thead class="bg-gray-50 text-gray-500 text-[10px] uppercase">
          <tr><th class="px-3 py-1.5 text-left">NIP</th><th class="px-3 py-1.5 text-left">Nama</th><th class="px-3 py-1.5 text-center">Scrape</th><th class="px-3 py-1.5 text-center">Aksi</th></tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          {#each pegawaiList as p (p.id)}
            <tr class="hover:bg-gray-50">
              <td class="px-3 py-1.5 text-gray-500 text-[10px]">{p.nip}</td>
              <td class="px-3 py-1.5 font-medium text-gray-900">{p.nama}</td>
              <td class="px-3 py-1.5 text-center">
                <button onclick={() => doScrape(p.nip)} disabled={scrapingNip === p.nip}
                        title="Scrape"
                        class="p-1 rounded-md border border-gray-200 hover:bg-yellow-50 text-yellow-600 transition-colors disabled:opacity-50">
                  {#if scrapingNip === p.nip}
                    <div class="w-3 h-3 border-2 border-yellow-200 border-t-yellow-600 rounded-full animate-spin mx-auto"></div>
                  {:else}
                    <Icon name="bolt" class="text-[10px]" />
                  {/if}
                </button>
              </td>
              <td class="px-3 py-1.5 text-center">
                <div class="flex items-center justify-center gap-0.5">
                  <button onclick={() => showEdit(p)} aria-label="Edit pegawai"
                          class="p-1 rounded-md border border-gray-200 hover:bg-blue-50 text-blue-600 transition-colors">
                    <Icon name="pen" class="text-[10px]" />
                  </button>
                  <button onclick={() => del(p.id)} aria-label="Nonaktifkan pegawai"
                          class="p-1 rounded-md border border-gray-200 hover:bg-red-50 text-red-600 transition-colors">
                    <Icon name="trash" class="text-[10px]" />
                  </button>
                </div>
              </td>
            </tr>
          {:else}
            <tr><td colspan="4" class="px-3 py-4 text-center text-gray-400 text-xs">Belum ada pegawai</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
    {/if}
  </div>
</Layout>

{#if showModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-3">
    <div class="absolute inset-0 bg-black/50" onclick={closeModal} role="presentation"></div>
    <div class="bg-white rounded-xl shadow-xl w-full max-w-sm p-4 relative z-10">
      <button onclick={closeModal} class="absolute top-2 right-2 p-0.5 rounded-md hover:bg-gray-100 text-gray-400" aria-label="Tutup">
        <Icon name="xmark" class="text-xs" />
      </button>
      <h3 class="text-sm font-semibold text-gray-900 mb-0.5">{editingId ? 'Edit' : 'Tambah'} Pegawai</h3>
      <p class="text-[10px] text-gray-500 mb-3">Isi data pegawai dan password Pusaka.</p>
      <div class="space-y-2">
        <div>
          <label for="pnip" class="block text-[10px] font-medium text-gray-700 mb-0.5">NIP <span class="text-red-500">*</span></label>
          <input id="pnip" type="text" bind:value={form.nip} class="w-full px-2 py-1.5 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-blue-500 outline-none" placeholder="Nomor Induk Pegawai" required />
        </div>
        <div>
          <label for="pnama" class="block text-[10px] font-medium text-gray-700 mb-0.5">Nama Lengkap <span class="text-red-500">*</span></label>
          <input id="pnama" type="text" bind:value={form.nama} class="w-full px-2 py-1.5 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-blue-500 outline-none" required />
        </div>
        <div>
          <label for="ppass" class="block text-[10px] font-medium text-gray-700 mb-0.5">Password Pusaka</label>
          <input id="ppass" type="password" bind:value={form.password_pusaka} class="w-full px-2 py-1.5 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-blue-500 outline-none" placeholder="Password login Pusaka Kemenag" />
        </div>
      </div>
      <div class="flex gap-1.5 mt-4">
        <button onclick={closeModal} disabled={busy} class="flex-1 px-3 py-1.5 border border-gray-300 rounded-md text-xs font-medium text-gray-700 hover:bg-gray-50 transition-colors disabled:opacity-50">Batal</button>
        <button onclick={save} disabled={busy}
                class="flex-1 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-xs font-medium transition-colors disabled:opacity-60">
          {#if saving}
            <Loading variant="button" size="sm" label="Menyimpan..." />
          {:else}
            Simpan
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}

<ConfirmDialog show={confirmShow} title={confirmTitle} message={confirmMessage}
               confirmLabel="Nonaktifkan" variant="danger"
               onConfirm={confirmAction} onCancel={() => confirmShow = false} />
