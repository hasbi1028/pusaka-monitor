<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import { toasts } from '$lib/stores/toast.js';
  import { onMount } from 'svelte';

  let username = $state('-');
  let role = $state('-');
  let oldPwd = $state('');
  let newPwd = $state('');

  onMount(async () => {
    try {
      const res = await fetch('/api/auth/me');
      const data = await res.json();
      if (data.success) {
        username = data.data.username || '-';
        role = data.data.role || '-';
      }
    } catch {}
  });

  async function resetOwn() {
    if (!oldPwd || !newPwd) { toasts.error('Isi password lama & baru'); return; }
    const res = await fetch('/api/me/reset-password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ old_password: oldPwd, new_password: newPwd })
    });
    const d = await res.json();
    if (d.success) { toasts.success('Password berhasil diubah'); oldPwd = ''; newPwd = ''; }
    else toasts.error(d.error || 'Gagal');
  }
</script>

<Layout title="Profil" activePage="profil">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Profil' }]} />
  <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
    <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700">
      <i class="fa-solid fa-circle-user mr-0.5"></i> Akun Saya
    </div>
    <div class="p-2.5 space-y-1">
      <div class="flex justify-between text-xs">
        <span class="text-gray-500">Username</span>
        <span class="font-semibold text-gray-900">{username}</span>
      </div>
      <div class="flex justify-between text-xs">
        <span class="text-gray-500">Role</span>
        <span class="text-gray-900">{role}</span>
      </div>
    </div>
  </div>
  
  <div class="bg-white rounded-lg border border-gray-200 shadow-sm">
    <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700">
      <i class="fa-solid fa-key mr-0.5"></i> Ganti Password
    </div>
    <div class="p-2.5 space-y-2">
      <div>
        <label for="old-pwd" class="block text-[10px] font-medium text-gray-700 mb-0.5">Password Lama</label>
        <input id="old-pwd" type="password" bind:value={oldPwd} class="w-full px-2 py-1.5 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-blue-500 outline-none" placeholder="Password saat ini" />
      </div>
      <div>
        <label for="new-pwd" class="block text-[10px] font-medium text-gray-700 mb-0.5">Password Baru</label>
        <input id="new-pwd" type="password" bind:value={newPwd} class="w-full px-2 py-1.5 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-blue-500 outline-none" placeholder="Minimal 4 karakter" />
      </div>
      <button onclick={resetOwn}
              class="w-full px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-xs font-medium transition-colors">
        <i class="fa-solid fa-check mr-0.5"></i> Simpan Password Baru
      </button>
    </div>
  </div>
</Layout>
