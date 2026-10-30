import { describe, expect, it } from 'vitest';
import { mockLiveTraceView } from '../test/fixtures/mockData';
import { orderLiveMessages } from './liveMessageOrdering';

describe('orderLiveMessages', () => {
  it('returns latest-first messages without mutating the source array', () => {
    const messages = mockLiveTraceView.messages;

    const ordered = orderLiveMessages(messages, 'latest-first');

    expect(ordered.map(({ message }) => message.id)).toEqual(
      [...messages].reverse().map((message) => message.id),
    );
    expect(messages.map((message) => message.id)).toEqual(
      mockLiveTraceView.messages.map((message) => message.id),
    );
    expect(ordered.map(({ turnIndex }) => turnIndex)).toEqual([3, 2, 1, 0]);
  });

  it('preserves source order for oldest-first messages', () => {
    const ordered = orderLiveMessages(mockLiveTraceView.messages, 'oldest-first');

    expect(ordered.map(({ message }) => message.id)).toEqual(
      mockLiveTraceView.messages.map((message) => message.id),
    );
    expect(ordered.map(({ turnIndex }) => turnIndex)).toEqual([0, 1, 2, 3]);
  });
});
