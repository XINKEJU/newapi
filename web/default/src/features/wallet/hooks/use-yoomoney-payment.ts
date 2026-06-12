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
import { useState, useCallback } from 'react'
import i18next from 'i18next'
import { toast } from 'sonner'
import { requestYooMoneyPayment, isApiSuccess } from '../api'

function isSafeUrl(value: string): boolean {
  const trimmed = value.trim()
  if (!trimmed) {
    return false
  }
  try {
    const u = new URL(trimmed)
    return u.protocol === 'http:' || u.protocol === 'https:'
  } catch {
    return false
  }
}

/**
 * Hook for YooMoney payment flow.
 *
 * YooMoney returns a pay_url that the user should be redirected to.
 */
export function useYooMoneyPayment() {
  const [processing, setProcessing] = useState(false)

  const processYooMoneyPayment = useCallback(
    async (amount: number) => {
      try {
        setProcessing(true)

        const response = await requestYooMoneyPayment({
          amount,
          payment_method: 'yoomoney',
        })

        if (!isApiSuccess(response)) {
          toast.error(response.message || i18next.t('Payment request failed'))
          return false
        }

        const url = (response as unknown as { data?: { pay_url?: string } }).data
          ?.pay_url

        if (url && isSafeUrl(url)) {
          toast.success(i18next.t('Redirecting to payment page...'))
          window.location.href = url
          return true
        }

        toast.error(i18next.t('Payment request failed'))
        return false
      } catch (_error) {
        toast.error(i18next.t('Payment request failed'))
        return false
      } finally {
        setProcessing(false)
      }
    },
    []
  )

  return { processing, processYooMoneyPayment }
}

/**
 * Hook for YooMoney subscription payment flow.
 */
export function useYooMoneySubscription() {
  const [processing, setProcessing] = useState(false)

  const processYooMoneySubscription = useCallback(
    async (planId: number) => {
      try {
        setProcessing(true)

        const response = await requestYooMoneySubscription({
          plan_id: planId,
        })

        if (!isApiSuccess(response)) {
          toast.error(response.message || i18next.t('Payment request failed'))
          return false
        }

        const url = (response as unknown as { data?: { pay_url?: string } }).data
          ?.pay_url

        if (url && isSafeUrl(url)) {
          toast.success(i18next.t('Redirecting to payment page...'))
          window.location.href = url
          return true
        }

        toast.error(i18next.t('Payment request failed'))
        return false
      } catch (_error) {
        toast.error(i18next.t('Payment request failed'))
        return false
      } finally {
        setProcessing(false)
      }
    },
    []
  )

  return { processing, processYooMoneySubscription }
}
