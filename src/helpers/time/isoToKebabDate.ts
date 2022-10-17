import {format} from 'date-fns'
import {DATEFNS_DATE_FORMAT_KEBAB} from '@/constants/date' 
import type {DateISO} from '@/types/DateISO'
import type {DateKebab} from '@/types/DateKebab'

export function isoToKebabDate(isoDate: DateISO) {
	return format(new Date(isoDate), DATEFNS_DATE_FORMAT_KEBAB) as DateKebab
}