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
import { useEffect, useRef, useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'

declare global {
  interface Window {
    TelegramLoginWidget?: {
      dataOnauth?: (user: TelegramUser) => void
    }
  }
}

export interface TelegramUser {
  id: number
  first_name: string
  last_name?: string
  username?: string
  photo_url?: string
  auth_date: number
  hash: string
}

interface TelegramLoginButtonProps {
  botName: string
  onAuth: (user: TelegramUser) => void
  className?: string
}

const WIDGET_SCRIPT = 'https://telegram.org/js/telegram-widget.js?22'
const WIDGET_TIMEOUT_MS = 3000

const TelegramIcon = () => (
  <svg width='18' height='18' viewBox='0 0 24 24' xmlns='http://www.w3.org/2000/svg'>
    <title>Telegram</title>
    <path
      d='M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69a.19.19 0 0 0-.05-.18c-.05-.03-.14-.03-.21-.02-.09.02-1.49.95-4.22 2.79-.4.27-.76.41-1.08.4-.36-.01-1.04-.2-1.55-.37-.63-.2-1.12-.31-1.08-.66.02-.18.27-.36.74-.55 2.92-1.27 4.86-2.11 5.83-2.51 2.78-1.16 3.35-1.36 3.73-1.36.08 0 .27.02.39.12.1.08.13.19.14.27.01.08.03.25.02.39z'
      fill='currentColor'
    />
  </svg>
)

export function TelegramLoginButton({
  botName,
  onAuth,
  className = '',
}: TelegramLoginButtonProps) {
  const { t } = useTranslation()
  const containerRef = useRef<HTMLDivElement>(null)
  const [widgetFailed, setWidgetFailed] = useState(false)
  const onAuthRef = useRef(onAuth)
  onAuthRef.current = onAuth

  const handleTelegramAuth = useCallback((user: TelegramUser) => {
    onAuthRef.current(user)
  }, [])

  useEffect(() => {
    const container = containerRef.current
    if (!container || !botName) return

    setWidgetFailed(false)

    window.TelegramLoginWidget = {
      dataOnauth: handleTelegramAuth,
    }

    container.innerHTML = ''

    const script = document.createElement('script')
    script.src = WIDGET_SCRIPT
    script.async = true
    script.setAttribute('data-telegram-login', botName)
    script.setAttribute('data-size', 'large')
    script.setAttribute('data-radius', '8')
    script.setAttribute('data-onauth', 'TelegramLoginWidget.dataOnauth(user)')
    script.setAttribute('data-request-access', 'write')

    container.appendChild(script)

    const timeoutId = setTimeout(() => {
      if (container.children.length <= 1) {
        setWidgetFailed(true)
      }
    }, WIDGET_TIMEOUT_MS)

    return () => {
      delete window.TelegramLoginWidget
      clearTimeout(timeoutId)
    }
  }, [botName, handleTelegramAuth])

  const handleFallbackClick = () => {
    window.open(`https://t.me/${botName}`, '_blank')
  }

  return (
    <div className={className}>
      {/* Primary: Telegram Login Widget */}
      <div
        ref={containerRef}
        style={{ minHeight: 44, display: 'flex', justifyContent: 'center' }}
      />

      {/* Fallback: shown when widget fails to load, or as a secondary option */}
      {widgetFailed && (
        <div className='mt-2 space-y-2'>
          <Button
            variant='outline'
            type='button'
            onClick={handleFallbackClick}
            className='h-11 w-full justify-center gap-2 rounded-lg'
          >
            <TelegramIcon />
            {t('Continue with Telegram')}
          </Button>
        </div>
      )}
    </div>
  )
}
