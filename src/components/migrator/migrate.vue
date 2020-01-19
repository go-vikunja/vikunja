<template>
	<div class="content">
		<h1>Migrate your data from other services to Vikunja</h1>
		<p>Click on the logo of one of the third-party services below to get started.</p>
		<div class="migration-services-overview">
			<router-link :to="{name: 'migrateWunderlist'}">
				<img src="/images/migration/wunderlist.png" alt="Wunderlist"/>
				Wunderlist
			</router-link>
		</div>
	</div>
</template>

<script>
	export default {
		name: 'migrate'
	}
</script>
