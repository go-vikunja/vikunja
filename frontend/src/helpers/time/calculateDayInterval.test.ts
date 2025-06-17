import {test, expect} from 'vitest'

import {calculateDayInterval} from './calculateDayInterval'

const days = {
    monday:    1,
    tuesday:   2,
    wednesday: 3,
    thursday:  4,
    friday:    5,
    saturday:  6,
    sunday:    0,
} as Record<string, number>

for (const n in days) {
    test(`today on a ${n}`, () => {
        expect(calculateDayInterval('today', days[n])).toBe(0)
    })
}

for (const n in days) {
    test(`tomorrow on a ${n}`, () => {
        expect(calculateDayInterval('tomorrow', days[n])).toBe(1)
    })
}

const nextMonday = {
    monday:    0,
    tuesday:   6,
    wednesday: 5,
    thursday:  4,
    friday:    3,
    saturday:  2,
    sunday:    1,
} as Record<string, number>

for (const n in nextMonday) {
    test(`next monday on a ${n}`, () => {
        expect(calculateDayInterval('nextMonday', days[n])).toBe(nextMonday[n])
    })
}

const thisWeekend = {
    monday:    5,
    tuesday:   4,
    wednesday: 3,
    thursday:  2,
    friday:    1,
    saturday:  0,
    sunday:    0,
} as Record<string, number>

for (const n in thisWeekend) {
    test(`this weekend on a ${n}`, () => {
        expect(calculateDayInterval('thisWeekend', days[n])).toBe(thisWeekend[n])
    })
}

const laterThisWeek = {
    monday:    2,
    tuesday:   2,
    wednesday: 2,
    thursday:  2,
    friday:    0,
    saturday:  0,
    sunday:    0,
} as Record<string, number>

for (const n in laterThisWeek) {
    test(`later this week on a ${n}`, () => {
        expect(calculateDayInterval('laterThisWeek', days[n])).toBe(laterThisWeek[n])
    })
}

const laterNextWeek = {
    monday:    7 + 2,
    tuesday:   7 + 2,
    wednesday: 7 + 2,
    thursday:  7 + 2,
    friday:    7 + 0,
    saturday:  7 + 0,
    sunday:    7 + 0,
} as Record<string, number>

for (const n in laterNextWeek) {
    test(`later next week on a ${n} (this week)`, () => {
        expect(calculateDayInterval('laterNextWeek', days[n])).toBe(laterNextWeek[n])
    })
}

for (const n in days) {
    test(`next week on a ${n}`, () => {
        expect(calculateDayInterval('nextWeek', days[n])).toBe(7)
    })
}
