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
import { useState, useRef, useEffect } from 'react'
import type { AxiosRequestConfig } from 'axios'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useAuthStore, type AuthUser } from '@/stores/auth-store'
import { api } from '@/lib/api'
import { getOAuthState } from '../api'
import {
  buildGitHubOAuthUrl,
  buildDiscordOAuthUrl,
  buildOIDCOAuthUrl,
  buildLinuxDOOAuthUrl,
  buildYandexOAuthUrl,
} from '../lib/oauth'
import { loadVKIDSDK } from '../lib/vkid-sdk'
import type { SystemStatus, CustomOAuthProviderInfo } from '../types'

type LogoutRequestConfig = AxiosRequestConfig & {
  skipErrorHandler?: boolean
}

/**
 * Hook for managing OAuth login
 */
export function useOAuthLogin(status: SystemStatus | null) {
  const { t } = useTranslation()
  const [isLoading, setIsLoading] = useState(false)
  const [githubButtonText, setGithubButtonText] = useState('')
  const [githubButtonDisabled, setGithubButtonDisabled] = useState(false)
  const [vkSDKReady, setVkSDKReady] = useState(false)
  const githubTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const { auth } = useAuthStore()

  useEffect(() => {
    setGithubButtonText(t('Continue with GitHub'))

    return () => {
      if (githubTimeoutRef.current) {
        clearTimeout(githubTimeoutRef.current)
      }
    }
  }, [t])

  // Pre-load VK ID SDK when VK OAuth is enabled
  useEffect(() => {
    if (!status?.vk_oauth || !status?.vk_client_id) return
    loadVKIDSDK()
      .then((sdk) => {
        sdk.Config.init({
          app: Number(status.vk_client_id),
          redirectUrl: `${window.location.origin}/oauth/vk`,
          responseMode: sdk.ConfigResponseMode.Callback,
          source: sdk.ConfigSource.LOWCODE,
          scope: 'email',
        })
        setVkSDKReady(true)
      })
      .catch(() => {
        // SDK failed to load; handleVKLogin will show an error
      })
  }, [status?.vk_oauth, status?.vk_client_id])

  const resetSession = async () => {
    try {
      auth.reset()
    } catch (_error) {
      // ignore store reset errors
    }
    try {
      await api.get('/api/user/logout', {
        skipErrorHandler: true,
      } as LogoutRequestConfig)
    } catch (_error) {
      // ignore logout errors
    }
  }

  const handleGitHubLogin = async () => {
    if (!status?.github_client_id) return
    if (githubButtonDisabled) return

    setIsLoading(true)
    setGithubButtonDisabled(true)
    setGithubButtonText(t('Redirecting to GitHub...'))

    if (githubTimeoutRef.current) {
      clearTimeout(githubTimeoutRef.current)
    }

    githubTimeoutRef.current = setTimeout(() => {
      setIsLoading(false)
      setGithubButtonText(
        t('Request timed out, please refresh and restart GitHub login')
      )
      setGithubButtonDisabled(true)
    }, 20000)

    try {
      await resetSession()
      const state = await getOAuthState()
      if (!state) {
        toast.error(t('Failed to initialize OAuth'))
        if (githubTimeoutRef.current) {
          clearTimeout(githubTimeoutRef.current)
        }
        setIsLoading(false)
        setGithubButtonText(t('Continue with GitHub'))
        setGithubButtonDisabled(false)
        return
      }

      const url = buildGitHubOAuthUrl(status.github_client_id, state)
      window.open(url, '_self')
    } catch (_error) {
      toast.error(t('Failed to start GitHub login'))
      if (githubTimeoutRef.current) {
        clearTimeout(githubTimeoutRef.current)
      }
      setIsLoading(false)
      setGithubButtonText(t('Continue with GitHub'))
      setGithubButtonDisabled(false)
    }
  }

  const handleDiscordLogin = async () => {
    if (!status?.discord_client_id) return

    setIsLoading(true)
    try {
      await resetSession()
      const state = await getOAuthState()
      if (!state) {
        toast.error(t('Failed to initialize OAuth'))
        return
      }

      const url = buildDiscordOAuthUrl(status.discord_client_id, state)
      window.open(url, '_self')
    } catch (_error) {
      toast.error(t('Failed to start Discord login'))
    } finally {
      setIsLoading(false)
    }
  }

  const handleOIDCLogin = async () => {
    if (!status?.oidc_authorization_endpoint || !status?.oidc_client_id) return

    setIsLoading(true)
    try {
      await resetSession()
      const state = await getOAuthState()
      if (!state) {
        toast.error(t('Failed to initialize OAuth'))
        return
      }

      const url = buildOIDCOAuthUrl(
        status.oidc_authorization_endpoint,
        status.oidc_client_id,
        state
      )
      window.open(url, '_self')
    } catch (_error) {
      toast.error(t('Failed to start OIDC login'))
    } finally {
      setIsLoading(false)
    }
  }

  const handleLinuxDOLogin = async () => {
    if (!status?.linuxdo_client_id) return

    setIsLoading(true)
    try {
      await resetSession()
      const state = await getOAuthState()
      if (!state) {
        toast.error(t('Failed to initialize OAuth'))
        return
      }

      const url = buildLinuxDOOAuthUrl(status.linuxdo_client_id, state)
      window.open(url, '_self')
    } catch (_error) {
      toast.error(t('Failed to start LinuxDO login'))
    } finally {
      setIsLoading(false)
    }
  }

  const handleTelegramLogin = async (telegramUser: {
    id: number
    first_name: string
    last_name?: string
    username?: string
    photo_url?: string
    auth_date: number
    hash: string
  }) => {
    setIsLoading(true)
    try {
      const res = await api.get('/api/oauth/telegram/login', {
        params: telegramUser,
      })

      if (res.data.success) {
        const loginUser = res.data.data as AuthUser
        useAuthStore.getState().auth.setUser(loginUser)
        try {
          if (typeof window !== 'undefined' && loginUser?.id != null) {
            window.localStorage.setItem('uid', String(loginUser.id))
          }
        } catch {
          // ignore storage errors
        }
        toast.success(t('Signed in successfully!'))
        window.location.href = '/dashboard'
      } else {
        toast.error(res.data.message || t('Failed to start Telegram login'))
      }
    } catch (_error) {
      toast.error(t('Failed to start Telegram login'))
    } finally {
      setIsLoading(false)
    }
  }

  const handleVKLogin = async () => {
    if (!status?.vk_client_id) {
      toast.error(t('VK OAuth is not properly configured'))
      return
    }

    if (!vkSDKReady || !window.VKIDSDK) {
      toast.error(t('VK ID SDK is not loaded yet, please try again'))
      return
    }

    setIsLoading(true)
    try {
      const VKID = window.VKIDSDK

      // Start login immediately (synchronous, user-initiated — avoids popup blocking)
      const loginPromise = VKID.Auth.login()

      // Reset session while the auth popup is open
      await resetSession()

      // Wait for the user to complete authentication
      const payload = await loginPromise
      const { code, device_id } = payload

      // Exchange code for access token (client-side via SDK)
      const tokenData = await VKID.Auth.exchangeCode(code, device_id)

      // Send access token to backend for verification and session creation
      const res = await api.post('/api/oauth/vk_sdk', {
        access_token: tokenData.access_token,
      })

      if (res.data.success) {
        const loginUser = res.data.data as AuthUser
        useAuthStore.getState().auth.setUser(loginUser)
        try {
          if (typeof window !== 'undefined' && loginUser?.id != null) {
            window.localStorage.setItem('uid', String(loginUser.id))
          }
        } catch {
          // ignore storage errors
        }
        toast.success(t('Signed in successfully!'))
        window.location.href = '/dashboard'
      } else {
        toast.error(res.data.message || t('Failed to start VK login'))
      }
    } catch (_error) {
      toast.error(t('Failed to start VK login'))
    } finally {
      setIsLoading(false)
    }
  }

  const handleYandexLogin = async () => {
    if (!status?.yandex_client_id) {
      toast.error(t('Yandex OAuth is not properly configured'))
      return
    }

    setIsLoading(true)
    try {
      await resetSession()
      const state = await getOAuthState()
      if (!state) {
        toast.error(t('Failed to initialize OAuth'))
        return
      }

      const url = buildYandexOAuthUrl(status.yandex_client_id, state)
      window.open(url, '_self')
    } catch (_error) {
      toast.error(t('Failed to start Yandex login'))
    } finally {
      setIsLoading(false)
    }
  }

  const handleCustomOAuthLogin = async (provider: CustomOAuthProviderInfo) => {
    if (!provider.authorization_endpoint || !provider.client_id) return

    setIsLoading(true)
    try {
      await resetSession()
      const state = await getOAuthState()
      if (!state) {
        toast.error(t('Failed to initialize OAuth'))
        return
      }

      const redirectUri = `${window.location.origin}/oauth/${provider.slug}`
      const url = new URL(provider.authorization_endpoint)
      url.searchParams.set('client_id', provider.client_id)
      url.searchParams.set('redirect_uri', redirectUri)
      url.searchParams.set('response_type', 'code')
      url.searchParams.set('state', state)
      if (provider.scopes) {
        url.searchParams.set('scope', provider.scopes)
      }

      window.open(url.toString(), '_self')
    } catch (_error) {
      toast.error(
        t('Failed to start {{provider}} login', { provider: provider.name })
      )
    } finally {
      setIsLoading(false)
    }
  }

  return {
    isLoading,
    githubButtonText,
    githubButtonDisabled,
    handleGitHubLogin,
    handleDiscordLogin,
    handleOIDCLogin,
    handleLinuxDOLogin,
    handleTelegramLogin,
    handleVKLogin,
    handleYandexLogin,
    handleCustomOAuthLogin,
  }
}
