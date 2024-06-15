<template>
	<Message variant="danger">
		<i18n-t
			keypath="loadingError.failed"
			scope="global"
		>
			<ButtonLink @click="reload">
				{{ $t('loadingError.tryAgain') }}
			</ButtonLink>
			<ButtonLink href="https://vikunja.io/contact/">
				{{ $t('loadingError.contact') }}
			</ButtonLink>
		</i18n-t>
	</Message>
</template>

<script lang="ts" setup>
import Message from '@/components/misc/Message.vue'
import ButtonLink from '@/components/misc/ButtonLink.vue'

defineOptions({
	inheritAttrs: false,
})

function reload() {
	window.location.reload()
}
</script>
