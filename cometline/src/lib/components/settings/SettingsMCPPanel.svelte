<script lang="ts">
	import SettingsToggle from './SettingsToggle.svelte';
	import SettingsButton from './SettingsButton.svelte';
	import SettingsField from './SettingsField.svelte';
	import {
		formatIdList,
		parseIdList,
		type CometMindMCPSettings,
		type CometMindSettings,
		type MCPServerConfig,
		type MCPTransport
	} from '$lib/cometmind-settings';
	import { importCursorMcpJson } from '$lib/settings/mcp-import';
	import {
		listMcpServers,
		listMcpTools,
		reconnectMcpServer,
		testMcpServer,
		type McpServerStatus,
		type McpToolInfo
	} from '$lib/client/cometmind';
	import { ChevronDown, ChevronRight, FileJson, Plus, RefreshCw, Upload } from '@lucide/svelte';
	import { onMount } from 'svelte';

	let { cometmind = $bindable() }: { cometmind: CometMindSettings } = $props();

	const transportOptions: { value: MCPTransport; label: string; hint: string }[] = [
		{ value: 'stdio', label: 'Local command', hint: 'Run npx, node, uvx, or another CLI on this machine.' },
		{ value: 'http', label: 'Remote URL (HTTP)', hint: 'Connect to a hosted MCP server over HTTP.' },
		{ value: 'sse', label: 'Remote URL (SSE)', hint: 'Legacy server-sent events transport.' }
	];

	let serverStatuses = $state<McpServerStatus[]>([]);
	let toolPreview = $state<McpToolInfo[]>([]);
	let mcpBusy = $state(false);
	let mcpStatus = $state('');
	let oauthStatus = $state<Record<string, boolean>>({});
	let envTexts = $state<Record<string, string>>({});
	let headerTexts = $state<Record<string, string>>({});
	let argsTexts = $state<Record<string, string>>({});
	let allowedToolsTexts = $state<Record<string, string>>({});
	let scopesTexts = $state<Record<string, string>>({});
	let expandedServerId = $state<string | null>(null);
	let advancedOpen = $state<Record<string, boolean>>({});
	let showPasteImport = $state(false);
	let pasteJsonText = $state('');

	function syncTextFieldsFromSettings() {
		const nextEnv: Record<string, string> = {};
		const nextHeaders: Record<string, string> = {};
		const nextArgs: Record<string, string> = {};
		const nextAllowed: Record<string, string> = {};
		const nextScopes: Record<string, string> = {};
		for (const server of cometmind.mcp.servers) {
			nextEnv[server.id] = formatEnv(server.env);
			nextHeaders[server.id] = formatEnv(server.headers);
			nextArgs[server.id] = (server.args ?? []).join(' ');
			nextAllowed[server.id] = formatIdList(server.allowedTools ?? []);
			nextScopes[server.id] = formatIdList(server.oauth?.scopes ?? []);
		}
		envTexts = nextEnv;
		headerTexts = nextHeaders;
		argsTexts = nextArgs;
		allowedToolsTexts = nextAllowed;
		scopesTexts = nextScopes;
	}

	onMount(() => {
		syncTextFieldsFromSettings();
		void refreshMcpRuntime();
	});

	function formatEnv(values: Record<string, string> | undefined): string {
		if (!values) return '';
		return Object.entries(values)
			.map(([key, value]) => `${key}=${value}`)
			.join('\n');
	}

	function parseEnv(raw: string): Record<string, string> {
		const out: Record<string, string> = {};
		for (const line of raw.split('\n')) {
			const trimmed = line.trim();
			if (!trimmed) continue;
			const idx = trimmed.indexOf('=');
			if (idx <= 0) continue;
			const key = trimmed.slice(0, idx).trim();
			const value = trimmed.slice(idx + 1).trim();
			if (key) out[key] = value;
		}
		return out;
	}

	function updateMcp(patch: Partial<CometMindMCPSettings>) {
		cometmind = {
			...cometmind,
			mcp: {
				...cometmind.mcp,
				...patch
			}
		};
	}

	function updateServer(serverId: string, patch: Partial<MCPServerConfig>) {
		updateMcp({
			servers: cometmind.mcp.servers.map((server) =>
				server.id === serverId ? { ...server, ...patch } : server
			)
		});
	}

	function addServer() {
		const id = `server-${cometmind.mcp.servers.length + 1}`;
		const server: MCPServerConfig = {
			id,
			name: `MCP Server ${cometmind.mcp.servers.length + 1}`,
			enabled: true,
			transport: 'stdio',
			command: '',
			args: [],
			env: {},
			url: '',
			headers: {}
		};
		updateMcp({ enabled: true, servers: [...cometmind.mcp.servers, server] });
		syncTextFieldsFromSettings();
		expandedServerId = id;
	}

	function removeServer(serverId: string) {
		updateMcp({ servers: cometmind.mcp.servers.filter((server) => server.id !== serverId) });
		if (expandedServerId === serverId) expandedServerId = null;
		syncTextFieldsFromSettings();
	}

	function toggleExpanded(serverId: string) {
		expandedServerId = expandedServerId === serverId ? null : serverId;
	}

	function toggleAdvanced(serverId: string) {
		advancedOpen = { ...advancedOpen, [serverId]: !advancedOpen[serverId] };
	}

	function statusFor(serverId: string): McpServerStatus | undefined {
		return serverStatuses.find((item) => item.id === serverId);
	}

	function toolsForServer(serverId: string): McpToolInfo[] {
		return toolPreview.filter((tool) => tool.server_id === serverId);
	}

	function statusLabel(status: McpServerStatus | undefined, server: MCPServerConfig): string {
		if (!cometmind.mcp.enabled) return 'Off';
		if (!server.enabled) return 'Disabled';
		if (!status) return 'Unknown';
		return status.status;
	}

	function statusClass(status: McpServerStatus | undefined, server: MCPServerConfig): string {
		const value = statusLabel(status, server);
		if (value === 'connected') return 'connected';
		if (value === 'error' || value === 'disconnected') return 'error';
		if (value === 'Disabled' || value === 'Off') return 'idle';
		return 'idle';
	}

	function connectionSummary(server: MCPServerConfig): string {
		if (server.transport === 'stdio') {
			const command = String(server.command ?? '').trim();
			const args = (server.args ?? []).join(' ');
			return [command, args].filter(Boolean).join(' ') || 'No command configured';
		}
		return String(server.url ?? '').trim() || 'No URL configured';
	}

	function transportHint(value: MCPTransport): string {
		return transportOptions.find((option) => option.value === value)?.hint ?? '';
	}

	function importServers(parsed: unknown): number {
		const imported = importCursorMcpJson(
			parsed,
			cometmind.mcp.servers.map((server) => server.id)
		);
		if (imported.length === 0) return 0;
		updateMcp({ enabled: true, servers: [...cometmind.mcp.servers, ...imported] });
		syncTextFieldsFromSettings();
		if (imported.length === 1) {
			expandedServerId = imported[0].id;
		}
		return imported.length;
	}

	async function refreshMcpRuntime() {
		mcpBusy = true;
		mcpStatus = '';
		try {
			const [servers, tools] = await Promise.all([listMcpServers(), listMcpTools()]);
			serverStatuses = servers;
			toolPreview = tools;
			const oauthEntries = await Promise.all(
				cometmind.mcp.servers.map(async (server) => {
					const status = await window.electronAPI?.getMcpOAuthStatus?.(server.id);
					return [server.id, Boolean(status?.authenticated)] as const;
				})
			);
			oauthStatus = Object.fromEntries(oauthEntries);
		} catch (err) {
			mcpStatus = err instanceof Error ? err.message : 'Failed to load MCP status';
		} finally {
			mcpBusy = false;
		}
	}

	async function onTestServer(serverId: string) {
		mcpBusy = true;
		mcpStatus = '';
		try {
			const result = await testMcpServer(serverId);
			mcpStatus = result.ok
				? `Connected: ${result.tool_count} tool(s) discovered. Save settings to apply permanently.`
				: result.error || 'Connection test failed';
		} catch (err) {
			mcpStatus = err instanceof Error ? err.message : 'Connection test failed';
		} finally {
			mcpBusy = false;
		}
	}

	async function onReconnectServer(serverId: string) {
		mcpBusy = true;
		mcpStatus = '';
		try {
			await reconnectMcpServer(serverId);
			mcpStatus = `Reconnected. Save settings if you changed configuration.`;
			await refreshMcpRuntime();
		} catch (err) {
			mcpStatus = err instanceof Error ? err.message : 'Reconnect failed';
		} finally {
			mcpBusy = false;
		}
	}

	async function onConnectOAuth(server: MCPServerConfig) {
		if (!window.electronAPI?.startMcpOAuth) {
			mcpStatus = 'OAuth connect is only available in the desktop app.';
			return;
		}
		if (!server.oauth?.clientId || !server.oauth.authorizationUrl || !server.oauth.tokenUrl) {
			mcpStatus = 'Fill OAuth client ID, authorization URL, and token URL first.';
			return;
		}
		mcpBusy = true;
		mcpStatus = '';
		try {
			const result = await window.electronAPI.startMcpOAuth({
				serverId: server.id,
				oauth: {
					clientId: server.oauth.clientId,
					scopes: server.oauth.scopes ?? [],
					authorizationUrl: server.oauth.authorizationUrl,
					tokenUrl: server.oauth.tokenUrl
				}
			});
			mcpStatus = result.message;
			oauthStatus = { ...oauthStatus, [server.id]: true };
			await onReconnectServer(server.id);
		} catch (err) {
			mcpStatus = err instanceof Error ? err.message : 'OAuth connect failed';
		} finally {
			mcpBusy = false;
		}
	}

	async function onImportMcpJson() {
		const input = document.createElement('input');
		input.type = 'file';
		input.accept = 'application/json,.json';
		input.onchange = async () => {
			const file = input.files?.[0];
			if (!file) return;
			try {
				const text = await file.text();
				const parsed = JSON.parse(text) as unknown;
				const count = importServers(parsed);
				if (count === 0) {
					mcpStatus = 'No MCP servers found in file.';
					return;
				}
				showPasteImport = false;
				pasteJsonText = '';
				mcpStatus = `Imported ${count} server(s). Save settings to apply.`;
			} catch (err) {
				mcpStatus = err instanceof Error ? err.message : 'Failed to import mcp.json';
			}
		};
		input.click();
	}

	function onPasteImport() {
		mcpStatus = '';
		try {
			const parsed = JSON.parse(pasteJsonText) as unknown;
			const count = importServers(parsed);
			if (count === 0) {
				mcpStatus = 'No MCP servers found in pasted JSON.';
				return;
			}
			showPasteImport = false;
			pasteJsonText = '';
			mcpStatus = `Imported ${count} server(s). Save settings to apply.`;
		} catch (err) {
			mcpStatus = err instanceof Error ? err.message : 'Invalid JSON';
		}
	}

	export function syncFields() {
		cometmind = {
			...cometmind,
			mcp: {
				...cometmind.mcp,
				servers: cometmind.mcp.servers.map((server) => ({
					...server,
					args: (argsTexts[server.id] ?? '')
						.split(/\s+/)
						.map((part) => part.trim())
						.filter(Boolean),
					env: parseEnv(envTexts[server.id] ?? ''),
					headers: parseEnv(headerTexts[server.id] ?? ''),
					allowedTools: parseIdList(allowedToolsTexts[server.id] ?? ''),
					oauth: server.oauth
						? {
								...server.oauth,
								scopes: parseIdList(scopesTexts[server.id] ?? '')
							}
						: undefined
				}))
			}
		};
	}

	function syncServerLists(serverId: string) {
		const current = cometmind.mcp.servers.find((s) => s.id === serverId);
		updateServer(serverId, {
			args: (argsTexts[serverId] ?? '')
				.split(/\s+/)
				.map((part) => part.trim())
				.filter(Boolean),
			env: parseEnv(envTexts[serverId] ?? ''),
			headers: parseEnv(headerTexts[serverId] ?? ''),
			allowedTools: parseIdList(allowedToolsTexts[serverId] ?? ''),
			oauth: current?.oauth
				? {
						...current.oauth,
						scopes: parseIdList(scopesTexts[serverId] ?? '')
					}
				: undefined
		});
	}
