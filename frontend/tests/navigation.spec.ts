import { test, expect } from '@playwright/test';

/**
 * Helper: login via the UI and wait for dashboard.
 */
async function login(page: any, username = 'admin', password = 'admin123') {
  await page.goto('/#/login');
  await page.waitForSelector('input[placeholder="Masukkan username"]', { timeout: 10_000 });
  await page.fill('input[placeholder="Masukkan username"]', username);
  await page.fill('input[placeholder="Masukkan password"]', password);
  await page.click('button[type="submit"]');
  await page.waitForFunction(
    () => window.location.hash === '#/' || window.location.hash === '#',
    { timeout: 15_000 },
  );
}

test.describe('Navigation', () => {
  test('sidebar navigation works (desktop)', async ({ page }) => {
    // Use desktop viewport so sidebar is visible
    await page.setViewportSize({ width: 1280, height: 720 });
    await login(page);

    // Wait for dashboard to load
    await expect(page.locator('text=Total')).toBeVisible({ timeout: 10_000 });

    // Click "Pegawai" in the desktop sidebar
    // The desktop sidebar is the one with class "hidden lg:flex"
    const sidebar = page.locator('aside.hidden.lg\\:flex');
    await sidebar.locator('a:has-text("Pegawai")').click();

    // Should navigate to /pegawai
    await page.waitForFunction(
      () => window.location.hash === '#/pegawai',
      { timeout: 10_000 },
    );

    // Pegawai page should show "Daftar Pegawai"
    await expect(page.locator('text=Daftar Pegawai')).toBeVisible({ timeout: 10_000 });
  });

  test('sidebar navigation to multiple pages', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await login(page);
    await expect(page.locator('text=Total')).toBeVisible({ timeout: 10_000 });

    const sidebar = page.locator('aside.hidden.lg\\:flex');

    // Navigate to Scrape
    await sidebar.locator('a:has-text("Scrape")').click();
    await page.waitForFunction(
      () => window.location.hash === '#/scrape',
      { timeout: 10_000 },
    );

    // Navigate to Pengaturan
    await sidebar.locator('a:has-text("Pengaturan")').click();
    await page.waitForFunction(
      () => window.location.hash === '#/settings',
      { timeout: 10_000 },
    );

    // Navigate back to Beranda
    await sidebar.locator('a:has-text("Beranda")').click();
    await page.waitForFunction(
      () => window.location.hash === '#/' || window.location.hash === '#',
      { timeout: 10_000 },
    );

    await expect(page.locator('text=Total')).toBeVisible();
  });

  test('bottom nav works on mobile', async ({ page }) => {
    // Use mobile viewport
    await page.setViewportSize({ width: 375, height: 812 });
    await login(page);

    // Wait for dashboard
    await expect(page.locator('text=Total')).toBeVisible({ timeout: 10_000 });

    // Bottom nav should be visible on mobile (class: fixed bottom-0)
    const bottomNav = page.locator('nav.fixed.bottom-0');
    await expect(bottomNav).toBeVisible();

    // Click "Pegawai" in bottom nav
    await bottomNav.locator('a:has-text("Pegawai")').click();
    await page.waitForFunction(
      () => window.location.hash === '#/pegawai',
      { timeout: 10_000 },
    );
    await expect(page.locator('text=Daftar Pegawai')).toBeVisible({ timeout: 10_000 });

    // Click "Beranda" to go back
    await bottomNav.locator('a:has-text("Beranda")').click();
    await page.waitForFunction(
      () => window.location.hash === '#/' || window.location.hash === '#',
      { timeout: 10_000 },
    );
    await expect(page.locator('text=Total')).toBeVisible();
  });

  test('mobile drawer opens and navigates', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 812 });
    await login(page);
    await expect(page.locator('text=Total')).toBeVisible({ timeout: 10_000 });

    // The hamburger menu button is in the Topbar
    // Click it to open the mobile drawer
    const menuBtn = page.locator('button').filter({ has: page.locator('i.fa-bars') });
    await menuBtn.click();

    // The mobile drawer should now be visible (translate-x-0)
    const drawer = page.locator('aside.fixed.top-0.left-0.z-50');
    await expect(drawer).toHaveClass(/translate-x-0/);

    // Click "Pegawai" in the drawer
    await drawer.locator('a:has-text("Pegawai")').click();
    await page.waitForFunction(
      () => window.location.hash === '#/pegawai',
      { timeout: 10_000 },
    );

    await expect(page.locator('text=Daftar Pegawai')).toBeVisible({ timeout: 10_000 });
  });
});
