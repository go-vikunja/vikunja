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
				<div class="admin-overview__card admin-overview__card--wide">
					<h2 class="admin-overview__card-title">
						{{ $t('admin.overview.license') }}
					</h2>
					<dl class="admin-overview__kv">
						<dt>{{ $t('admin.overview.licenseValidUntil') }}</dt>
						<dd>
							<time :datetime="formatISO(data.license.expiresAt)">{{ formatDisplayDate(data.license.expiresAt) }}</time>
							<span
								v-if="expiresInDays !== null"
								class="admin-overview__hint"
							>
								({{ $t('admin.overview.licenseExpiresIn', {days: expiresInDays}) }})
							</span>
						</dd>
						<dt>{{ $t('admin.overview.licenseLastVerified') }}</dt>
						<dd>
							<time
								v-if="data.license.validatedAt"
								:datetime="formatISO(data.license.validatedAt)"
							>
								{{ formatDateSince(data.license.validatedAt) }}
							</time>
							<span v-else>{{ $t('admin.overview.licenseNever') }}</span>
							<span
								v-if="data.license.lastCheckFailed"
								class="has-text-danger admin-overview__hint"
							>
								({{ $t('admin.overview.licenseLastCheckFailed') }})
							</span>
						</dd>
						<template v-if="data.license.features.length">
							<dt>{{ $t('admin.overview.licenseFeatures') }}</dt>
							<dd>{{ data.license.features.join(', ') }}</dd>
						</template>
						<template v-if="data.license.instanceId">
							<dt>{{ $t('admin.overview.licenseInstance') }}</dt>
							<dd><code>{{ data.license.instanceId }}</code></dd>
						</template>
					</dl>
					<p class="admin-overview__card-action">
						<a
							href="https://console.vikunja.io"
							target="_blank"
							rel="noopener noreferrer"
						>
							{{ $t('admin.overview.licenseManage') }}
						</a>
					</p>
				</div>
			</div>
		</div>
	</Card>
</template>

<script setup lang="ts">
import {ref, computed, onMounted} from 'vue'
import dayjs from 'dayjs'
import Card from '@/components/misc/Card.vue'
import AdminOverviewService from '@/services/admin/overviewService'
import type {IAdminOverview} from '@/modelTypes/IAdminOverview'
import {error} from '@/message'
import {formatDisplayDate, formatDateSince, formatISO} from '@/helpers/time/formatDate'

const adminOverviewService = new AdminOverviewService()

const data = ref<IAdminOverview | null>(null)
const loading = ref(false)

const expiresInDays = computed<number | null>(() => {
	const expiresAt = data.value?.license?.expiresAt
	if (!expiresAt) return null
	return Math.max(0, dayjs(expiresAt).diff(dayjs(), 'day'))
})

onMounted(async () => {
	loading.value = true
	try {
		data.value = await adminOverviewService.getOverview()
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
	grid-template-columns: 1fr;
	gap: 1rem;
}

@media (width >= 600px) {
	.admin-overview__grid {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}
}

@media (width >= 1000px) {
	.admin-overview__grid {
		grid-template-columns: repeat(5, minmax(0, 1fr));
	}
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

.admin-overview__card--wide {
	grid-column: 1 / -1;
}

.admin-overview__kv {
	display: grid;
	grid-template-columns: max-content 1fr;
	column-gap: 1rem;
	row-gap: 0.25rem;
	margin-block-start: 1rem;
	font-size: 0.9rem;

	dt {
		font-weight: 600;
		color: var(--grey-700);
	}

	dd {
		margin: 0;
	}
}

.admin-overview__hint {
	color: var(--grey-600);
	margin-inline-start: 0.25rem;
}

.admin-overview__card-action {
	margin-block-start: 1rem;
}

.admin-overview__card-list {
	list-style: none;
	padding: 0;
	margin: 0;
}
</style>