</script>

<div class="settings-section">
	<div class="settings-section-heading">
		<h3>MCP servers</h3>
		<p>
			Connect external tool servers so CometMind can search, browse, and interact with services
			beyond built-in tools. Import from Cursor or add servers manually.
		</p>
	</div>

	<SettingsToggle
		label="Use MCP tools in chat"
		description="Discover tools from configured servers when the sidecar starts."
		bind:checked={cometmind.mcp.enabled}
	/>

	{#if mcpStatus}
		<p class="settings-field-hint" class:error={mcpStatus.toLowerCase().includes('fail') || mcpStatus.toLowerCase().includes('invalid')}>
			{mcpStatus}
		</p>
	{/if}

	{#if cometmind.mcp.servers.length === 0}
		<div class="mcp-empty">
			<div class="import-card">
				<div class="import-card-icon" aria-hidden="true">
					<FileJson size={20} strokeWidth={1.75} />
				</div>
				<div class="import-card-copy">
					<strong>Import from Cursor</strong>
					<p>Use an existing <code>mcp.json</code> file — the same format Cursor uses for MCP servers.</p>
				</div>
				<div class="import-card-actions">
					<SettingsButton variant="primary" disabled={mcpBusy} onclick={onImportMcpJson}>
						<Upload size={14} strokeWidth={2} />
						Choose file
					</SettingsButton>
					<SettingsButton
						variant="secondary"
						onclick={() => {
							showPasteImport = !showPasteImport;
						}}
					>
						Paste JSON
					</SettingsButton>
				</div>
			</div>

			{#if showPasteImport}
				<div class="paste-import">
					<SettingsField
						label="Paste mcp.json contents"
						note="Cursor-style JSON with a top-level mcpServers object."
					>
						<textarea
							bind:value={pasteJsonText}
							rows="6"
							placeholder={'{\n  "mcpServers": {\n    "github": {\n      "command": "npx",\n      "args": ["-y", "@modelcontextprotocol/server-github"],\n      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "..." }\n    }\n  }\n}'}
							spellcheck="false"
						></textarea>
					</SettingsField>
					<div class="mcp-toolbar">
						<SettingsButton variant="primary" disabled={!pasteJsonText.trim()} onclick={onPasteImport}>
							Import servers
						</SettingsButton>
						<SettingsButton
							variant="secondary"
							onclick={() => {
								showPasteImport = false;
								pasteJsonText = '';
							}}
						>
							Cancel
						</SettingsButton>
					</div>
				</div>
			{/if}

			<div class="mcp-empty-footer">
				<span>Or start from scratch</span>
				<SettingsButton variant="secondary" onclick={addServer}>
					<Plus size={14} strokeWidth={2} />
					Add custom server
				</SettingsButton>
			</div>
		</div>
	{:else}
		<div class="mcp-toolbar">
			<SettingsButton variant="secondary" disabled={mcpBusy} onclick={onImportMcpJson}>
				<Upload size={14} strokeWidth={2} />
				Import more
			</SettingsButton>
			<SettingsButton variant="secondary" onclick={addServer}>
				<Plus size={14} strokeWidth={2} />
				Add server
			</SettingsButton>
			<SettingsButton variant="secondary" disabled={mcpBusy} onclick={refreshMcpRuntime}>
				<RefreshCw size={14} strokeWidth={2} class={mcpBusy ? 'spin' : ''} />
				{mcpBusy ? 'Refreshing…' : 'Refresh status'}
			</SettingsButton>
		</div>

		<div class="mcp-server-list">
			<div class="mcp-server-list-header">
				<span>Configured servers</span>
				<strong>{cometmind.mcp.servers.length}</strong>
			</div>

			{#each cometmind.mcp.servers as server (server.id)}
				{@const status = statusFor(server.id)}
				{@const expanded = expandedServerId === server.id}
				<div class="mcp-server-item" class:expanded>
					<div class="mcp-server-row-wrap">
						<button
							type="button"
							class="mcp-server-row"
							aria-expanded={expanded}
							onclick={() => toggleExpanded(server.id)}
						>
							<span class="row-chevron" aria-hidden="true">
								{#if expanded}
									<ChevronDown size={16} strokeWidth={2} />
								{:else}
									<ChevronRight size={16} strokeWidth={2} />
								{/if}
							</span>
							<span class="row-main">
								<span class="row-title">
									<strong>{server.name}</strong>
									<span class="status-badge {statusClass(status, server)}">
										{statusLabel(status, server)}
										{#if status?.tool_count}
											· {status.tool_count} tools
										{/if}
									</span>
								</span>
								<span class="row-summary">{connectionSummary(server)}</span>
							</span>
						</button>
						<label class="row-toggle" title="Enable this server">
							<input
								type="checkbox"
								checked={server.enabled}
								onchange={(e) => updateServer(server.id, { enabled: e.currentTarget.checked })}
							/>
							<span>On</span>
						</label>
					</div>

					{#if status?.last_error && !expanded}
						<p class="row-error">{status.last_error}</p>
					{/if}

					{#if expanded}
						{@const serverTools = toolsForServer(server.id)}
						<div class="mcp-server-editor">
							<SettingsField label="Display name">
								<input
									type="text"
									value={server.name}
									oninput={(e) => updateServer(server.id, { name: e.currentTarget.value })}
								/>
							</SettingsField>

							<SettingsField
								label="Connection type"
								note={transportHint(server.transport)}
							>
								<select
									value={server.transport}
									onchange={(e) =>
										updateServer(server.id, {
											transport: e.currentTarget.value as MCPTransport
										})}
								>
									{#each transportOptions as option (option.value)}
										<option value={option.value}>{option.label}</option>
									{/each}
								</select>
							</SettingsField>

							{#if server.transport === 'stdio'}
								<SettingsField label="Command">
									<input
										type="text"
										value={server.command ?? ''}
										oninput={(e) => updateServer(server.id, { command: e.currentTarget.value })}
										placeholder="npx"
										spellcheck="false"
									/>
								</SettingsField>
								<SettingsField label="Arguments" note="Space-separated, same as Cursor mcp.json args.">
									<input
										type="text"
										bind:value={argsTexts[server.id]}
										onchange={() => syncServerLists(server.id)}
										onblur={() => syncServerLists(server.id)}
										placeholder="-y @modelcontextprotocol/server-filesystem /path/to/dir"
										spellcheck="false"
									/>
								</SettingsField>
							{:else}
								<SettingsField label="Server URL">
									<input
										type="text"
										value={server.url ?? ''}
										oninput={(e) => updateServer(server.id, { url: e.currentTarget.value })}
										placeholder="https://example.com/mcp"
										spellcheck="false"
									/>
								</SettingsField>
							{/if}

							<div class="editor-actions">
								<SettingsButton variant="primary" disabled={mcpBusy} onclick={() => onTestServer(server.id)}>
									Test connection
								</SettingsButton>
								{#if statusLabel(status, server) === 'error' || statusLabel(status, server) === 'disconnected'}
									<SettingsButton
										variant="secondary"
										disabled={mcpBusy}
										onclick={() => onReconnectServer(server.id)}
									>
										Reconnect
									</SettingsButton>
								{/if}
								<SettingsButton variant="danger" onclick={() => removeServer(server.id)}>
									Remove
								</SettingsButton>
							</div>

							{#if status?.last_error}
								<p class="settings-field-hint error">{status.last_error}</p>
							{/if}

							{#if serverTools.length > 0}
								<div class="server-tools">
									<div class="server-tools-header">
										<span>Discovered tools</span>
										<strong>{serverTools.length}</strong>
									</div>
									{#each serverTools as tool (tool.registry_name)}
										<div class="tool-row">
											<strong>{tool.tool_name}</strong>
											<p>{tool.description || tool.registry_name}</p>
										</div>
									{/each}
								</div>
							{/if}

							<button
								type="button"
								class="advanced-toggle"
								aria-expanded={advancedOpen[server.id] ?? false}
								onclick={() => toggleAdvanced(server.id)}
							>
								{#if advancedOpen[server.id]}
									<ChevronDown size={14} strokeWidth={2} />
								{:else}
									<ChevronRight size={14} strokeWidth={2} />
								{/if}
								Advanced settings
							</button>

							{#if advancedOpen[server.id]}
								<div class="advanced-panel">
									<SettingsField label="Server ID" note="Used in tool names like mcp_{server.id}_tool_name. Auto-generated on import.">
										<input type="text" value={server.id} readonly spellcheck="false" />
									</SettingsField>

									{#if server.transport === 'stdio'}
										<SettingsField label="Environment variables" note="One KEY=value per line.">
											<textarea
												bind:value={envTexts[server.id]}
												onchange={() => syncServerLists(server.id)}
												onblur={() => syncServerLists(server.id)}
												rows="3"
												spellcheck="false"
											></textarea>
										</SettingsField>
									{:else}
										<SettingsField label="Headers" note="One KEY=value per line. Use for API keys or Bearer tokens.">
											<textarea
												bind:value={headerTexts[server.id]}
												onchange={() => syncServerLists(server.id)}
												onblur={() => syncServerLists(server.id)}
												rows="3"
												spellcheck="false"
											></textarea>
										</SettingsField>

										<div class="oauth-block">
											<p class="advanced-label">OAuth (optional)</p>
											<p class="settings-field-hint">
												Tokens are stored in <code>~/.cometmind/mcp-oauth/</code>, not in settings JSON.
											</p>
											<SettingsField label="Client ID">
												<input
													type="text"
													value={server.oauth?.clientId ?? ''}
													oninput={(e) =>
														updateServer(server.id, {
															oauth: {
																...(server.oauth ?? {}),
																clientId: e.currentTarget.value
															}
														})}
													spellcheck="false"
												/>
											</SettingsField>
											<SettingsField label="Authorization URL">
												<input
													type="text"
													value={server.oauth?.authorizationUrl ?? ''}
													oninput={(e) =>
														updateServer(server.id, {
															oauth: {
																...(server.oauth ?? {}),
																authorizationUrl: e.currentTarget.value
															}
														})}
													spellcheck="false"
												/>
											</SettingsField>
											<SettingsField label="Token URL">
												<input
													type="text"
													value={server.oauth?.tokenUrl ?? ''}
													oninput={(e) =>
														updateServer(server.id, {
															oauth: {
																...(server.oauth ?? {}),
																tokenUrl: e.currentTarget.value
															}
														})}
													spellcheck="false"
												/>
											</SettingsField>
											<SettingsField label="Scopes" note="One scope per line.">
												<textarea
													bind:value={scopesTexts[server.id]}
													onchange={() => syncServerLists(server.id)}
													onblur={() => syncServerLists(server.id)}
													rows="2"
													spellcheck="false"
												></textarea>
											</SettingsField>
											<div class="oauth-actions">
												<SettingsButton
													variant="secondary"
													disabled={mcpBusy}
													onclick={() => onConnectOAuth(server)}
												>
													Connect with OAuth
												</SettingsButton>
												<span class="oauth-status">
													{oauthStatus[server.id] ? 'OAuth token saved' : 'Not connected'}
												</span>
											</div>
										</div>
									{/if}

									<SettingsField
										label="Allowed tools"
										note="Optional filter. Leave empty to expose every tool from this server."
									>
										<textarea
											bind:value={allowedToolsTexts[server.id]}
											onchange={() => syncServerLists(server.id)}
											onblur={() => syncServerLists(server.id)}
											rows="2"
											placeholder="tool_one&#10;tool_two"
											spellcheck="false"
										></textarea>
									</SettingsField>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			{/each}
		</div>

		{#if toolPreview.length > 0}
			<p class="settings-field-hint mcp-footnote">
				{toolPreview.length} tool(s) registered across all servers. Save settings to apply changes.
			</p>
		{/if}
	{/if}
</div>

<style>
	.mcp-empty {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.import-card {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 12px 14px;
		padding: 14px;
		border: 1px solid var(--border-soft);
		border-radius: 14px;
		background: rgba(255, 255, 255, 0.72);
	}

	.import-card-icon {
		display: flex;
		align-items: flex-start;
		justify-content: center;
		padding-top: 2px;
		color: var(--text-muted);
	}

	.import-card-copy strong {
		display: block;
		font-size: 13px;
		margin-bottom: 4px;
	}

	.import-card-copy p {
		margin: 0;
		font-size: 12px;
		line-height: 1.45;
		color: var(--text-muted);
	}

	.import-card-actions {
		grid-column: 1 / -1;
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.paste-import {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 12px;
		border: 1px dashed rgba(0, 0, 0, 0.12);
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.45);
	}

	.mcp-empty-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		font-size: 12px;
		color: var(--text-muted);
	}

	.mcp-toolbar,
	.editor-actions,
	.oauth-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		align-items: center;
	}

	.mcp-server-list {
		border: 1px solid var(--border-soft);
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.58);
		overflow: hidden;
	}

	.mcp-server-list-header {
		display: flex;
		justify-content: space-between;
		padding: 9px 11px;
		border-bottom: 1px solid var(--border-soft);
		background: rgba(250, 248, 244, 0.94);
		font-size: 12px;
		font-weight: 650;
		color: var(--text-main);
	}

	.mcp-server-item {
		border-bottom: 1px solid rgba(0, 0, 0, 0.06);
	}

	.mcp-server-item:last-child {
		border-bottom: 0;
	}

	.mcp-server-row-wrap {
		display: grid;
		grid-template-columns: 1fr auto;
		align-items: center;
		gap: 8px;
		padding: 10px 11px;
	}

	.mcp-server-row-wrap:hover {
		background: rgba(15, 23, 42, 0.03);
	}

	.mcp-server-item.expanded .mcp-server-row-wrap {
		background: rgba(15, 23, 42, 0.04);
	}

	.mcp-server-row {
		min-width: 0;
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 8px 10px;
		align-items: center;
		padding: 0;
		border: none;
		background: transparent;
		text-align: left;
		cursor: pointer;
		font: inherit;
		color: inherit;
	}

	.mcp-server-row:hover {
		background: transparent;
	}

	.row-chevron {
		display: flex;
		color: var(--text-muted);
	}

	.row-main {
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 3px;
	}

	.row-title {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
	}

	.row-title strong {
		font-size: 13px;
	}

	.row-summary {
		font-size: 11px;
		color: var(--text-muted);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.row-toggle {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding-right: 2px;
		font-size: 11px;
		font-weight: 600;
		color: var(--text-muted);
		cursor: pointer;
		flex-shrink: 0;
	}

	.row-error {
		margin: 0;
		padding: 0 11px 8px 37px;
		font-size: 11px;
		color: #b42318;
	}

	.mcp-server-editor {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 0 11px 14px 37px;
	}

	.status-badge {
		display: inline-block;
		padding: 1px 8px;
		border-radius: 999px;
		font-size: 10px;
		font-weight: 650;
		text-transform: capitalize;
		background: rgba(15, 23, 42, 0.06);
		color: var(--text-muted);
	}

	.status-badge.connected {
		background: rgba(47, 111, 79, 0.12);
		color: #2f6f4f;
	}

	.status-badge.error {
		background: rgba(180, 35, 24, 0.12);
		color: #b42318;
	}

	.server-tools {
		border: 1px solid rgba(0, 0, 0, 0.08);
		border-radius: 10px;
		overflow: auto;
		max-height: 180px;
	}

	.server-tools-header {
		display: flex;
		justify-content: space-between;
		padding: 7px 10px;
		border-bottom: 1px solid rgba(0, 0, 0, 0.06);
		font-size: 11px;
		font-weight: 650;
		background: rgba(250, 248, 244, 0.9);
	}

	.tool-row {
		padding: 7px 10px;
		border-bottom: 1px solid rgba(0, 0, 0, 0.05);
	}

	.tool-row:last-child {
		border-bottom: 0;
	}

	.tool-row strong {
		display: block;
		font-size: 12px;
	}

	.tool-row p {
		margin: 2px 0 0;
		font-size: 11px;
		color: var(--text-muted);
	}

	.advanced-toggle {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 0;
		border: none;
		background: transparent;
		font: inherit;
		font-size: 12px;
		font-weight: 600;
		color: var(--text-muted);
		cursor: pointer;
	}

	.advanced-toggle:hover {
		color: var(--text-main);
	}

	.advanced-panel {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 12px;
		border: 1px dashed rgba(0, 0, 0, 0.1);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.5);
	}

	.advanced-label {
		margin: 0;
		font-size: 12px;
		font-weight: 650;
		color: var(--text-main);
	}

	.oauth-block {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.oauth-status {
		font-size: 11px;
		color: var(--text-muted);
	}

	.settings-field-hint.error,
	.settings-field-hint.error {
		color: #b42318;
	}

	.mcp-footnote {
		margin-top: 4px;
	}

	textarea,
	input,
	select {
		width: 100%;
	}

	textarea {
		resize: vertical;
		min-height: 72px;
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 12px;
	}

	:global(.spin) {
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
