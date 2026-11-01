import React, { useState, useId } from 'react';

export type TooltipPlacement = 'top' | 'bottom' | 'left' | 'right';

export interface InfoTooltipProps {
  readonly content: string;
  readonly placement?: TooltipPlacement | undefined;
  readonly ariaLabel?: string | undefined;
  readonly className?: string | undefined;
}

export const InfoTooltip: React.FC<InfoTooltipProps> = ({
  content,
  placement = 'top',
  ariaLabel = 'Information',
  className,
}) => {
  const [isVisible, setIsVisible] = useState(false);
  const tooltipId = useId();

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
    isVisible ? ' is-visible' : ''
  }`;

  return (
    <span className={containerClasses}>
      <button
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
      <span
        id={tooltipId}
        role="tooltip"
        className={bubbleClasses}
      >
        {content}
      </span>
    </span>
  );
};
