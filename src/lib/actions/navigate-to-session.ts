import { goto } from '$app/navigation';
import { modelStore } from '$lib/stores/model.svelte';
import { sessionStore } from '$lib/stores/session.svelte';
import { shellStore } from '$lib/stores/shell.svelte';
import type { Session } from '$lib/types';

/** Activate a session and keep the sidebar workspace highlight in sync. */
export function navigateToSession(session: Session) {
	sessionStore.selectSession(session);
	modelStore.selectFromSession(session);
	if (session.workspace_path && session.workspace_path !== shellStore.workspacePath) {
		void window.electronAPI?.setWorkspacePath?.(session.workspace_path);
		shellStore.setWorkspacePath(session.workspace_path);
	}
	void goto(`/session/${session.id}`);
}
