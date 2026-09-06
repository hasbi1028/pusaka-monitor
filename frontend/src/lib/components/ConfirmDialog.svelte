<script>
  let { show = false, title = 'Konfirmasi', message = '', confirmLabel = 'Ya', cancelLabel = 'Batal', variant = 'danger', onConfirm = () => {}, onCancel = () => {} } = $props();

  const variantClasses = {
    danger: 'bg-red-600 hover:bg-red-700',
    warning: 'bg-yellow-500 hover:bg-yellow-600',
    primary: 'bg-blue-600 hover:bg-blue-700'
  };
</script>

{#if show}
  <div class="fixed inset-0 z-[9999] flex items-center justify-center bg-black/50 backdrop-blur-sm" role="dialog" aria-modal="true" tabindex="-1"
       onkeydown={(e) => { if (e.key === 'Escape') onCancel(); }}>
    <div class="bg-white rounded-xl shadow-2xl w-full max-w-sm mx-4 overflow-hidden"
         style="animation: confirmPop 0.15s ease-out">
      <div class="px-5 pt-5 pb-3">
        <h3 class="text-sm font-semibold text-gray-900">{title}</h3>
        <p class="mt-1.5 text-xs text-gray-600 leading-relaxed">{message}</p>
      </div>
      <div class="px-5 pb-4 flex items-center justify-end gap-2">
        <button onclick={onCancel}
                class="px-3 py-1.5 text-xs font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md transition-colors">
          {cancelLabel}
        </button>
        <button onclick={() => { onConfirm(); onCancel(); }}
                class="px-3 py-1.5 text-xs font-medium text-white {variantClasses[variant]} rounded-md transition-colors">
          {confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  @keyframes confirmPop {
    from { opacity: 0; transform: scale(0.95); }
    to { opacity: 1; transform: scale(1); }
  }
</style>
