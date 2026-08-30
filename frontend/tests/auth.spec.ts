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
  // Wait for redirect to dashboard (hash route changes to /)
  await page.waitForFunction(
    () => window.location.hash === '#/' || window.location.hash === '#',
    { timeout: 15_000 },
  );
}

test.describe('Auth Flow', () => {
  test('shows login page when not authenticated', async ({ page }) => {
    await page.goto('/#/login');

    // The login form should be visible
    await expect(page.locator('h1')).toHaveText('Pusaka Monitor');
    await expect(page.locator('input[placeholder="Masukkan username"]')).toBeVisible();
    await expect(page.locator('input[placeholder="Masukkan password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toHaveText('Masuk');
  });

  test('login with valid credentials redirects to dashboard', async ({ page }) => {
    await login(page);

    // Should be on the dashboard now
    const hash = page.url().split('#')[1] || '';
    expect(hash === '/' || hash === '').toBeTruthy();

    // Dashboard should have stat cards
    await expect(page.locator('text=Total')).toBeVisible({ timeout: 10_000 });
    await expect(page.locator('text=Hadir')).toBeVisible();
  });

  test('accessing protected route without auth redirects to login', async ({ page }) => {
    // Clear any auth cookies by visiting the app fresh
    await page.goto('/#/dashboard');

    // The app should redirect unauthenticated users to /login
    await page.waitForFunction(
      () => window.location.hash === '#/login',
      { timeout: 15_000 },
    );

    // Verify we see the login form
    await expect(page.locator('input[placeholder="Masukkan username"]')).toBeVisible();
  });

  test('logout redirects to login', async ({ page }) => {
    await login(page);

    // Wait for dashboard to load
    await expect(page.locator('text=Total')).toBeVisible({ timeout: 10_000 });

    // Click the "Keluar" (logout) button in the desktop sidebar
    // Desktop sidebar is visible at lg breakpoint (1024px+)
    await page.setViewportSize({ width: 1280, height: 720 });
    const keluarBtn = page.locator('button:has-text("Keluar")').first();
    await keluarBtn.click();

    // Should redirect to login
    await page.waitForFunction(
      () => window.location.hash === '#/login',
      { timeout: 15_000 },
    );

    await expect(page.locator('input[placeholder="Masukkan username"]')).toBeVisible();
  });

  test('logged in user on /login redirects to dashboard', async ({ page }) => {
    await login(page);

    // Wait for dashboard
    await expect(page.locator('text=Total')).toBeVisible({ timeout: 10_000 });

    // Now navigate to /login — should redirect back to dashboard
    await page.goto('/#/login');
    await page.waitForFunction(
      () => window.location.hash === '#/' || window.location.hash === '#',
      { timeout: 15_000 },
    );

    // Dashboard elements should still be visible
    await expect(page.locator('text=Total')).toBeVisible();
  });

  test('register page is accessible without auth', async ({ page }) => {
    await page.goto('/#/register');

    // Register form should be visible
    await expect(page.locator('h1')).toHaveText('Daftar Akun Baru');
    await expect(page.locator('input').first()).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toHaveText('Daftar');
  });
});
