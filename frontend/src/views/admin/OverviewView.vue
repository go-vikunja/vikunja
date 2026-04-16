<template>
	<Card :title="$t('admin.overview.title')">
		<div class="admin-overview">
			<p v-if="loading">
				{{ $t('misc.loading') }}
			</p>
			<div
				v-else-if="data"
				class="admin-overview__grid"
			>
				<div class="admin-overview__card">
					<h2 class="admin-overview__card-title">
						{{ $t('admin.overview.users') }}
					</h2>
					<p class="admin-overview__card-value">
						{{ data.users }}
					</p>
				</div>
				<div class="admin-overview__card">
					<h2 class="admin-overview__card-title">
						{{ $t('admin.overview.projects') }}
					</h2>
					<p class="admin-overview__card-value">
						{{ data.projects }}
					</p>
				</div>
				<div class="admin-overview__card">
					<h2 class="admin-overview__card-title">
						{{ $t('admin.overview.tasks') }}
					</h2>
					<p class="admin-overview__card-value">
						{{ data.tasks }}
					</p>
				</div>
				<div class="admin-overview__card">
					<h2 class="admin-overview__card-title">
						{{ $t('admin.overview.shares') }}
					</h2>
					<ul class="admin-overview__card-list">
						<li>{{ $t('admin.overview.linkShares') }}: {{ data.shares.linkShares }}</li>
						<li>{{ $t('admin.overview.teamShares') }}: {{ data.shares.teamShares }}</li>
						<li>{{ $t('admin.overview.userShares') }}: {{ data.shares.userShares }}</li>
					</ul>
				</div>
				<div class="admin-overview__card">
					<h2 class="admin-overview__card-title">
						{{ $t('admin.overview.version') }}
					</h2>
					<p class="admin-overview__card-value">
						{{ data.version }}
					</p>
				</div>
			</div>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import Card from '@/components/misc/Card.vue'
import {getAdminOverview, type AdminOverview} from '@/services/admin/overviewService'
import {error} from '@/message'

const data = ref<AdminOverview | null>(null)
const loading = ref(false)

onMounted(async () => {
	loading.value = true
	try {
		data.value = await getAdminOverview()
	} catch (e) {
		error(e)
	} finally {
		loading.value = false
	}
})
</script>

<style lang="scss" scoped>
.admin-overview__grid {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
	gap: 1rem;
}

.admin-overview__card {
	background: var(--white);
	border: 1px solid var(--grey-200);
	border-radius: 6px;
	padding: 1.25rem;
}

.admin-overview__card-title {
	font-size: 0.85rem;
	color: var(--grey-600);
	text-transform: uppercase;
	letter-spacing: 0.03em;
	margin-block-end: 0.5rem;
}

.admin-overview__card-value {
	font-size: 1.75rem;
	font-weight: 600;
}

.admin-overview__card-list {
	list-style: none;
	padding: 0;
	margin: 0;
}
</style>
