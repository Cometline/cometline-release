import { describe, expect, it } from 'vitest';
import type { ChatItem, ProviderConfig } from '$lib/types';
import {
	estimateChatContextTokens,
	estimateTokensFromText,
	formatContextPercent,
	formatContextUsageTokens,
	formatContextWindow,
	resolveContextWindow
} from './context-window';

describe('context-window', () => {
	it('prefers per-model metadata', () => {
		const provider: ProviderConfig = {
			id: 'anthropic',
			name: 'Anthropic',
			method: 'anthropic',
			enabled: true,
			baseURL: '',
			apiKey: '',
			selectedModel: '',
			models: [],
			enabledModels: [],
			modelMetadata: { 'claude-sonnet': { contextWindow: 200_000 } }
		};
		expect(resolveContextWindow(provider, 'claude-sonnet')).toBe(200_000);
	});

	it('formats large windows compactly', () => {
		expect(formatContextWindow(200_000)).toBe('200k');
		expect(formatContextWindow(1_280_000)).toBe('1.3M');
	});

	it('estimates tokens from text with chars/4 heuristic', () => {
		expect(estimateTokensFromText('')).toBe(0);
		expect(estimateTokensFromText('abcd')).toBe(1);
		expect(estimateTokensFromText('a'.repeat(400))).toBe(100);
	});

	it('estimates transcript tokens from chat items', () => {
		const items: ChatItem[] = [
			{ id: '1', type: 'user', text: 'hello world' },
			{ id: '2', type: 'assistant', text: 'hi there' }
		];
		expect(estimateChatContextTokens(items)).toBeGreaterThan(0);
	});

	it('formats usage tooltip values', () => {
		expect(formatContextUsageTokens(180_400)).toBe('180.4K');
		expect(formatContextPercent(180_400, 200_000)).toBe('90.2');
	});
});
