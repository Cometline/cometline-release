import { describe, expect, it } from 'vitest';
import { buildJobExecutionPrompt } from './build-job-execution-prompt';

describe('buildJobExecutionPrompt', () => {
	it('includes description, DoD, and periodic update guidance', () => {
		const prompt = buildJobExecutionPrompt({
			id: 'job-1',
			description: 'Fix auth',
			definition_of_done: 'tests pass'
		});
		expect(prompt).toContain('Please work on: Fix auth');
		expect(prompt).toContain('Definition of done: tests pass');
		expect(prompt).toContain('after each meaningful milestone');
		expect(prompt).toContain('job_id "job-1"');
		expect(prompt).not.toContain('Previous progress');
	});

	it('includes previous progress when present', () => {
		const prompt = buildJobExecutionPrompt({
			id: 'job-2',
			description: 'Ship feature',
			progress: 'Middleware done; tests red.'
		});
		expect(prompt).toContain('Previous progress (from an earlier attempt):');
		expect(prompt).toContain('Middleware done; tests red.');
		expect(prompt).toContain('Continue from here.');
	});
});
