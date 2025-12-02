<script lang="ts">
	import { Settings as SettingsIcon, Info, ExternalLink } from 'lucide-svelte';
	import { browser } from '$app/environment';

	let apiUrl = $state('http://localhost:8080');
	let darkMode = $state(false);

	if (browser) {
		apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
		darkMode = document.documentElement.classList.contains('dark');
	}

	function toggleDarkMode() {
		darkMode = !darkMode;
		if (browser) {
			if (darkMode) {
				document.documentElement.classList.add('dark');
				localStorage.setItem('theme', 'dark');
			} else {
				document.documentElement.classList.remove('dark');
				localStorage.setItem('theme', 'light');
			}
		}
	}
</script>

<svelte:head>
	<title>Settings - Telos Idea Matrix</title>
</svelte:head>

<div class="space-y-6 max-w-4xl">
	<div>
		<h1 class="text-4xl font-bold mb-2">Settings</h1>
		<p class="text-surface-600-300-token">Configure your Telos Idea Matrix experience</p>
	</div>

	<!-- Appearance -->
	<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
		<h2 class="text-2xl font-bold mb-4 flex items-center gap-2">
			<SettingsIcon size={24} />
			Appearance
		</h2>

		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<div>
					<p class="font-medium">Dark Mode</p>
					<p class="text-sm text-surface-600-300-token">
						Toggle between light and dark themes
					</p>
				</div>
				<label class="flex items-center cursor-pointer">
					<input
						type="checkbox"
						class="sr-only peer"
						checked={darkMode}
						onchange={toggleDarkMode}
					/>
					<div
						class="relative w-11 h-6 bg-surface-300-600-token peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-surface-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"
					></div>
				</label>
			</div>
		</div>
	</div>

	<!-- API Configuration -->
	<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
		<h2 class="text-2xl font-bold mb-4">API Configuration</h2>

		<div class="space-y-4">
			<div>
				<label for="api-url" class="label">
					<span class="font-medium">API URL</span>
				</label>
				<input
					id="api-url"
					type="text"
					class="input"
					value={apiUrl}
					readonly
					disabled
				/>
				<p class="text-sm text-surface-600-300-token mt-2">
					Configure this in your <code class="code">.env</code> file using
					<code class="code">VITE_API_URL</code>
				</p>
			</div>
		</div>
	</div>

	<!-- About Telos -->
	<div class="card p-6 bg-surface-100-800-token border border-surface-300-600-token">
		<h2 class="text-2xl font-bold mb-4 flex items-center gap-2">
			<Info size={24} />
			About Telos Idea Matrix
		</h2>

		<div class="space-y-4">
			<p class="text-surface-600-300-token">
				Telos Idea Matrix helps you evaluate ideas against your personal goals and values,
				providing objective analysis to combat decision paralysis and context-switching.
			</p>

			<div class="space-y-2">
				<h3 class="font-semibold">How it works:</h3>
				<ul class="list-disc list-inside space-y-1 text-sm text-surface-600-300-token">
					<li>Define your goals, strategies, and failure patterns in a <code class="code">telos.md</code> file</li>
					<li>Capture ideas instantly and get automatic scoring</li>
					<li>Review pattern detection for context-switching, perfectionism, and more</li>
					<li>Make informed decisions based on alignment scores</li>
				</ul>
			</div>

			<div class="flex gap-4 pt-4">
				<a
					href="https://github.com/ryacub/telos-idea-matrix"
					target="_blank"
					rel="noopener noreferrer"
					class="btn variant-ghost-surface"
				>
					<ExternalLink size={16} />
					View on GitHub
				</a>
				<a
					href="https://github.com/ryacub/telos-idea-matrix/blob/main/README.md"
					target="_blank"
					rel="noopener noreferrer"
					class="btn variant-ghost-surface"
				>
					<ExternalLink size={16} />
					Documentation
				</a>
			</div>
		</div>
	</div>

	<!-- Version Info -->
	<div class="card p-4 bg-surface-100-800-token border border-surface-300-600-token">
		<p class="text-sm text-surface-600-300-token text-center">
			Telos Idea Matrix v1.0.0 | Built with SvelteKit & Go
		</p>
	</div>
</div>
