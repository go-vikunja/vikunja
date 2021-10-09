<template>
	<div class="content">
		<h1>{{ $t('migrate.titleService', {name: name}) }}</h1>
		<p>{{ $t('migrate.descriptionDo') }}</p>
		<template v-if="isMigrating === false && message === '' && lastMigrationDate === null">
			<template v-if="isFileMigrator">
				<p>{{ $t('migrate.importUpload', {name: name}) }}</p>
				<input
					@change="migrate"
					class="is-hidden"
					ref="uploadInput"
					type="file"
				/>
				<x-button
					:loading="migrationService.loading"
					:disabled="migrationService.loading"
					@click="$refs.uploadInput.click()"
				>
					{{ $t('migrate.upload') }}
				</x-button>
			</template>
			<template v-else>
				<p>{{ $t('migrate.authorize', {name: name}) }}</p>
				<x-button
					:loading="migrationService.loading"
					:disabled="migrationService.loading"
					:href="authUrl"
				>
					{{ $t('migrate.getStarted') }}
				</x-button>
			</template>
		</template>
		<div
			class="migration-in-progress-container"
			v-else-if="isMigrating === true && message === '' && lastMigrationDate === null">
			<div class="migration-in-progress">
				<img :alt="name" :src="serviceIconSource"/>
				<div class="progress-dots">
					<span v-for="i in progressDotsCount" :key="i" />
				</div>
				<img alt="Vikunja" :src="logoUrl">
			</div>
			<p>{{ $t('migrate.inProgress') }}</p>
		</div>
		<div v-else-if="lastMigrationDate">
			<p>
				{{ $t('migrate.alreadyMigrated1', {name: name, date: formatDate(lastMigrationDate)}) }}<br/>
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
import AbstractMigrationService from '../../services/migrator/abstractMigration'
import AbstractMigrationFileService from '../../services/migrator/abstractMigrationFile'
import {SERVICE_ICONS} from '../../helpers/migrator'

import logoUrl from '@/assets/logo.svg'

const PROGRESS_DOTS_COUNT = 8

export default {
	name: 'migration',
	data() {
		return {
			progressDotsCount: PROGRESS_DOTS_COUNT,
			authUrl: '',
			isMigrating: false,
			lastMigrationDate: null,
			message: '',
			migratorAuthCode: '',
			migrationService: null,
			logoUrl,
		}
	},
	props: {
		name: {
			type: String,
			required: true,
		},
		identifier: {
			type: String,
			required: true,
		},
		isFileMigrator: {
			type: Boolean,
			default: false,
		},
	},
	computed: {
		serviceIconSource() {
			return SERVICE_ICONS[this.identifier]()
		},
	},
	created() {
		this.message = ''

		if (this.isFileMigrator) {
			this.migrationService = new AbstractMigrationFileService(this.identifier)
			return
		}
		
		this.migrationService = new AbstractMigrationService(this.identifier)
		this.getAuthUrl()

		if (typeof this.$route.query.code !== 'undefined' || location.hash.startsWith('#token=')) {
			if (location.hash.startsWith('#token=')) {
				this.migratorAuthCode = location.hash.substring(7)
				console.debug(location.hash.substring(7))
			} else {
				this.migratorAuthCode = this.$route.query.code
			}
			this.migrationService.getStatus()
				.then(r => {
					if (r.time) {
						if (typeof r.time === 'string' && r.time.startsWith('0001-')) {
							this.lastMigrationDate = null
						} else {
							this.lastMigrationDate = new Date(r.time)
						}

						if (this.lastMigrationDate) {
							return
						}
					}
					this.migrate()
				})
		}
	},
	methods: {
		getAuthUrl() {
			this.migrationService.getAuthUrl()
				.then(r => {
					this.authUrl = r.url
				})
		},
		migrate() {
			this.isMigrating = true
			this.lastMigrationDate = null
			this.message = ''

			if (this.isFileMigrator) {
				return this.migrateFile()
			}

			this.migrationService.migrate({code: this.migratorAuthCode})
				.then(r => {
					this.message = r.message
					this.$store.dispatch('namespaces/loadNamespaces')
				})
				.finally(() => {
					this.isMigrating = false
				})
		},
		migrateFile() {
			if (this.$refs.uploadInput.files.length === 0) {
				return
			}

			this.migrationService.migrate(this.$refs.uploadInput.files[0])
				.then(r => {
					this.message = r.message
					this.$store.dispatch('namespaces/loadNamespaces')
				})
				.finally(() => {
					this.isMigrating = false
				})
		},
	},
}
</script>
