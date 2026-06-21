<script lang="ts">
	import { slide } from 'svelte/transition';
	import { Brain, ChevronDown } from '@lucide/svelte';
	import type { InjectedMemory } from '$lib/conversation/thinking-attribution';

	const FOLD_IN = { duration: 180 };

	let {
		memories,
		expanded,
		onToggle,
		nested = false,
		contentOnly = false
	}: {
		memories: InjectedMemory[];
		expanded: boolean;
		onToggle: () => void;
		nested?: boolean;
		contentOnly?: boolean;
	} = $props();
</script>

<div class="fold-panel memory-panel" class:nested class:content-only={contentOnly}>
	{#if !contentOnly}
		<button
			type="button"
			class="fold-toggle memory-toggle"
			aria-expanded={expanded}
			onclick={onToggle}
		>
			<Brain size={13} />
			<span>Memories used · {memories.length}</span>
			<ChevronDown size={13} class={expanded ? 'expanded' : ''} />
		</button>
	{/if}
	{#if contentOnly || expanded}
		<div class="fold-body memory-body" transition:slide={FOLD_IN}>
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

<style>
	/* Base .fold-panel / .fold-toggle / .fold-body styles live in
	   src/lib/styles/fold-panel.css. Only component-specific overrides here. */
	.fold-panel.nested {
		align-self: stretch;
	}

	.fold-panel.nested .fold-toggle {
		align-self: stretch;
	}

	.fold-panel.content-only .memory-body {
		border: none;
		background: transparent;
		padding: 0;
	}

	.memory-body {
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
</style>
