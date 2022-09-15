<template>
	<span
		:style="{backgroundColor: color }"
		class="color-bubble"
	></span>
</template>

<script lang="ts" setup>
defineProps({
	color: {
		type: String,
		required: true,
	},
})
</script>

<style scoped>
.color-bubble {
	display: inline-block;
	border-radius: 100%;
	margin-right: 4px;
	height: 10px;
	width: 10px;
	flex-shrink: 0;
}
</style>