<template>
	<div class="admin-shell">
		<div class="admin-shell__header">
			<h1 class="admin-shell__title">
				{{ $t('admin.title') }}
			</h1>
			<nav class="admin-shell__nav">
				<router-link :to="{name: 'admin.overview'}" class="admin-shell__nav-item">
					{{ $t('admin.nav.overview') }}
				</router-link>
				<router-link :to="{name: 'admin.users'}" class="admin-shell__nav-item">
					{{ $t('admin.nav.users') }}
				</router-link>
				<router-link :to="{name: 'admin.projects'}" class="admin-shell__nav-item">
					{{ $t('admin.nav.projects') }}
				</router-link>
			</nav>
		</div>
		<div class="admin-shell__body">
			<router-view />
		</div>
	</div>
</template>

<script setup lang="ts">
</script>

<style lang="scss" scoped>
.admin-shell {
	max-width: 1200px;
	margin: 0 auto;
	padding: 2rem 1rem;
}

.admin-shell__header {
	margin-bottom: 2rem;
}

.admin-shell__title {
	font-size: 1.75rem;
	font-weight: 600;
	margin-bottom: 0.75rem;
}

.admin-shell__nav {
	display: flex;
	gap: 1rem;
	border-bottom: 1px solid var(--grey-200);
}

.admin-shell__nav-item {
	padding: 0.5rem 0.75rem;
	color: var(--text);
	text-decoration: none;
	border-bottom: 2px solid transparent;

	&.router-link-active {
		border-bottom-color: var(--primary);
		color: var(--primary);
	}
}

.admin-shell__body {
	padding-top: 1.5rem;
}
</style>
