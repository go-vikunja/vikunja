<template>
	<message variant="danger">
		<i18n-t keypath="loadingError.failed">
			<a @click="reload">{{ $t('loadingError.tryAgain') }}</a>
			<a href="https://vikunja.io/contact/" rel="noreferrer noopener nofollow" target="_blank">{{ $t('loadingError.contact') }}</a>
		</i18n-t>
	</message>
</template>

<script>
import Message from '@/components/misc/message'

export default {
	name: 'error',
	components: {Message},
	methods: {
		reload() {
			window.location.reload()
		},
	},
}
</script>
