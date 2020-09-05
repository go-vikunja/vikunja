<template>
	<div class="content has-text-centered">
		<h1>Not found</h1>
		<p>The page you requested does not exist.</p>
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
