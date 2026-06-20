<script lang="ts">
	import ThinkingIndicator from './ThinkingIndicator.svelte';

	/** Inline-sized comet indicator (alias for ThinkingIndicator). */
	let {
		size = 12,
		color,
		label = 'Thinking'
	}: {
		size?: number;
		color?: string;
		label?: string;
	} = $props();
</script>

<ThinkingIndicator {size} {color} {label} />
