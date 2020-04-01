<template>
	<div class="label-wrapper">
		<span class="tag" v-for="label in labels" :style="{'background': label.hex_color, 'color': label.textColor}" :key="label.id">
			<span>{{ label.title }}</span>
		</span>
	</div>
</template>

<script>
	export default {
		name: 'labels',
		props: {
			labels: {
				type: Array,
				required: true,
			}
		}
	}
</script>

<style scoped>
	.label-wrapper {
		display: inline;
	}
</style>