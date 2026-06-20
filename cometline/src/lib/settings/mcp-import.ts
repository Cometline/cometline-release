import type { MCPServerConfig, MCPTransport } from './schema';

type CursorMcpEntry = {
	command?: string;
	args?: string[];
	env?: Record<string, string>;
	url?: string;
	headers?: Record<string, string>;
};

function slugify(name: string): string {
	return (
		name
			.toLowerCase()
			.replace(/[^a-z0-9]+/g, '-')
			.replace(/^-+|-+$/g, '') || 'server'
	);
}

function inferTransport(entry: CursorMcpEntry): MCPTransport {
	const url = String(entry.url ?? '').trim();
	if (!url) return 'stdio';
	if (url.includes('/sse') || url.endsWith('/sse')) return 'sse';
	return 'http';
}

/** Import servers from Cursor-style `{ "mcpServers": { ... } }` JSON. */
export function importCursorMcpJson(raw: unknown, existingIds: string[] = []): MCPServerConfig[] {
	if (!raw || typeof raw !== 'object') return [];
	const record = raw as Record<string, unknown>;
	const serversRaw = record.mcpServers;
	if (!serversRaw || typeof serversRaw !== 'object') return [];

	const used = new Set(existingIds.map((id) => id.trim()).filter(Boolean));
	const out: MCPServerConfig[] = [];

	for (const [name, value] of Object.entries(serversRaw as Record<string, CursorMcpEntry>)) {
		if (!value || typeof value !== 'object') continue;
		let id = slugify(name);
		let suffix = 2;
		while (used.has(id)) {
			id = `${slugify(name)}-${suffix++}`;
		}
		used.add(id);
		const transport = inferTransport(value);
		out.push({
			id,
			name: name.trim() || id,
			enabled: true,
			transport,
			command: String(value.command ?? '').trim(),
			args: Array.isArray(value.args)
				? value.args.map((part) => String(part).trim()).filter(Boolean)
				: [],
			env: value.env ?? {},
			url: String(value.url ?? '').trim(),
			headers: value.headers ?? {}
		});
	}
	return out;
}
