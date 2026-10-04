export type RiskRating = 'low' | 'medium' | 'high' | 'critical'
export type RiskStatus = 'open' | 'mitigating' | 'accepted' | 'closed'

export const RISK_RATINGS: RiskRating[] = ['low', 'medium', 'high', 'critical']
export const RISK_STATUSES: RiskStatus[] = ['open', 'mitigating', 'accepted', 'closed']

// The inclusive score bands. The server decides the rating of a stored risk (pkg/models/risk.go,
// riskRatingBands); this table only drives the live preview of the form, and a test on each side pins
// the same boundaries so they cannot drift apart.
export const RISK_RATING_BANDS: {rating: RiskRating, min: number, max: number}[] = [
	{rating: 'low', min: 1, max: 4},
	{rating: 'medium', min: 5, max: 9},
	{rating: 'high', min: 10, max: 16},
	{rating: 'critical', min: 17, max: 25},
]

export const MIN_RISK_LEVEL = 1
export const MAX_RISK_LEVEL = 5

export function riskScore(probability: number, impact: number): number {
	return probability * impact
}

export function riskRating(score: number): RiskRating | null {
	return RISK_RATING_BANDS.find(band => score >= band.min && score <= band.max)?.rating ?? null
}

export function isRiskStatus(value: unknown): value is RiskStatus {
	return typeof value === 'string' && (RISK_STATUSES as string[]).includes(value)
}

export function isRiskRating(value: unknown): value is RiskRating {
	return typeof value === 'string' && (RISK_RATINGS as string[]).includes(value)
}

// Any status can follow any other, so a status can move to every one except itself.
export function statusTargets(current: RiskStatus): RiskStatus[] {
	return RISK_STATUSES.filter(status => status !== current)
}

export type RiskNoteKind = 'close' | 'reopen' | 'plain'

// What the note of a status change is for. Closing offers a resolution, moving away from closed is a
// reopening, everything else is a plain remark for the history.
export function riskNoteKind(from: RiskStatus, to: RiskStatus): RiskNoteKind {
	if (to === 'closed' && from !== 'closed') return 'close'
	if (from === 'closed' && to !== 'closed') return 'reopen'
	return 'plain'
}

// A closed risk is reopened with one click, as an open risk and without a note.
export const REOPEN_STATUS: RiskStatus = 'open'

// Not closed and the due date is in the past. The same comparison the server makes for the overdue filter.
export function isRiskOverdue(status: RiskStatus, dueDate: string | null | undefined, now = new Date()): boolean {
	if (status === 'closed' || !dueDate) {
		return false
	}
	const due = new Date(dueDate)
	return !Number.isNaN(due.getTime()) && due.getTime() < now.getTime()
}

// Colours of a rating, as CSS variables of the theme, with a plain fallback for the print view.
export const RISK_RATING_COLORS: Record<RiskRating, {background: string, text: string, print: string}> = {
	low: {background: 'var(--success)', text: 'var(--white)', print: '#8bc34a'},
	medium: {background: 'var(--warning)', text: 'var(--grey-900)', print: '#ffc107'},
	high: {background: 'var(--orange, #f57c00)', text: 'var(--white)', print: '#fb8c00'},
	critical: {background: 'var(--danger)', text: 'var(--white)', print: '#e53935'},
}
