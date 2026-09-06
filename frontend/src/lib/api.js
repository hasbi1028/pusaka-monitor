// API Client for Pusaka Monitor — with stale-while-revalidate caching

const BASE = '';

// ─── Client-side cache (stale-while-revalidate) ─────────────────────────────
const _cache = new Map(); // key → { data, expiresAt }
const _inflight = new Map(); // key → Promise (dedup concurrent requests)
const CACHE_TTL = 15_000; // 15 seconds — fast but fresh enough for scrape data

function cacheKey(url) { return url; }

function getCached(url) {
  const k = cacheKey(url);
  const entry = _cache.get(k);
  if (!entry) return null;
  if (Date.now() < entry.expiresAt) return entry.data; // fresh
  return entry.data; // stale — return but revalidate in background
}

function setCache(url, data) {
  const k = cacheKey(url);
  _cache.set(k, { data, expiresAt: Date.now() + CACHE_TTL });
}

async function request(url, options = {}) {
  // For mutations (POST/PUT/DELETE), skip cache
  if (options.method && options.method !== 'GET') {
    const res = await fetch(BASE + url, {
      headers: { 'Content-Type': 'application/json', ...options.headers },
      ...options
    });
    return res.json();
  }

  // For GET requests: stale-while-revalidate
  const cached = getCached(url);
  const inflight = _inflight.get(url);

  if (inflight) return inflight;

  const promise = fetch(BASE + url, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options
  })
    .then(res => res.json())
    .then(data => {
      setCache(url, data);
      _inflight.delete(url);
      return data;
    })
    .catch(err => {
      _inflight.delete(url);
      // If network fails and we have stale data, return it
      if (cached) return cached;
      throw err;
    });

  _inflight.set(url, promise);
  return promise;
}

// Manual cache invalidation
export function invalidateCache(pattern) {
  for (const k of _cache.keys()) {
    if (!pattern || k.includes(pattern)) {
      _cache.delete(k);
    }
  }
}

// Auth
export const auth = {
  login: (username, password) => request('/api/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) }),
  register: (data) => request('/api/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  logout: () => request('/api/auth/logout', { method: 'POST' }),
  me: () => request('/api/auth/me'),
  resetPassword: (old_password, new_password) => request('/api/me/reset-password', { method: 'POST', body: JSON.stringify({ old_password, new_password }) })
};

// Dashboard
export const dashboard = {
  get: (tanggal) => request('/api/dashboard?tanggal=' + tanggal),
  getBulanPegawai: (bulan, tahun) => request('/api/dashboard/bulan-pegawai?bulan=' + bulan + '&tahun=' + tahun),
  getBulanDetail: (bulan, tahun) => request('/api/dashboard/bulan-detail?bulan=' + bulan + '&tahun=' + tahun),
  exportCSV: (bulan, tahun) => { window.location.href = '/api/admin/export-csv?bulan=' + bulan + '&tahun=' + tahun; }
};

// Pegawai
export const pegawai = {
  list: () => request('/api/pegawai'),
  create: (data) => request('/api/pegawai', { method: 'POST', body: JSON.stringify(data) }),
  update: (id, data) => request('/api/pegawai/' + id, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id) => request('/api/pegawai/' + id, { method: 'DELETE' })
};

// Scrape
export const scrape = {
  createJobs: (mode) => request('/api/scrape?mode=' + mode, { method: 'POST' }),
  status: () => request('/api/scrape/status'),
  retryFailed: () => request('/api/scrape/retry', { method: 'POST' }),
  cancelAll: () => request('/api/scrape/cancel-all', { method: 'POST' }),
  cancelOne: (id) => request('/api/scrape/' + id + '/cancel', { method: 'POST' }),
  scrapeOne: (nip) => request('/api/scrape/pegawai/' + nip, { method: 'POST' })
};

// Superadmin
export const superadmin = {
  listUsers: () => request('/api/superadmin/users'),
  resetPassword: (id, new_password) => request('/api/superadmin/users/' + id + '/reset-password', { method: 'POST', body: JSON.stringify({ new_password }) }),
  getConcurrency: () => request('/api/admin/concurrency'),
  setConcurrency: (n) => request('/api/admin/concurrency', { method: 'POST', body: JSON.stringify({ concurrency: n }) }),
  importPegawai: async (file) => {
    const formData = new FormData();
    formData.append('file', file);
    const res = await fetch('/api/admin/import-pegawai', { method: 'POST', body: formData });
    return res.json();
  },
  listInstansi: () => request('/api/superadmin/instansi'),
  updateInstansi: (id, data) => request('/api/superadmin/instansi/' + id, { method: 'PUT', body: JSON.stringify(data) })
};

// Approval
export const approval = {
  list: () => request('/superadmin/approval'),
  approve: (id) => request('/api/approval/' + id, { method: 'POST', body: JSON.stringify({ action: 'approve' }) }),
  reject: (id, reason) => request('/api/approval/' + id, { method: 'POST', body: JSON.stringify({ action: 'reject', reason }) })
};

// Instansi
export const instansi = {
  getSettings: () => request('/api/instansi/settings'),
  updateSettings: (data) => request('/api/instansi/settings', { method: 'PUT', body: JSON.stringify(data) })
};

// WA Groups
export const waGroups = {
  list: () => request('/api/admin/wa/groups')
};

// Recap
export const rekap = {
  kirimWA: (tanggal, instansiId) => {
    let url = '/api/rekap/kirim-wa?tanggal=' + (tanggal || '');
    if (instansiId) url += '&instansi_id=' + instansiId;
    return request(url, { method: 'POST' });
  },
  kirimTelegram: (tanggal, instansiId) => {
    let url = '/api/rekap/kirim-telegram?tanggal=' + (tanggal || '');
    if (instansiId) url += '&instansi_id=' + instansiId;
    return request(url, { method: 'POST' });
  },
  preview: (tanggal, instansiId) => {
    let url = '/api/admin/recap/preview?tanggal=' + (tanggal || '');
    if (instansiId) url += '&instansi_id=' + instansiId;
    return url;
  }
};

// ─── Prefetch utility (requestIdleCallback) ──────────────────────────────────
// Prefetch data during browser idle time — makes navigation feel instant
const _prefetched = new Set();
function _doPrefetch(urls) {
  urls.forEach(url => {
    if (!_prefetched.has(url)) {
      _prefetched.add(url);
      request(url); // populates cache
    }
  });
}
export function prefetch(urls) {
  if (!Array.isArray(urls) || urls.length === 0) return;
  if ('requestIdleCallback' in window) {
    requestIdleCallback(() => _doPrefetch(urls), { timeout: 2000 });
  } else {
    setTimeout(() => _doPrefetch(urls), 100);
  }
}
