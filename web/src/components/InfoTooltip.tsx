import React, { useCallback, useId, useLayoutEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';

export type TooltipPlacement = 'top' | 'bottom' | 'left' | 'right';

export interface InfoTooltipProps {
  readonly content: string;
  readonly placement?: TooltipPlacement | undefined;
  readonly ariaLabel?: string | undefined;
  readonly className?: string | undefined;
  readonly portal?: boolean | undefined;
}

const VIEWPORT_MARGIN = 8;
const TOOLTIP_GAP = 8;
const FALLBACK_BUBBLE_WIDTH = 260;
const FALLBACK_BUBBLE_HEIGHT = 64;

export const InfoTooltip: React.FC<InfoTooltipProps> = ({
  content,
  placement = 'top',
  ariaLabel = 'Information',
  className,
  portal = false,
}) => {
  const [isVisible, setIsVisible] = useState(false);
  const [portalStyle, setPortalStyle] = useState<React.CSSProperties | undefined>(undefined);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const bubbleRef = useRef<HTMLSpanElement>(null);
  const tooltipId = useId();

  const updatePortalPosition = useCallback(() => {
    const triggerRect = triggerRef.current?.getBoundingClientRect();
    if (!triggerRect) return;

    const bubbleRect = bubbleRef.current?.getBoundingClientRect();
    const hasBubbleSize = Boolean(bubbleRect && bubbleRect.width > 0 && bubbleRect.height > 0);
    const bubbleWidth = hasBubbleSize ? bubbleRect!.width : FALLBACK_BUBBLE_WIDTH;
    const bubbleHeight = hasBubbleSize ? bubbleRect!.height : FALLBACK_BUBBLE_HEIGHT;
    const viewportWidth = window.innerWidth || 0;
    const viewportHeight = window.innerHeight || 0;

    let top: number;
    let left: number;

    switch (placement) {
      case 'bottom':
        top = triggerRect.bottom + TOOLTIP_GAP;
        left = triggerRect.left + triggerRect.width / 2 - bubbleWidth / 2;
        break;
      case 'left':
        top = triggerRect.top + triggerRect.height / 2 - bubbleHeight / 2;
        left = triggerRect.left - bubbleWidth - TOOLTIP_GAP;
        break;
      case 'right':
        top = triggerRect.top + triggerRect.height / 2 - bubbleHeight / 2;
        left = triggerRect.right + TOOLTIP_GAP;
        break;
      case 'top':
      default:
        top = triggerRect.top - bubbleHeight - TOOLTIP_GAP;
        left = triggerRect.left + triggerRect.width / 2 - bubbleWidth / 2;
        break;
    }

    if (placement === 'left' && left < VIEWPORT_MARGIN && triggerRect.right + TOOLTIP_GAP + bubbleWidth <= viewportWidth - VIEWPORT_MARGIN) {
      left = triggerRect.right + TOOLTIP_GAP;
    } else if (placement === 'right' && left + bubbleWidth > viewportWidth - VIEWPORT_MARGIN && triggerRect.left - TOOLTIP_GAP - bubbleWidth >= VIEWPORT_MARGIN) {
      left = triggerRect.left - bubbleWidth - TOOLTIP_GAP;
    } else if (placement === 'top' && top < VIEWPORT_MARGIN && triggerRect.bottom + TOOLTIP_GAP + bubbleHeight <= viewportHeight - VIEWPORT_MARGIN) {
      top = triggerRect.bottom + TOOLTIP_GAP;
    } else if (placement === 'bottom' && top + bubbleHeight > viewportHeight - VIEWPORT_MARGIN && triggerRect.top - TOOLTIP_GAP - bubbleHeight >= VIEWPORT_MARGIN) {
      top = triggerRect.top - bubbleHeight - TOOLTIP_GAP;
    }

    const maxLeft = Math.max(VIEWPORT_MARGIN, viewportWidth - bubbleWidth - VIEWPORT_MARGIN);
    const maxTop = Math.max(VIEWPORT_MARGIN, viewportHeight - bubbleHeight - VIEWPORT_MARGIN);
    const clampedLeft = Math.min(Math.max(VIEWPORT_MARGIN, left), maxLeft);
    const clampedTop = Math.min(Math.max(VIEWPORT_MARGIN, top), maxTop);

    setPortalStyle({
      top: `${Math.round(clampedTop)}px`,
      left: `${Math.round(clampedLeft)}px`,
    });
  }, [placement]);

  useLayoutEffect(() => {
    if (!portal || !isVisible || !content) return;
    updatePortalPosition();
    window.addEventListener('resize', updatePortalPosition);
    window.addEventListener('scroll', updatePortalPosition, true);
    return () => {
      window.removeEventListener('resize', updatePortalPosition);
      window.removeEventListener('scroll', updatePortalPosition, true);
    };
  }, [portal, isVisible, content, updatePortalPosition]);

  if (!content) {
    return null;
  }

  const handleKeyDown = (event: React.KeyboardEvent<HTMLButtonElement>) => {
    if (event.key === 'Escape') {
      setIsVisible(false);
      event.currentTarget.blur();
    }
  };

  const containerClasses = className
    ? `info-tooltip-container ${className}`
    : 'info-tooltip-container';

  const bubbleClasses = `info-tooltip-bubble placement-${placement}${
    portal ? ' is-portaled' : ''
  }${isVisible ? ' is-visible' : ''}`;

  const bubbleElement = (
    <span
      ref={bubbleRef}
      id={tooltipId}
      role="tooltip"
      className={bubbleClasses}
      style={portal ? portalStyle : undefined}
    >
      {content}
    </span>
  );

  return (
    <span className={containerClasses}>
      <button
        ref={triggerRef}
        type="button"
        className="info-tooltip-trigger"
        aria-label={ariaLabel}
        aria-describedby={tooltipId}
        onClick={(event) => {
          event.stopPropagation();
        }}
        onMouseEnter={() => setIsVisible(true)}
        onMouseLeave={() => setIsVisible(false)}
        onFocus={() => setIsVisible(true)}
        onBlur={() => setIsVisible(false)}
        onKeyDown={handleKeyDown}
      >
        <svg
          className="info-tooltip-glyph"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden="true"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="16" x2="12" y2="12" />
          <line x1="12" y1="8" x2="12" y2="8" />
        </svg>
      </button>
      {portal && typeof document !== 'undefined'
        ? createPortal(bubbleElement, document.body)
        : bubbleElement}
    </span>
  );
};
