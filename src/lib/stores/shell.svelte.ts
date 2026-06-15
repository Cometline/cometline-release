import { getActiveSessionId } from '$lib/active-session';

export type SessionWebPanel = { url: string; visible: boolean };
export type FocusedPane = 'chat' | 'web';

function createShellStore() {
	let sidebarOpen = $state(true);
	let settingsOpen = $state(false);
	let composerPhase = $state<'centered' | 'docked'>('centered');
	let workspacePath = $state('/');
	let bootMessage = $state('');
	let fullscreen = $state(false);
	let webPanelsBySession = $state<Record<string, SessionWebPanel>>({});
	let focusedPane = $state<FocusedPane>('chat');

	function activeSessionId(): string | null {
		return getActiveSessionId();
	}

	function panelForActiveSession(): SessionWebPanel | null {
		const id = activeSessionId();
		if (!id) return null;
		return webPanelsBySession[id] ?? null;
	}

	function syncWebPanelOpen(open: boolean) {
		window.electronAPI?.setWebPanelOpen?.(open);
	}

	function syncWebPanelOpenForActiveSession() {
		const panel = panelForActiveSession();
		syncWebPanelOpen(Boolean(panel?.visible));
	}

	return {
		get sidebarOpen() {
			return sidebarOpen;
		},
		get fullscreen() {
			return fullscreen;
		},
		get settingsOpen() {
			return settingsOpen;
		},
		get composerPhase() {
			return composerPhase;
		},
		get workspacePath() {
			return workspacePath;
		},
		get bootMessage() {
			return bootMessage;
		},
		get focusedPane() {
			return focusedPane;
		},
		get webPanelOpen() {
			const panel = panelForActiveSession();
			return Boolean(panel?.visible);
		},
		get webPanelUrl() {
			return panelForActiveSession()?.url ?? null;
		},
		get hasWebPanelForSession() {
			return panelForActiveSession() !== null;
		},
		setWorkspacePath(path: string) {
			workspacePath = path;
		},
		setBootMessage(message: string) {
			bootMessage = message;
		},
		setFullscreen(value: boolean) {
			fullscreen = value;
		},
		toggleSidebar() {
			sidebarOpen = !sidebarOpen;
		},
		openSidebar() {
			sidebarOpen = true;
		},
		closeSidebar() {
			sidebarOpen = false;
		},
		openSettings() {
			settingsOpen = true;
		},
		closeSettings() {
			settingsOpen = false;
		},
		dockComposer() {
			composerPhase = 'docked';
		},
		centerComposer() {
			composerPhase = 'centered';
		},
		setFocusedPane(pane: FocusedPane) {
			focusedPane = pane;
		},
		onActiveSessionChange() {
			focusedPane = 'chat';
			syncWebPanelOpenForActiveSession();
		},
		openWebPanel(url: string, sessionId: string) {
			webPanelsBySession = {
				...webPanelsBySession,
				[sessionId]: { url, visible: true }
			};
			focusedPane = 'web';
			syncWebPanelOpen(true);
		},
		toggleWebPanel() {
			const sessionId = activeSessionId();
			if (!sessionId) return;
			const panel = webPanelsBySession[sessionId];
			if (!panel) return;
			const visible = !panel.visible;
			webPanelsBySession = {
				...webPanelsBySession,
				[sessionId]: { ...panel, visible }
			};
			focusedPane = visible ? 'web' : 'chat';
			syncWebPanelOpen(visible);
		},
		closeWebPanel() {
			const sessionId = activeSessionId();
			if (!sessionId || !webPanelsBySession[sessionId]) return;
			const next = { ...webPanelsBySession };
			delete next[sessionId];
			webPanelsBySession = next;
			focusedPane = 'chat';
			syncWebPanelOpen(false);
		},
		clearWebPanelForSession(sessionId: string) {
			if (!webPanelsBySession[sessionId]) return;
			const next = { ...webPanelsBySession };
			delete next[sessionId];
			webPanelsBySession = next;
			if (activeSessionId() === sessionId) {
				focusedPane = 'chat';
				syncWebPanelOpen(false);
			}
		}
	};
}

export const shellStore = createShellStore();
