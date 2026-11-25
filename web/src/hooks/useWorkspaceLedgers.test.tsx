import { afterEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook } from '@testing-library/react';
import * as apiClient from '../api/client';
import { useWorkspaceLedgers } from './useWorkspaceLedgers';

describe('workspace ledger request teardown', () => {
  afterEach(() => vi.restoreAllMocks());

  it.each(['success', 'rejection'] as const)('stops processing a pending %s after unmount', async (outcome) => {
    type Result = Awaited<ReturnType<typeof apiClient.fetchWorkspaceLedgers>>;
    let resolve!: (value: Result) => void;
    let reject!: (error: Error) => void;
    const promise = new Promise<Result>((ok, fail) => {
      resolve = ok;
      reject = fail;
    });
    const fetchLedgers = vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockReturnValueOnce(promise);
    const types = ['feat'];
    const prefixes: string[] = [];
    const { unmount } = renderHook(() => useWorkspaceLedgers(types, prefixes, 'Could not load workspaces'));
    expect(fetchLedgers).toHaveBeenCalledOnce();
    unmount();

    // Observable reads occur when the hook processes the payload or derives the error.
    const readData = vi.fn(() => []);
    const readMessage = vi.fn(() => 'Superseded read failed');
    const error = new Error();
    Object.defineProperty(error, 'message', { get: readMessage });
    await act(async () => {
      if (outcome === 'success') resolve({ get data() { return readData(); } });
      else reject(error);
    });
    expect(readData).not.toHaveBeenCalled();
    expect(readMessage).not.toHaveBeenCalled();
  });
});
