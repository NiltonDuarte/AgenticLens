import type { ContextHealthStatus } from '../domain/types';

export type AsyncState<T> =
  | { readonly status: 'idle' }
  | { readonly status: 'loading' }
  | { readonly status: 'success'; readonly data: T }
  | { readonly status: 'error'; readonly error: ApiError };

export interface ApiConfig {
  readonly baseUrl: string;
}

export interface ApiError {
  readonly message: string;
  readonly statusCode: number | null;
}

export interface HealthBadgeVariant {
  readonly status: ContextHealthStatus | string;
  readonly bgClass: string;
  readonly textClass: string;
  readonly borderClass: string;
}
