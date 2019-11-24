<template>
	<div class="control repeat-after-input columns">
		<div class="column">
			<p>
				Each
			</p>
		</div>
		<div class="column is-two-fifths">
			<input class="input" placeholder="Specify an amount..." v-model="repeatAfter.amount" @change="updateData"/>
		</div>
		<div class="column is-two-fifths">
			<div class="select">
				<select v-model="repeatAfter.type" @change="updateData">
					<option value="hours">Hours</option>
					<option value="days">Days</option>
					<option value="weeks">Weeks</option>
					<option value="months">Months</option>
					<option value="years">Years</option>
				</select>
			</div>
		</div>
	</div>

</template>

<script>
	export default {
		name: 'repeatAfter',
		data() {
			return {
				repeatAfter: {},
			}
		},
		props: {
			value: {
				default: () => {},
				required: true,
			}
		},
		watch: {
			value(newVal) {
				this.repeatAfter = newVal
			},
		},
		mounted() {
			this.repeatAfter = this.value
		},
		methods: {
			updateData() {
				this.$emit('input', this.repeatAfter)
				this.$emit('change')
			}
		},
	}
</script>

<style scoped>
	p {
		padding-top: 6px;
	}
</style>