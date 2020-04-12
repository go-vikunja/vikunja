import UserModel from './user'
import {merge} from 'lodash'

export default class TeamMemberModel extends UserModel {
	defaults() {
		return merge(
			super.defaults(),
			{
				admin: false,
				teamId: 0,
			}
		)
	}
}