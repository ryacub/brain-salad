<script lang="ts">
	import type { Idea } from '$lib/types';
	import { Lightbulb, Trash2, Edit, Calendar } from 'lucide-svelte';
	import PatternBadge from './PatternBadge.svelte';

	interface Props {
		idea: Idea;
		onDelete?: (id: string) => void;
		onEdit?: (id: string) => void;
	}

	let { idea, onDelete, onEdit }: Props = $props();

	function getScoreColor(score: number): string {
		if (score >= 8) return 'text-success-500';
		if (score >= 6) return 'text-warning-500';
		return 'text-error-500';
	}

	function getScoreBgColor(score: number): string {
		if (score >= 8) return 'bg-success-500/10';
		if (score >= 6) return 'bg-warning-500/10';
		return 'bg-error-500/10';
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}
</script>

<div
	class="card p-4 hover:shadow-xl transition-shadow duration-200 bg-surface-100-800-token border border-surface-300-600-token"
>
	<div class="flex justify-between items-start gap-4">
		<div class="flex-1">
			<div class="flex items-start gap-3">
				<div class="text-primary-500 mt-1">
					<Lightbulb size={20} />
				</div>
				<div class="flex-1">
					<a href="/ideas/{idea.id}" class="hover:text-primary-500">
						<p class="text-base font-medium mb-2">{idea.content}</p>
					</a>

					{#if idea.patterns && idea.patterns.length > 0}
						<div class="flex flex-wrap gap-2 mb-3">
							{#each idea.patterns as pattern}
								<PatternBadge {pattern} />
							{/each}
						</div>
					{/if}

					<div class="flex items-center gap-4 text-sm text-surface-600-300-token">
						<div class="flex items-center gap-1">
							<Calendar size={14} />
							<span>{formatDate(idea.created_at)}</span>
						</div>
						<span class="chip variant-soft-surface text-xs">{idea.status}</span>
					</div>
				</div>
			</div>
		</div>

		<div class="flex flex-col items-end gap-3">
			<div class={`${getScoreBgColor(idea.final_score)} ${getScoreColor(idea.final_score)} px-3 py-2 rounded-lg font-bold text-lg`}>
				{idea.final_score.toFixed(1)}
			</div>

			<div class="flex gap-2">
				{#if onEdit}
					<button
						onclick={() => onEdit?.(idea.id)}
						class="btn btn-sm variant-ghost-surface"
						aria-label="Edit idea"
					>
						<Edit size={16} />
					</button>
				{/if}
				{#if onDelete}
					<button
						onclick={() => onDelete?.(idea.id)}
						class="btn btn-sm variant-ghost-error"
						aria-label="Delete idea"
					>
						<Trash2 size={16} />
					</button>
				{/if}
			</div>
		</div>
	</div>

	{#if idea.recommendation}
		<div class="mt-3 pt-3 border-t border-surface-300-600-token">
			<p class="text-sm text-surface-600-300-token italic">{idea.recommendation}</p>
		</div>
	{/if}
</div>
