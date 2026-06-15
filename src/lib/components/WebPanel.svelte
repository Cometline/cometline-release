<script lang="ts">
	import { ArrowLeft, ArrowRight, RotateCw, X } from '@lucide/svelte';
	import { shellStore } from '$lib/stores/shell.svelte';
	import { isWebPanelUrl, openLink } from '$lib/open-link';
	import { openExternalLink } from '$lib/external-link';

	type WebviewElement = HTMLElement & {
		src: string;
		goBack(): void;
		goForward(): void;
		reload(): void;
		canGoBack(): boolean;
		canGoForward(): boolean;
		getURL(): string;
		getTitle(): string;
	};

	let webviewEl = $state<WebviewElement | null>(null);
	let canGoBack = $state(false);
	let canGoForward = $state(false);
	let loading = $state(false);
	let displayUrl = $state('');
	let pageTitle = $state('');

	function updateNavigationState() {
		const el = webviewEl;
		if (!el) return;
		canGoBack = el.canGoBack();
		canGoForward = el.canGoForward();
		try {
			displayUrl = el.getURL() || shellStore.webPanelUrl || '';
		} catch {
			displayUrl = shellStore.webPanelUrl || '';
		}
		try {
			pageTitle = el.getTitle() || '';
		} catch {
			pageTitle = '';
		}
	}

	function onBack() {
		if (!webviewEl?.canGoBack()) return;
		webviewEl.goBack();
	}

	function onForward() {
		if (!webviewEl?.canGoForward()) return;
		webviewEl.goForward();
	}

	function onReload() {
		webviewEl?.reload();
	}

	function onClose() {
		shellStore.closeWebPanel();
	}

	function onNewWindow(event: Event & { url?: string; preventDefault?: () => void }) {
		event.preventDefault?.();
		const url = event.url;
		if (!url) return;
		if (isWebPanelUrl(url)) {
			openLink(url);
			return;
		}
		openExternalLink(url);
	}

	function attachWebview(el: WebviewElement) {
		el.setAttribute(
			'sandbox',
			'allow-scripts allow-same-origin allow-popups allow-forms'
		);
		const onNavigate = () => {
			updateNavigationState();
		};
		const onStartLoading = () => {
			loading = true;
		};
		const onStopLoading = () => {
			loading = false;
			updateNavigationState();
		};
		const onTitleUpdated = (event: Event & { title?: string }) => {
			pageTitle = event.title ?? '';
		};

		el.addEventListener('did-navigate', onNavigate);
		el.addEventListener('did-navigate-in-page', onNavigate);
		el.addEventListener('did-start-loading', onStartLoading);
		el.addEventListener('did-stop-loading', onStopLoading);
		el.addEventListener('page-title-updated', onTitleUpdated);
		el.addEventListener('new-window', onNewWindow);

		return () => {
			el.removeEventListener('did-navigate', onNavigate);
			el.removeEventListener('did-navigate-in-page', onNavigate);
			el.removeEventListener('did-start-loading', onStartLoading);
			el.removeEventListener('did-stop-loading', onStopLoading);
			el.removeEventListener('page-title-updated', onTitleUpdated);
			el.removeEventListener('new-window', onNewWindow);
		};
	}

	$effect(() => {
		const el = webviewEl;
		const url = shellStore.webPanelOpen ? shellStore.webPanelUrl : null;
		if (!el || !url) return;
		let current = '';
		try {
			current = el.getURL?.() ?? el.src ?? '';
		} catch {
			current = el.src ?? '';
		}
		if (current !== url) {
			el.src = url;
			displayUrl = url;
		}
	});

	$effect(() => {
		const el = webviewEl;
		if (!el) return;
		return attachWebview(el);
	});

	$effect(() => {
		if (!shellStore.webPanelOpen) {
			loading = false;
			canGoBack = false;
			canGoForward = false;
			pageTitle = '';
		}
	});
</script>

