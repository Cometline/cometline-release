<script lang="ts">
	import { X } from '@lucide/svelte';
	import { imageDataURL } from '$lib/files/images';
	import type { ImageAttachment } from '$lib/types';

	let {
		images,
		onRemove
	}: {
		images: ImageAttachment[];
		onRemove: (id: string) => void;
	} = $props();
</script>

{#if images.length > 0}
	<div class="image-attachments" aria-label="Attached images">
		{#each images as image (image.id)}
			<div class="image-attachment">
				<img src={imageDataURL(image)} alt={image.name ?? 'Attached image'} />
				<button
					type="button"
					class="image-remove"
					aria-label={`Remove ${image.name ?? 'image'}`}
					onclick={() => image.id && onRemove(image.id)}
				>
					<X size={12} stroke-width={2} />
				</button>
			</div>
		{/each}
	</div>
{/if}

<style>
	.image-attachments {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-top: -2px;
	}

	.image-attachment {
		position: relative;
		width: 58px;
		height: 58px;
		border: 1px solid var(--border-soft);
		border-radius: 11px;
		background: rgba(15, 23, 42, 0.04);
		overflow: hidden;
	}

	.image-attachment img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}

	.image-remove {
		position: absolute;
		top: 4px;
		right: 4px;
		display: grid;
		place-items: center;
		width: 18px;
		height: 18px;
		padding: 0;
		border: none;
		border-radius: 999px;
		background: rgba(15, 23, 42, 0.72);
		color: white;
		cursor: pointer;
	}

	.image-remove:hover {
		background: rgba(180, 35, 24, 0.9);
	}
</style>
