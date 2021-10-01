<template>
	<transition name="fade">
		<filters
			v-if="visibleInternal"
			v-model="value"
			ref="filters"
		/>
	</transition>
</template>

<script>
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import Filters from '../../../components/list/partials/filters'

export default {
	name: 'filter-popup',
	components: {
		Filters,
	},
	props: {
		modelValue: {
			required: true,
		},
		visible: {
			type: Boolean,
			default: false,
		},
	},
	emits: ['update:modelValue'],
	data() {
		return {
			visibleInternal: false,
		}
	},
	computed: {
		value: {
			get() {
				return this.modelValue
			},
			set(value) {
				this.$emit('update:modelValue', value)
			},
		},
	},
	mounted() {
		document.addEventListener('click', this.hidePopup)
	},
	beforeUnmount() {
		document.removeEventListener('click', this.hidePopup)
	},
	watch: {
		modelValue: {
			handler(value) {
				this.params = value
			},
			immediate: true,
		},
		visible() {
			this.visibleInternal = !this.visibleInternal
		},
	},
	methods: {
		hidePopup(e) {
			if (!this.visibleInternal) {
				return
			}

			closeWhenClickedOutside(e, this.$refs.filters.$el, () => {
				this.visibleInternal = false
			})
		},
	},
}
</script>
