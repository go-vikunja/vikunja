<template>
	<BaseButton
		class="dropdown-item"
		:disabled="disabled"
		:class="{'is-disabled': disabled}"
	>
		<span
			v-if="icon"
			class="icon is-small"
			:class="iconClass"
		>
			<Icon :icon="icon" />
		</span>
		<span>
			<slot />
		</span>
	</BaseButton>
</template>

<script lang="ts" setup>
import BaseButton, {type BaseButtonProps} from '@/components/base//BaseButton.vue'
import Icon from '@/components/misc/Icon'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

export interface DropDownItemProps extends /* @vue-ignore */ BaseButtonProps {
	icon?: IconProp,
	iconClass?: object | string,
	disabled?: boolean,
}

defineProps<DropDownItemProps>()
</script>

<style scoped lang="scss">
.dropdown-item {
	color: var(--text);
	font-size: 0.875rem;
	line-height: 1.5;
	padding: $item-padding;
	position: relative;
	text-align: inherit;
	white-space: nowrap;
	inline-size: 100%;
	display: flex;
	align-items: center;
	justify-content: left !important;

	&.is-active {
		background-color: var(--link);
		color: var(--link-invert);
	}

	&:hover:not(.is-disabled) {
		background-color: var(--grey-100);
	}

	&.is-disabled {
		cursor: not-allowed;
	}
}

.icon {
	padding-inline-end: .5rem;
	color: var(--grey-300);
}

.has-text-danger .icon {
	color: var(--danger) !important;
}
</style>
