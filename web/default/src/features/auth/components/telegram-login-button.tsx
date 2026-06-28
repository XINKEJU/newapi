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
import { useEffect, useRef, useCallback } from 'react'

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

export function TelegramLoginButton({
  botName,
  onAuth,
  className = '',
}: TelegramLoginButtonProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const onAuthRef = useRef(onAuth)
  onAuthRef.current = onAuth

  const handleTelegramAuth = useCallback((user: TelegramUser) => {
    onAuthRef.current(user)
  }, [])

  useEffect(() => {
    const container = containerRef.current
    if (!container || !botName) return

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

    return () => {
      delete window.TelegramLoginWidget
    }
  }, [botName, handleTelegramAuth])

  return (
    <div
      ref={containerRef}
      className={className}
      style={{ minHeight: 44, display: 'flex', justifyContent: 'center' }}
    />
  )
}
