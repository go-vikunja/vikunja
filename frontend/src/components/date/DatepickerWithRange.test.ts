import {describe, it, expect, beforeEach, afterEach} from 'vitest'
import {mount} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import DatepickerWithRange from './DatepickerWithRange.vue'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'
import en from '@/i18n/lang/en.json'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountPicker() {
    return mount(DatepickerWithRange, {
        props: {modelValue: {dateFrom: '', dateTo: ''}},
        global: {
            plugins: [i18n],
            stubs: ['RouterLink', 'Modal', 'XButton', 'BaseButton', 'Popup', 'flat-pickr'],
        },
    })
}

describe('DatepickerWithRange predefined ranges', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
    })
    it('selects Last Week range', async () => {
        const wrapper = mountPicker()
        ;(wrapper.vm as any).setDateRange(['now/w-1w', 'now/w'])
        await wrapper.vm.$nextTick()
        const last = wrapper.emitted('update:modelValue')?.pop()?.[0]
        expect(last).toEqual({dateFrom: 'now/w-1w', dateTo: 'now/w'})
    })

    it('selects Last Month range', async () => {
        const wrapper = mountPicker()
        ;(wrapper.vm as any).setDateRange(['now/M-1M', 'now/M'])
        await wrapper.vm.$nextTick()
        const last = wrapper.emitted('update:modelValue')?.pop()?.[0]
        expect(last).toEqual({dateFrom: 'now/M-1M', dateTo: 'now/M'})
    })

    // A cleared range (the Custom option) comes back as null via v-model; the
    // modelValue watcher must coerce it, not call null.toISOString().
    it('accepts a null modelValue without crashing', async () => {
        const wrapper = mountPicker()
        await wrapper.setProps({modelValue: {dateFrom: 'now/w', dateTo: 'now/w+1w'}})
        await wrapper.vm.$nextTick()
        expect((wrapper.vm as any).from).toBe('now/w')

        await wrapper.setProps({modelValue: {dateFrom: null, dateTo: null}})
        await wrapper.vm.$nextTick()
        expect((wrapper.vm as any).from).toBe('')
        expect((wrapper.vm as any).to).toBe('')
    })
})

describe('DatepickerWithRange Jalali picker', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
        globalI18n.global.locale.value = 'fa-IR'
        useAuthStore().setUserSettings({timezone: 'Asia/Tehran'} as never)
    })

    afterEach(() => {
        globalI18n.global.locale.value = 'en'
    })

    function mountJalaliPicker() {
        return mount(DatepickerWithRange, {
            props: {modelValue: {dateFrom: '2026-09-07T12:00:00Z', dateTo: '2026-09-10T12:00:00Z'}},
            global: {
                plugins: [i18n],
                stubs: {
                    RouterLink: true,
                    Modal: true,
                    XButton: true,
                    BaseButton: true,
                    // The default stub drops named slots; render #content so the grid mounts.
                    Popup: {template: '<div><slot name="content" :is-open="true" /></div>'},
                    'flat-pickr': true,
                },
            },
        })
    }

    function dayButton(wrapper: ReturnType<typeof mountJalaliPicker>, label: string) {
        const found = wrapper.findAll('.jalali-calendar__day').find((button) => button.text() === label)
        if (!found) {
            throw new Error(`day button ${label} not found`)
        }
        return found
    }

    it('renders the Jalali grid instead of flatpickr', () => {
        const wrapper = mountJalaliPicker()
        expect(wrapper.find('.jalali-calendar').exists()).toBe(true)
        expect(wrapper.find('flat-pickr-stub').exists()).toBe(false)
    })

    it('picks a Jalali range as Gregorian ISO strings', async () => {
        const wrapper = mountJalaliPicker()

        // First click only arms the range, like flatpickr's range mode.
        await dayButton(wrapper, '۱۲').trigger('click')
        expect(wrapper.emitted('update:modelValue')).toBeUndefined()

        await dayButton(wrapper, '۱۴').trigger('click')
        const last = wrapper.emitted('update:modelValue')?.pop()?.[0] as {dateFrom: string, dateTo: string}
        // Start of Shahrivar 12, end of Shahrivar 14, Asia/Tehran wall-clock.
        expect(last).toEqual({
            dateFrom: '2026-09-02T20:30:00.000Z',
            dateTo: '2026-09-05T20:29:00.000Z',
        })
    })

    it('keeps datemath presets working', async () => {
        const wrapper = mountJalaliPicker()
        ;(wrapper.vm as any).setDateRange(['now/w', 'now/w+1w'])
        await wrapper.vm.$nextTick()
        const last = wrapper.emitted('update:modelValue')?.pop()?.[0]
        expect(last).toEqual({dateFrom: 'now/w', dateTo: 'now/w+1w'})
    })
})
