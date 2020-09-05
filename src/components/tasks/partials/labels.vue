<template>
	<div class="label-wrapper">
		<span
			:key="label.id"
			:style="{'background': label.hexColor, 'color': label.textColor}"
			class="tag"
			v-for="label in labels">
			<span>{{ label.title }}</span>
		</span>
	</div>
</template>

<script>
export default {
	name: 'labels',
	props: {
		labels: {
			required: true,
		},
	},
}
</script>

<style scoped>
.label-wrapper {
	display: inline;
}
</style>