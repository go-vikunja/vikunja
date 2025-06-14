import AbstractService from './abstractService'
import type {IPasswordUpdate} from '@/modelTypes/IPasswordUpdate'

export default class PasswordUpdateService extends AbstractService<IPasswordUpdate> {
	constructor() {
		super({
			update: '/user/password',
		})
	}
}
