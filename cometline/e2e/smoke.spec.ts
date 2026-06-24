import { test, expect } from '@playwright/test';
import { visitReadyApp } from './fixtures/test-helpers';

test('home page loads with a ready composer', async ({ page }) => {
	await visitReadyApp(page);
	await expect(page.getByRole('textbox', { name: 'Message input' })).toBeVisible();
});
