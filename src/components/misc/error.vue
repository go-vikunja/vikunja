<template>
	<div class="notification is-danger">
		<i18n path="loadingError.failed">
			<a @click="() => location.reload()">{{ $t('loadingError.tryAgain') }}</a>
			<a href="https://vikunja.io/contact/">{{ $t('loadingError.contact') }}</a>
		</i18n>
	</div>
</template>

<script>
export default {
	name: 'error',
}
</script>
