<script>
  import { goto } from '$app/navigation';

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

<div class="min-h-screen bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center p-4">
  <div class="w-full max-w-sm">
    <div class="bg-white rounded-2xl shadow-2xl p-8">
      <div class="text-center mb-6">
        <div class="w-16 h-16 bg-blue-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
          <i class="fa-solid fa-shield-halved text-blue-600 text-2xl"></i>
        </div>
        <h1 class="text-xl font-bold text-gray-900">Pusaka Monitor</h1>
        <p class="text-sm text-gray-500 mt-1">Masuk ke akun Anda</p>
      </div>
      
      {#if error}
        <div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700 flex items-center gap-2">
          <i class="fa-solid fa-circle-exclamation"></i>
          <span>{error}</span>
        </div>
      {/if}
      
      <form onsubmit={(e) => { e.preventDefault(); handleLogin(); }} class="space-y-4">
        <div>
          <label for="username" class="block text-sm font-medium text-gray-700 mb-1">Username</label>
          <input id="username" type="text" bind:value={username}
                 class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                 placeholder="Masukkan username" required />
        </div>
        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
          <input id="password" type="password" bind:value={password}
                 class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                 placeholder="Masukkan password" required />
        </div>
        <button type="submit" disabled={loading}
                class="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-medium py-2.5 px-4 rounded-lg text-sm transition-colors">
          {loading ? 'Masuk...' : 'Masuk'}
        </button>
      </form>
      
      <div class="text-center mt-4 text-sm text-gray-500">
        Belum punya akun?
        <a href="/register" class="text-blue-600 hover:text-blue-700 font-medium">Daftar</a>
      </div>
    </div>
    
    <p class="text-center text-xs text-blue-200 mt-4 px-4">
      <i class="fa-solid fa-circle-info"></i>
      Aplikasi ini <strong>HANYA membaca</strong> riwayat presensi dari Pusaka Kemenag.
    </p>
  </div>
</div>
