<template>
	<component
		:is="componentNodeName"
		v-bind="elementBindings"
		:to="to"
		class="dropdown-item">
		<span class="icon" v-if="icon">
			<icon :icon="icon"/>
		</span>
		<span>
			<slot></slot>
		</span>
	</component>
</template>

<script lang="ts" setup>
import {ref, useAttrs, watchEffect} from 'vue'

const props = defineProps<{
	to?: object,
	icon?: string | string[],
}>()

const componentNodeName = ref<Node['nodeName']>('a')
const elementBindings = ref({})

const attrs = useAttrs()
watchEffect(() => {
	let nodeName = 'a'

	if (props.to) {
		nodeName = 'router-link'
	}

	if ('href' in attrs) {
		nodeName = 'BaseButton'
	}

	componentNodeName.value = nodeName
	elementBindings.value = {
		...attrs,
	}
})
</script>

<style scoped lang="scss">
.dropdown-item {
	color: var(--text);
	display: block;
	font-size: 0.875rem;
	line-height: 1.5;
	padding: $item-padding;
	position: relative;
}

a.dropdown-item,
button.dropdown-item {
	text-align: inherit;
	white-space: nowrap;
	width: 100%;
	display: flex;
	align-items: center;
	justify-content: left !important;

	&:hover {
		background-color: var(--grey-100) !important;
	}

	&.is-active {
		background-color: var(--link);
		color: var(--link-invert);
	}

	.icon {
		padding-right: .5rem;
	}

	.icon:not(.has-text-success) {
		color: var(--grey-300) !important;
	}

	&.has-text-danger .icon {
		color: var(--danger) !important;
	}

	&.is-disabled {
		cursor: not-allowed;

		&:hover {
			background-color: transparent;
		}
	}
}

</style>
