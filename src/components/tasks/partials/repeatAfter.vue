<template>
	<div class="control repeat-after-input columns">
		<p class="column is-1">
			Each
		</p>
		<div class="column is-7 field has-addons">
			<div class="control">
				<input
						class="input"
						placeholder="Specify an amount..."
						v-model="repeatAfter.amount"
						@change="updateData"/>
			</div>
			<div class="control">
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
		<fancycheckbox
				class="column"
				@change="updateData"
				v-model="task.repeatFromCurrentDate"
				v-tooltip="'When marking the task as done, all dates will be set relative to the current date rather than the date they had before.'"
		>
			Repeat from current date
		</fancycheckbox>
	</div>
</template>

<script>
	import Fancycheckbox from '../../input/fancycheckbox'

	export default {
		name: 'repeatAfter',
		components: {Fancycheckbox},
		data() {
			return {
				task: {},
				repeatAfter: {
					amount: 0,
					type: '',
				},
			}
		},
		props: {
			value: {
				default: () => {
				},
				required: true,
			}
		},
		watch: {
			value(newVal) {
				this.task = newVal
				if (typeof newVal.repeatAfter !== 'undefined') {
					this.repeatAfter = newVal.repeatAfter
				}
			},
		},
		mounted() {
			this.task = this.value
			if (typeof this.value.repeatAfter !== 'undefined') {
				this.repeatAfter = this.value.repeatAfter
			}
		},
		methods: {
			updateData() {
				this.task.repeatAfter = this.repeatAfter
				this.$emit('input', this.task)
				this.$emit('change')
			}
		},
	}
</script>

<style scoped lang="scss">
	p {
		padding-top: 6px;
	}

	.field.has-addons {

		margin-bottom: .5rem;

		.control .select select {
			height: 2.5em;
		}
	}

	.columns {
		align-items: center;
	}
</style>