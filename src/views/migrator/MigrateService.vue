<template>
	<div class="content">
		<h1>{{ $t('migrate.titleService', {name: migrator.name}) }}</h1>
		<p>{{ $t('migrate.descriptionDo') }}</p>

		<template v-if="message === '' && lastMigrationDate === null">
			<template v-if="isMigrating === false">
				<template v-if="migrator.isFileMigrator">
					<p>{{ $t('migrate.importUpload', {name: migrator.name}) }}</p>
					<input
						@change="migrate"
						class="is-hidden"
						ref="uploadInput"
						type="file"
					/>
					<x-button
						:loading="migrationService.loading"
						:disabled="migrationService.loading || null"
						@click="$refs.uploadInput.click()"
					>
						{{ $t('migrate.upload') }}
					</x-button>
				</template>
				<template v-else>
					<p>{{ $t('migrate.authorize', {name: migrator.name}) }}</p>
					<x-button
						:loading="migrationService.loading"
						:disabled="migrationService.loading || null"
						:href="authUrl"
					>
						{{ $t('migrate.getStarted') }}
					</x-button>
				</template>
			</template>
			<div
				v-else
				class="migration-in-progress-container"
			>
				<div class="migration-in-progress">
					<img :alt="migrator.name" :src="migrator.icon"/>
					<div class="progress-dots">
						<span v-for="i in progressDotsCount" :key="i" />
					</div>
					<Logo alt="Vikunja" />
				</div>
				<p>{{ $t('migrate.inProgress') }}</p>
			</div>
		</template>
		<div v-else-if="lastMigrationDate">
			<p>
				{{ $t('migrate.alreadyMigrated1', {name: migrator.name, date: formatDate(lastMigrationDate)}) }}<br/>
				{{ $t('migrate.alreadyMigrated2') }}
			</p>
			<div class="buttons">
				<x-button @click="migrate">{{ $t('migrate.confirm') }}</x-button>
				<x-button :to="{name: 'home'}" type="tertary" class="has-text-danger">{{ $t('misc.cancel') }}</x-button>
			</div>
		</div>
		<div v-else>
			<div class="message is-primary">
				<div class="message-body">
					{{ message }}
				</div>
			</div>
			<x-button :to="{name: 'home'}">{{ $t('misc.refresh') }}</x-button>
		</div>
	</div>
</template>

<script>
import AbstractMigrationService from '@/services/migrator/abstractMigration'
import AbstractMigrationFileService from '@/services/migrator/abstractMigrationFile'
import Logo from '@/assets/logo.svg?component'

import {MIGRATORS} from './migrators'

const PROGRESS_DOTS_COUNT = 8

export default {
	name: 'MigrateService',

	components: { Logo },

	data() {
		return {
			progressDotsCount: PROGRESS_DOTS_COUNT,
			authUrl: '',
			isMigrating: false,
			lastMigrationDate: null,
			message: '',
			migratorAuthCode: '',
			migrationService: null,
		}
	},

	computed: {
		migrator() {
			return MIGRATORS[this.$route.params.service]
		},
	},

	beforeRouteEnter(to) {
		if (MIGRATORS[to.params.service] === undefined) {
			return { name: 'not-found' }
		}
	},

	created() {
		this.initMigration()
	},

	mounted() {
		this.setTitle(this.$t('migrate.titleService', {name: this.migrator.name}))
	},

	methods: {
		async initMigration() {
			this.migrationService = this.migrator.isFileMigrator
				? new AbstractMigrationFileService(this.migrator.id)
				: new AbstractMigrationService(this.migrator.id)

			if (this.migrator.isFileMigrator) {
				return
			}
			
			this.authUrl = await this.migrationService.getAuthUrl().then(({url}) => url)

			this.migratorAuthCode = location.hash.startsWith('#token=')
				? location.hash.substring(7)
				: this.$route.query.code

			if (!this.migratorAuthCode) {
				return
			}
			const {time} = await this.migrationService.getStatus()
			if (time) {
				this.lastMigrationDate = typeof time === 'string' && time?.startsWith('0001-')
					? null
					: new Date(time)

				if (this.lastMigrationDate) {
					return
				}
			}
			await this.migrate()
		},

		async migrate() {
			this.isMigrating = true
			this.lastMigrationDate = null
			this.message = ''

			let migrationConfig = { code: this.migratorAuthCode }

			if (this.migrator.isFileMigrator) {
				if (this.$refs.uploadInput.files.length === 0) {
					return
				}
				migrationConfig = this.$refs.uploadInput.files[0]
			}

			try {
				const { message } = await this.migrationService.migrate(migrationConfig)
				this.message = message
				return this.$store.dispatch('namespaces/loadNamespaces')
			} finally {
				this.isMigrating = false
			}
		},
	},
}
</script>

<style lang="scss" scoped>
.migration-in-progress-container {
  max-width: 400px;
  margin: 4rem auto 0;
  text-align: center;
}

.migration-in-progress {
  text-align: center;
  display: flex;
  max-width: 400px;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;

  img {
    display: block;
    max-height: 100px;
  }
}

.progress-dots {
	height: 40px;
	width: 140px;
	overflow: visible;

	span {
		transition: all 500ms ease;
		background: $grey-500;
		height: 10px;
		width: 10px;
		display: inline-block;
		border-radius: 10px;
		animation: wave 2s ease infinite;
		margin-right: 5px;

		&:nth-child(1) {
			animation-delay: 0;
		}

		&:nth-child(2) {
			animation-delay: 100ms;
		}

		&:nth-child(3) {
			animation-delay: 200ms;
		}

		&:nth-child(4) {
			animation-delay: 300ms;
		}

		&:nth-child(5) {
			animation-delay: 400ms;
		}

		&:nth-child(6) {
			animation-delay: 500ms;
		}

		&:nth-child(7) {
			animation-delay: 600ms;
		}

		&:nth-child(8) {
			animation-delay: 700ms;
		}
	}
}

@keyframes wave {
	0%, 40%, 100% {
		transform: translate(0, 0);
		background-color: $primary;
	}
	10% {
		transform: translate(0, -15px);
		background-color: $primary-dark;
	}
}
</style>