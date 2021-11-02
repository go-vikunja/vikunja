<template>
	<div class="notification is-danger">
		<i18n-t keypath="loadingError.failed">
			<a @click="reload">{{ $t('loadingError.tryAgain') }}</a>
			<a href="https://vikunja.io/contact/" rel="noreferrer noopener nofollow" target="_blank">{{ $t('loadingError.contact') }}</a>
		</i18n-t>
	</div>
</template>

<script>
export default {
	name: 'error',
	methods: {
		reload() {
			window.location.reload()
		},
	}
}
</script>
