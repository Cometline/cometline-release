<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		description = '',
		children
	}: {
		title: string;
		description?: string;
		children: Snippet;
	} = $props();
</script>

<section class="section">
	<div class="section-heading">
		<h3>{title}</h3>
		{#if description}
			<p>{description}</p>
		{/if}
	</div>
	<div class="section-body">
		{@render children()}
	</div>
</section>

<style>
	.section {
		display: grid;
		gap: 12px;
	}

	.section-heading h3 {
		margin: 0;
		font-size: 13px;
		font-weight: 650;
		color: var(--text-main);
	}

	.section-heading p {
		margin: 4px 0 0;
		font-size: 11px;
		line-height: 1.45;
		color: var(--text-muted);
	}

	.section-body {
		display: grid;
		gap: 12px;
	}
</style>
