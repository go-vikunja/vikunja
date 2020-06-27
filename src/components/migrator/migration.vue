<template>
	<div class="content">
		<h1>Import your data from {{ name }} to Vikunja</h1>
		<p>Vikunja will import all lists, tasks, notes, reminders and files you have access to.</p>
		<template v-if="isMigrating === false && message === '' && lastMigrationDate === 0">
			<p>To authorize Vikunja to access your {{ name }} Account, click the button below.</p>
			<a :href="authUrl" class="button is-primary" :class="{'is-loading': migrationService.loading}" :disabled="migrationService.loading">Get Started</a>
		</template>
		<div class="migration-in-progress-container" v-else-if="isMigrating === true && message === '' && lastMigrationDate === 0">
			<div class="migration-in-progress">
				<img :src="`/images/migration/${identifier}.png`" :alt="name"/>
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
				<img src="/images/logo.svg" alt="Vikunja">
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
				<button class="button is-primary" @click="migrate">I am sure, please start migrating now!</button>
				<router-link :to="{name: 'home'}" class="button is-danger is-outlined">Cancel</router-link>
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
	import AbstractMigrationService from "../../services/migrator/abstractMigrationService";

	export default {
		name: 'migration',
		data() {
			return {
				authUrl: '',
				isMigrating: false,
				lastMigrationDate: null,
				message: '',
				wunderlistCode: '',
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

			if(typeof this.$route.query.code !== 'undefined') {
				this.wunderlistCode = this.$route.query.code
				this.migrationService.getStatus()
					.then(r => {
						if(r.time) {
							this.lastMigrationDate = new Date(r.time)
							return
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
				this.lastMigrationDate = 0
				this.migrationService.migrate({code: this.wunderlistCode})
					.then(r => {
						this.message = r.message
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
