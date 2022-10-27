<template>
	<message variant="danger">
		<i18n-t keypath="loadingError.failed" scope="global">
			<ButtonLink @click="reload">{{ $t('loadingError.tryAgain') }}</ButtonLink>
			<ButtonLink href="https://vikunja.io/contact/">{{ $t('loadingError.contact') }}</ButtonLink>
		</i18n-t>
	</message>
</template>

<script lang="ts">
export default {
	inheritAttrs: false,
}
</script>

<script lang="ts" setup>
import Message from '@/components/misc/message.vue'
import ButtonLink from '@/components/misc/ButtonLink.vue'

function reload() {
	window.location.reload()
}
</script>
