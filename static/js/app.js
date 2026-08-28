// Pusaka Monitor - Shared JavaScript

// Toggle Sidebar (Mobile)
function toggleSidebar() {
    const sidebar = document.getElementById('sidebar');
    const overlay = document.getElementById('sidebarOverlay');
    sidebar.classList.toggle('show');
    overlay.classList.toggle('show');
}

// Logout
async function logout() {
    await fetch('/api/auth/logout', { method: 'POST' });
    window.location.href = '/login';
}

// Load Dashboard
async function loadDashboard() {
    const tanggal = document.getElementById('datePicker')?.value;
    if (!tanggal) return;
    
    try {
        const res = await fetch(`/api/dashboard?tanggal=${tanggal}`);
        const data = await res.json();
        if (data.success) {
            document.getElementById('total').textContent = data.data.rekap.total || 0;
            document.getElementById('hadir').textContent = data.data.rekap.hadir || 0;
            document.getElementById('telat').textContent = data.data.rekap.terlambat || 0;
            document.getElementById('alfa').textContent = data.data.rekap.tidak_hadir || 0;

            const tbody = document.getElementById('absensiTable');
            const countBadge = document.getElementById('tableCount');
            
            if (data.data.absensi.length === 0) {
                tbody.innerHTML = '<tr><td colspan="4" class="text-center text-muted py-4"><i class="bi bi-inbox" style="font-size:2rem;"></i><br>Belum ada data presensi</td></tr>';
                countBadge.textContent = '0';
            } else {
                tbody.innerHTML = data.data.absensi.map(a => `
                    <tr>
                        <td>
                            <div class="fw-semibold">${a.nama}</div>
                            <small class="text-muted">${a.nip}</small>
                        </td>
                        <td class="text-center">${a.jam_masuk || '-'}</td>
                        <td class="text-center">${a.jam_pulang || '-'}</td>
                        <td class="text-center">
                            <span class="badge ${getStatusBadge(a.status)}">${a.status || '-'}</span>
                        </td>
                    </tr>
                `).join('');
                countBadge.textContent = data.data.absensi.length;
            }
        }
    } catch (err) {
        console.error('Error loading dashboard:', err);
    }
}

// Get status badge class
function getStatusBadge(status) {
    switch(status) {
        case 'Hadir': return 'bg-success';
        case 'Terlambat': return 'bg-danger';
        case 'Telat Ringan': return 'bg-warning text-dark';
        default: return 'bg-secondary';
    }
}

// Show loading
function showLoading(btn) {
    btn.disabled = true;
    btn.innerHTML = '<span class="spinner-border spinner-border-sm"></span> Memproses...';
}

// Hide loading
function hideLoading(btn, text) {
    btn.disabled = false;
    btn.innerHTML = text;
}

// Show toast notification
function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    toast.className = `alert alert-${type} position-fixed top-0 end-0 m-3`;
    toast.style.zIndex = '9999';
    toast.style.minWidth = '300px';
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
}
