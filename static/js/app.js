// Pusaka Monitor - Shared JavaScript (Basecoat UI)

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

// Show loading on button
function showLoading(btn) {
    btn.disabled = true;
    btn.dataset.original = btn.innerHTML;
    btn.innerHTML = '<i class="bi bi-arrow-repeat animate-spin"></i> Memproses...';
}

// Hide loading
function hideLoading(btn) {
    btn.disabled = false;
    if (btn.dataset.original) btn.innerHTML = btn.dataset.original;
}

// Show toast notification (Basecoat alert, fixed top-right)
function showToast(message, type = 'success') {
    const variant = type === 'danger' ? 'destructive' : type === 'warning' ? 'secondary' : 'default';
    const toast = document.createElement('div');
    toast.className = 'alert';
    if (variant !== 'default') toast.setAttribute('data-variant', variant);
    toast.style.position = 'fixed';
    toast.style.top = '36px';
    toast.style.right = '12px';
    toast.style.zIndex = '9999';
    toast.style.minWidth = '280px';
    toast.style.maxWidth = '90vw';
    toast.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)';
    toast.innerHTML = `<section>${message}</section>`;
    document.body.appendChild(toast);
    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transition = 'opacity 0.3s';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}
