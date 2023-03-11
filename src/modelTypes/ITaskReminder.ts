import type { IAbstract } from './IAbstract'
import type { IReminderPeriodRelativeTo } from '@/types/IReminderPeriodRelativeTo'

export interface ITaskReminder extends IAbstract {
	reminder: Date
	relativePeriod: number
	relativeTo: IReminderPeriodRelativeTo
}