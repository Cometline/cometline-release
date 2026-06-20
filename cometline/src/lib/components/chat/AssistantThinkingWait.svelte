<script lang="ts">
	import ThinkingIndicator from '$lib/components/ThinkingIndicator.svelte';

	let {
		label,
		detail,
		color
	}: {
		label: string;
		detail: string;
		color?: string;
	} = $props();
</script>

<div class="assistant-thinking-wait" aria-live="polite" aria-busy="true">
	<ThinkingIndicator size={24} {label} {color} />
	<div class="assistant-thinking-copy">
		<p>{detail}</p>
	</div>
</div>

<style>
	.assistant-thinking-wait {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		padding: 8px 2px 2px;
	}

	.assistant-thinking-copy {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
		padding-top: 2px;
	}

	.assistant-thinking-copy p {
		margin: 0;
		font-size: 12px;
		line-height: 1.45;
		color: var(--text-soft, rgba(0, 0, 0, 0.55));
	}
</style>
