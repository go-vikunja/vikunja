<template>
	<BaseButton class="dropdown-item">
		<span class="icon" v-if="icon">
			<Icon :icon="icon"/>
		</span>
		<span>
			<slot  />
		</span>
	</BaseButton>
</template>

<script lang="ts" setup>
import BaseButton, { type BaseButtonProps } from '@/components/base//BaseButton.vue'
import Icon from '@/components/misc/Icon'
import type { IconProp } from '@fortawesome/fontawesome-svg-core'

export interface DropDownItemProps extends BaseButtonProps {
	icon?: IconProp,
}

defineProps<DropDownItemProps>()
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
