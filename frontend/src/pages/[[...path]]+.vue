<template>
	<div class="content has-text-centered">
		<h1>{{ $t('404.title') }}</h1>
		<p>{{ $t('404.text') }}</p>
	</div>
</template>

<script setup lang="ts">
import {useTitle} from '@/composables/useTitle'

useTitle(() => '404')
</script>
