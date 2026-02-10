import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN.ts'
import enUS from './locales/en-US.ts'

const messages = {
  'zh-CN': zhCN,
  'en-US': enUS
}

const supportedLocales = ['zh-CN', 'en-US'] as const

// 从 localStorage 读取语言，不在支持列表则回退到中文
const savedLocaleRaw = localStorage.getItem('locale') || 'zh-CN'
const savedLocale = supportedLocales.includes(savedLocaleRaw as any) ? savedLocaleRaw : 'zh-CN'

const i18n = createI18n({
  legacy: false,
  locale: savedLocale,
  fallbackLocale: 'zh-CN',
  globalInjection: true,
  messages
})

export default i18n