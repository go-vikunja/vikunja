import {ref, watch, onMounted} from 'vue'

export function useTheme() {
  const theme = ref<'auto'|'light'|'dark'>('auto')

  onMounted(() => {
    const saved = localStorage.getItem('app-theme')
    if (saved === 'light' || saved === 'dark') theme.value = saved
  })

  watch(theme, val => {
    document.documentElement.classList.toggle('dark-mode', val === 'dark')
    localStorage.setItem('app-theme', val)
  }, { flush: 'post' })

  return theme
}
