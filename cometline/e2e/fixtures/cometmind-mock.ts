import type { Page, Route } from '@playwright/test';

const BASE_URL = 'http://127.0.0.1:7700';

type MockSession = {
	id: string;
	workspace_id: string;
	workspace_path: string;
	title: string;
	model_id: string;
	provider_id: string;
	status: 'active' | 'archived';
	token_usage: { input: number; output: number };
	pinned: boolean;
	created_at: number;
	updated_at: number;
};

function json(route: Route, status: number, body: unknown) {
	return route.fulfill({
		status,
		contentType: 'application/json',
		body: JSON.stringify(body)
	});
}

function sse(route: Route, events: unknown[]) {
	const body = events.map((event) => `data: ${JSON.stringify(event)}\n\n`).join('');
	return route.fulfill({
		status: 200,
		contentType: 'text/event-stream',
		body
	});
}

function createSessionRecord(id: string, body: Record<string, unknown> = {}): MockSession {
	const now = Date.now();
	return {
		id,
		workspace_id: 'ws-e2e',
		workspace_path: String(body.workspace_path ?? '/'),
		title: 'New chat',
		model_id: String(body.model_id ?? 'claude-sonnet-4-5'),
		provider_id: String(body.provider_id ?? 'anthropic'),
		status: 'active',
		token_usage: { input: 0, output: 0 },
		pinned: false,
		created_at: now,
		updated_at: now
	};
}

export async function installCometMindMock(
	page: Page,
	options: { healthy?: boolean; sessionId?: string } = {}
) {
	const healthy = options.healthy ?? true;
	const sessionId = options.sessionId ?? 'e2e-session-1';
	let session = createSessionRecord(sessionId);

	await page.route(`${BASE_URL}/**`, async (route) => {
		const request = route.request();
		const url = new URL(request.url());
		const { pathname } = url;
		const method = request.method();

		if (pathname === '/api/v1/health') {
			if (!healthy) {
				return json(route, 503, { status: 'error' });
			}
			return json(route, 200, { status: 'ok' });
		}

		if (pathname === '/api/v1/sessions' && method === 'GET') {
			return json(route, 200, { sessions: [session] });
		}

		if (pathname === '/api/v1/sessions' && method === 'POST') {
			const body = (request.postDataJSON() ?? {}) as Record<string, unknown>;
			session = createSessionRecord(sessionId, body);
			return json(route, 201, session);
		}

		if (pathname === '/api/v1/workspaces' && method === 'POST') {
			return json(route, 200, {
				id: session.workspace_id,
				path: session.workspace_path
			});
		}

		if (pathname === '/api/v1/jobs' && method === 'GET') {
			return json(route, 200, { jobs: [] });
		}

		const sessionMatch = pathname.match(/^\/api\/v1\/sessions\/([^/]+)(?:\/(.+))?$/);
		if (sessionMatch) {
			const [, id, action] = sessionMatch;
			if (id !== sessionId) {
				return json(route, 404, { message: 'Session not found' });
			}

			if (!action && method === 'GET') {
				return json(route, 200, session);
			}

			if (action === 'messages' && method === 'GET') {
				return json(route, 200, { session_id: sessionId, items: [] });
			}

			if (action === 'message' && method === 'POST') {
				return sse(route, [
					{ type: 'text_delta', text: 'Mock assistant reply.' },
					{ type: 'done' }
				]);
			}
		}

		return json(route, 404, { message: `Unhandled mock route: ${method} ${pathname}` });
	});
}
