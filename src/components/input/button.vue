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
		:disabled="disabled"
		@click="click"
		:href="href !== '' ? href : false"
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