import type { ChatItem } from '$lib/stores/chat.svelte';
import type { ThinkingAttribution } from './thinking-attribution';
import type { ChatTurnPayload } from '$lib/actions/start-chat';
import type { JobResource } from '$lib/client/cometmind';

type AssistantItem = Extract<ChatItem, { type: 'assistant' }>;
type ToolItem = Extract<ChatItem, { type: 'tool' }>;

export interface AssistantStackFoldController {
	thinkingExpanded: (
		assistant: AssistantItem,
		segmentKey: string,
		segmentIndex: number,
		pending?: boolean
	) => boolean;
	toggleThinking: (
		assistant: AssistantItem,
		segmentKey: string,
		segmentIndex: number,
		pending?: boolean
	) => void;
	activityGroupExpanded: (assistantId: string, assistant: AssistantItem) => boolean;
	toggleActivityGroup: (assistantId: string, assistant: AssistantItem) => void;
	memoryInThinkingExpanded: (segmentKey: string) => boolean;
	toggleMemoryInThinking: (segmentKey: string) => void;
	toolOutputExpanded: (tool: ToolItem) => boolean;
	toggleToolOutput: (id: string) => void;
	subagentExpanded: (id: string) => boolean;
	toggleSubagent: (id: string) => void;
}

export interface AssistantStackContext {
	threadItems: readonly ChatItem[];
	thinkingForAssistant: ThinkingAttribution;
	streamingAssistantId: string | null;
	sessionStreaming: boolean;
	sessionId: string;
	now: number;
	heroGlowColor: string;
	copiedId: string | null;
	fold: AssistantStackFoldController;
	toolFoldLabel: (tool: ToolItem) => string;
	onCopyMessage: (id: string, text: string) => void | Promise<void>;
	onNotifyAgent?: (payload: ChatTurnPayload) => void | Promise<void>;
	onStartJob?: (job: JobResource) => void | Promise<void>;
}

export function assistantStackBindings(
	ctx: AssistantStackContext,
	item: AssistantItem,
	showActivitySpinner: boolean
) {
	return {
		item,
		threadItems: ctx.threadItems,
		thinkingForAssistant: ctx.thinkingForAssistant,
		streamingAssistantId: ctx.streamingAssistantId,
		sessionStreaming: ctx.sessionStreaming,
		sessionId: ctx.sessionId,
		now: ctx.now,
		heroGlowColor: ctx.heroGlowColor,
		copiedId: ctx.copiedId,
		showActivitySpinner,
		toolFoldLabel: ctx.toolFoldLabel,
		onCopyMessage: ctx.onCopyMessage,
		onNotifyAgent: ctx.onNotifyAgent,
		onStartJob: ctx.onStartJob,
		thinkingExpanded: ctx.fold.thinkingExpanded,
		toggleThinking: ctx.fold.toggleThinking,
		activityGroupExpanded: ctx.fold.activityGroupExpanded,
		toggleActivityGroup: ctx.fold.toggleActivityGroup,
		memoryInThinkingExpanded: ctx.fold.memoryInThinkingExpanded,
		toggleMemoryInThinking: ctx.fold.toggleMemoryInThinking,
		toolOutputExpanded: ctx.fold.toolOutputExpanded,
		toggleToolOutput: ctx.fold.toggleToolOutput,
		subagentExpanded: ctx.fold.subagentExpanded,
		toggleSubagent: ctx.fold.toggleSubagent
	};
}
