/**
 * VK ID SDK loader and type declarations
 * @see https://vkcom.github.io/vkid-web-sdk/docs/
 */

export interface VKIDAuthPayload {
  code: string
  device_id: string
  state?: string
}

export interface VKIDTokenResponse {
  access_token: string
  refresh_token?: string
  id_token?: string
  token_type?: string
  expires_in?: number
  user_id?: string
  state?: string
  scope?: string
}

export interface VKIDConfig {
  app: number
  redirectUrl: string
  responseMode?: string
  source?: string
  scope?: string
  state?: string
  codeVerifier?: string
}

export interface VKIDSDK {
  Config: {
    init(config: VKIDConfig): void
  }
  Auth: {
    login(): Promise<VKIDAuthPayload>
    exchangeCode(code: string, deviceId: string): Promise<VKIDTokenResponse>
  }
  ConfigResponseMode: {
    Callback: string
    Redirect: string
  }
  ConfigSource: {
    LOWCODE: string
  }
  Scheme: {
    LIGHT: string
    DARK: string
  }
  Languages: {
    RUS: number
  }
}

declare global {
  interface Window {
    VKIDSDK?: VKIDSDK
  }
}

const SDK_SCRIPT_URL =
  'https://unpkg.com/@vkid/sdk@<3.0.0/dist-sdk/umd/index.js'

let sdkLoadPromise: Promise<VKIDSDK> | null = null

/**
 * Loads the VK ID SDK script. Returns a cached promise so the SDK
 * is only loaded once even if called multiple times.
 */
export function loadVKIDSDK(): Promise<VKIDSDK> {
  if (window.VKIDSDK) {
    return Promise.resolve(window.VKIDSDK)
  }
  if (sdkLoadPromise) {
    return sdkLoadPromise
  }
  sdkLoadPromise = new Promise<VKIDSDK>((resolve, reject) => {
    const script = document.createElement('script')
    script.src = SDK_SCRIPT_URL
    script.async = true
    script.onload = () => {
      if (window.VKIDSDK) {
        resolve(window.VKIDSDK)
      } else {
        sdkLoadPromise = null
        reject(new Error('VK ID SDK loaded but not initialized'))
      }
    }
    script.onerror = () => {
      sdkLoadPromise = null
      reject(new Error('Failed to load VK ID SDK script'))
    }
    document.head.appendChild(script)
  })
  return sdkLoadPromise
}
