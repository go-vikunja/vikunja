<template>
	<div class="content">
		<h1>Import your data from Wunderlist to Vikunja</h1>
		<p>Vikunja will import all folders, lists, tasks, notes, reminders and files you have access to.</p>
		<template v-if="isMigrating === false && message === ''">
			<p>To authorize Vikunja to access your Wunderlist Account, click the button below.</p>
			<a :href="authUrl" class="button is-primary" :class="{'is-loading': migrationService.loading}" :disabled="migrationService.loading">Get Started</a>
		</template>
		<div class="migration-in-progress-container" v-else-if="isMigrating === true && message === ''">
			<div class="migration-in-progress">
				<img src="/images/migration/wunderlist.png" alt="Wunderlist Logo"/>
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
				<img src="/images/logo.svg" alt="Vikunja Logo">
			</div>
			<p>Migration in progress, hang tight...</p>
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
	import WunderlistMigrationService from '../../services/migrator/wunderlist'
	import message from '../../message'

	export default {
		name: 'wunderlist',
		data() {
			return {
				migrationService: WunderlistMigrationService,
				authUrl: '',
				isMigrating: false,
				message: '',
				wunderlistCode: '',
			}
		},
		created() {
			this.migrationService = new WunderlistMigrationService()
			this.getAuthUrl()
			this.message = ''

			if(typeof this.$route.query.code !== 'undefined') {
				this.isMigrating = true
				this.wunderlistCode = this.$route.query.code
				this.migrate()
			}
		},
		methods: {
			getAuthUrl() {
				this.migrationService.getAuthUrl()
					.then(r => {
						this.authUrl = r.url
					})
					.catch(e => {
						message.error(e, this)
					})
			},
			migrate() {
				this.migrationService.migrate({code: this.wunderlistCode})
					.then(r => {
						this.message = r.message
					})
					.catch(e => {
						message.error(e, this)
					})
					.finally(() => {
						this.isMigrating = false
					})
			},
		},
	}
</script>
