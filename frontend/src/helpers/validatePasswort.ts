export function validatePassword(password: string, validateMinLength: boolean = true): string | true {
	if (password === '') {
		return 'user.auth.passwordRequired'
	}

	if (validateMinLength && password.length < 8) {
		return 'user.auth.passwordNotMin'
	}
	
	if (validateMinLength && password.length > 72) {
		return 'user.auth.passwordNotMax'
	}
	
	return true
}
