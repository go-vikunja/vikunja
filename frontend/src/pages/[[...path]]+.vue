<template>
	<div class="content has-text-centered">
		<h1>{{ $t('404.title') }}</h1>
		<p>{{ $t('404.text') }}</p>
	</div>
</template>

<script setup lang="ts">
import {useTitle} from '@/composables/useTitle'

// regarding the name of this component, see
// https://github.com/posva/unplugin-vue-router/issues/412#issuecomment-2150643118
definePage({
	name: 'not-found',
})

defineOptions({ name: 'NotFound' })

useTitle(() => '404')
</script>
