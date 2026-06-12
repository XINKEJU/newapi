/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/ru'
import 'dayjs/locale/zh'
import 'dayjs/locale/fr'
import 'dayjs/locale/ja'
import 'dayjs/locale/vi'

dayjs.extend(relativeTime)

/**
 * Map i18n language codes to dayjs locale identifiers.
 * Call this whenever the app language changes.
 */
export function setDayjsLocale(lang: string): void {
  const localeMap: Record<string, string> = {
    ru: 'ru',
    zh: 'zh',
    fr: 'fr',
    ja: 'ja',
    vi: 'vi',
    en: 'en',
  }
  const locale = localeMap[lang] ?? 'en'
  dayjs.locale(locale)
}

export default dayjs
