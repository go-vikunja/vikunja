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

<script>
import LoadingComponent from '@/components/misc/loading.vue'
import ErrorComponent from '@/components/misc/error.vue'

import {LOADING} from '@/store/mutation-types'
import {mapState} from 'vuex'

export default {
	name: 'description',
	components: {
		editor: () => ({
			component: import('@/components/input/editor.vue'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
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
		value: {
			required: true,
		},
		attachmentUpload: {
			required: true,
		},
		canWrite: {
			required: true,
		},
	},
	watch: {
		value: {
			handler(value) {
				this.task = value
			},
			immediate: true,
		},
	},
	methods: {
		save() {
			this.saving = true

			this.$store.dispatch('tasks/update', this.task)
				.then(t => {
					this.task = t
					this.$emit('input', t)
					this.saved = true
					setTimeout(() => {
						this.saved = false
					}, 2000)
				})
				.catch(e => {
					this.$message.error(e)
				})
				.finally(() => {
					this.saving = false
				})
		},
	},
}
</script>

