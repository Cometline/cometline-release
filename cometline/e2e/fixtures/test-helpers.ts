import { expect, type Page } from '@playwright/test';
import { seedAppState } from './app-seed';
import { installCometMindMock } from './cometmind-mock';

export async function visitReadyApp(page: Page) {
	await seedAppState(page);
	await installCometMindMock(page);
	await page.goto('/');

	await expect(page.getByRole('textbox', { name: 'Message input' })).toBeVisible({
		timeout: 15_000
	});
	await expect(page.getByText('Starting CometMind…')).toHaveCount(0, { timeout: 15_000 });
	await expect(page.getByRole('heading', { name: 'Cannot reach CometMind' })).toHaveCount(0);
	await expect(page.getByRole('button', { name: 'Select model' })).not.toContainText(
		'No enabled models',
		{ timeout: 15_000 }
	);
}

export async function typeComposerMessage(page: Page, text: string) {
	const input = page.getByRole('textbox', { name: 'Message input' });
	await input.click();
	await input.fill(text);
	await expect(page.getByRole('button', { name: 'Send' })).toBeEnabled({ timeout: 15_000 });
}
