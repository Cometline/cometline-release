import { describe, expect, it } from 'vitest';

import { defaultSettings } from '$lib/settings/schema';

import { runtimeActionForSettingsSave, saveStatusMessage } from './settings-save';

describe('runtimeActionForSettingsSave', () => {
	it('returns reload for provider changes unrelated to memory providers', () => {
		const persisted = defaultSettings();
		const next = defaultSettings();
		next.providers[0] = { ...next.providers[0], enabled: true, apiKey: 'new-key' };

		expect(runtimeActionForSettingsSave(persisted, next)).toBe('reload');
	});

	it('returns restart for memory settings changes', () => {
		const persisted = defaultSettings();
		const next = defaultSettings();
		next.cometmind.memory.extractionModel = 'text-embedding-3-large';

		expect(runtimeActionForSettingsSave(persisted, next)).toBe('restart');
	});

	it('returns restart when a memory provider entry changes', () => {
		const persisted = defaultSettings();
		const next = defaultSettings();
		next.cometmind.memory.embedding.providerId = 'openai';
		next.providers = next.providers.map((provider) =>
			provider.id === 'openai' ? { ...provider, baseURL: 'http://localhost:11434/v1' } : provider
		);

		expect(runtimeActionForSettingsSave(persisted, next)).toBe('restart');
	});
});

describe('saveStatusMessage', () => {
	it('reports reload distinctly from restart', () => {
		expect(saveStatusMessage('agent', 'reload')).toBe('Changes saved. CometMind reloaded.');
		expect(saveStatusMessage('agent', 'restart')).toBe('Changes saved. CometMind restarted.');
		expect(saveStatusMessage('agent', 'none')).toBe('Changes saved.');
	});
});
