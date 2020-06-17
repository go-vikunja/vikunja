<template>
	<div class="fancycheckbox" :class="{'is-disabled': disabled}">
		<input @change="updateData" type="checkbox" :id="checkBoxId" :checked="checked" style="display: none;" :disabled="disabled">
		<label :for="checkBoxId" class="check">
			<svg width="18px" height="18px" viewBox="0 0 18 18">
				<path d="M1,9 L1,3.5 C1,2 2,1 3.5,1 L14.5,1 C16,1 17,2 17,3.5 L17,14.5 C17,16 16,17 14.5,17 L3.5,17 C2,17 1,16 1,14.5 L1,9 Z"></path>
				<polyline points="1 9 7 14 15 4"></polyline>
			</svg>
			<span>
				<slot></slot>
			</span>
		</label>
	</div>
</template>

<script>
	export default {
		name: 'fancycheckbox',
		data() {
			return {
				checked: false,
				checkBoxId: '',
			}
		},
		props: {
			value: {
				required: false,
			},
			disabled: {
				type: Boolean,
				required: false,
				default: false,
			},
		},
		watch: {
			value(newVal) {
				this.checked = newVal
			},
		},
		mounted() {
			this.checked = this.value
		},
		created() {
			this.checkBoxId = 'fancycheckbox' + Math.random()
		},
		methods: {
			updateData(e) {
				this.checked = e.target.checked
				this.$emit('input', this.checked)
				this.$emit('change', e.target.checked)
			},
		},
	}
</script>
