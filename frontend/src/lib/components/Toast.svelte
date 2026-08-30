<script>
  import { toasts } from '../stores/toast.js';
  
  const colors = {
    success: 'bg-green-50 text-green-800 border-green-300',
    danger: 'bg-red-50 text-red-800 border-red-300',
    warning: 'bg-yellow-50 text-yellow-800 border-yellow-300'
  };
  const icons = {
    success: '✅',
    danger: '❌',
    warning: '⚠️'
  };
</script>

<div class="fixed top-4 right-4 z-[9999] flex flex-col gap-2 max-w-xs">
  {#each $toasts as toast (toast.id)}
    <div class="toast-enter flex items-center gap-2 px-4 py-3 rounded-lg border text-sm font-medium shadow-lg {colors[toast.type] || colors.success}">
      <span>{icons[toast.type] || ''}</span>
      <span>{toast.message}</span>
    </div>
  {/each}
</div>
