import type { IPasswordUpdate } from '@/models/passwordUpdate'
import AbstractService from './abstractService'

export default class PasswordUpdateService extends AbstractService<IPasswordUpdate> {
	constructor() {
		super({
			update: '/user/password',
		})
	}
}