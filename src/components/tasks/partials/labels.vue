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

<style lang="scss" scoped>
.label-wrapper {
	display: inline;
}

.tag {
	& + & {
		margin-left: 0.5rem;
	}
}
</style>