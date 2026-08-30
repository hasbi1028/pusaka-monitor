// API Client for Pusaka Monitor

const BASE = '';

async function request(url, options = {}) {
  const res = await fetch(BASE + url, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options
  });
  return res.json();
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
  }
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
