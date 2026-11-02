<template>
	<div
		v-if="shares.length > 0"
		class="revoke-link-shares"
	>
		<p>{{ $t('project.share.accessCleanup.linkShares.intro') }}</p>
		<ul class="link-share-list">
			<li
				v-for="{share} in shares"
				:key="share.id"
				data-cy="revocable-link-share"
			>
				<span :class="{'is-italic': !share.name}">{{ share.name || $t('project.share.links.noName') }}</span>
				· {{ permissionLabel(share.permission) }}
				· {{ $t('project.share.accessCleanup.linkShares.created', {date: formatDateShort(share.created)}) }}
			</li>
		</ul>
		<FancyCheckbox v-model="revoke">
			{{ $t('project.share.accessCleanup.linkShares.revoke', shares.length) }}
		</FancyCheckbox>
	</div>
</template>

<script setup lang="ts">
import {useI18n} from 'vue-i18n'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import type {RevocableLinkShare} from '@/client/queries/linkShares'
import {normalizeSharePermission} from '@/client/queries/projectShares'
import {PERMISSIONS} from '@/constants/permissions'
import {formatDateShort} from '@/helpers/time/formatDate'

defineProps<{
	shares: RevocableLinkShare[]
}>()

const revoke = defineModel<boolean>('revoke', {required: true})

const {t} = useI18n({useScope: 'global'})

function permissionLabel(permission: number | undefined) {
	switch (normalizeSharePermission(permission)) {
		case PERMISSIONS.ADMIN: return t('project.share.permission.admin')
		case PERMISSIONS.READ_WRITE: return t('project.share.permission.readWrite')
		default: return t('project.share.permission.read')
	}
}
</script>

<style lang="scss" scoped>
.revoke-link-shares {
	margin-block-start: 1rem;
}

.link-share-list {
	display: inline-block;
	text-align: start;
	margin-inline-start: 0;
	margin-block: 0 .5rem;
	padding-inline-start: 1.25rem;
}
</style>
