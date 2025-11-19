<script lang="ts">
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { api } from '$lib/api';
	import type { Idea } from '$lib/types';
	import { Send, Loader2 } from 'lucide-svelte';

	interface Props {
		idea?: Idea;
		onSuccess?: (idea: Idea) => void;
		onCancel?: () => void;
	}

	let { idea, onSuccess, onCancel }: Props = $props();

	let content = $state(idea?.content || '');
	const queryClient = useQueryClient();

	const createIdeaMutation = createMutation({
		mutationFn: (content: string) => api.ideas.create({ content }),
		onSuccess: (data) => {
			queryClient.invalidateQueries({ queryKey: ['ideas'] });
			content = '';
			onSuccess?.(data);
		}
	});

	const updateIdeaMutation = createMutation({
		mutationFn: (data: { id: string; content: string }) =>
			api.ideas.update(data.id, { content: data.content }),
		onSuccess: (data) => {
			queryClient.invalidateQueries({ queryKey: ['ideas'] });
			queryClient.invalidateQueries({ queryKey: ['idea', data.id] });
			onSuccess?.(data);
		}
	});

	function handleSubmit(e: Event) {
		e.preventDefault();
		if (!content.trim()) return;

		if (idea) {
			updateIdeaMutation.mutate({ id: idea.id, content });
		} else {
			createIdeaMutation.mutate(content);
		}
	}

	const isPending = $derived($createIdeaMutation.isPending || $updateIdeaMutation.isPending);
	const error = $derived($createIdeaMutation.error || $updateIdeaMutation.error);
</script>

<form onsubmit={handleSubmit} class="space-y-4">
	<div class="form-group">
		<label for="idea-content" class="label">
			<span class="text-lg font-semibold">
				{idea ? 'Edit Idea' : 'Capture Your Idea'}
			</span>
		</label>
		<textarea
			id="idea-content"
			bind:value={content}
			placeholder="Enter your idea here... (e.g., 'Build a Rust CLI tool for personal productivity')"
			class="textarea resize-none"
			rows="4"
			disabled={isPending}
			required
		/>
	</div>

	{#if error}
		<div class="alert variant-filled-error">
			<p>{error.message || 'An error occurred'}</p>
		</div>
	{/if}

	<div class="flex gap-2 justify-end">
		{#if onCancel}
			<button type="button" onclick={onCancel} class="btn variant-ghost-surface" disabled={isPending}>
				Cancel
			</button>
		{/if}
		<button type="submit" class="btn variant-filled-primary" disabled={isPending || !content.trim()}>
			{#if isPending}
				<Loader2 size={16} class="animate-spin" />
			{:else}
				<Send size={16} />
			{/if}
			<span>{idea ? 'Update' : 'Analyze & Save'}</span>
		</button>
	</div>
</form>
