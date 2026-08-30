<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import { superadmin } from '$lib/api.js';
  import { toasts } from '$lib/stores/toast.js';

  let users = $state([]);

  async function loadUsers() {
    const res = await superadmin.listUsers();
    if (res.success) users = res.data;
  }

  function resetUser(id, username) {
    const pwd = prompt('Password baru untuk ' + username + ' (minimal 4 karakter):');
    if (!pwd || pwd.length < 4) { if (pwd !== null) toasts.error('Minimal 4 karakter'); return; }
    superadmin.resetPassword(id, pwd).then(d => {
      if (d.success) toasts.success('Password ' + username + ' direset');
      else toasts.error(d.error || 'Gagal');
    });
  }

  loadUsers();
</script>

<Layout title="User Management" activePage="users">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Users' }]} />
  <div class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
    <div class="px-4 py-2.5 border-b border-gray-200 text-sm font-semibold text-gray-700">
      <i class="fa-solid fa-users mr-1"></i> Daftar User
    </div>
    <div class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 text-gray-500 text-xs uppercase">
          <tr><th class="px-4 py-2 text-left">Username</th><th class="px-4 py-2 text-center">Role</th><th class="px-4 py-2 text-left">Instansi</th><th class="px-4 py-2 text-right">Aksi</th></tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          {#each users as u}
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
                <button onclick={() => resetUser(u.id, u.username)}
                        class="px-3 py-1 border border-gray-200 rounded-lg text-xs font-medium hover:bg-gray-50 transition-colors">
                  <i class="fa-solid fa-key mr-1"></i>Reset
                </button>
              </td>
            </tr>
          {:else}
            <tr><td colspan="4" class="px-4 py-6 text-center text-gray-400 text-sm">Memuat...</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</Layout>
