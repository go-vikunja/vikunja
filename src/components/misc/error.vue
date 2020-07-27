<template>
	<div class="notification is-danger">
		Loading failed, please <a @click="() => location.reload()">try again</a>.
		If the error persists, please <a href="https://vikunja.io/contact/">contact us</a>.
	</div>
</template>

<script>
	export default {
		name: 'error'
	}
</script>
