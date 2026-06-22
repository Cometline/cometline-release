import { describe, expect, it } from 'vitest';
import type { JobResource } from '$lib/client/cometmind';
import {
	filterArchivedJobs,
	groupJobsByColumn,
	isJobReady,
	isJobScheduledNotReady,
	sortJobs
} from './group-jobs';

function job(overrides: Partial<JobResource> & Pick<JobResource, 'id' | 'status'>): JobResource {
	return {
		description: 'test',
		definition_of_done: '',
		progress: '',
		priority: 0,
		created_by: 'user',
		created_at: 1,
		updated_at: 1,
		...overrides
	};
}

describe('sortJobs', () => {
	it('sorts by priority desc then updated_at asc', () => {
		const jobs = [
			job({ id: 'a', status: 'todo', priority: 1, updated_at: 20 }),
			job({ id: 'b', status: 'todo', priority: 5, updated_at: 30 }),
			job({ id: 'c', status: 'todo', priority: 5, updated_at: 10 })
		];
		expect(sortJobs(jobs).map((j) => j.id)).toEqual(['c', 'b', 'a']);
	});
});

describe('isJobScheduledNotReady', () => {
	it('returns true when scheduled_at is in the future', () => {
		const j = job({ id: 'x', status: 'todo', scheduled_at: 2_000 });
		expect(isJobScheduledNotReady(j, 1_000)).toBe(true);
		expect(isJobReady(j, 1_000)).toBe(false);
	});

	it('returns false when scheduled_at is past or absent', () => {
		const j = job({ id: 'x', status: 'todo', scheduled_at: 500 });
		expect(isJobScheduledNotReady(j, 1_000)).toBe(false);
		expect(isJobReady(j, 1_000)).toBe(true);
	});
});

describe('groupJobsByColumn', () => {
	it('groups active jobs by status', () => {
		const jobs = [
			job({ id: 't', status: 'todo' }),
			job({ id: 'o', status: 'ongoing' }),
			job({ id: 'd', status: 'done' }),
			job({ id: 'del', status: 'todo', deleted_at: 99 })
		];
		const grouped = groupJobsByColumn(jobs);
		expect(grouped.todo.map((j) => j.id)).toEqual(['t']);
		expect(grouped.ongoing.map((j) => j.id)).toEqual(['o']);
		expect(grouped.done.map((j) => j.id)).toEqual(['d']);
	});
});

describe('filterArchivedJobs', () => {
	it('returns only soft-deleted jobs', () => {
		const jobs = [
			job({ id: 'a', status: 'todo' }),
			job({ id: 'b', status: 'done', deleted_at: 100 })
		];
		expect(filterArchivedJobs(jobs).map((j) => j.id)).toEqual(['b']);
	});
});
