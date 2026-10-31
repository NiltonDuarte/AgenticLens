import type { LiveMessageFact } from './live';

export type LiveMessageOrder = 'latest-first' | 'oldest-first';

export interface OrderedLiveMessage {
  readonly message: LiveMessageFact;
  readonly turnIndex: number;
}

export const orderLiveMessages = (
  messages: readonly LiveMessageFact[],
  order: LiveMessageOrder,
): readonly OrderedLiveMessage[] => {
  const indexedMessages = messages.map((message, turnIndex) => ({ message, turnIndex }));
  return order === 'latest-first' ? indexedMessages.reverse() : indexedMessages;
};
