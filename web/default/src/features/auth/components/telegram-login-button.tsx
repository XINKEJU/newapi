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
import { useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'

interface TelegramLoginButtonProps {
  botName: string
  className?: string
}

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
  botName: _botName,
  className = '',
}: TelegramLoginButtonProps) {
  const { t } = useTranslation()

  const handleTelegramClick = useCallback(() => {
    window.open('/api/oauth/telegram', '_self')
  }, [])

  return (
    <div className={className}>
      <Button
        variant='outline'
        type='button'
        onClick={handleTelegramClick}
        className='h-11 w-full justify-center gap-2 rounded-lg'
      >
        <TelegramIcon />
        {t('Continue with Telegram')}
      </Button>
    </div>
  )
}
