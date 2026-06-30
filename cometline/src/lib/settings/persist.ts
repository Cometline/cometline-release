import type { MemorySettings } from '$lib/client/cometmind';
import { runStorageRetentionAndSyncSessions } from '$lib/retention/storage-retention-sync';
import { normalizeSettings, validateSettings } from '$lib/settings/schema';
import type { RuntimeApplyAction } from '$lib/settings/settings-save';
import type { ProviderSettings } from '$lib/types';
import { putMemorySettings } from '$lib/client/cometmind';
import { connectionState } from '$lib/stores/runtime.svelte';

export interface PersistSettingsOptions {
	runtimeAction?: RuntimeApplyAction;
	restartCometMind?: boolean;
	memory?: MemorySettings;
}

export async function persistSettings(
	draft: ProviderSettings,
	options: PersistSettingsOptions = {}
): Promise<{ settings: ProviderSettings; memory?: MemorySettings }> {
	const runtimeAction = options.runtimeAction ?? (options.restartCometMind === false ? 'none' : 'restart');
	const normalized = validateSettings(normalizeSettings(draft));

	let saved: ProviderSettings;
	if (window.electronAPI?.saveProviderSettings) {
		saved = await window.electronAPI.saveProviderSettings(normalized, {
			runtimeAction
		});
	} else {
		localStorage.setItem('cometline-settings', JSON.stringify(normalized));
		saved = normalized;
	}

	let memory: MemorySettings | undefined;
	if (options.memory) {
		memory = await putMemorySettings(options.memory);
	}

	if (connectionState.status === 'ready') {
		await runStorageRetentionAndSyncSessions();
	}

	if (runtimeAction === 'restart') {
		connectionState.reconnect();
	}

	return { settings: saved, memory };
}
