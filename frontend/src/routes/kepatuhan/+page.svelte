<script>
  import Layout from '$lib/components/Layout.svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import Icon from '$lib/components/Icon.svelte';
</script>

<Layout title="Kepatuhan" activePage="kepatuhan">
  <Breadcrumb items={[{ label: 'Beranda', href: '/dashboard' }, { label: 'Kepatuhan' }]} />

  <div class="bg-green-50 border border-green-200 rounded-lg p-3 mb-2 text-center">
    <div class="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-1.5">
      <Icon name="shield-halved" class="text-green-600" />
    </div>
    <h2 class="text-sm font-bold text-green-800">Aplikasi Read-Only</h2>
    <p class="text-xs text-green-700 mt-0.5">Pusaka Monitor <strong>HANYA membaca</strong> data — tidak membuat, tidak mengubah, tidak memalsukan absen.</p>
  </div>

  <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
    <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700">
      <Icon name="list-check" class="mr-0.5" /> Yang Dilakukan Aplikasi
    </div>
    <ul class="p-3 space-y-1.5 text-xs text-gray-600">
      <li class="flex gap-1.5"><Icon name="check" class="text-green-500 mt-0.5" /> Login dengan NIP + password milik pegawai sendiri ke Pusaka Kemenag</li>
      <li class="flex gap-1.5"><Icon name="check" class="text-green-500 mt-0.5" /> Membaca (GET) halaman riwayat presensi bulan berjalan</li>
      <li class="flex gap-1.5"><Icon name="check" class="text-green-500 mt-0.5" /> Menyalin tampilannya ke database lokal untuk rekap dan laporan</li>
      <li class="flex gap-1.5"><Icon name="check" class="text-green-500 mt-0.5" /> Menampilkan ulang data tersebut di dashboard, laporan, dan gambar rekap WA</li>
    </ul>
  </div>

  <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
    <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700">
      <Icon name="ban" class="mr-0.5" /> Yang TIDAK PERNAH Dilakukan
    </div>
    <ul class="p-3 space-y-1.5 text-xs text-gray-600">
      <li class="flex gap-1.5"><Icon name="xmark" class="text-red-500 mt-0.5" /> Mengirim absen masuk/pulang ke Pusaka (tidak ada endpoint tulis yang disentuh)</li>
      <li class="flex gap-1.5"><Icon name="xmark" class="text-red-500 mt-0.5" /> Membuat, mengubah, atau menghapus data apa pun di server Pusaka</li>
      <li class="flex gap-1.5"><Icon name="xmark" class="text-red-500 mt-0.5" /> Memalsukan jam kehadiran — semua angka berasal dari riwayat resmi Pusaka</li>
      <li class="flex gap-1.5"><Icon name="xmark" class="text-red-500 mt-0.5" /> Membagikan password Pusaka pegawai ke pihak lain (tersimpan terenkripsi AES-256)</li>
    </ul>
  </div>

  <div class="bg-white rounded-lg border border-gray-200 shadow-sm mb-2">
    <div class="px-3 py-1.5 border-b border-gray-200 text-xs font-semibold text-gray-700">
      <Icon name="magnifying-glass" class="mr-0.5" /> Cara Memverifikasi
    </div>
    <div class="p-3 text-xs text-gray-600 space-y-1.5">
      <p>1. Bandingkan angka di aplikasi dengan halaman riwayat di Pusaka Kemenag — harus identik.</p>
      <p>2. Setiap pengambilan tercatat di log server (waktu, NIP, hasil) dan statusnya terlihat di halaman Scrape.</p>
      <p>3. Kode sumber endpoint yang diakses hanya: <code class="bg-gray-100 px-1 rounded text-[10px]">GET /api/auth/csrf</code>, <code class="bg-gray-100 px-1 rounded text-[10px]">GET /api/auth/session</code>, <code class="bg-gray-100 px-1 rounded text-[10px]">GET riwayat-presensi</code>, dan <code class="bg-gray-100 px-1 rounded text-[10px]">POST login</code> (autentikasi itu sendiri).</p>
    </div>
  </div>

  <div class="bg-amber-50 border border-amber-200 rounded-lg p-3 text-[11px] text-amber-800">
    <Icon name="triangle-exclamation" class="mr-0.5" />
    Data di aplikasi ini adalah <strong>salinan tampilan</strong> untuk monitoring internal, bukan pengganti absensi resmi. Kehadiran resmi tetap tercatat di Pusaka Kemenag melalui mekanisme resmi (hadir fisik / aplikasi yang berwenang).
  </div>
</Layout>
