import AbstractModel from "./abstractModel";

export default class PasswordResetModel extends AbstractModel {
	constructor(data) {
		super(data)
		
		this.token = localStorage.getItem('passwordResetToken')
	}
	
	defaults() {
		return {
			token: '',
			new_password: '',
			email: '',
		}
	}
}