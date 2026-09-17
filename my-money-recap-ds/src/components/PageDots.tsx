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
  return (
    <div className="mmr-pagedots" data-on-light={onLight ? 'true' : 'false'} role="tablist" aria-label="Pages">
      {Array.from({ length: count }, (_, i) =>
        onSelect ? (
          <button
            key={i} type="button" className="mmr-pagedot mmr-pressable" data-tappable="true"
            data-active={i === active ? 'true' : 'false'} role="tab" aria-selected={i === active} aria-label={`Page ${i + 1} of ${count}`}
            onClick={() => onSelect(i)}
          />
        ) : (
          <span key={i} className="mmr-pagedot" data-active={i === active ? 'true' : 'false'} role="tab" aria-selected={i === active} />
        ),
      )}
    </div>
  );
}
