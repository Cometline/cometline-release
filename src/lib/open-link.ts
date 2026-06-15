import { openExternalLink } from '$lib/external-link';
import { shellStore } from '$lib/stores/shell.svelte';

export function isWebPanelUrl(rawUrl: string): boolean {
	try {
		const parsed = new URL(String(rawUrl));
		return parsed.protocol === 'http:' || parsed.protocol === 'https:';
	} catch {
		return false;
	}
}

/** Opens http(s) links in the in-app web panel; mailto and dev fallback stay external. */
export function openLink(rawUrl: string): void {
	if (!rawUrl) return;
	try {
		const parsed = new URL(String(rawUrl));
		if (parsed.protocol === 'mailto:') {
			openExternalLink(rawUrl);
			return;
		}
		if (parsed.protocol === 'http:' || parsed.protocol === 'https:') {
			if (window.electronAPI?.setWebPanelOpen) {
				shellStore.openWebPanel(String(rawUrl));
				return;
			}
			openExternalLink(rawUrl);
		}
	} catch {
		// Ignore malformed URLs.
	}
}
