import { test, expect } from '@playwright/test';
import { seedAppState } from './fixtures/app-seed';
import { installCometMindMock } from './fixtures/cometmind-mock';

test.describe('Runtime overlay', () => {
	test('shows an error when CometMind is unreachable', async ({ page }) => {
		await seedAppState(page);
		await installCometMindMock(page, { healthy: false });
		await page.goto('/');

		await expect(page.getByRole('alert')).toBeVisible({ timeout: 15_000 });
		await expect(page.getByRole('heading', { name: 'Cannot reach CometMind' })).toBeVisible();
		await expect(page.getByRole('button', { name: /Retry connection/i })).toBeVisible();
	});
});
