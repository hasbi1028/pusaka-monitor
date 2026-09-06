<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import Loading from '$lib/components/Loading.svelte';
  import { superadmin } from '$lib/api.js';
  import { toasts } from '$lib/stores/toast.js';
  import Icon from '$lib/components/Icon.svelte';

  let users = $state([]);
  let loading = $state(true);

  // Reset password modal
  let showResetModal = $state(false);
  let resetUserId = $state('');
  let resetUsername = $state('');
  let resetPassword = $state('');
  let resetBusy = $state(false);

  async function loadUsers() {
    loading = true;
    const res = await superadmin.listUsers();
    if (res.success) users = res.data;
    loading = false;
  }

  function openReset(id, username) {
    resetUserId = id;
    resetUsername = username;
    resetPassword = '';
    showResetModal = true;
  }

  function closeReset() {
    showResetModal = false;
    resetUserId = '';
    resetUsername = '';
    resetPassword = '';
  }

  async function doReset() {
    if (resetPassword.length < 4) { toasts.error('Minimal 4 karakter'); return; }
    resetBusy = true;
    const res = await superadmin.resetPassword(resetUserId, resetPassword);
    resetBusy = false;
    if (res.success) { toasts.success('Password ' + resetUsername + ' direset'); closeReset(); }
    else toasts.error(res.error || 'Gagal');
  }

  loadUsers();
</script>

<Layout title="User Management" activePage="users">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Users' }]} />
  <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
    <div class="px-4 py-2.5 border-b border-gray-200 text-sm font-semibold text-gray-700">
      <Icon name="users" class="mr-1" /> Daftar User
    </div>
    <div class="overflow-x-auto">
    {#if loading}
      <Loading variant="spinner" label="Memuat daftar user..." />
    {:else}
      <table class="w-full text-sm">
        <thead class="bg-gray-50 text-gray-500 text-xs uppercase">
          <tr><th class="px-4 py-2 text-left">Username</th><th class="px-4 py-2 text-center">Role</th><th class="px-4 py-2 text-left">Instansi</th><th class="px-4 py-2 text-right">Aksi</th></tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          {#each users as u (u.id)}
            <tr class="hover:bg-gray-50">
              <td class="px-4 py-2 font-medium text-gray-900">{u.username}</td>
              <td class="px-4 py-2 text-center">
                <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
                      class:bg-red-100={u.role==='superadmin'} class:text-red-700={u.role==='superadmin'}
                      class:bg-gray-100={u.role!=='superadmin'} class:text-gray-600={u.role!=='superadmin'}>
                  {u.role}
                </span>
              </td>
              <td class="px-4 py-2 text-gray-500 text-xs">{u.instansi_nama || '-'}</td>
              <td class="px-4 py-2 text-right">
                <button onclick={() => openReset(u.id, u.username)}
                        class="px-3 py-1 border border-gray-200 rounded-lg text-xs font-medium hover:bg-gray-50 transition-colors">
                  <Icon name="key" class="mr-1" />Reset
                </button>
              </td>
            </tr>
          {:else}
            <tr><td colspan="4" class="px-4 py-6 text-center text-gray-400 text-sm">Tidak ada user</td></tr>
          {/each}
        </tbody>
      </table>
    {/if}
    </div>
  </div>
</Layout>

{#if showResetModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-3">
    <div class="absolute inset-0 bg-black/50" onclick={closeReset} role="presentation"></div>
    <div class="bg-white rounded-xl shadow-xl w-full max-w-sm p-4 relative z-10">
      <button onclick={closeReset} class="absolute top-2 right-2 p-0.5 rounded-md hover:bg-gray-100 text-gray-400" aria-label="Tutup">
        <Icon name="xmark" class="text-xs" />
      </button>
      <h3 class="text-sm font-semibold text-gray-900 mb-0.5">Reset Password</h3>
      <p class="text-[10px] text-gray-500 mb-3">Password baru untuk <strong>{resetUsername}</strong></p>
      <input type="password" bind:value={resetPassword} placeholder="Minimal 4 karakter"
             class="w-full px-2 py-1.5 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-blue-500 outline-none mb-3"
             onkeydown={(e) => { if (e.key === 'Enter') doReset(); }} />
      <div class="flex gap-1.5">
        <button onclick={closeReset} disabled={resetBusy}
                class="flex-1 px-3 py-1.5 border border-gray-300 rounded-md text-xs font-medium text-gray-700 hover:bg-gray-50 transition-colors disabled:opacity-50">Batal</button>
        <button onclick={doReset} disabled={resetBusy}
                class="flex-1 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-xs font-medium transition-colors disabled:opacity-60">
          {#if resetBusy}
            <span class="inline-flex items-center gap-1"><span class="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"></span> Menyimpan...</span>
          {:else}
            Reset
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}
