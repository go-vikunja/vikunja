import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { Agent as HttpAgent } from 'http';
import { Agent as HttpsAgent } from 'https';
import { config } from '../config/index.js';
import { logger } from '../utils/logger.js';
import { mapVikunjaError } from '../utils/errors.js';

/**
 * Retry configuration
 */
const MAX_RETRIES = 3;
const RETRY_DELAY_MS = 1000;

/**
 * Check if error is retryable
 */
function isRetryableError(error: AxiosError): boolean {
  // Don't retry 4xx errors (client errors)
  if (error.response && error.response.status >= 400 && error.response.status < 500) {
    return false;
  }
  // Retry network errors and 5xx errors
  return true;
}

/**
 * Sleep for specified milliseconds
 */
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Vikunja API client with connection pooling and retries
 */
export class VikunjaClient {
  private readonly axios: AxiosInstance;

  constructor() {
    this.axios = axios.create({
      baseURL: config.vikunjaApiUrl,
      timeout: 5000,
      headers: {
        'Content-Type': 'application/json',
      },
      // Enable connection pooling
      httpAgent: new HttpAgent({ keepAlive: true }),
      httpsAgent: new HttpsAgent({ keepAlive: true }),
    });

    // Request interceptor for logging
    this.axios.interceptors.request.use(
      (config) => {
        logger.debug(`Vikunja API request: ${config.method?.toUpperCase()} ${config.url ?? ''}`);
        return config;
      },
      (error: AxiosError) => {
        logger.error('Vikunja API request error', { error: error.message });
        return Promise.reject(error);
      }
    );

    // Response interceptor for logging
    this.axios.interceptors.response.use(
      (response: AxiosResponse) => {
        logger.debug(
          `Vikunja API response: ${response.status} ${response.config.url ?? ''}`
        );
        return response;
      },
      (error: AxiosError) => {
        logger.error('Vikunja API response error', {
          status: error.response?.status,
          url: error.config?.url,
        });
        return Promise.reject(error);
      }
    );
  }

  /**
   * Set authentication token (deprecated - use token parameter in API methods instead)
   * @deprecated This method causes race conditions in concurrent requests. Pass token to each API method instead.
   */
  setToken(_token: string): void {
    logger.warn('setToken() is deprecated and should not be used. Pass token to API methods instead.');
  }

  /**
   * Get request config with auth header
   */
  private getConfig(token?: string, config?: AxiosRequestConfig): AxiosRequestConfig {
    const headers: Record<string, string> = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    return { ...config, headers };
  }

  /**
   * Perform request with retries
   */
  private async requestWithRetries<T>(
    fn: () => Promise<AxiosResponse<T>>,
    retriesLeft = MAX_RETRIES
  ): Promise<T> {
    try {
      const response = await fn();
      return response.data;
    } catch (error) {
      if (error instanceof Error && 'isAxiosError' in error) {
        const axiosError = error as AxiosError;

        // If retryable and retries left, try again
        if (isRetryableError(axiosError) && retriesLeft > 0) {
          const delay = RETRY_DELAY_MS * (MAX_RETRIES - retriesLeft + 1);
          logger.warn(
            `Retrying Vikunja API request in ${delay}ms (${retriesLeft} retries left)`
          );
          await sleep(delay);
          return this.requestWithRetries(fn, retriesLeft - 1);
        }

        // Map to MCP error
        throw mapVikunjaError({
          response: axiosError.response
            ? {
                status: axiosError.response.status,
                data: axiosError.response.data as { message?: string; code?: number },
              }
            : undefined,
          message: axiosError.message,
        });
      }
      throw error;
    }
  }

  /**
   * GET request
   */
  async get<T>(path: string, params?: Record<string, unknown>, token?: string): Promise<T> {
    return this.requestWithRetries(() =>
      this.axios.get<T>(path, this.getConfig(token, { params }))
    );
  }

  /**
   * POST request
   */
  async post<T>(path: string, data?: unknown, token?: string): Promise<T> {
    return this.requestWithRetries(() =>
      this.axios.post<T>(path, data, this.getConfig(token))
    );
  }

  /**
   * PUT request
   */
  async put<T>(path: string, data?: unknown, token?: string): Promise<T> {
    return this.requestWithRetries(() =>
      this.axios.put<T>(path, data, this.getConfig(token))
    );
  }

  /**
   * DELETE request
   */
  async delete<T>(path: string, token?: string): Promise<T> {
    return this.requestWithRetries(() =>
      this.axios.delete<T>(path, this.getConfig(token))
    );
  }
}
