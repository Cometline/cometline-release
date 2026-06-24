<script lang="ts">
	import { FileText, Loader } from '@lucide/svelte';
	import SlashCommandMenu from '$lib/components/composer/SlashCommandMenu.svelte';
	import type { createComposerMentionsController } from '$lib/components/composer/composer-mentions.svelte';

	type MentionsController = ReturnType<typeof createComposerMentionsController>;

	let {
		mentions,
		menuRef = $bindable<HTMLDivElement | null>(null)
	}: {
		mentions: MentionsController;
		menuRef?: HTMLDivElement | null;
	} = $props();
</script>

{#if mentions.mentionMenuOpen}
	<SlashCommandMenu ariaLabel="Workspace files" class="mention-menu" bind:menuRef>
		{#if !mentions.fileIndexReady && mentions.filteredMentionFiles.length === 0}
			<p class="skill-command-loading">
				<Loader size={13} stroke-width={2} class="mention-spinner" />
				<span>Indexing workspace…</span>
			</p>
		{:else if mentions.useServerSearch && mentions.mentionServerLoading && mentions.filteredMentionFiles.length === 0}
			<p class="skill-command-loading">
				<Loader size={13} stroke-width={2} class="mention-spinner" />
				<span>Searching…</span>
			</p>
		{:else if mentions.fileIndex?.error && mentions.filteredMentionFiles.length === 0}
			<p class="skill-command-empty">Could not index workspace.</p>
		{:else if mentions.filteredMentionFiles.length === 0}
			<p class="skill-command-empty">No matching files.</p>
		{:else}
			{#each mentions.filteredMentionFiles as path, index (path)}
				<button
					type="button"
					class="skill-command-option mention-option"
					class:highlighted={index === mentions.mentionHighlight}
					data-mention-index={index}
					role="option"
					aria-selected={index === mentions.mentionHighlight}
					onpointerenter={() => (mentions.mentionHighlight = index)}
					onpointerdown={(e) => {
						e.preventDefault();
						mentions.selectMentionFile(path);
					}}
				>
					<FileText size={14} stroke-width={1.8} />
					<span class="mention-path">{path}</span>
				</button>
			{/each}
		{/if}
		{#if mentions.mentionTruncated && !mentions.mentionQuery.trim()}
			<p class="mention-hint">
				Showing first {mentions.fileIndex?.files.length ?? 0}. Type to search all files.
			</p>
		{/if}
	</SlashCommandMenu>
{/if}
