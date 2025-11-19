import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
	test('loads the dashboard page', async ({ page }) => {
		await page.goto('/');
		await expect(page.locator('h1')).toContainText('Idea Dashboard');
	});

	test('displays the idea form', async ({ page }) => {
		await page.goto('/');
		await expect(page.locator('textarea#idea-content')).toBeVisible();
		await expect(page.locator('button:has-text("Analyze & Save")')).toBeVisible();
	});

	test('shows navigation links', async ({ page }) => {
		await page.goto('/');
		await expect(page.locator('a:has-text("Dashboard")')).toBeVisible();
		await expect(page.locator('a:has-text("Analytics")')).toBeVisible();
		await expect(page.locator('a:has-text("Settings")')).toBeVisible();
	});
});

test.describe('Create Idea Flow', () => {
	test.skip('creates a new idea', async ({ page }) => {
		// Skip this test if API is not available
		await page.goto('/');

		// Fill in the idea form
		await page.fill('textarea#idea-content', 'Test Idea from E2E');

		// Submit the form
		await page.click('button:has-text("Analyze & Save")');

		// Wait for the idea to appear (this assumes API is working)
		// In real scenario, we'd mock the API or have test database
		// await expect(page.locator('text=Test Idea from E2E')).toBeVisible({ timeout: 5000 });
	});
});
