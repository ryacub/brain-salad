import { test, expect } from '@playwright/test';

test.describe('Settings Page', () => {
	test('loads the settings page', async ({ page }) => {
		await page.goto('/settings');
		await expect(page.locator('h1')).toContainText('Settings');
	});

	test('displays appearance settings', async ({ page }) => {
		await page.goto('/settings');
		await expect(page.locator('text=Dark Mode')).toBeVisible();
	});

	test('displays API configuration', async ({ page }) => {
		await page.goto('/settings');
		await expect(page.locator('text=API Configuration')).toBeVisible();
		await expect(page.locator('input#api-url')).toBeVisible();
	});

	test('displays about section', async ({ page }) => {
		await page.goto('/settings');
		await expect(page.locator('text=About Telos Idea Matrix')).toBeVisible();
	});
});
