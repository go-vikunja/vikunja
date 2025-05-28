import AbstractService from './abstractService'
import type {IUserWebhookSetting, IWebhookNotificationType} from '@/modelTypes/IUserWebhookSetting'
import UserWebhookSettingModel from '@/models/userWebhookSetting'

export default class UserWebhookSettingsService extends AbstractService<IUserWebhookSetting> {
	constructor() {
		super({
			getAll: '/user/settings/webhooks',
		})
	}

	modelFactory(data: Partial<IUserWebhookSetting>) {
		return new UserWebhookSettingModel(data)
	}

	/**
	 * Get all webhook settings for the current user
	 */
	async getAll(): Promise<IUserWebhookSetting[]> {
		const response = await this.http.get('/user/settings/webhooks')
		if (!Array.isArray(response.data)) {
			return []
		}
		return response.data.map(s => this.modelFactory(s))
	}

	/**
	 * Get a webhook setting by notification type
	 */
	async getByType(notificationType: string): Promise<IUserWebhookSetting> {
		const response = await this.http.get(`/user/settings/webhooks/${encodeURIComponent(notificationType)}`)
		return this.modelFactory(response.data as IUserWebhookSetting)
	}

	/**
	 * Create or update a webhook setting
	 */
	async saveByType(notificationType: string, targetUrl: string, enabled = true): Promise<IUserWebhookSetting> {
		const response = await this.http.put(`/user/settings/webhooks/${encodeURIComponent(notificationType)}`, {
			target_url: targetUrl,
			enabled,
		})
		return this.modelFactory(response.data as IUserWebhookSetting)
	}

	/**
	 * Delete a webhook setting
	 */
	async deleteByType(notificationType: string): Promise<void> {
		await this.http.delete(`/user/settings/webhooks/${encodeURIComponent(notificationType)}`)
	}

	/**
	 * Get all available webhook notification types
	 */
	async getAvailableTypes(): Promise<IWebhookNotificationType[]> {
		const response = await this.http.get('/user/settings/webhooks/types')
		if (!Array.isArray(response.data)) {
			return []
		}
		return response.data as IWebhookNotificationType[]
	}
}
