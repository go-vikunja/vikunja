
import AbstractModel from './abstractModel'

export default class UserNameModel extends AbstractModel {
	defaults() {
		return {
			name: '',
		}
	}
}