<template>
	<card :title="$t('user.deletion.title')" v-if="userDeletionEnabled">
		<template v-if="deletionScheduledAt !== null">
			<form @submit.prevent="cancelDeletion()">
				<p>
					{{
						$t('user.deletion.scheduled', {
							date: formatDateShort(deletionScheduledAt),
							dateSince: formatDateSince(deletionScheduledAt),
						})
					}}
				</p>
				<p>
					{{ $t('user.deletion.scheduledCancelText') }}
				</p>
				<div class="field">
					<label class="label" for="currentPasswordAccountDelete">
						{{ $t('user.settings.currentPassword') }}
					</label>
					<div class="control">
						<input
							class="input"
							:class="{'is-danger': errPasswordRequired}"
							id="currentPasswordAccountDelete"
							:placeholder="$t('user.settings.currentPasswordPlaceholder')"
							type="password"
							v-model="password"
							@keyup="() => errPasswordRequired = password === ''"
							ref="passwordInput"
						/>
					</div>
					<p class="help is-danger" v-if="errPasswordRequired">
						{{ $t('user.deletion.passwordRequired') }}
					</p>
				</div>
			</form>

			<x-button
				:loading="accountDeleteService.loading"
				@click="cancelDeletion()"
				class="is-fullwidth mt-4">
				{{ $t('user.deletion.scheduledCancelConfirm') }}
			</x-button>
		</template>
		<template v-else>
			<form @submit.prevent="deleteAccount()">
				<p>
					{{ $t('user.deletion.text1') }}
				</p>
				<p>
					{{ $t('user.deletion.text2') }}
				</p>
				<div class="field">
					<label class="label" for="currentPasswordAccountDelete">
						{{ $t('user.settings.currentPassword') }}
					</label>
					<div class="control">
						<input
							class="input"
							:class="{'is-danger': errPasswordRequired}"
							id="currentPasswordAccountDelete"
							:placeholder="$t('user.settings.currentPasswordPlaceholder')"
							type="password"
							v-model="password"
							@keyup="() => errPasswordRequired = password === ''"
							ref="passwordInput"
						/>
					</div>
					<p class="help is-danger" v-if="errPasswordRequired">
						{{ $t('user.deletion.passwordRequired') }}
					</p>
				</div>
			</form>

			<x-button
				:loading="accountDeleteService.loading"
				@click="deleteAccount()"
				class="is-fullwidth mt-4 is-danger">
				{{ $t('user.deletion.confirm') }}
			</x-button>
		</template>
	</card>
</template>

<script>
import AccountDeleteService from '@/services/accountDelete'
import {mapState} from 'vuex'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

export default {
	name: 'user-settings-deletion',
	data() {
		return {
			accountDeleteService: new AccountDeleteService(),
			password: '',
			errPasswordRequired: false,
		}
	},
	computed: mapState({
		userDeletionEnabled: state => state.config.userDeletionEnabled,
		deletionScheduledAt: state => parseDateOrNull(state.auth.info?.deletionScheduledAt),
	}),
	mounted() {
		this.setTitle(`${this.$t('user.deletion.title')} - ${this.$t('user.settings.title')}`)
	},
	methods: {
		async deleteAccount() {
			if (this.password === '') {
				this.errPasswordRequired = true
				this.$refs.passwordInput.focus()
				return
			}

			await this.accountDeleteService.request(this.password)
			this.$message.success({message: this.$t('user.deletion.requestSuccess')})
			this.password = ''
		},

		async cancelDeletion() {
			if (this.password === '') {
				this.errPasswordRequired = true
				this.$refs.passwordInput.focus()
				return
			}

			await this.accountDeleteService.cancel(this.password)
			this.$message.success({message: this.$t('user.deletion.scheduledCancelSuccess')})
			this.$store.dispatch('auth/refreshUserInfo')
			this.password = ''
		},
	},
}
</script>
