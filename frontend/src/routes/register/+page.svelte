<script>
  import { goto } from '$app/navigation';
  import { toasts } from '$lib/stores/toast.js';

  let form = $state({
    nama_instansi: '',
    jns_instansi: 'madrasah',
    kabupaten: '',
    provinsi: 'Sulawesi Tenggara',
    telepon: '',
    username: '',
    password: ''
  });
  let error = $state('');
  let success = $state('');
  let loading = $state(false);

  async function handleRegister() {
    loading = true;
    error = '';
    success = '';
    try {
      const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form)
      });
      const data = await res.json();
      if (data.success) {
        success = 'Pendaftaran berhasil! Menunggu persetujuan admin.';
        setTimeout(() => goto('/login'), 2000);
      } else {
        error = data.error || 'Gagal mendaftar';
      }
    } catch (e) {
      error = 'Koneksi gagal';
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center p-4">
  <div class="w-full max-w-md">
    <div class="bg-white rounded-2xl shadow-2xl p-8">
      <div class="text-center mb-6">
        <div class="w-16 h-16 bg-blue-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
          <i class="fa-solid fa-user-plus text-blue-600 text-2xl"></i>
        </div>
        <h1 class="text-xl font-bold text-gray-900">Daftar Akun Baru</h1>
        <p class="text-sm text-gray-500 mt-1">Isi data instansi Anda</p>
      </div>
      
      {#if error}
        <div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700 flex items-center gap-2">
          <i class="fa-solid fa-circle-exclamation"></i><span>{error}</span>
        </div>
      {/if}
      {#if success}
        <div class="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-700 flex items-center gap-2">
          <i class="fa-solid fa-circle-check"></i><span>{success}</span>
        </div>
      {/if}
      
      <form onsubmit={(e) => { e.preventDefault(); handleRegister(); }} class="space-y-3">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Nama Instansi <span class="text-red-500">*</span></label>
          <input type="text" bind:value={form.nama_instansi} class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" required />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Jenis Instansi <span class="text-red-500">*</span></label>
          <select bind:value={form.jns_instansi} class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm bg-white focus:ring-2 focus:ring-blue-500 outline-none">
            <option value="madrasah">Madrasah</option>
            <option value="kua">KUA</option>
            <option value="kemenag">Kantor Kemenag</option>
          </select>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Kabupaten <span class="text-red-500">*</span></label>
            <input type="text" bind:value={form.kabupaten} class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" required />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Provinsi <span class="text-red-500">*</span></label>
            <input type="text" bind:value={form.provinsi} class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" required />
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Telepon</label>
          <input type="text" bind:value={form.telepon} class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Username <span class="text-red-500">*</span></label>
          <input type="text" bind:value={form.username} class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" required />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Password <span class="text-red-500">*</span></label>
          <input type="password" bind:value={form.password} class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" required />
        </div>
        <button type="submit" disabled={loading}
                class="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-medium py-2.5 px-4 rounded-lg text-sm transition-colors">
          {loading ? 'Mendaftar...' : 'Daftar'}
        </button>
      </form>
      
      <div class="text-center mt-4 text-sm text-gray-500">
        Sudah punya akun?
        <a href="/login" class="text-blue-600 hover:text-blue-700 font-medium">Masuk</a>
      </div>
    </div>
  </div>
</div>
