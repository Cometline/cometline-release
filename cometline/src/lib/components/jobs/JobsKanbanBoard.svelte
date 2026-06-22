<script lang="ts">
	import type { GroupedJobs } from '$lib/jobs/group-jobs';
	import type { JobResource } from '$lib/client/cometmind';
	import JobsKanbanColumn from './JobsKanbanColumn.svelte';

	let {
		grouped,
		selectedJobId = null,
		onSelectJob,
		onAddJob
	}: {
		grouped: GroupedJobs;
		selectedJobId?: string | null;
		onSelectJob: (job: JobResource) => void;
		onAddJob: () => void;
	} = $props();

	const columns = [
		{ key: 'todo' as const, title: 'Todo', showAdd: true },
		{ key: 'ongoing' as const, title: 'Ongoing', showAdd: false },
		{ key: 'done' as const, title: 'Done', showAdd: false }
	];
</script>

<div class="kanban-board">
	{#each columns as column (column.key)}
		<JobsKanbanColumn
			title={column.title}
			jobs={grouped[column.key]}
			{selectedJobId}
			showAdd={column.showAdd}
			onSelectJob={onSelectJob}
			onAdd={column.showAdd ? onAddJob : undefined}
		/>
	{/each}
</div>

<style>
	.kanban-board {
		display: grid;
		grid-template-columns: repeat(3, minmax(220px, 1fr));
		gap: 12px;
		min-height: 0;
		flex: 1;
		overflow-x: auto;
		padding-bottom: 4px;
	}

	@media (max-width: 900px) {
		.kanban-board {
			grid-template-columns: repeat(3, minmax(260px, 1fr));
			scroll-snap-type: x mandatory;
		}

		.kanban-board :global(.kanban-column) {
			scroll-snap-align: start;
		}
	}
</style>
