import type { ProviderSettings } from '$lib/types';
import type { SettingsSection } from '$lib/components/settings/settings-controller.svelte';

export function providersNeedRestart(
	persisted: ProviderSettings,
	next: ProviderSettings
): boolean {
	return JSON.stringify(persisted.providers) !== JSON.stringify(next.providers);
}

export function cometmindNeedsRestart(
	persisted: ProviderSettings,
	next: ProviderSettings
): boolean {
	return JSON.stringify(persisted.cometmind) !== JSON.stringify(next.cometmind);
}

export function saveStatusMessage(
	section: SettingsSection,
	restartCometMind: boolean,
	iconVariantChanged = false
): string {
	const restartNote = restartCometMind ? ' CometMind restarted.' : '';
	switch (section) {
		case 'models':
			return restartCometMind ? `Changes saved.${restartNote}` : 'Changes saved.';
		case 'agent':
			return `Changes saved.${restartNote}`;
		case 'appearance':
			return iconVariantChanged || restartCometMind
				? `Changes saved.${restartNote}`
				: 'Changes saved.';
		case 'app':
			return 'Changes saved.';
		case 'memory':
			return `Changes saved.${restartNote}`;
		default:
			return 'Changes saved.';
	}
}
