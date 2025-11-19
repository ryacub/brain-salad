// API Client for Telos Idea Matrix
import type {
	Idea,
	ListIdeasResponse,
	AnalyticsStats,
	CreateIdeaRequest,
	UpdateIdeaRequest,
	AnalyzeRequest,
	AnalyzeResponse
} from './types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

class ApiError extends Error {
	constructor(
		message: string,
		public status: number
	) {
		super(message);
		this.name = 'ApiError';
	}
}

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
	const url = `${API_BASE_URL}${endpoint}`;

	try {
		const response = await fetch(url, {
			...options,
			headers: {
				'Content-Type': 'application/json',
				...options?.headers
			}
		});

		if (!response.ok) {
			const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
			throw new ApiError(errorData.error || `HTTP ${response.status}`, response.status);
		}

		// Handle 204 No Content
		if (response.status === 204) {
			return null as T;
		}

		return await response.json();
	} catch (error) {
		if (error instanceof ApiError) {
			throw error;
		}
		throw new ApiError('Network error', 0);
	}
}

export const api = {
	// Health check
	health: async (): Promise<{ status: string }> => {
		return fetchAPI('/health');
	},

	// Ideas
	ideas: {
		list: async (params?: {
			status?: string;
			limit?: number;
			offset?: number;
		}): Promise<ListIdeasResponse> => {
			const searchParams = new URLSearchParams();
			if (params?.status) searchParams.append('status', params.status);
			if (params?.limit) searchParams.append('limit', params.limit.toString());
			if (params?.offset) searchParams.append('offset', params.offset.toString());

			const query = searchParams.toString();
			return fetchAPI(`/api/v1/ideas${query ? `?${query}` : ''}`);
		},

		get: async (id: string): Promise<Idea> => {
			return fetchAPI(`/api/v1/ideas/${id}`);
		},

		create: async (data: CreateIdeaRequest): Promise<Idea> => {
			return fetchAPI('/api/v1/ideas', {
				method: 'POST',
				body: JSON.stringify(data)
			});
		},

		update: async (id: string, data: UpdateIdeaRequest): Promise<Idea> => {
			return fetchAPI(`/api/v1/ideas/${id}`, {
				method: 'PUT',
				body: JSON.stringify(data)
			});
		},

		delete: async (id: string): Promise<void> => {
			return fetchAPI(`/api/v1/ideas/${id}`, {
				method: 'DELETE'
			});
		}
	},

	// Analysis
	analyze: async (data: AnalyzeRequest): Promise<AnalyzeResponse> => {
		return fetchAPI('/api/v1/analyze', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	},

	// Analytics
	analytics: {
		stats: async (): Promise<AnalyticsStats> => {
			return fetchAPI('/api/v1/analytics/stats');
		}
	}
};
