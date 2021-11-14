<template>
	<div class="content-widescreen">
		<div class="user-settings">
			<nav class="navigation">
				<ul>
					<li>
						<router-link :to="{name: 'user.settings.general'}">
							{{ $t('user.settings.general.title') }}
						</router-link>
					</li>
					<li v-if="isLocalUser">
						<router-link :to="{name: 'user.settings.password-update'}">
							{{ $t('user.settings.newPasswordTitle') }}
						</router-link>
					</li>
					<li v-if="isLocalUser">
						<router-link :to="{name: 'user.settings.email-update'}">
							{{ $t('user.settings.updateEmailTitle') }}
						</router-link>
					</li>
					<li>
						<router-link :to="{name: 'user.settings.avatar'}">
							{{ $t('user.settings.avatar.title') }}
						</router-link>
					</li>
					<li v-if="totpEnabled">
						<router-link :to="{name: 'user.settings.totp'}">
							{{ $t('user.settings.totp.title') }}
						</router-link>
					</li>
					<li>
						<router-link :to="{name: 'user.settings.data-export'}">
							{{ $t('user.export.title') }}
						</router-link>
					</li>
					<li v-if="migratorsEnabled">
						<router-link :to="{name: 'migrate.start'}">
							{{ $t('migrate.title') }}
						</router-link>
					</li>
					<li v-if="caldavEnabled">
						<router-link :to="{name: 'user.settings.caldav'}">
							{{ $t('user.settings.caldav.title') }}
						</router-link>
					</li>
					<li>
						<router-link :to="{name: 'user.settings.deletion'}">
							{{ $t('user.deletion.title') }}
						</router-link>
					</li>
				</ul>
			</nav>
			<section class="view">
				<router-view/>
			</section>
		</div>
	</div>
</template>

<script setup>
import {computed} from 'vue'
import { store } from '@/store'
import { useI18n } from 'vue-i18n'
import { useTitle } from '@/composables/useTitle'

const { t } = useI18n()
useTitle(() => t('user.settings.title'))

const totpEnabled = computed(() => store.state.config.totpEnabled)
const caldavEnabled = computed(() => store.state.config.caldavEnabled)
const migratorsEnabled = computed(() => store.getters['config/migratorsEnabled'])
const isLocalUser = computed(() => store.state.auth.info?.isLocalUser)
</script>

<style lang="scss" scoped>
.user-settings {
	display: flex;

	.navigation {
		width: 25%;
		padding-right: 1rem;

		a {
			display: block;
			padding: .5rem;
			color: $dark;
			width: 100%;
			border-left: 3px solid transparent;

			&:hover, &.router-link-active {
				background: $white;
				border-color: $primary;
			}
		}
	}

	.view {
		width: 75%;
	}

	@media screen and (max-width: $tablet) {
		flex-direction: column;

		.navigation, .view {
			width: 100%;
			padding-left: 0;
		}

		.view {
			padding-top: 1rem;
		}
	}
}
</style>
