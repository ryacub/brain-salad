<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { api } from '$lib/api';
	import ScoreChart from '$lib/components/ScoreChart.svelte';
	import { Loader2, AlertCircle, BarChart3 } from 'lucide-svelte';

	const statsQuery = createQuery({
		queryKey: ['analytics', 'stats'],
		queryFn: () => api.analytics.stats()
	});

	const ideasQuery = createQuery({
		queryKey: ['ideas'],
		queryFn: () => api.ideas.list({})
	});

	const scoreDistribution = $derived(() => {
		if (!$ideasQuery.data) return { high: 0, medium: 0, low: 0 };

		const ideas = $ideasQuery.data.ideas;
		return {
			high: ideas.filter((i) => i.final_score >= 8).length,
			medium: ideas.filter((i) => i.final_score >= 6 && i.final_score < 8).length,
			low: ideas.filter((i) => i.final_score < 6).length
		};
	});

	const dist = $derived(scoreDistribution());
</script>

<svelte:head>
	<title>Analytics - Telos Idea Matrix</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h1 class="text-4xl font-bold mb-2">Analytics Dashboard</h1>
		<p class="text-surface-600-300-token">
			Track your idea patterns and make data-driven decisions
		</p>
	</div>

	{#if $statsQuery.isPending}
		<div class="flex justify-center items-center py-12">
			<Loader2 size={32} class="animate-spin text-primary-500" />
		</div>
	{:else if $statsQuery.isError}
		<div class="alert variant-filled-error">
			<AlertCircle size={20} />
			<p>Error loading analytics: {$statsQuery.error?.message || 'Unknown error'}</p>
		</div>
	{:else if $statsQuery.data}
		<!-- Overview Stats -->
		<ScoreChart stats={$statsQuery.data} />

		<!-- Score Distribution -->
		<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
			<div class="flex items-center gap-2 mb-6">
				<BarChart3 size={24} />
				<h2 class="text-2xl font-bold">Score Distribution</h2>
			</div>

			<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
				<!-- High Scores -->
				<div class="p-4 rounded-lg bg-success-500/10 border border-success-500/30">
					<div class="flex items-center justify-between mb-2">
						<span class="text-sm font-medium text-success-500">High (8-10)</span>
						<span class="text-2xl font-bold text-success-500">{dist.high}</span>
					</div>
					<div class="bg-success-500/20 h-2 rounded-full overflow-hidden">
						<div
							class="h-full bg-success-500"
							style="width: {$statsQuery.data.total_ideas > 0 ? (dist.high / $statsQuery.data.total_ideas) * 100 : 0}%"
						></div>
					</div>
					<p class="text-xs text-surface-600-300-token mt-1">
						Highly aligned ideas
					</p>
				</div>

				<!-- Medium Scores -->
				<div class="p-4 rounded-lg bg-warning-500/10 border border-warning-500/30">
					<div class="flex items-center justify-between mb-2">
						<span class="text-sm font-medium text-warning-500">Medium (6-7.9)</span>
						<span class="text-2xl font-bold text-warning-500">{dist.medium}</span>
					</div>
					<div class="bg-warning-500/20 h-2 rounded-full overflow-hidden">
						<div
							class="h-full bg-warning-500"
							style="width: {$statsQuery.data.total_ideas > 0 ? (dist.medium / $statsQuery.data.total_ideas) * 100 : 0}%"
						></div>
					</div>
					<p class="text-xs text-surface-600-300-token mt-1">
						Worth considering
					</p>
				</div>

				<!-- Low Scores -->
				<div class="p-4 rounded-lg bg-error-500/10 border border-error-500/30">
					<div class="flex items-center justify-between mb-2">
						<span class="text-sm font-medium text-error-500">Low (0-5.9)</span>
						<span class="text-2xl font-bold text-error-500">{dist.low}</span>
					</div>
					<div class="bg-error-500/20 h-2 rounded-full overflow-hidden">
						<div
							class="h-full bg-error-500"
							style="width: {$statsQuery.data.total_ideas > 0 ? (dist.low / $statsQuery.data.total_ideas) * 100 : 0}%"
						></div>
					</div>
					<p class="text-xs text-surface-600-300-token mt-1">
						Likely distractions
					</p>
				</div>
			</div>
		</div>

		<!-- Insights -->
		<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
			<h2 class="text-2xl font-bold mb-4">Insights</h2>
			<div class="space-y-3">
				{#if $statsQuery.data.average_score >= 8}
					<div class="p-4 rounded-lg bg-success-500/10 border border-success-500/30">
						<p class="text-success-500 font-medium">
							Excellent! Your ideas are highly aligned with your goals.
						</p>
					</div>
				{:else if $statsQuery.data.average_score >= 6}
					<div class="p-4 rounded-lg bg-warning-500/10 border border-warning-500/30">
						<p class="text-warning-500 font-medium">
							Good progress! Consider focusing on ideas that better align with your mission.
						</p>
					</div>
				{:else}
					<div class="p-4 rounded-lg bg-error-500/10 border border-error-500/30">
						<p class="text-error-500 font-medium">
							Many ideas are misaligned. Review your telos.md file and focus on ideas with
							higher scores.
						</p>
					</div>
				{/if}

				{#if dist.high > 0}
					<div class="p-4 rounded-lg bg-primary-500/10 border border-primary-500/30">
						<p class="text-primary-500">
							You have {dist.high} high-scoring
							{dist.high === 1 ? 'idea' : 'ideas'}. These are strong candidates for
							immediate action!
						</p>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
