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
const WIDGET_TIMEOUT_MS = 5000

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

    // Set up the global callback before Telegram script renders the widget
    window.TelegramLoginWidget = {
      dataOnauth: handleTelegramAuth,
    }

    // Clean any previous widget
    container.innerHTML = ''

    // Create the Telegram login widget script element
    const script = document.createElement('script')
    script.src = WIDGET_SCRIPT
    script.async = true
    script.setAttribute('data-telegram-login', botName)
    script.setAttribute('data-size', 'large')
    script.setAttribute('data-radius', '8')
    script.setAttribute('data-onauth', 'TelegramLoginWidget.dataOnauth(user)')
    script.setAttribute('data-request-access', 'write')

    container.appendChild(script)

    // Set a timeout: if the widget doesn't render within the timeout, show fallback
    const timeoutId = setTimeout(() => {
      // Check if the widget rendered an iframe (script tag + iframe = 2 children)
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
    window.location.href = '/api/oauth/telegram'
  }

  if (widgetFailed) {
    return (
      <div className={className} style={{ display: 'flex', justifyContent: 'center' }}>
        <Button
          variant='outline'
          type='button'
          onClick={handleFallbackClick}
          className='h-11 w-full justify-center gap-2 rounded-lg'
        >
          <svg width='18' height='18' viewBox='0 0 24 24' fill='#26A5E4'>
            <path d='M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.562 8.161c-.18 1.897-.962 6.502-1.359 8.627-.168.9-.5 1.201-.82 1.23-.697.064-1.226-.46-1.901-.903-1.056-.692-1.653-1.123-2.678-1.798-1.185-.781-.417-1.21.258-1.911.177-.184 3.247-2.977 3.307-3.23.007-.032.015-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.139-5.062 3.345-.479.329-.913.489-1.302.481-.428-.009-1.252-.242-1.865-.441-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.831-2.529 6.998-3.015 3.333-1.386 4.025-1.627 4.476-1.635.099-.002.321.023.465.139.121.098.154.228.17.32.016.092.036.302.02.466z'/>
          </svg>
          {t('Continue with Telegram')}
        </Button>
      </div>
    )
  }

  return (
    <div
      ref={containerRef}
      className={className}
      style={{ minHeight: 44, display: 'flex', justifyContent: 'center' }}
    />
  )
}