<div class="web-panel" class:open={shellStore.webPanelOpen} aria-hidden={!shellStore.webPanelOpen}>
	<div class="web-panel-inner">
		<header class="web-panel-toolbar">
			<div class="nav-actions">
				<button type="button" class="icon-button" disabled={!canGoBack} onclick={onBack} aria-label="Back">
					<ArrowLeft size={16} />
				</button>
				<button
					type="button"
					class="icon-button"
					disabled={!canGoForward}
					onclick={onForward}
					aria-label="Forward"
				>
					<ArrowRight size={16} />
				</button>
				<button type="button" class="icon-button" onclick={onReload} aria-label="Reload">
					<RotateCw size={16} class={loading ? 'spin' : ''} />
				</button>
			</div>
			<div class="url-display" title={displayUrl}>
				{#if pageTitle}
					<span class="page-title">{pageTitle}</span>
				{/if}
				<span class="page-url">{displayUrl}</span>
			</div>
			<button type="button" class="icon-button close-button" onclick={onClose} aria-label="Close panel">
				<X size={16} />
			</button>
		</header>
		<div class="web-panel-content">
			<!-- Electron webview tag; inert in plain browser dev without Electron. -->
			<webview bind:this={webviewEl} class="web-panel-view"></webview>
		</div>
	</div>
</div>

<style>
	.web-panel {
		flex: 0 0 0;
		width: 0;
		min-width: 0;
		overflow: hidden;
		transition: width var(--duration-fast) var(--ease-smooth);
	}

	.web-panel.open {
		flex: 0 0 var(--web-panel-width);
		width: var(--web-panel-width);
	}

	.web-panel-inner {
		width: var(--web-panel-width);
		height: 100%;
		display: flex;
		flex-direction: column;
		margin: var(--content-panel-inset);
		margin-left: 0;
		border: 1px solid var(--border-soft);
		border-radius: var(--radius-window);
		background: var(--panel-bg);
		box-shadow: var(--shadow-card);
		transform: translateX(100%);
		transition: transform var(--duration-fast) var(--ease-smooth);
		overflow: hidden;
	}

	.web-panel.open .web-panel-inner {
		transform: translateX(0);
	}

	.web-panel-toolbar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 10px;
		border-bottom: 1px solid var(--border-soft);
		background: rgba(250, 250, 249, 0.95);
		min-height: 44px;
	}

	.nav-actions {
		display: flex;
		align-items: center;
		gap: 4px;
		flex-shrink: 0;
	}

	.icon-button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		padding: 0;
		border: none;
		border-radius: 8px;
		background: transparent;
		color: var(--text-main);
		cursor: pointer;
	}

	.icon-button:hover:not(:disabled) {
		background: rgba(15, 23, 42, 0.06);
	}

	.icon-button:disabled {
		opacity: 0.35;
		cursor: default;
	}

	.url-display {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 1px;
		padding: 0 4px;
	}

	.page-title {
		font-size: 12px;
		font-weight: 600;
		color: var(--text-main);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.page-url {
		font-size: 11px;
		color: var(--text-muted);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.close-button {
		flex-shrink: 0;
	}

	.web-panel-content {
		flex: 1;
		min-height: 0;
		position: relative;
		background: #fff;
	}

	.web-panel-view {
		display: inline-flex;
		width: 100%;
		height: 100%;
		border: none;
	}

	:global(.spin) {
		animation: web-panel-spin 0.8s linear infinite;
	}

	@keyframes web-panel-spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (max-width: 900px) {
		.web-panel {
			position: fixed;
			inset: 0;
			z-index: 40;
			width: 100% !important;
			flex: none !important;
			pointer-events: none;
		}

		.web-panel.open {
			pointer-events: auto;
		}

		.web-panel-inner {
			width: 100%;
			height: 100%;
			margin: 0;
			border: none;
			border-radius: 0;
			box-shadow: none;
		}
	}
</style>
