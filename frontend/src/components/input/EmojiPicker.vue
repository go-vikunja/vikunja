<script setup lang="ts">
import {computed, nextTick, onMounted, ref, shallowRef, useId, useTemplateRef, watch} from 'vue'
import {useStorage} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import {
	filterEmojis,
	loadEmojiCatalog,
	loadEmojis,
	type EmojiCatalogEntry,
	type EmojiEntry,
} from '@/components/input/editor/emoji/emojiData'

const emit = defineEmits<{
	select: [emoji: string]
}>()

const {t} = useI18n({useScope: 'global'})

const categories = [
	{
		group: -1,
		icon: '🕒',
		key: 'recent',
	},
	{
		group: 0,
		icon: '😀',
		key: 'smileys',
	},
	{
		group: 1,
		icon: '👋',
		key: 'people',
	},
	{
		group: 3,
		icon: '🌿',
		key: 'nature',
	},
	{
		group: 4,
		icon: '🍋',
		key: 'food',
	},
	{
		group: 5,
		icon: '🚀',
		key: 'travel',
	},
	{
		group: 6,
		icon: '⚽',
		key: 'activities',
	},
	{
		group: 7,
		icon: '💡',
		key: 'objects',
	},
	{
		group: 8,
		icon: '❤️',
		key: 'symbols',
	},
	{
		group: 9,
		icon: '🏁',
		key: 'flags',
	},
]
const COLUMNS = 8
const tones = [
	{
		glyph: '👋',
		label: 'misc.default',
	},
	{
		glyph: '👋🏻',
		label: 'input.emojiPicker.tones.light',
	},
	{
		glyph: '👋🏼',
		label: 'input.emojiPicker.tones.mediumLight',
	},
	{
		glyph: '👋🏽',
		label: 'input.emojiPicker.tones.medium',
	},
	{
		glyph: '👋🏾',
		label: 'input.emojiPicker.tones.mediumDark',
	},
	{
		glyph: '👋🏿',
		label: 'input.emojiPicker.tones.dark',
	},
]
const skinTone = useStorage('emoji-picker-skin-tone', 0)
const recent = useStorage<string[]>('emoji-picker-recent', [])
const catalog = shallowRef<EmojiCatalogEntry[]>([])
const searchIndex = shallowRef<EmojiEntry[]>([])
const loading = ref(true)
const failed = ref(false)
const query = ref('')
const group = ref(recent.value.length ? -1 : 0)
const activeIndex = ref(0)
const search = useTemplateRef<HTMLInputElement>('search')
const results = useTemplateRef<HTMLElement>('results')
const headingId = useId()
const resultsId = useId()

const searchQuery = computed(() => query.value.trim().replace(/^:|:$/g, ''))
const catalogByEmoji = computed(() => new Map(catalog.value.map(entry => [entry.emoji, entry])))
const visibleEmojis = computed(() => {
	if (searchQuery.value) {
		return [...new Set(filterEmojis(searchIndex.value, searchQuery.value, Infinity)
			.flatMap(match => catalogByEmoji.value.get(match.emoji) ?? []))]
	}
	if (group.value === -1) {
		return recent.value.flatMap(emoji => {
			const entry = catalog.value.find(entry => entry.emoji === emoji
				|| entry.skins?.some(skin => skin.emoji === emoji))
			return entry ? [{
				...entry,
				emoji,
				skins: undefined,
			}] : []
		})
	}
	return catalog.value.filter(entry => entry.group === group.value)
})
const categoryKey = computed(() => categories.find(category => category.group === group.value)?.key ?? 'smileys')

function glyph(entry: EmojiCatalogEntry) {
	return entry.skins?.find(skin => skin.tone === skinTone.value)?.emoji ?? entry.emoji
}

function select(entry: EmojiCatalogEntry) {
	const emoji = glyph(entry)
	recent.value = [emoji, ...recent.value.filter(value => value !== emoji)].slice(0, 24)
	emit('select', emoji)
}

function chooseCategory(value: number) {
	group.value = value
	query.value = ''
}

watch(visibleEmojis, () => {
	activeIndex.value = 0
	if (results.value) results.value.scrollTop = 0
})

async function focusEmoji(index: number) {
	activeIndex.value = Math.max(0, Math.min(index, visibleEmojis.value.length - 1))
	await nextTick()
	const button = results.value?.querySelectorAll<HTMLButtonElement>('.emoji-button')[activeIndex.value]
	button?.focus()
	button?.scrollIntoView({block: 'nearest'})
}

function onResultsKeydown(event: KeyboardEvent) {
	if (!(event.target instanceof HTMLButtonElement)) return
	const rtl = getComputedStyle(event.currentTarget as HTMLElement).direction === 'rtl'
	const movement: Record<string, number> = {
		ArrowRight: rtl ? -1 : 1,
		ArrowLeft: rtl ? 1 : -1,
		ArrowDown: COLUMNS,
		ArrowUp: -COLUMNS,
		Home: -visibleEmojis.value.length,
		End: visibleEmojis.value.length,
	}
	if (!(event.key in movement)) return
	event.preventDefault()
	if (event.key === 'ArrowUp' && activeIndex.value < COLUMNS) {
		search.value?.focus()
		return
	}
	focusEmoji(activeIndex.value + movement[event.key])
}

async function load() {
	loading.value = true
	failed.value = false
	try {
		const [entries, emojiIndex] = await Promise.all([loadEmojiCatalog(), loadEmojis()])
		// group 2 holds lone skin tone and hair components, not postable emoji
		catalog.value = entries.filter(entry => entry.group !== 2)
		searchIndex.value = emojiIndex
	} catch {
		failed.value = true
	} finally {
		loading.value = false
	}
}

