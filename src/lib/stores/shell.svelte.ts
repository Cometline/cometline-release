function createShellStore() {
	let sidebarOpen = $state(true);
	let settingsOpen = $state(false);
	let composerPhase = $state<'centered' | 'docked'>('centered');
	let workspacePath = $state('/');
	let bootMessage = $state('');
	let fullscreen = $state(false);
	let webPanelOpen = $state(false);
	let webPanelUrl = $state<string | null>(null);

	function syncWebPanelOpen(open: boolean) {
		window.electronAPI?.setWebPanelOpen?.(open);
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
		get webPanelOpen() {
			return webPanelOpen;
		},
		get webPanelUrl() {
			return webPanelUrl;
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
		openWebPanel(url: string) {
			webPanelUrl = url;
			if (!webPanelOpen) {
				webPanelOpen = true;
				syncWebPanelOpen(true);
			}
		},
		closeWebPanel() {
			if (!webPanelOpen) return;
			webPanelOpen = false;
			webPanelUrl = null;
			syncWebPanelOpen(false);
		}
	};
}

export const shellStore = createShellStore();
