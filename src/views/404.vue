<template>
	<div class="content has-text-centered">
		<h1>{{ $t('404.title') }}</h1>
		<p>{{ $t('404.text') }}</p>
	</div>
</template>

<script>
export default {
	name: '404',
	mounted() {
		this.setTitle('404')
	},
}
</script>
