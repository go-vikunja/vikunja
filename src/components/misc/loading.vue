<template>
	<div class="loader-container is-loading"></div>
</template>

<script>
	export default {
		name: 'loading'
	}
</script>

<style scoped>
	.loader-container {
		height: 100%;
		min-height: 200px;
		width: 100%;
		min-width: 600px;
		max-width: 100vw;
	}
</style>