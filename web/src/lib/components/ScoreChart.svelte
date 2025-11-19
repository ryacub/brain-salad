<script lang="ts">
	import type { AnalyticsStats } from '$lib/types';
	import { TrendingUp, Target, Award } from 'lucide-svelte';

	interface Props {
		stats: AnalyticsStats;
	}

	let { stats }: Props = $props();

	const scorePercentage = $derived((stats.average_score / 10) * 100);
	const activePercentage = $derived(
		stats.total_ideas > 0 ? (stats.active_ideas / stats.total_ideas) * 100 : 0
	);

	function getScoreColor(score: number): string {
		if (score >= 8) return 'bg-success-500';
		if (score >= 6) return 'bg-warning-500';
		return 'bg-error-500';
	}
</script>

<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
	<!-- Total Ideas -->
	<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
		<div class="flex items-center justify-between mb-4">
			<h3 class="text-lg font-semibold">Total Ideas</h3>
			<div class="text-primary-500">
				<Target size={24} />
			</div>
		</div>
		<p class="text-3xl font-bold">{stats.total_ideas}</p>
		<p class="text-sm text-surface-600-300-token mt-2">
			{stats.active_ideas} active ({activePercentage.toFixed(0)}%)
		</p>
	</div>

	<!-- Average Score -->
	<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
		<div class="flex items-center justify-between mb-4">
			<h3 class="text-lg font-semibold">Average Score</h3>
			<div class="text-warning-500">
				<TrendingUp size={24} />
			</div>
		</div>
		<p class="text-3xl font-bold">{stats.average_score.toFixed(1)}</p>
		<div class="mt-3">
			<div class="bg-surface-300-600-token h-2 rounded-full overflow-hidden">
				<div
					class={`h-full transition-all duration-500 ${getScoreColor(stats.average_score)}`}
					style="width: {scorePercentage}%"
				></div>
			</div>
		</div>
	</div>

	<!-- Score Range -->
	<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
		<div class="flex items-center justify-between mb-4">
			<h3 class="text-lg font-semibold">Score Range</h3>
			<div class="text-success-500">
				<Award size={24} />
			</div>
		</div>
		<div class="flex items-baseline gap-2">
			<div>
				<p class="text-sm text-surface-600-300-token">High</p>
				<p class="text-2xl font-bold text-success-500">{stats.high_score.toFixed(1)}</p>
			</div>
			<span class="text-surface-600-300-token">-</span>
			<div>
				<p class="text-sm text-surface-600-300-token">Low</p>
				<p class="text-2xl font-bold text-error-500">{stats.low_score.toFixed(1)}</p>
			</div>
		</div>
	</div>
</div>
