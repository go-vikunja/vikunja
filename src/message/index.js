export const getErrorText = (r, $t) => {

	if (r.response && r.response.data) {
		if(r.response.data.code) {
			const path = `error.${r.response.data.code}`
			const message = $t(path)

			// If message and path are equal no translation exists for that error code
			if (path !== message) {
				return [
					r.message,
					message,
				]
			}
		}

		if (r.response.data.message) {
			return [
				r.message,
				r.response.data.message,
			]
		}
	}

	return [r.message]
}

export default {
	error(e, context, $t, actions = []) {
		context.$notify({
			type: 'error',
			title: $t('error.error'),
			text: getErrorText(e, $t),
			actions: actions,
		})
	},
	success(e, context, $t, actions = []) {
		context.$notify({
			type: 'success',
			title: $t('error.success'),
			text: getErrorText(e, $t),
			data: {
				actions: actions,
			},
		})
	},
}