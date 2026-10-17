import type { ContextHealthStatus } from '../domain/types';

export type AsyncState<T> =
  | { readonly status: 'idle' }
  | { readonly status: 'loading' }
  | { readonly status: 'success'; readonly data: T; readonly isMock: boolean }
  | { readonly status: 'error'; readonly error: ApiError };

export interface ApiConfig {
  readonly baseUrl: string;
  readonly preferMock: boolean;
}

export interface ApiError {
  readonly message: string;
  readonly statusCode: number | null;
  readonly isFallback: boolean;
}

export interface HealthBadgeVariant {
  readonly status: ContextHealthStatus | string;
  readonly bgClass: string;
  readonly textClass: string;
  readonly borderClass: string;
}
