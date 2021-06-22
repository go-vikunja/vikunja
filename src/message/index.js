const getText = t => {

	if (t.response && t.response.data && t.response.data.message) {
		return [
			t.message,
			t.response.data.message
		]
	}

	return [t.message]
}

export default {
	error(e, context, actions = []) {
		context.$notify({
			type: 'error',
			title: 'Error',
			text: getText(e),
			actions: actions,
		})
	},
	success(e, context, actions = []) {
		context.$notify({
			type: 'success',
			title: 'Success',
			text: getText(e),
			data: {
				actions: actions,
			},
		})
	},
}