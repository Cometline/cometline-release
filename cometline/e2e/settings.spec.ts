import { test, expect } from '@playwright/test';
import { visitReadyApp } from './fixtures/test-helpers';

test.describe('Settings modal', () => {
	test('opens and closes from the sidebar', async ({ page }) => {
		await visitReadyApp(page);

		await page.getByRole('button', { name: 'Settings' }).click();
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();

		await dialog.getByRole('button', { name: 'Close settings' }).click();
		await expect(dialog).not.toBeVisible();
	});
});
