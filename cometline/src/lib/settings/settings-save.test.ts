import { describe, expect, it } from 'vitest';
import { cometmindNeedsRestart, providersNeedRestart, saveStatusMessage } from './settings-save';
import { defaultSettings } from '$lib/settings/schema';
import type { ProviderSettings } from '$lib/types';

const base = (): ProviderSettings => defaultSettings();

describe('settings-save helpers', () => {
	it('detects provider restart need', () => {
		const a = base();
		const b = { ...a, providers: [{ id: 'p1' }] as ProviderSettings['providers'] };
		expect(providersNeedRestart(a, b)).toBe(true);
	});

	it('formats save status with restart note', () => {
		expect(saveStatusMessage('memory', true)).toContain('CometMind restarted');
	});

	it('detects cometmind config changes', () => {
		const a = base();
		const b = {
			...a,
			cometmind: {
				...a.cometmind,
				storage: { ...a.cometmind.storage, retentionDays: 99 }
			}
		};
		expect(cometmindNeedsRestart(a, b)).toBe(true);
	});
});
