<script lang="ts">
	import { slide } from 'svelte/transition';
	import { Brain, ChevronDown, LoaderCircle } from '@lucide/svelte';
	import type { InjectedMemory } from '$lib/conversation/thinking-attribution';

	const FOLD_IN = { duration: 180 };

	let {
		text,
		pending = false,
		memories,
		expanded,
		memoryExpanded,
		showSpinner = false,
		onToggle,
		onToggleMemory
	}: {
		text: string;
		pending?: boolean;
		memories?: InjectedMemory[];
		expanded: boolean;
		memoryExpanded: boolean;
		showSpinner?: boolean;
		onToggle: () => void;
		onToggleMemory: () => void;
	} = $props();

	function thinkingLabel(memoriesList?: InjectedMemory[]) {
		if (!memoriesList?.length) return 'Thinking';
		return `Thinking · ${memoriesList.length} memor${memoriesList.length === 1 ? 'y' : 'ies'}`;
	}

	function autoScrollBottom(node: HTMLElement, value: string) {
		const pin = (content: string) => {
			void content;
			node.scrollTop = node.scrollHeight;
		};
		pin(value);
		return {
			update(content: string) {
				pin(content);
			}
		};
	}
</script>

<div class="fold-panel thinking-panel">
	<button
		type="button"
		class="fold-toggle thinking-toggle"
		aria-expanded={expanded}
		onclick={onToggle}
	>
		<Brain size={13} />
		<span>{thinkingLabel(memories)}</span>
		{#if showSpinner}
			<LoaderCircle size={12} class="spin" />
		{/if}
		<ChevronDown size={13} class={expanded ? 'expanded' : ''} />
	</button>
	{#if expanded}
		<div class="fold-body thinking-body" transition:slide={FOLD_IN}>
			{#if memories?.length}
				<div class="thinking-memories">
					<button
						type="button"
						class="fold-toggle memory-toggle"
						aria-expanded={memoryExpanded}
						onclick={onToggleMemory}
					>
						<Brain size={12} />
						<span>Memories used · {memories.length}</span>
						<ChevronDown size={12} class={memoryExpanded ? 'expanded' : ''} />
					</button>
					{#if memoryExpanded}
						<div class="thinking-memory-body" transition:slide={FOLD_IN}>
							<div class="memory-chips">
								{#each memories as mem (mem.id)}
									<span class="memory-chip" title={mem.content}>
										{mem.kind}: {mem.content}
									</span>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			{/if}
			<div class="thinking-reasoning">
				<p use:autoScrollBottom={text}>
					{text || 'Thinking…'}
				</p>
			</div>
		</div>
	{/if}
</div>

<style>
	.thinking-memories {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.memory-toggle {
		font-size: 11px;
		padding: 4px 9px;
	}

	.thinking-memory-body {
		border: 1px solid var(--border-soft);
		border-radius: 10px;
		padding: 8px 10px;
		background: rgba(0, 102, 204, 0.04);
	}

	.memory-chips {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.memory-chip {
		min-width: 0;
		max-width: 100%;
		overflow-wrap: anywhere;
		word-break: break-word;
		white-space: normal;
		padding: 5px 10px;
		border-radius: 10px;
		background: rgba(0, 102, 204, 0.08);
		color: var(--text-main);
		font-size: 11px;
		line-height: 1.45;
	}

	.fold-panel {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.fold-toggle {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		align-self: flex-start;
		border: 1px solid var(--border-soft);
		background: rgba(255, 255, 255, 0.72);
		color: var(--text-muted);
		border-radius: 999px;
		padding: 5px 10px;
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
	}

	.fold-toggle:hover {
		background: rgba(255, 255, 255, 0.92);
		color: var(--text-main);
	}

	.fold-toggle :global(svg.expanded) {
		transform: rotate(180deg);
	}

	.fold-body {
		border: 1px solid var(--border-soft);
		background: rgba(255, 255, 255, 0.68);
		border-radius: 12px;
		padding: 10px 12px;
		color: var(--text-muted);
		box-shadow: 0 6px 18px rgba(15, 23, 42, 0.04);
	}

	.thinking-body {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.thinking-reasoning p {
		margin: 0;
		font-size: 12px;
		line-height: 1.5;
		white-space: pre-wrap;
		word-break: break-word;
		max-height: 220px;
		overflow: auto;
		scrollbar-gutter: stable;
		color: var(--text-muted);
	}
</style>
