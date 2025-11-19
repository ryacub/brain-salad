<script lang="ts">
	import '../app.postcss';
	import { QueryClient, QueryClientProvider } from '@tanstack/svelte-query';
	import { AppShell } from '@skeletonlabs/skeleton';
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();

	const queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				staleTime: 1000 * 60 * 5,
				refetchOnWindowFocus: false
			}
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Telos Idea Matrix</title>
</svelte:head>

<QueryClientProvider client={queryClient}>
	<AppShell>
		<svelte:fragment slot="header">
			<header class="bg-surface-50-900-token border-b border-surface-300-600-token">
				<div class="container mx-auto px-4 py-4">
					<div class="flex justify-between items-center">
						<h1 class="text-2xl font-bold">
							<a href="/" class="hover:text-primary-500">Telos Idea Matrix</a>
						</h1>
						<nav>
							<ul class="flex gap-4">
								<li><a href="/" class="hover:text-primary-500">Dashboard</a></li>
								<li><a href="/analytics" class="hover:text-primary-500">Analytics</a></li>
								<li><a href="/settings" class="hover:text-primary-500">Settings</a></li>
							</ul>
						</nav>
					</div>
				</div>
			</header>
		</svelte:fragment>

		<div class="container mx-auto px-4 py-8">
			{@render children()}
		</div>
	</AppShell>
</QueryClientProvider>
