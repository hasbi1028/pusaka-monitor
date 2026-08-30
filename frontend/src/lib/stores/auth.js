// Auth store — Svelte 5 compatible
import { writable } from 'svelte/store';

function createAuthStore() {
  const { subscribe, set, update } = writable({
    user: null,
    loading: true,
    isLoggedIn: false
  });

  return {
    subscribe,
    async checkAuth() {
      try {
        const res = await fetch('/api/auth/me');
        if (res.ok) {
          const data = await res.json();
          if (data.success) {
            set({ user: data.data, loading: false, isLoggedIn: true });
            return;
          }
        }
        set({ user: null, loading: false, isLoggedIn: false });
      } catch {
        set({ user: null, loading: false, isLoggedIn: false });
      }
    },
    setUser(user) {
      update(s => ({ ...s, user, isLoggedIn: !!user }));
    },
    logout() {
      set({ user: null, loading: false, isLoggedIn: false });
    }
  };
}

export const auth = createAuthStore();
