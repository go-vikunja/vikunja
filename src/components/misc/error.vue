<template>
	<message variant="danger">
		<i18n-t keypath="loadingError.failed">
			<BaseButton @click="reload">{{ $t('loadingError.tryAgain') }}</BaseButton>
			<BaseButton href="https://vikunja.io/contact/">{{ $t('loadingError.contact') }}</BaseButton>
		</i18n-t>
	</message>
</template>

<script lang="ts" setup>
import Message from '@/components/misc/message.vue'
import BaseButton from '@/components/base/BaseButton.vue'

function reload() {
	window.location.reload()
}
</script>
