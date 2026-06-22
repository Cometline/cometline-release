import type { JobResource } from '$lib/client/cometmind';

export type JobExecutionPromptInput = Pick<JobResource, 'id' | 'description'> &
	Partial<Pick<JobResource, 'definition_of_done' | 'progress'>>;

export function buildJobExecutionPrompt(job: JobExecutionPromptInput): string {
	const dod = job.definition_of_done?.trim() || '(none specified)';
	const progress = job.progress?.trim();
	const progressBlock = progress
		? `\nPrevious progress (from an earlier attempt):\n${progress}\n\nContinue from here.\n`
		: '';
	return `Please work on: ${job.description}\n\nDefinition of done: ${dod}${progressBlock}\nWhile working, call update_job with progress after each meaningful milestone (and before long tool runs) so another session can resume if this one stops. When finished, call complete_job with a final progress summary.\n\n(Use job_id "${job.id}" when calling job tools.)`;
}
