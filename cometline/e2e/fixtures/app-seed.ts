import type { Page } from '@playwright/test';
import { defaultSettings } from '../../src/lib/settings/schema.ts';

const LOCAL_SETTINGS_KEY = 'cometline-settings';

export function buildE2eSettings() {
	const settings = defaultSettings();
	const anthropic = settings.providers.find((provider) => provider.id === 'anthropic');
	if (!anthropic) {
		throw new Error('Expected anthropic provider in default settings');
	}

	anthropic.enabled = true;
	anthropic.models = ['claude-sonnet-4-5'];
	anthropic.enabledModels = ['claude-sonnet-4-5'];
	anthropic.selectedModel = 'claude-sonnet-4-5';

	settings.activeProviderId = 'anthropic';
	settings.defaultProviderId = 'anthropic';
	settings.defaultModelId = 'claude-sonnet-4-5';
	settings.app.hasSeenIntro = true;
	settings.app.hasCompletedSetup = true;
	settings.cometmind.jobs.notifications.enabled = false;

	return settings;
}

export async function seedAppState(page: Page) {
	const settings = buildE2eSettings();
	await page.addInitScript(
		({ key, value }) => {
			localStorage.setItem(key, value);
		},
		{ key: LOCAL_SETTINGS_KEY, value: JSON.stringify(settings) }
	);
}
