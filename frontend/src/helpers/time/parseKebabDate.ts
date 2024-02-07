import {parse} from 'date-fns'
import {DATEFNS_DATE_FORMAT_KEBAB} from '@/constants/date'
import type {DateKebab} from '@/types/DateKebab'

export function parseKebabDate(date: DateKebab): Date {
	return parse(date, DATEFNS_DATE_FORMAT_KEBAB, new Date())
}