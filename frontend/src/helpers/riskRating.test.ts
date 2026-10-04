import {describe, expect, it} from 'vitest'

import {
	isRiskOverdue,
	isRiskRating,
	isRiskStatus,
	REOPEN_STATUS,
	RISK_RATING_BANDS,
	RISK_RATINGS,
	RISK_STATUSES,
	riskNoteKind,
	riskRating,
	riskScore,
	statusTargets,
} from './riskRating'

describe('riskRating', () => {
	// The same boundaries as TestRiskRating in pkg/models/risk_test.go.
	it.each([
		[1, 'low'], [4, 'low'],
		[5, 'medium'], [9, 'medium'],
		[10, 'high'], [16, 'high'],
		[17, 'critical'], [25, 'critical'],
	])('score %i is %s', (score, rating) => {
		expect(riskRating(score)).toBe(rating)
	})

	it.each([0, 26, -3, 100])('score %i has no rating', score => {
		expect(riskRating(score)).toBeNull()
	})

	it('every score from 1 to 25 has exactly one rating', () => {
		for (let score = 1; score <= 25; score++) {
			const matches = RISK_RATING_BANDS.filter(b => score >= b.min && score <= b.max)
			expect(matches, `score ${score}`).toHaveLength(1)
		}
	})

	it('bands are ordered and list every rating', () => {
		expect(RISK_RATING_BANDS.map(b => b.rating)).toEqual(RISK_RATINGS)
		expect(RISK_RATING_BANDS[0].min).toBe(1)
		expect(RISK_RATING_BANDS[RISK_RATING_BANDS.length - 1].max).toBe(25)
	})
})

describe('riskScore', () => {
	it('is probability times impact', () => {
		expect(riskScore(4, 5)).toBe(20)
		expect(riskScore(1, 1)).toBe(1)
		expect(riskScore(5, 5)).toBe(25)
	})
})

describe('status changes', () => {
	it.each(RISK_STATUSES)('%s can move to every other status and not to itself', current => {
		const targets = statusTargets(current)
		expect(targets).toHaveLength(RISK_STATUSES.length - 1)
		expect(targets).not.toContain(current)
	})

	it('closing offers a resolution note', () => {
		expect(riskNoteKind('open', 'closed')).toBe('close')
		expect(riskNoteKind('mitigating', 'closed')).toBe('close')
		expect(riskNoteKind('accepted', 'closed')).toBe('close')
	})

	it('moving away from closed is a reopening', () => {
		expect(riskNoteKind('closed', 'open')).toBe('reopen')
		expect(riskNoteKind('closed', 'mitigating')).toBe('reopen')
		expect(riskNoteKind('closed', 'accepted')).toBe('reopen')
	})

	it('every other change has a plain note', () => {
		expect(riskNoteKind('open', 'mitigating')).toBe('plain')
		expect(riskNoteKind('accepted', 'open')).toBe('plain')
		expect(riskNoteKind('closed', 'closed')).toBe('plain')
	})

	it('reopens as open', () => {
		expect(REOPEN_STATUS).toBe('open')
		expect(statusTargets('closed')).toContain(REOPEN_STATUS)
	})
})

describe('isRiskOverdue', () => {
	const now = new Date('2026-10-04T12:00:00Z')

	it('is overdue when the due date passed and the risk is not closed', () => {
		expect(isRiskOverdue('open', '2026-10-03T00:00:00Z', now)).toBe(true)
		expect(isRiskOverdue('mitigating', '2026-10-04T11:59:59Z', now)).toBe(true)
		expect(isRiskOverdue('accepted', '2026-10-01T00:00:00Z', now)).toBe(true)
	})

	it('is not overdue before the due date, without one, or when closed', () => {
		expect(isRiskOverdue('open', '2026-10-05T00:00:00Z', now)).toBe(false)
		expect(isRiskOverdue('open', null, now)).toBe(false)
		expect(isRiskOverdue('open', undefined, now)).toBe(false)
		expect(isRiskOverdue('closed', '2026-01-01T00:00:00Z', now)).toBe(false)
	})

	it('ignores a due date that is not a date', () => {
		expect(isRiskOverdue('open', 'soon', now)).toBe(false)
	})
})

describe('type guards', () => {
	it('know the statuses and ratings', () => {
		expect(RISK_STATUSES).toEqual(['open', 'mitigating', 'accepted', 'closed'])
		expect(isRiskStatus('closed')).toBe(true)
		expect(isRiskStatus('Closed')).toBe(false)
		expect(isRiskStatus(undefined)).toBe(false)
		expect(isRiskRating('critical')).toBe(true)
		expect(isRiskRating('huge')).toBe(false)
	})
})
