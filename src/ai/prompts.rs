// AI prompt templates for different analysis tasks
// This file contains specialized prompts for different types of AI analysis

pub struct AnalysisPrompts;

impl AnalysisPrompts {
    pub const MISSION_ALIGNMENT: &'static str = r#"Analyze how well this idea aligns with the user's Telos missions.

Idea: {idea}
Missions: {missions}
Goals: {goals}

Rate alignment on a scale of 1-10 and explain your reasoning.

Focus on:
- How directly this advances the stated missions
- Whether this moves toward key goals
- Mission priority alignment

Respond with: Score (1-10) + Brief explanation"#;

    pub const PATTERN_DETECTION: &'static str = r#"Detect behavioral patterns in this idea based on the user's documented challenges.

Idea: {idea}
Known Challenges: {challenges}
Current Strategy: {strategy}

Look for these patterns:
1. Context-switching (new tech/framework temptations)
2. Perfectionism (over-engineering, scope creep)
3. Procrastination (consumption traps, future-phrasing)
4. Accountability avoidance (solo-only projects)

For each pattern detected:
- Severity level (low/medium/high/critical)
- Specific indicators
- Counter-suggestions

Respond in structured format."#;

    pub const STRATEGIC_FIT: &'static str = r#"Evaluate how this idea fits with the user's current strategy.

Idea: {idea}
Current Strategy: {strategy}
Current Stack: {stack}
Timeline: {timeline}

Assess:
- Stack compliance (0-10)
- Shipping habit alignment (0-10)
- Public accountability (0-10)
- Revenue testing fit (0-10)

Overall strategic fit: Score (0-10) + Recommendations"#;

    pub const RECOMMENDATION_ENGINE: &'static str = r#"Based on all analysis, provide a clear recommendation.

Idea: {idea}
Mission Score: {mission_score}
Pattern Score: {pattern_score}
Strategic Score: {strategic_score}

Recommendation: PRIORITIZE NOW / GOOD ALIGNMENT / CONSIDER LATER / AVOID FOR NOW

Provide:
1. Clear recommendation with reasoning
2. Specific next steps (2-3 actions)
3. Timeline suggestion
4. Success criteria

Keep it actionable and specific."#;

    pub const CONTEXTUAL_FACTORS: &'static str = r#"Consider contextual factors for this idea.

Current context:
- Deadline pressure: {deadline_pressure}
- Energy levels: {energy}
- Resource availability: {resources}
- Recent patterns: {recent_patterns}

How do these factors affect the timing and feasibility of this idea?

Adjust the recommendation based on current context."#;
}
