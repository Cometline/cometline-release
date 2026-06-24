import { test, expect } from '@playwright/test';
import { typeComposerMessage, visitReadyApp } from './fixtures/test-helpers';

test.describe('Composer', () => {
	test('sends a first message and navigates to the new session', async ({ page }) => {
		await visitReadyApp(page);

		await typeComposerMessage(page, 'Hello from Playwright');
		await page.getByRole('button', { name: 'Send' }).click();

		await expect(page).toHaveURL(/\/session\/e2e-session-1/, { timeout: 15_000 });
		await expect(page.getByText('Hello from Playwright')).toBeVisible({ timeout: 15_000 });
	});
});
