<template>
	<div class="content">
		<h1>Import your data from {{ name }} to Vikunja</h1>
		<p>Vikunja will import all lists, tasks, notes, reminders and files you have access to.</p>
		<template v-if="isMigrating === false && message === '' && lastMigrationDate === null">
			<p>To authorize Vikunja to access your {{ name }} Account, click the button below.</p>
			<a
				:class="{'is-loading': migrationService.loading}"
				:disabled="migrationService.loading"
				:href="authUrl"
				class="button is-primary">
				Get Started
			</a>
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
			<p>Importing in progress, hang tight...</p>
		</div>
		<div v-else-if="lastMigrationDate">
			<p>
				It looks like you've already imported your stuff from {{ name }} at {{ formatDate(lastMigrationDate) }}.<br/>
				Importing again is possible, but might create duplicates.
				Are you sure?
			</p>
			<div class="buttons">
				<button @click="migrate" class="button is-primary">I am sure, please start migrating now!</button>
				<router-link :to="{name: 'home'}" class="button is-text has-text-danger is-inverted noshadow underline-none">Cancel</router-link>
			</div>
		</div>
		<div v-else>
			<div class="message is-primary">
				<div class="message-body">
					{{ message }}
				</div>
			</div>
			<router-link :to="{name: 'home'}" class="button is-primary">Refresh</router-link>
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
				console.log(location.hash.substring(7))
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
					this.error(e, this)
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
					this.error(e, this)
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
					this.error(e, this)
				})
				.finally(() => {
					this.isMigrating = false
				})
		},
	},
}
</script>
