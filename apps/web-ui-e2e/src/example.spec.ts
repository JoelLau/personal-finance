import { test, expect } from '@playwright/test';

test('renders', async ({ page }) => {
  await page.goto('/');

  expect(await page.locator('html').innerText()).toContain(
    'dashboard-page works!',
  );
});
