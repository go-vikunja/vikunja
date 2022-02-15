<template>
	<div>
		<h3>
			<span class="icon is-grey">
				<icon icon="align-left"/>
			</span>
			{{ $t('task.attributes.description') }}
			<transition name="fade">
				<span class="is-small is-inline-flex" v-if="loading && saving">
					<span class="loader is-inline-block mr-2"></span>
					{{ $t('misc.saving') }}
				</span>
				<span class="is-small has-text-success" v-else-if="!loading && saved">
					<icon icon="check"/>
					{{ $t('misc.saved') }}
				</span>
			</transition>
		</h3>
		<editor
			:is-edit-enabled="canWrite"
			:upload-callback="attachmentUpload"
			:upload-enabled="true"
			@change="save"
			:placeholder="$t('task.description.placeholder')"
			:empty-text="$t('task.description.empty')"
			:show-save="true"
			v-model="task.description"
		/>
	</div>
</template>

<script lang="ts">
import AsyncEditor from '@/components/input/AsyncEditor'

import {LOADING} from '@/store/mutation-types'
import {mapState} from 'vuex'

export default {
	name: 'description',
	components: {
		Editor: AsyncEditor,
	},
	data() {
		return {
			task: {description: ''},
			saved: false,
			saving: false, // Since loading is global state, this variable ensures we're only showing the saving icon when saving the description.
		}
	},
	computed: mapState({
		loading: LOADING,
	}),
	props: {
		modelValue: {
			required: true,
		},
		attachmentUpload: {
			required: true,
		},
		canWrite: {
			required: true,
		},
	},
	emits: ['update:modelValue'],
	watch: {
		modelValue: {
			handler(value) {
				this.task = value
			},
			immediate: true,
		},
	},
	methods: {
		async save() {
			this.saving = true

			try {
				this.task = await this.$store.dispatch('tasks/update', this.task)
				this.$emit('update:modelValue', this.task)
				this.saved = true
				setTimeout(() => {
					this.saved = false
				}, 2000)
			} finally {
				this.saving = false
			}
		},
	},
}
</script>

