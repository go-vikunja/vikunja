<template>
	<a
		class="button"
		:class="{
			'is-loading': loading,
			'has-no-shadow': !shadow,
			'is-primary': type === 'primary',
			'is-outlined': type === 'secondary',
			'is-text is-inverted has-no-shadow underline-none':
				type === 'tertary',
		}"
		:disabled="disabled || null"
		@click="click"
		:href="href !== '' ? href : null"
	>
		<icon :icon="icon" v-if="showIconOnly"/>
		<span class="icon is-small" v-else-if="icon !== ''">
			<icon :icon="icon"/>
		</span>
		<slot></slot>
	</a>
</template>

<script>
export default {
	name: 'x-button',
	props: {
		type: {
			type: String,
			default: 'primary',
		},
		href: {
			type: String,
			default: '',
		},
		to: {
			default: false,
		},
		icon: {
			default: '',
		},
		loading: {
			type: Boolean,
			default: false,
		},
		shadow: {
			type: Boolean,
			default: true,
		},
		disabled: {
			type: Boolean,
			default: false,
		},
	},
	emits: ['click'],
	computed: {
		showIconOnly() {
			return this.icon !== '' && typeof this.$slots.default === 'undefined'
		},
	},
	methods: {
		click(e) {
			if (this.disabled) {
				return
			}

			if (this.to !== false) {
				this.$router.push(this.to)
			}

			this.$emit('click', e)
		},
	},
}
</script>

<style lang="scss" scoped>
.button {
  transition: all $transition;
  border: 0;
  text-transform: uppercase;
  font-size: 0.85rem;
  font-weight: bold;
  height: $button-height;
  box-shadow: var(--shadow-sm);

  &.is-hovered,
  &:hover {
    box-shadow: var(--shadow-md);
  }

  &.fullheight {
    padding-right: 7px;
    height: 100%;
  }

  &.is-active,
  &.is-focused,
  &:active,
  &:focus,
  &:focus:not(:active) {
    box-shadow: var(--shadow-xs) !important;
  }

  &.is-primary.is-outlined:hover {
    color: var(--white);
  }

  &.is-small {
    border-radius: $radius;
  }
}

.underline-none {
  text-decoration: none !important;
}
</style>