async function retry() {
	await load()
	search.value?.focus()
}

onMounted(async () => {
	load()
	await nextTick()
	search.value?.focus()
})
</script>

<template>
	<section
		class="emoji-picker"
		role="region"
		:aria-label="t('input.emojiPicker.title')"
	>
		<div class="emoji-toolbar">
			<input
				ref="search"
				v-model="query"
				type="search"
				class="input search"
				:placeholder="t('input.emojiPicker.search')"
				:aria-label="t('input.emojiPicker.search')"
				:aria-controls="resultsId"
				@keydown.down.prevent="focusEmoji(0)"
			>
			<select
				v-model="skinTone"
				class="emoji-tone"
				:aria-label="t('input.emojiPicker.skinTone')"
			>
				<option
					v-for="(tone, index) in tones"
					:key="index"
					:value="index"
				>
					{{ tone.glyph }} {{ t(tone.label) }}
				</option>
			</select>
		</div>
		<nav
			class="emoji-categories"
			:aria-label="t('input.emojiPicker.categories')"
		>
			<button
				v-for="category in categories"
				:key="category.key"
				type="button"
				:aria-label="t(`input.emojiPicker.groups.${category.key}`)"
				:title="t(`input.emojiPicker.groups.${category.key}`)"
				:aria-pressed="!searchQuery && group === category.group"
				@click="chooseCategory(category.group)"
			>
				<span aria-hidden="true">{{ category.icon }}</span>
			</button>
		</nav>
		<h3
			:id="headingId"
			class="emoji-heading"
		>
			{{ searchQuery ? t('input.emojiPicker.results') : t(`input.emojiPicker.groups.${categoryKey}`) }}
		</h3>
		<div
			:id="resultsId"
			ref="results"
			class="emoji-results"
			role="group"
			:aria-labelledby="headingId"
			:aria-busy="loading"
			@keydown="onResultsKeydown"
		>
			<p
				v-if="loading || failed || !visibleEmojis.length"
				class="emoji-status"
				role="status"
			>
				<template v-if="loading">
					{{ t('misc.loading') }}
				</template>
				<template v-else-if="failed">
					{{ t('input.emojiPicker.loadError') }}
					<XButton
						variant="secondary"
						:shadow="false"
						@click="retry"
					>
						{{ t('project.background.retry') }}
					</XButton>
				</template>
				<template v-else>
					{{ t('input.editor.emoji.empty') }}
				</template>
			</p>
			<button
				v-for="(entry, index) in visibleEmojis"
				v-else
				:key="entry.emoji"
				type="button"
				class="emoji-button"
				:aria-label="entry.annotation"
				:title="entry.annotation"
				:tabindex="index === activeIndex ? 0 : -1"
				@focus="activeIndex = index"
				@click="select(entry)"
			>
				{{ glyph(entry) }}
			</button>
		</div>
	</section>
</template>

<style scoped lang="scss">
.emoji-picker {
	display: flex;
	flex-direction: column;
	inline-size: min(344px, calc(100vw - 16px));
	block-size: 400px;
	max-block-size: calc(100dvh - 88px);
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	background: var(--scheme-main);
	color: var(--grey-900);
	box-shadow: var(--shadow-md);
	overflow: hidden;
}

.emoji-toolbar {
	display: flex;
	gap: .5rem;
	padding: .75rem;

	.search {
		min-inline-size: 0;
		block-size: 34px;
		font-size: 1rem;
	}
}

.emoji-tone {
	inline-size: 3rem;
	flex-shrink: 0;
	border: 1px solid var(--grey-300);
	border-radius: $radius;
	background: var(--scheme-main);
	color: var(--grey-900);
	font-size: 1rem;
}

.emoji-categories {
	display: flex;
	padding: 0 .25rem;
	border-block-end: 1px solid var(--grey-200);

	button {
		flex: 1;
		min-inline-size: 0;
		block-size: 36px;
		border-radius: $radius $radius 0 0;
		font-size: 1rem;

		&[aria-pressed="true"] {
			box-shadow: inset 0 -2px var(--primary);
			background: var(--grey-100);
		}
	}
}

// Nested to beat Bulma's `.content h3:not(:first-child)` margin in task comments
.emoji-picker .emoji-heading {
	margin: 0;
	padding: .75rem .75rem .5rem;
	font-size: .875rem;
	// Overrides the global `h3 {font-weight: 400 !important}` in theme.scss
	font-weight: 600 !important;
}

.emoji-results {
	display: grid;
	grid-template-columns: repeat(v-bind(COLUMNS), minmax(0, 38px));
	justify-content: space-between;
	align-content: start;
	min-block-size: 0;
	padding: 0 .5rem .5rem;
	overflow-y: auto;
	overscroll-behavior: contain;
}

.emoji-categories button,
.emoji-button {
	padding: 0;
	border: 0;
	background: transparent;
	color: inherit;
	cursor: pointer;

	&:hover {
		background: var(--grey-100);
	}

	&:focus-visible {
		outline: 2px solid var(--primary);
		outline-offset: -2px;
	}
}

.emoji-button {
	block-size: 38px;
	border-radius: $radius;
	font-family: 'Apple Color Emoji', 'Segoe UI Emoji', 'Noto Color Emoji', sans-serif;
	font-size: 22px;
	line-height: 22px;
}

.emoji-status {
	grid-column: 1 / -1;
	padding: 1rem .25rem;
	color: var(--grey-600);
	font-size: .875rem;
}
</style>
