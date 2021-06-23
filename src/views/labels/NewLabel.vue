<template>
	<create-edit
		:title="$t('label.create.title')"
		@create="newLabel()"
		:create-disabled="label.title === ''"
	>
		<div class="field">
			<label class="label" for="labelTitle">{{ $t('label.attributes.title') }}</label>
			<div
				class="control is-expanded"
				:class="{ 'is-loading': loading }"
			>
				<input
					:class="{ disabled: loading }"
					class="input"
					:placeholder="$t('label.attributes.titlePlaceholder')"
					type="text"
					id="labelTitle"
					v-focus
					v-model="label.title"
					@keyup.enter="newLabel()"
				/>
			</div>
		</div>
		<p class="help is-danger" v-if="showError && label.title === ''">
			{{ $t('label.create.titleRequired') }}
		</p>
		<div class="field">
			<label class="label">{{ $t('label.attributes.color') }}</label>
			<div class="control">
				<color-picker v-model="label.hexColor"/>
			</div>
		</div>
	</create-edit>
</template>

<script>
import labelModel from '../../models/label'
import LabelModel from '../../models/label'
import CreateEdit from '@/components/misc/create-edit'
import ColorPicker from '../../components/input/colorPicker'
import {mapState} from 'vuex'
import {LOADING, LOADING_MODULE} from '@/store/mutation-types'

export default {
	name: 'NewLabel',
	data() {
		return {
			label: labelModel,
			showError: false,
		}
	},
	components: {
		CreateEdit,
		ColorPicker,
	},
	created() {
		this.label = new LabelModel()
	},
	mounted() {
		this.setTitle(this.$t('label.create.title'))
	},
	computed: mapState({
		loading: state => state[LOADING] && state[LOADING_MODULE] === 'labels',
	}),
	methods: {
		newLabel() {
			if (this.label.title === '') {
				this.showError = true
				return
			}
			this.showError = false

			this.$store.dispatch('labels/createLabel', this.label)
				.then(r => {
					this.$router.push({
						name: 'labels.index',
						params: {id: r.id},
					})
					this.success({message: this.$t('label.create.success')})
				})
				.catch((e) => {
					this.error(e)
				})
		},
	},
}
</script>
