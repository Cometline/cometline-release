import { describe, expect, it } from 'vitest';
import { filterWorkspaceOptions, normalizeWorkspacePath } from './slash-commands';

describe('normalizeWorkspacePath', () => {
	it('normalizes slashes and trailing separators', () => {
		expect(normalizeWorkspacePath('/foo/bar/')).toBe('/foo/bar');
		expect(normalizeWorkspacePath('C:\\foo\\bar')).toBe('C:/foo/bar');
	});
});

describe('filterWorkspaceOptions deletable', () => {
	const paths = ['/Users/me/project-a', '/Users/me/project-b'];

	it('marks workspaces deletable when session count is zero', () => {
		const counts = new Map([
			['/Users/me/project-a', 0],
			['/Users/me/project-b', 2]
		]);
		const options = filterWorkspaceOptions('', paths, counts);
		const workspaceOptions = options.filter((option) => option.kind === 'workspace');
		expect(workspaceOptions[0]?.deletable).toBe(true);
		expect(workspaceOptions[1]?.deletable).toBe(false);
	});

	it('treats paths missing from the map as deletable', () => {
		const options = filterWorkspaceOptions('', ['/Users/me/recent-only'], new Map());
		const workspace = options.find((option) => option.kind === 'workspace');
		expect(workspace?.deletable).toBe(true);
	});

	it('does not mark browse option as deletable workspace row', () => {
		const options = filterWorkspaceOptions('', paths, new Map());
		expect(options.at(-1)?.kind).toBe('browse');
	});
});
