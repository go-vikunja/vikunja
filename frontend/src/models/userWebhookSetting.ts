import AbstractModel from './abstractModel'

import type {IUserWebhookSetting} from '@/modelTypes/IUserWebhookSetting'

export default class UserWebhookSettingModel extends AbstractModel<IUserWebhookSetting> implements IUserWebhookSetting {
	id = 0
	userId = 0
	notificationType = ''
	enabled = true
	targetUrl = ''
	created = new Date()
	updated = new Date()

	constructor(data: Partial<IUserWebhookSetting> = {}) {
		super()
		this.assignData(data)
	}
}
