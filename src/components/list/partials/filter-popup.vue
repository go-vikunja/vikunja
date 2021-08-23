<template>
	<transition name="fade">
		<filters
			@change="change"
			v-if="visibleInternal"
			v-model="params"
			ref="filters"
		/>
	</transition>
</template>

<script>
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import Filters from '../../../components/list/partials/filters'

export default {
	name: 'filter-popup',
	emits: ['update:modelValue', 'change'],
	data() {
		return {
			params: null,
			visibleInternal: false,
		}
	},
	components: {
		Filters,
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
	props: {
		modelValue: {
			required: true,
		},
		visible: {
			type: Boolean,
			default: false,
		},
	},
	methods: {
		change() {
			this.$emit('change', this.params)
			this.$emit('update:modelValue', this.params)
		},
		hidePopup(e) {
			if (this.visibleInternal) {
				closeWhenClickedOutside(e, this.$refs.filters.$el, () => {
					this.visibleInternal = false
				})
			}
		},
	},
}
</script>
