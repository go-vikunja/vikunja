export default {
	error(e, context, actions = []) {
		// Build the notification text from error response
		let err = e.message
		if (e.response && e.response.data && e.response.data.message) {
			err += '<br/>' + e.response.data.message
		}

		// Fire a notification
		context.$notify({
			type: 'error',
			title: 'Error',
			text: err,
			actions: actions,
		})
	},
	success(e, context, actions = []) {
		// Build the notification text from error response
		let err = e.message
		if (e.response && e.response.data && e.response.data.message) {
			err += '<br/>' + e.response.data.message
		}

		// Fire a notification
		context.$notify({
			type: 'success',
			title: 'Success',
			text: err,
			data: {
				actions: actions,
			},
		})
	},
}