<template>
	<div class="content">
		<h1>{{ $t('migrate.titleService', { name: name }) }}</h1>
		<p>{{ $t('migrate.descriptionDo') }}</p>
		<template v-if="isMigrating === false && message === '' && lastMigrationDate === null">
			<p>{{ $t('migrate.authorize', {name: name}) }}</p>
			<x-button
				:loading="migrationService.loading"
				:disabled="migrationService.loading"
				:href="authUrl"
			>
				{{ $t('migrate.getStarted') }}
			</x-button>
		</template>
		<div
			class="migration-in-progress-container"
			v-else-if="isMigrating === true && message === '' && lastMigrationDate === null">
			<div class="migration-in-progress">
				<img :alt="name" :src="`/images/migration/${identifier}.png`"/>
				<div class="progress-dots">
					<span></span>
					<span></span>
					<span></span>
					<span></span>
					<span></span>
					<span></span>
					<span></span>
					<span></span>
				</div>
				<img alt="Vikunja" src="/images/logo.svg">
			</div>
			<p>{{ $t('migrate.inProgress') }}</p>
		</div>
		<div v-else-if="lastMigrationDate">
			<p>
				{{ $t('migrate.alreadyMigrated1', { name: name, date: formatDate(lastMigrationDate) }) }}<br/>
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
import AbstractMigrationService from '../../services/migrator/abstractMigrationService'

export default {
	name: 'migration',
	data() {
		return {
			authUrl: '',
			isMigrating: false,
			lastMigrationDate: null,
			message: '',
			migratorAuthCode: '',
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
	},
	created() {
		this.migrationService = new AbstractMigrationService(this.identifier)
		this.getAuthUrl()
		this.message = ''

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
				.catch(e => {
					this.error(e)
				})
		}
	},
	methods: {
		getAuthUrl() {
			this.migrationService.getAuthUrl()
				.then(r => {
					this.authUrl = r.url
				})
				.catch(e => {
					this.error(e)
				})
		},
		migrate() {
			this.isMigrating = true
			this.lastMigrationDate = null
			this.message = ''
			this.migrationService.migrate({code: this.migratorAuthCode})
				.then(r => {
					this.message = r.message
					this.$store.dispatch('namespaces/loadNamespaces')
				})
				.catch(e => {
					this.error(e)
				})
				.finally(() => {
					this.isMigrating = false
				})
		},
	},
}
</script>
