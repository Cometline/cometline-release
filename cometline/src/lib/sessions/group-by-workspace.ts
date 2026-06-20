import type { Session } from '$lib/types';

export const PINNED_GROUP_KEY = '__pinned__';

export interface WorkspaceSessionGroup {
	workspacePath: string;
	label: string;
	sessions: Session[];
}

export interface SidebarSessionLayout {
	pinnedSessions: Session[];
	workspaceGroups: WorkspaceSessionGroup[];
}

/** Returns the final path segment used as a short directory label. */
export function workspaceLabel(path: string): string {
	const parts = path.split(/[/\\]/).filter(Boolean);
	return parts[parts.length - 1] || path;
}

/** Sort sessions by most recently updated first. */
export function sortSessionsByRecency(sessions: Session[]): Session[] {
	return [...sessions].sort((a, b) => (b.updated_at ?? 0) - (a.updated_at ?? 0));
}

/** Split sessions into pinned and unpinned lists. */
export function partitionPinnedSessions(sessions: Session[]): {
	pinned: Session[];
	unpinned: Session[];
} {
	const pinned: Session[] = [];
	const unpinned: Session[] = [];
	for (const session of sessions) {
		if (session.pinned) pinned.push(session);
		else unpinned.push(session);
	}
	return { pinned, unpinned };
}

/**
 * Groups unpinned sessions by workspace path. Groups are ordered with the active
 * workspace first, then by the most recently updated session in each group.
 */
export function groupSessionsByWorkspace(
	sessions: Session[],
	activeWorkspacePath = ''
): WorkspaceSessionGroup[] {
	const groups = new Map<string, WorkspaceSessionGroup>();
	const mostRecent = new Map<string, number>();

	for (const session of sessions) {
		const path = session.workspace_path;
		let group = groups.get(path);
		if (!group) {
			group = { workspacePath: path, label: workspaceLabel(path), sessions: [] };
			groups.set(path, group);
		}
		group.sessions.push(session);
		const updatedAt = session.updated_at ?? 0;
		mostRecent.set(path, Math.max(mostRecent.get(path) ?? 0, updatedAt));
	}

	return Array.from(groups.values())
		.map((group) => ({
			...group,
			sessions: sortSessionsByRecency(group.sessions)
		}))
		.sort((a, b) => {
			if (a.workspacePath === activeWorkspacePath) return -1;
			if (b.workspacePath === activeWorkspacePath) return 1;
			return (mostRecent.get(b.workspacePath) ?? 0) - (mostRecent.get(a.workspacePath) ?? 0);
		});
}

/**
 * Builds sidebar layout: a global pinned section followed by workspace groups
 * containing only unpinned sessions.
 */
export function layoutSessionsForSidebar(
	sessions: Session[],
	activeWorkspacePath = ''
): SidebarSessionLayout {
	const { pinned, unpinned } = partitionPinnedSessions(sessions);
	return {
		pinnedSessions: sortSessionsByRecency(pinned),
		workspaceGroups: groupSessionsByWorkspace(unpinned, activeWorkspacePath)
	};
}

/** Flatten sessions in sidebar group order for keyboard navigation. */
export function flattenSessionsInSidebarOrder(
	sessions: Session[],
	orderWorkspacePath = ''
): Session[] {
	const { pinnedSessions, workspaceGroups } = layoutSessionsForSidebar(
		sessions,
		orderWorkspacePath
	);
	return [
		...pinnedSessions,
		...workspaceGroups.flatMap((group) => group.sessions)
	];
}
