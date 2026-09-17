import { ensureStyles } from '../styles/ensureStyles';

export interface PageDotsProps {
  count: number;
  active: number;
  /** Use on the yellow ground. */
  onLight?: boolean;
  /** Makes each dot a control (PRD D46): the drawn dot keeps its size and a centred 44pt hit area is added. */
  onSelect?: (index: number) => void;
}

/** 6px dots, the active one stretched to 22px. Omit on shareable screens so screenshots stay clean. */
export function PageDots({ count, active, onLight = false, onSelect }: PageDotsProps) {
  ensureStyles();
  const dots = Array.from({ length: count }, (_, i) => i);
  return (
    <div
      className="mmr-pagedots"
      data-on-light={onLight ? 'true' : 'false'}
      {...(onSelect
        ? { role: 'tablist' as const, 'aria-label': 'Pages' }
        : { role: 'img' as const, 'aria-label': `Page ${active + 1} of ${count}` })}
    >
      {dots.map((i) =>
        onSelect ? (
          <button
            key={`dot-${i}`}
            type="button"
            className="mmr-pagedot mmr-pressable"
            data-tappable="true"
            data-active={i === active ? 'true' : 'false'}
            role="tab"
            aria-selected={i === active}
            aria-label={`Page ${i + 1} of ${count}`}
            onClick={() => onSelect(i)}
          />
        ) : (
          <span key={`dot-${i}`} className="mmr-pagedot" data-active={i === active ? 'true' : 'false'} />
        ),
      )}
    </div>
  );
}
