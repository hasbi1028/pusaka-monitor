<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { auth } from '$lib/stores/auth.js';
  import '../app.css';

  let { children } = $props();

  const publicPaths = ['/login', '/register'];

  let user = $state(null);
  let loading = $state(true);
  let isLoggedIn = $state(false);

  onMount(() => {
    // Use the auth store — single source of truth
    const unsub = auth.subscribe(s => {
      user = s.user;
      isLoggedIn = s.isLoggedIn;
      loading = s.loading;
    });

    auth.checkAuth();

    return unsub;
  });

  // Export for child components to use
  export function getAuth() {
    return { user, isLoggedIn, loading };
  }

  export function logout() {
    auth.logout();
  }

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
