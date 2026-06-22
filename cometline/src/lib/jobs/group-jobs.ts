import type { JobResource } from '$lib/client/cometmind';

export type JobColumn = 'todo' | 'ongoing' | 'done';

export type GroupedJobs = Record<JobColumn, JobResource[]>;

export function sortJobs(jobs: JobResource[]): JobResource[] {
	return [...jobs].sort((a, b) => {
		const priorityDiff = (b.priority ?? 0) - (a.priority ?? 0);
		if (priorityDiff !== 0) return priorityDiff;
		return (a.updated_at ?? 0) - (b.updated_at ?? 0);
	});
}

export function isJobScheduledNotReady(job: JobResource, now = Date.now()): boolean {
	return job.scheduled_at != null && job.scheduled_at > now;
}

export function isJobReady(job: JobResource, now = Date.now()): boolean {
	if (job.status !== 'todo' || job.deleted_at) return false;
	return !isJobScheduledNotReady(job, now);
}

export function groupJobsByColumn(jobs: JobResource[]): GroupedJobs {
	const active = jobs.filter((job) => !job.deleted_at);
	return {
		todo: sortJobs(active.filter((job) => job.status === 'todo')),
		ongoing: sortJobs(active.filter((job) => job.status === 'ongoing')),
		done: sortJobs(active.filter((job) => job.status === 'done'))
	};
}

export function filterArchivedJobs(jobs: JobResource[]): JobResource[] {
	return sortJobs(jobs.filter((job) => job.deleted_at != null));
}

export function truncateWorkspacePath(path: string, maxLength = 28): string {
	const normalized = path.replace(/\\/g, '/');
	if (normalized.length <= maxLength) return normalized;
	const parts = normalized.split('/').filter(Boolean);
	const name = parts[parts.length - 1] ?? normalized;
	if (name.length <= maxLength) return `…/${name}`;
	return `…${name.slice(-(maxLength - 1))}`;
}

export function formatScheduledLabel(scheduledAt: number, now = Date.now()): string {
	const diff = scheduledAt - now;
	if (diff <= 0) return 'Ready';
	const minutes = Math.round(diff / 60_000);
	if (minutes < 60) return `in ${minutes}m`;
	const hours = Math.round(minutes / 60);
	if (hours < 48) return `in ${hours}h`;
	const days = Math.round(hours / 24);
	return `in ${days}d`;
}
