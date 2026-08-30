// Toast store
import { writable } from 'svelte/store';

let id = 0;

function createToastStore() {
  const { subscribe, update } = writable([]);

  return {
    subscribe,
    show(message, type = 'success') {
      const toast = { id: ++id, message, type };
      update(toasts => [...toasts, toast]);
      setTimeout(() => {
        update(toasts => toasts.filter(t => t.id !== toast.id));
      }, 3000);
    },
    success(msg) { this.show(msg, 'success'); },
    error(msg) { this.show(msg, 'danger'); },
    warning(msg) { this.show(msg, 'warning'); }
  };
}

export const toasts = createToastStore();
