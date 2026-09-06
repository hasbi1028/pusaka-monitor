<script>
  // Loading — reusable spinner / skeleton / overlay per Svelte runes best practices.
  // Props:
  //   variant = 'spinner' | 'skeleton' | 'overlay' | 'button'
  //   label   = teks di bawah spinner (default bullish)
  //   size    = 'sm' | 'md' | 'lg'  (ukuran spinner)
  //   rows    = jumlah baris skeleton (default 5)
  //   cols    = jumlah kolom skeleton (default 4)
  let { variant = 'spinner', label = 'Memuat data...', size = 'md', rows = 5, cols = 4 } = $props();

  const spinnerSize = {
    sm: 'w-5 h-5 border-2',
    md: 'w-8 h-8 border-3',
    lg: 'w-12 h-12 border-4'
  };
  let spinnerCls = $derived(`${spinnerSize[size]} border-blue-200 border-t-blue-600 rounded-full animate-spin`);

  // Cuplik suang — tombol kecil berputar untuk state "menyimpan/sedang memproses"
  let isInline = $derived(variant === 'button');
</script>

{#if variant === 'spinner'}
  <div class="flex flex-col items-center justify-center py-8" role="status" aria-live="polite">
    <div class="{spinnerCls}"></div>
    {#if label}
      <p class="mt-3 text-xs text-gray-500 flex items-center gap-1">
        {label}
      </p>
    {/if}
  </div>
{:else if variant === 'skeleton'}
  <div class="animate-pulse" role="status" aria-live="polite">
    <div class="grid gap-2 mb-2" style="grid-template-columns: repeat({cols}, minmax(0,1fr))">
      {#each Array(4) as _}
        <div class="bg-white rounded-lg border border-gray-200 p-2 shadow-sm">
          <div class="h-4 bg-gray-200 rounded w-10 mb-1"></div>
          <div class="h-2.5 bg-gray-100 rounded w-14"></div>
        </div>
      {/each}
    </div>
    <div class="bg-white rounded-lg border border-gray-200 shadow-sm p-3">
      <div class="h-3.5 bg-gray-200 rounded w-28 mb-3"></div>
      <div class="space-y-2">
        {#each Array(rows) as _}
          <div class="flex gap-3">
            <div class="h-3.5 bg-gray-100 rounded flex-1"></div>
            <div class="h-3.5 bg-gray-100 rounded w-14"></div>
            <div class="h-3.5 bg-gray-100 rounded w-14"></div>
          </div>
        {/each}
      </div>
    </div>
  </div>
{:else if variant === 'overlay'}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-sm" role="status" aria-live="polite">
    <div class="bg-white rounded-xl shadow-xl p-6 flex flex-col items-center" style="animation: pmPop 0.18s ease-out">
      <div class="{spinnerCls}"></div>
      {#if label}
        <p class="mt-3 text-sm font-medium text-gray-700">{label}</p>
      {/if}
    </div>
  </div>
{:else if isInline}
  <span class="inline-flex items-center gap-1.5" role="status" aria-live="polite">
    <div class="w-3.5 h-3.5 border-2 border-current/25 border-t-current rounded-full animate-spin"></div>
    {#if label}<span class="text-current">{label}</span>{/if}
  </span>
{/if}

<style>
  @keyframes pmPop {
    from { transform: scale(0.92); opacity: 0; }
    to   { transform: scale(1);   opacity: 1; }
  }
</style>