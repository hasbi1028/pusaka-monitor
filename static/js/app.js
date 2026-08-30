// Pusaka Monitor — Shared JavaScript (Flowbite + Vanilla JS)

// ========== TOAST ==========
function showToast(message, type = 'success') {
    const colors = {
        success: 'text-green-800 bg-green-50 border-green-300',
        danger:  'text-red-800 bg-red-50 border-red-300',
        warning: 'text-yellow-800 bg-yellow-50 border-yellow-300',
        info:    'text-blue-800 bg-blue-50 border-blue-300'
    };
    const icons = {
        success: '✅',
        danger:  '❌',
        warning: '⚠️',
        info:    'ℹ️'
    };
    const container = document.getElementById('toast-container');
    if (!container) return;
    const toast = document.createElement('div');
    toast.className = `flex items-center gap-2 px-4 py-3 rounded-lg border text-sm font-medium shadow-lg transition-all duration-300 opacity-0 translate-y-2 ${colors[type] || colors.success}`;
    toast.innerHTML = `<span>${icons[type] || ''}</span><span>${message}</span>`;
    container.appendChild(toast);
    requestAnimationFrame(() => { toast.classList.remove('opacity-0', 'translate-y-2'); });
    setTimeout(() => {
        toast.classList.add('opacity-0', 'translate-y-2');
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// ========== SIDEBAR TOGGLE (mobile) ==========
function toggleDrawer() {
    const drawer = document.getElementById('sidebar-drawer');
    const backdrop = document.getElementById('sidebar-backdrop');
    if (drawer && backdrop) {
        drawer.classList.toggle('translate-x-0');
        drawer.classList.toggle('-translate-x-full');
        backdrop.classList.toggle('hidden');
    }
}

function closeDrawer() {
    const drawer = document.getElementById('sidebar-drawer');
    const backdrop = document.getElementById('sidebar-backdrop');
    if (drawer) { drawer.classList.add('-translate-x-full'); drawer.classList.remove('translate-x-0'); }
    if (backdrop) backdrop.classList.add('hidden');
}

// ========== LOGOUT ==========
async function logout() {
    try { await fetch('/api/auth/logout', { method: 'POST' }); } catch {}
    window.location.href = '/login';
}

// ========== LOADING BUTTON ==========
function showLoading(btn) {
    btn.disabled = true;
    btn.dataset.original = btn.innerHTML;
    btn.innerHTML = '<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg> Memproses...';
}
function hideLoading(btn) {
    btn.disabled = false;
    if (btn.dataset.original) btn.innerHTML = btn.dataset.original;
}
