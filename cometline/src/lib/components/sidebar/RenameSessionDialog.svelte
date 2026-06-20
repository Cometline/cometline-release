<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	import { onMount } from 'svelte';
	import type { Session } from '$lib/types';

	let {
		session,
		renaming = false,
		onCancel,
		onConfirm
	}: {
		session: Session;
		renaming?: boolean;
		onCancel: () => void;
		onConfirm: (title: string) => void;
	} = $props();

	let titleInput = $state<HTMLInputElement | null>(null);
	let title = $state('');

	onMount(() => {
		title = session.title || '';
		titleInput?.focus();
		titleInput?.select();
	});

	function submit() {
		onConfirm(title.trim());
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			event.preventDefault();
			submit();
		}
		if (event.key === 'Escape') {
			event.preventDefault();
			onCancel();
		}
	}
</script>

<div class="rename-dialog" transition:fly={{ y: 8, duration: 140 }}>
	<div class="rename-copy">
		<strong>Rename session</strong>
		<span>Choose a name for this chat.</span>
	</div>
	<input
		bind:this={titleInput}
		class="rename-input"
		type="text"
		bind:value={title}
		placeholder="Untitled"
		maxlength={200}
		onkeydown={handleKeydown}
	/>
	<div class="rename-actions">
		<button class="cancel-rename" onclick={onCancel}>Cancel</button>
		<button class="confirm-rename" onclick={submit} disabled={renaming}>Save</button>
	</div>
</div>
<button
	class="rename-scrim"
	aria-label="Cancel rename"
	onclick={onCancel}
	transition:fade={{ duration: 100 }}
></button>

<style>
	.rename-dialog {
		position: absolute;
		left: 12px;
		right: 12px;
		bottom: 56px;
		z-index: 20;
		display: grid;
		gap: 10px;
		padding: 12px;
		border: 1px solid var(--border-soft);
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.98);
		box-shadow: var(--shadow-card);
	}

	.rename-copy {
		display: grid;
		gap: 4px;
		font-size: 12px;
		color: var(--text-muted);
	}

	.rename-copy strong {
		color: var(--text-main);
	}

	.rename-input {
		width: 100%;
		padding: 8px 10px;
		border: 1px solid var(--border-soft);
		border-radius: 8px;
		background: var(--panel-bg, #fff);
		color: var(--text-main);
		font: inherit;
		font-size: 13px;
		line-height: 1.35;
	}

	.rename-input:focus {
		outline: none;
		border-color: var(--accent);
		box-shadow: 0 0 0 2px rgba(0, 102, 204, 0.12);
	}

	.rename-actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
	}

	.cancel-rename,
	.confirm-rename {
		border: none;
		border-radius: 8px;
		padding: 7px 10px;
		font: inherit;
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
	}

	.cancel-rename {
		background: rgba(15, 23, 42, 0.05);
		color: var(--text-main);
	}

	.confirm-rename {
		background: var(--accent);
		color: white;
	}

	.confirm-rename:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.rename-scrim {
		position: absolute;
		inset: 0;
		z-index: 15;
		border: none;
		background: rgba(15, 23, 42, 0.12);
		cursor: pointer;
	}
</style>
