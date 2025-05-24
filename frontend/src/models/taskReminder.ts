import AbstractModel from './abstractModel'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import type {IReminderPeriodRelativeTo} from '@/types/IReminderPeriodRelativeTo'

export default class TaskReminderModel extends AbstractModel<ITaskReminder> implements ITaskReminder {
	reminder: Date | null
	relativePeriod = 0
	relativeTo: IReminderPeriodRelativeTo | null = null

	constructor(data: Partial<ITaskReminder> = {}) {
		super()
		this.assignData(data)
		this.reminder = parseDateOrNull(data.reminder)
		if (this.relativeTo === '') {
			this.relativeTo = null
		}
	}

}
