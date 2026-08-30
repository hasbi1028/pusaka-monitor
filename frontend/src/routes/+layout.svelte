<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import '../app.css';

  let { children } = $props();

  const publicPaths = ['/login', '/register'];

  // Auth state using $state rune
  let user = $state(null);
  let loading = $state(true);
  let isLoggedIn = $state(false);

  async function checkAuth() {
    try {
      const res = await fetch('/api/auth/me');
      if (res.ok) {
        const data = await res.json();
        if (data.success) {
          user = data.data;
          isLoggedIn = true;
          loading = false;
          return;
        }
      }
      user = null;
      isLoggedIn = false;
      loading = false;
    } catch {
      user = null;
      isLoggedIn = false;
      loading = false;
    }
  }

  // Export for child components to use
  export function getAuth() {
    return { user, isLoggedIn, loading };
  }

  export function logout() {
    user = null;
    isLoggedIn = false;
  }

  onMount(() => {
    checkAuth();
  });

  // Redirect based on auth state
  $effect(() => {
    if (loading) return;

    const currentPath = $page.url.pathname;
    const isPublic = publicPaths.includes(currentPath);

    if (!isLoggedIn && !isPublic) {
      goto('/login');
    } else if (isLoggedIn && isPublic) {
      goto('/dashboard');
    }
  });
</script>

{#if loading}
  <div class="min-h-screen bg-gray-50 flex items-center justify-center">
    <div class="text-center">
      <div class="w-12 h-12 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin mx-auto mb-4"></div>
      <p class="text-sm text-gray-500">Memuat...</p>
    </div>
  </div>
{:else if isLoggedIn || publicPaths.includes($page.url.pathname)}
  {@render children()}
{:else}
  <div class="min-h-screen bg-gray-50 flex items-center justify-center">
    <div class="text-center">
      <p class="text-sm text-gray-500">Mengalihkan ke login...</p>
    </div>
  </div>
{/if}
