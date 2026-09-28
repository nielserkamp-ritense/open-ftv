// API Client with configurable base URL

import {ApiError} from "@/lib/api/errors.ts";
import {
  VITE_API_BASE_URL, VITE_VLIERDAM_OUTWAY_URL, VITE_BRP_API_KEY,
  VITE_VLIERDAM_PDP1_URL, VITE_DOELBINDING, VITE_DOELBINDING_BRP, VITE_DOELBINDING_BRV,
} from "@/lib/env.ts";

const API_BASE_URL = VITE_API_BASE_URL
const VLIERDAM_API_GATEWAY_URL = VITE_VLIERDAM_OUTWAY_URL
const BRP_API_KEY = VITE_BRP_API_KEY
const DOELBINDING = VITE_DOELBINDING
const DOELBINDING_BRP = VITE_DOELBINDING_BRP
const DOELBINDING_BRV = VITE_DOELBINDING_BRV

interface RequestConfig extends RequestInit {
  params?: Record<string, string | number | boolean>
}

class ApiClient {
  private baseURL: string
  // docker container name of the backend behind baseURL, shown in error toasts.
  private readonly source: string
  // extra headers (e.g. X-Api-Key, Doelbinding) sent on every request.
  private readonly extraHeaders: Record<string, string>

  constructor(baseURL: string, source: string, extraHeaders: Record<string, string> = {}) {
    this.baseURL = baseURL.endsWith('/') ? baseURL.slice(0, -1) : baseURL
    this.source = source
    this.extraHeaders = extraHeaders
  }

  private async request<T>(
    endpoint: string,
    config: RequestConfig = {}
  ): Promise<T> {
    const { params, ...fetchConfig } = config

    let url = `${this.baseURL}${endpoint}`

    // Add query parameters if provided
    if (params) {
      const searchParams = new URLSearchParams()
      Object.entries(params).forEach(([key, value]) => {
        searchParams.append(key, String(value))
      })
      url += `?${searchParams.toString()}`
    }

    let response: Response
    try {
      response = await fetch(url, {
        cache: 'no-store',
        ...fetchConfig,
        headers: {
          'Content-Type': 'application/json',
          ...this.extraHeaders,
          ...fetchConfig.headers,
        },
      })
    } catch (err) {
      // Connection refused / DNS failure / offline - never reached an HTTP response at all.
      throw new ApiError(
        err instanceof Error ? err.message : 'Network error',
        undefined,
        undefined,
        this.source
      )
    }

    if (!response.ok) {
      const raw = await response.text()
      let errorData: unknown = undefined;
      if (typeof raw === "string" && raw.trim() !== "") {
        try {
          errorData = JSON.parse(raw);
        } catch (err) {
          errorData = { parseError: true, raw };
        }
      }

      throw new ApiError(
        `API Error: ${response.status} ${response.statusText}`,
        response.status,
        errorData,
        this.source
      )
    }

    // Handle 204 No Content
    if (response.status === 204) {
      return undefined as T
    }

    return response.json()
  }

  async get<T>(endpoint: string, config?: RequestConfig): Promise<T> {
    return this.request<T>(endpoint, { ...config, method: 'GET' })
  }

  async post<T>(
    endpoint: string,
    data?: unknown,
    config?: RequestConfig
  ): Promise<T> {
    return this.request<T>(endpoint, {
      ...config,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  async put<T>(
    endpoint: string,
    data?: unknown,
    config?: RequestConfig
  ): Promise<T> {
    return this.request<T>(endpoint, {
      ...config,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  async patch<T>(
    endpoint: string,
    data?: unknown,
    config?: RequestConfig
  ): Promise<T> {
    return this.request<T>(endpoint, {
      ...config,
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined,
    })
  }

  async delete<T>(endpoint: string, config?: RequestConfig): Promise<T> {
    return this.request<T>(endpoint, { ...config, method: 'DELETE' })
  }
}

export const apiClient = new ApiClient(API_BASE_URL, 'openftv-vlierdam-dataspace', {'Doelbinding': DOELBINDING})

export const brpApiClient = new ApiClient(
  VLIERDAM_API_GATEWAY_URL,
  'openftv-rvig-dataspace',
  {'Doelbinding': DOELBINDING_BRP, ...(BRP_API_KEY ? {'X-Api-Key': BRP_API_KEY} : {})},
)

export const rdwApiClient = new ApiClient(VLIERDAM_API_GATEWAY_URL, 'openftv-rdw-dataspace', {'Doelbinding': DOELBINDING_BRV})

// Only reachable against the local docker mock stack - null when unconfigured (e.g. in
// production against real backends).
const gatewayClient = (baseURL: string, source: string): ApiClient | null =>
  baseURL ? new ApiClient(baseURL, source) : null

// Vlierdam's own PDP1 (Cedar) - gates who may submit a laadpaal aanvraag (Scenario 1). Called
// directly - no inway hop, since it's an internal decision point, not a cross-organizational boundary.
export const vlierdamPdp1Client = gatewayClient(VITE_VLIERDAM_PDP1_URL, 'openftv-vlierdam-pdp1')
