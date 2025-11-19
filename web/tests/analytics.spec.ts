import { test, expect } from '@playwright/test';

test.describe('Analytics Page', () => {
	test('loads the analytics page', async ({ page }) => {
		await page.goto('/analytics');
		await expect(page.locator('h1')).toContainText('Analytics Dashboard');
	});

	test('displays analytics description', async ({ page }) => {
		await page.goto('/analytics');
		await expect(page.locator('text=Track your idea patterns')).toBeVisible();
	});
});
