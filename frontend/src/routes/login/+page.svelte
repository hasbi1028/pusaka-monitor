<script>
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/Icon.svelte';

  let username = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleLogin() {
    loading = true;
    error = '';
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });
      const data = await res.json();
      if (data.success) {
        // Full reload so layout re-checks auth state
        window.location.href = '/dashboard';
      } else {
        error = data.error || 'Login gagal';
      }
    } catch (e) {
      error = 'Koneksi gagal';
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center p-3">
  <div class="w-full max-w-xs">
    <div class="bg-white rounded-xl shadow-2xl p-5">
      <div class="text-center mb-4">
        <div class="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center mx-auto mb-2">
          <Icon name="shield-halved" class="text-blue-600 text-lg" />
        </div>
        <h1 class="text-base font-bold text-gray-900">Pusaka Monitor</h1>
        <p class="text-[10px] text-gray-500 mt-0.5">Masuk ke akun Anda</p>
      </div>
      
      {#if error}
        <div class="mb-3 p-2 bg-red-50 border border-red-200 rounded-md text-[10px] text-red-700 flex items-center gap-1.5">
          <Icon name="circle-exclamation" />
          <span>{error}</span>
        </div>
      {/if}
      
      <form onsubmit={(e) => { e.preventDefault(); handleLogin(); }} class="space-y-2.5">
        <div>
          <label for="username" class="block text-[10px] font-medium text-gray-700 mb-0.5">Username</label>
          <input id="username" type="text" bind:value={username}
                 class="w-full px-2 py-1.5 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                 placeholder="Masukkan username" required />
        </div>
        <div>
          <label for="password" class="block text-[10px] font-medium text-gray-700 mb-0.5">Password</label>
          <input id="password" type="password" bind:value={password}
                 class="w-full px-2 py-1.5 border border-gray-300 rounded-md text-xs focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                 placeholder="Masukkan password" required />
        </div>
        <button type="submit" disabled={loading}
                class="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-medium py-1.5 px-3 rounded-md text-xs transition-colors">
          {loading ? 'Masuk...' : 'Masuk'}
        </button>
      </form>
      
      <div class="text-center mt-3 text-[10px] text-gray-500">
        Belum punya akun?
        <a href="/register" class="text-blue-600 hover:text-blue-700 font-medium">Daftar</a>
      </div>
    </div>
    
    <p class="text-center text-[9px] text-blue-200 mt-2 px-3 leading-relaxed">
      <Icon name="shield-halved" />
      Aplikasi <strong>READ-ONLY</strong>: hanya membaca riwayat Pusaka Kemenag.<br />
      Tidak membuat / mengubah absen — bukan alat absen fiktif.
    </p>
  </div>
</div>
