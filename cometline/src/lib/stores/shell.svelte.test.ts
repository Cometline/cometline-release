import { beforeEach, describe, expect, it, vi } from 'vitest';

const getActiveSessionId = vi.hoisted(() => vi.fn<() => string | null>(() => null));

vi.mock('$lib/active-session', () => ({
	getActiveSessionId
}));

import { shellStore } from './shell.svelte';

describe('shellStore default vs active workspace', () => {
	beforeEach(() => {
		getActiveSessionId.mockReturnValue(null);
		shellStore.initializeDefaultWorkspace('/default');
	});

	it('initializeDefaultWorkspace sets default and active in sync', () => {
		expect(shellStore.defaultWorkspacePath).toBe('/default');
		expect(shellStore.workspacePath).toBe('/default');
		expect(shellStore.sidebarOrderWorkspacePath).toBe('/default');
	});

	it('setDefaultWorkspacePath syncs active when no session is open', () => {
		shellStore.commitActiveWorkspace('/temporary');

		shellStore.setDefaultWorkspacePath('/new-default');

		expect(shellStore.defaultWorkspacePath).toBe('/new-default');
		expect(shellStore.workspacePath).toBe('/new-default');
		expect(shellStore.sidebarOrderWorkspacePath).toBe('/new-default');
	});

	it('setDefaultWorkspacePath leaves active unchanged when a session is open', () => {
		shellStore.commitActiveWorkspace('/session-ws');
		getActiveSessionId.mockReturnValue('sess-1');

		shellStore.setDefaultWorkspacePath('/new-default');

		expect(shellStore.defaultWorkspacePath).toBe('/new-default');
		expect(shellStore.workspacePath).toBe('/session-ws');
		expect(shellStore.sidebarOrderWorkspacePath).toBe('/session-ws');
	});

	it('commitActiveWorkspace updates active and sidebar order without touching default', () => {
		shellStore.commitActiveWorkspace('/fork-target');

		expect(shellStore.defaultWorkspacePath).toBe('/default');
		expect(shellStore.workspacePath).toBe('/fork-target');
		expect(shellStore.sidebarOrderWorkspacePath).toBe('/fork-target');
		expect(shellStore.sidebarOrderDiscordActive).toBe(false);
	});

	it('resetActiveToDefault restores active and sidebar order to default', () => {
		shellStore.commitActiveWorkspace('/fork-target');

		shellStore.resetActiveToDefault();

		expect(shellStore.defaultWorkspacePath).toBe('/default');
		expect(shellStore.workspacePath).toBe('/default');
		expect(shellStore.sidebarOrderWorkspacePath).toBe('/default');
	});
});
