<template>
	<card v-if="caldavEnabled" :title="$t('user.settings.caldav.title')">
		<p>
			{{ $t('user.settings.caldav.howTo') }}
		</p>
		<div class="field has-addons no-input-mobile">
			<div class="control is-expanded">
				<input type="text" v-model="caldavUrl" class="input" readonly/>
			</div>
			<div class="control">
				<x-button
					@click="copy(caldavUrl)"
					:shadow="false"
					v-tooltip="$t('misc.copy')"
					icon="paste"
				/>
			</div>
		</div>
		<p>
			<a href="https://vikunja.io/docs/caldav/" rel="noreferrer noopener nofollow" target="_blank">
				{{ $t('user.settings.caldav.more') }}
			</a>
		</p>
	</card>
</template>

<script lang="ts">
import {defineComponent} from 'vue'
import copy from 'copy-to-clipboard'
import {mapState} from 'vuex'
import {CALDAV_DOCS} from '@/urls'

export default defineComponent({
	name: 'user-settings-caldav',
	data() {
		return {
			caldavDocsUrl: CALDAV_DOCS,
		}
	},
	mounted() {
		this.setTitle(`${this.$t('user.settings.caldav.title')} - ${this.$t('user.settings.title')}`)
	},
	computed: {
		caldavUrl() {
			return `${this.$store.getters['config/apiBase']}/dav/principals/${this.userInfo.username}/`
		},
		...mapState('config', ['caldavEnabled']),
		...mapState({
			userInfo: state => state.auth.info,
		}),
	},
	methods: {
		copy,
	},
})
</script>
