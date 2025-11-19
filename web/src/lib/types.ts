// API Types matching Go backend

export interface DetectedPattern {
	name: string;
	description: string;
	severity: string;
}

export interface ScoreBreakdown {
	mission_alignment: number;
	strategy_fit: number;
	anti_pattern_penalty: number;
}

export interface Analysis {
	raw_score: number;
	final_score: number;
	score_breakdown: ScoreBreakdown;
	detected_patterns: DetectedPattern[];
	recommendation: string;
}

export interface Idea {
	id: string;
	content: string;
	raw_score: number;
	final_score: number;
	patterns: string[];
	recommendation: string;
	analysis?: Analysis;
	created_at: string;
	reviewed_at?: string;
	status: 'active' | 'archived' | 'completed';
}

export interface ListIdeasResponse {
	ideas: Idea[];
	total: number;
	limit: number;
	offset: number;
}

export interface AnalyticsStats {
	total_ideas: number;
	active_ideas: number;
	average_score: number;
	high_score: number;
	low_score: number;
}

export interface CreateIdeaRequest {
	content: string;
}

export interface UpdateIdeaRequest {
	content?: string;
	status?: string;
}

export interface AnalyzeRequest {
	content: string;
}

export interface AnalyzeResponse {
	analysis: Analysis;
}
