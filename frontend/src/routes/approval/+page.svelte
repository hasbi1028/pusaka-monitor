<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import Loading from '$lib/components/Loading.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import { approval } from '$lib/api.js';
  import { toasts } from '$lib/stores/toast.js';
  import Icon from '$lib/components/Icon.svelte';

  let pending = $state([]);
  let recent = $state([]);
  let loading = $state(true);
  let busyId = $state('');

  // Confirm dialog state
  let confirmShow = $state(false);
  let confirmTitle = $state('');
  let confirmMessage = $state('');
  let confirmAction = $state(() => {});

  async function loadApproval() {
    loading = true;
    const res = await approval.list();
    if (res.success) {
      pending = res.data.pending;
      recent = res.data.recent;
    }
    loading = false;
  }

  // Optimistic approve — move from pending to recent instantly
  async function approve(id) {
    if (busyId) return;
    confirmTitle = 'Setujui Pendaftaran';
    confirmMessage = 'Setujui pendaftaran ini?';
    confirmAction = async () => {
      busyId = id;
      const item = pending.find(p => p.id === id);
      const prevPending = [...pending];
      const prevRecent = [...recent];
      pending = pending.filter(p => p.id !== id);
      if (item) recent = [{ ...item, status: 'approved' }, ...recent];
      try {
        const res = await approval.approve(id);
        if (res.success) toasts.success('Pendaftaran disetujui');
        else { pending = prevPending; recent = prevRecent; toasts.error(res.error || 'Gagal'); }
      } catch {
        pending = prevPending; recent = prevRecent;
        toasts.error('Gagal koneksi');
      }
      busyId = '';
    };
    confirmShow = true;
  }

  // Optimistic reject — move from pending to recent instantly
  async function reject(id) {
    if (busyId) return;
    confirmTitle = 'Tolak Pendaftaran';
    confirmMessage = 'Tolak pendaftaran ini?';
    confirmAction = async () => {
      busyId = id;
      const item = pending.find(p => p.id === id);
      const prevPending = [...pending];
      const prevRecent = [...recent];
      pending = pending.filter(p => p.id !== id);
      if (item) recent = [{ ...item, status: 'rejected' }, ...recent];
      try {
        const res = await approval.reject(id, '');
        if (res.success) toasts.error('Pendaftaran ditolak');
        else { pending = prevPending; recent = prevRecent; toasts.error(res.error || 'Gagal'); }
      } catch {
        pending = prevPending; recent = prevRecent;
        toasts.error('Gagal koneksi');
      }
      busyId = '';
    };
    confirmShow = true;
  }

  loadApproval();
</script>

<Layout title="Approval" activePage="settings">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Approval' }]} />
  <h3 class="text-sm font-semibold text-gray-500 mb-3"><Icon name="hourglass-half" class="mr-1" /> Menunggu Persetujuan</h3>
  {#if loading}
    <Loading variant="spinner" size="sm" label="Memuat pengajuan..." />
  {:else}
  <div class="space-y-3 mb-6">
    {#each pending as item (item.id)}
      <div class="bg-white rounded-xl border border-gray-200 shadow-sm p-4">
        <div class="flex items-start justify-between mb-3">
          <div>
            <h4 class="text-sm font-semibold text-gray-900">{item.nama}</h4>
            <p class="text-xs text-gray-500">{item.jns_instansi?.toUpperCase()} · {item.kabupaten}, {item.provinsi}</p>
            <p class="text-xs text-gray-400 mt-1">Username: {item.username || '-'} · WA: {item.telepon || '-'}</p>
          </div>
          <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600">Pending</span>
        </div>
        <div class="flex gap-2">
          <button onclick={() => approve(item.id)} disabled={busyId === item.id}
                  class="flex-1 px-3 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg text-xs font-medium transition-colors disabled:opacity-60">
            {#if busyId === item.id}
              <Loading variant="button" label="Menyetujui..." />
            {:else}
              <Icon name="check" class="mr-1" />Setujui
            {/if}
          </button>
          <button onclick={() => reject(item.id)} disabled={busyId === item.id}
                  class="flex-1 px-3 py-2 bg-red-50 hover:bg-red-100 text-red-700 rounded-lg text-xs font-medium transition-colors disabled:opacity-60">
            {#if busyId === item.id}
              <Loading variant="button" label="Menolak..." />
            {:else}
              <Icon name="xmark" class="mr-1" />Tolak
            {/if}
          </button>
        </div>
      </div>
    {:else}
      <div class="bg-green-50 border border-green-200 rounded-xl p-4 text-center text-green-700 text-sm">
        <Icon name="circle-check" class="mr-1" /> Tidak ada pendaftaran menunggu
      </div>
    {/each}
  </div>
  {/if}
  
  <hr class="border-gray-200 mb-4" />
  
  <h3 class="text-sm font-semibold text-gray-500 mb-3"><Icon name="clock-rotate-left" class="mr-1" /> Riwayat</h3>
  <div class="space-y-2">
    {#each recent as item (item.id)}
      <div class="bg-white rounded-xl border border-gray-200 shadow-sm px-4 py-2.5 flex items-center justify-between">
        <div>
          <span class="text-sm font-medium text-gray-900">{item.nama}</span>
          <span class="text-xs text-gray-400 ml-2">{item.kabupaten}</span>
        </div>
        <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
              class:bg-green-100={item.status==='approved'} class:text-green-700={item.status==='approved'}
              class:bg-red-100={item.status!=='approved'} class:text-red-700={item.status!=='approved'}>
          {item.status}
        </span>
      </div>
    {:else}
      <div class="text-center text-gray-400 text-sm py-3">Belum ada riwayat</div>
    {/each}
  </div>
</Layout>

<ConfirmDialog show={confirmShow} title={confirmTitle} message={confirmMessage}
               confirmLabel="Ya, Lanjut" variant="warning"
               onConfirm={confirmAction} onCancel={() => confirmShow = false} />
