<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { api } from '$lib/api';
	import PatternBadge from '$lib/components/PatternBadge.svelte';
	import {
		Loader2,
		AlertCircle,
		Calendar,
		TrendingUp,
		Edit,
		Trash2,
		ArrowLeft
	} from 'lucide-svelte';

	const id = $derived($page.params.id);
	const queryClient = useQueryClient();

	const ideaQuery = createQuery({
		queryKey: () => ['idea', id],
		queryFn: () => api.ideas.get(id)
	});

	const deleteMutation = createMutation({
		mutationFn: (id: string) => api.ideas.delete(id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['ideas'] });
			goto('/');
		}
	});

	function handleDelete() {
		if (confirm('Are you sure you want to delete this idea?')) {
			deleteMutation.mutate(id);
		}
	}

	function getScoreColor(score: number): string {
		if (score >= 8) return 'text-success-500';
		if (score >= 6) return 'text-warning-500';
		return 'text-error-500';
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<svelte:head>
	<title>Idea Details - Telos Idea Matrix</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center gap-4">
		<a href="/" class="btn btn-sm variant-ghost-surface">
			<ArrowLeft size={16} />
			Back
		</a>
		<h1 class="text-3xl font-bold">Idea Details</h1>
	</div>

	{#if $ideaQuery.isPending}
		<div class="flex justify-center items-center py-12">
			<Loader2 size={32} class="animate-spin text-primary-500" />
		</div>
	{:else if $ideaQuery.isError}
		<div class="alert variant-filled-error">
			<AlertCircle size={20} />
			<p>Error loading idea: {$ideaQuery.error?.message || 'Unknown error'}</p>
		</div>
	{:else if $ideaQuery.data}
		{@const idea = $ideaQuery.data}
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Main Content -->
			<div class="lg:col-span-2 space-y-6">
				<!-- Idea Content -->
				<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
					<div class="flex justify-between items-start mb-4">
						<h2 class="text-2xl font-bold flex-1">{idea.content}</h2>
						<div class="flex gap-2">
							<a href="/ideas/{idea.id}/edit" class="btn btn-sm variant-ghost-surface">
								<Edit size={16} />
								Edit
							</a>
							<button
								onclick={handleDelete}
								class="btn btn-sm variant-ghost-error"
								disabled={$deleteMutation.isPending}
							>
								<Trash2 size={16} />
								Delete
							</button>
						</div>
					</div>

					<div class="flex items-center gap-4 text-sm text-surface-600-300-token">
						<div class="flex items-center gap-1">
							<Calendar size={14} />
							<span>{formatDate(idea.created_at)}</span>
						</div>
						<span class="chip variant-soft-surface text-xs">{idea.status}</span>
					</div>
				</div>

				<!-- Analysis Details -->
				{#if idea.analysis}
					<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
						<h3 class="text-xl font-bold mb-4">Score Breakdown</h3>

						<div class="space-y-4">
							<!-- Mission Alignment -->
							<div>
								<div class="flex justify-between mb-2">
									<span class="font-medium">Mission Alignment (40%)</span>
									<span class="font-bold">
										{idea.analysis.score_breakdown.mission_alignment.toFixed(1)}
									</span>
								</div>
								<div class="bg-surface-300-600-token h-2 rounded-full overflow-hidden">
									<div
										class="h-full bg-primary-500 transition-all duration-500"
										style="width: {(idea.analysis.score_breakdown.mission_alignment / 10) * 100}%"
									></div>
								</div>
							</div>

							<!-- Strategy Fit -->
							<div>
								<div class="flex justify-between mb-2">
									<span class="font-medium">Strategy Fit (25%)</span>
									<span class="font-bold">
										{idea.analysis.score_breakdown.strategy_fit.toFixed(1)}
									</span>
								</div>
								<div class="bg-surface-300-600-token h-2 rounded-full overflow-hidden">
									<div
										class="h-full bg-secondary-500 transition-all duration-500"
										style="width: {(idea.analysis.score_breakdown.strategy_fit / 10) * 100}%"
									></div>
								</div>
							</div>

							<!-- Anti-Pattern Penalty -->
							<div>
								<div class="flex justify-between mb-2">
									<span class="font-medium">Anti-Pattern Penalty (35%)</span>
									<span class="font-bold text-error-500">
										-{Math.abs(idea.analysis.score_breakdown.anti_pattern_penalty).toFixed(1)}
									</span>
								</div>
								<div class="bg-surface-300-600-token h-2 rounded-full overflow-hidden">
									<div
										class="h-full bg-error-500 transition-all duration-500"
										style="width: {(Math.abs(idea.analysis.score_breakdown.anti_pattern_penalty) / 10) * 100}%"
									></div>
								</div>
							</div>
						</div>
					</div>

					<!-- Detected Patterns -->
					{#if idea.analysis.detected_patterns && idea.analysis.detected_patterns.length > 0}
						<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
							<h3 class="text-xl font-bold mb-4">Detected Patterns</h3>
							<div class="space-y-3">
								{#each idea.analysis.detected_patterns as pattern}
									<div class="flex items-start gap-3">
										<PatternBadge pattern={pattern.name} />
										<div class="flex-1">
											<p class="text-sm">{pattern.description}</p>
											<p class="text-xs text-surface-600-300-token mt-1">
												Severity: {pattern.severity}
											</p>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				{/if}

				<!-- Recommendation -->
				{#if idea.recommendation}
					<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
						<h3 class="text-xl font-bold mb-3">Recommendation</h3>
						<p class="text-surface-600-300-token italic">{idea.recommendation}</p>
					</div>
				{/if}
			</div>

			<!-- Sidebar -->
			<div class="space-y-6">
				<!-- Score Card -->
				<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
					<div class="flex items-center gap-2 text-surface-600-300-token mb-3">
						<TrendingUp size={20} />
						<h3 class="font-semibold">Final Score</h3>
					</div>
					<div class="text-center">
						<p class={`text-5xl font-bold ${getScoreColor(idea.final_score)}`}>
							{idea.final_score.toFixed(1)}
						</p>
						<p class="text-sm text-surface-600-300-token mt-2">out of 10</p>
					</div>
				</div>

				<!-- Raw Score -->
				<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
					<h3 class="font-semibold mb-3">Raw Score</h3>
					<p class="text-3xl font-bold text-surface-600-300-token">
						{idea.raw_score.toFixed(1)}
					</p>
				</div>
			</div>
		</div>
	{/if}
</div>
