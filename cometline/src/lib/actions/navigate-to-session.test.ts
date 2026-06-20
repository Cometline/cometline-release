import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Session } from '$lib/types';

const mocks = vi.hoisted(() => ({
	goto: vi.fn().mockResolvedValue(undefined),
	selectSession: vi.fn(),
	selectFromSession: vi.fn(),
	setWorkspacePath: vi.fn(),
	setSidebarOrderWorkspacePath: vi.fn(),
	setSidebarOrderDiscordActive: vi.fn()
}));

vi.mock('$app/navigation', () => ({ goto: mocks.goto }));
vi.mock('$lib/stores/session.svelte', () => ({
	sessionStore: { selectSession: mocks.selectSession }
}));
vi.mock('$lib/stores/model.svelte', () => ({
	modelStore: { selectFromSession: mocks.selectFromSession }
}));
vi.mock('$lib/stores/shell.svelte', () => ({
	shellStore: {
		workspacePath: '/ws-a',
		setWorkspacePath: mocks.setWorkspacePath,
		setSidebarOrderWorkspacePath: mocks.setSidebarOrderWorkspacePath,
		setSidebarOrderDiscordActive: mocks.setSidebarOrderDiscordActive
	}
}));

import { navigateToSession } from './navigate-to-session';

function session(overrides: Partial<Session> = {}): Session {
	return {
		id: 'sess-1',
		workspace_id: 'ws-1',
		workspace_path: '/ws-b',
		title: 'Test',
		model_id: 'm',
		provider_id: 'p',
		status: 'active',
		token_usage: {
			input_tokens: 0,
			output_tokens: 0,
			cache_read: 0,
			cache_write: 0
		},
		pinned: false,
		created_at: 0,
		updated_at: 0,
		...overrides
	};
}

describe('navigateToSession sidebar order', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.stubGlobal('window', { electronAPI: { setWorkspacePath: vi.fn() } });
	});

	it('commits sidebar order for unpinned sessions by default', () => {
		navigateToSession(session());

		expect(mocks.setSidebarOrderWorkspacePath).toHaveBeenCalledWith('/ws-b');
		expect(mocks.setSidebarOrderDiscordActive).toHaveBeenCalledWith(false);
	});

	it('does not commit sidebar order for pinned sessions by default', () => {
		navigateToSession(session({ pinned: true }));

		expect(mocks.setSidebarOrderWorkspacePath).not.toHaveBeenCalled();
		expect(mocks.setSidebarOrderDiscordActive).not.toHaveBeenCalled();
	});

	it('allows explicit commitSidebarOrder override for pinned sessions', () => {
		navigateToSession(session({ pinned: true }), { commitSidebarOrder: true });

		expect(mocks.setSidebarOrderWorkspacePath).toHaveBeenCalledWith('/ws-b');
		expect(mocks.setSidebarOrderDiscordActive).toHaveBeenCalledWith(false);
	});

	it('still opens the session and updates composer workspace when pinned', () => {
		navigateToSession(session({ pinned: true }));

		expect(mocks.selectSession).toHaveBeenCalled();
		expect(mocks.selectFromSession).toHaveBeenCalled();
		expect(mocks.setWorkspacePath).toHaveBeenCalledWith('/ws-b');
		expect(mocks.goto).toHaveBeenCalledWith('/session/sess-1');
	});
});
