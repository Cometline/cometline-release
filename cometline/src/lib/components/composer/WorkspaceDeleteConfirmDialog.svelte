<script lang="ts">
	import { fly, fade } from 'svelte/transition';

	let {
		path,
		deleting = false,
		onCancel,
		onConfirm
	}: {
		path: string;
		deleting?: boolean;
		onCancel: () => void;
		onConfirm: () => void;
	} = $props();
</script>

<div class="workspace-delete-layer" aria-modal="true" role="dialog" aria-labelledby="workspace-delete-title">
	<button
		type="button"
		class="delete-scrim"
		aria-label="Cancel remove workspace"
		onclick={onCancel}
		transition:fade={{ duration: 100 }}
	></button>
	<div class="delete-confirm" transition:fly={{ y: 8, duration: 140 }}>
		<div class="delete-copy">
			<strong id="workspace-delete-title">Remove workspace from list?</strong>
			<span class="delete-path" title={path}>{path}</span>
			<span>This removes the path from /change and CometMind registrations. Files on disk are not deleted.</span>
		</div>
		<div class="delete-actions">
			<button type="button" class="cancel-delete" onclick={onCancel}>Cancel</button>
			<button type="button" class="confirm-delete" onclick={onConfirm} disabled={deleting}>
				Remove
			</button>
		</div>
	</div>
</div>

<style>
	.workspace-delete-layer {
		position: absolute;
		inset: 0;
		z-index: 50;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 16px;
		border-radius: inherit;
		pointer-events: none;
	}

	.delete-scrim {
		position: absolute;
		inset: 0;
		z-index: 0;
		border: none;
		border-radius: inherit;
		background: rgba(15, 23, 42, 0.28);
		cursor: pointer;
		pointer-events: auto;
	}

	.delete-confirm {
		position: relative;
		z-index: 1;
		width: min(420px, 100%);
		display: grid;
		gap: 10px;
		padding: 14px;
		border: 1px solid var(--border-soft);
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.98);
		box-shadow: var(--shadow-card);
		max-height: calc(100% - 32px);
		overflow: auto;
		pointer-events: auto;
	}

	.delete-copy {
		display: grid;
		gap: 4px;
		font-size: 12px;
		color: var(--text-muted);
	}

	.delete-copy strong {
		color: var(--text-main);
	}

	.delete-path {
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 11px;
		color: var(--text-main);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.delete-actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
	}

	.cancel-delete,
	.confirm-delete {
		border: none;
		border-radius: 8px;
		padding: 7px 10px;
		font: inherit;
		font-size: 12px;
		font-weight: 600;
		cursor: pointer;
	}

	.cancel-delete {
		background: rgba(15, 23, 42, 0.05);
		color: var(--text-main);
	}

	.confirm-delete {
		background: #b42318;
		color: white;
	}

	.confirm-delete:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
