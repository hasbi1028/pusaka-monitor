<script>
  import Icon from '$lib/components/Icon.svelte';
  let { message = 'Gagal memuat data', onRetry = null, icon = 'triangle-exclamation' } = $props();
</script>

<div class="flex flex-col items-center justify-center py-10 text-center">
  <div class="w-10 h-10 rounded-full bg-red-50 flex items-center justify-center mb-3">
    <Icon name={icon} class="text-red-400 text-lg" />
  </div>
  <p class="text-xs text-gray-600 font-medium mb-1">{message}</p>
  <p class="text-[10px] text-gray-400 mb-3">Periksa koneksi internet atau coba lagi</p>
  {#if onRetry}
    <button onclick={onRetry}
            class="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium rounded-md transition-colors flex items-center gap-1.5">
      <Icon name="rotate-right" class="text-[10px]" />
      Coba Lagi
    </button>
  {/if}
</div>
