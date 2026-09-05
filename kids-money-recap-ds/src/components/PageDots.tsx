import { useStyles } from '../styles/useStyles';

export interface PageDotsProps { count: number; active: number; /** Use on the yellow ground. */ onLight?: boolean; }
/** 6px dots, the active one stretched to 22px. Omit on shareable screens so screenshots stay clean. */
export function PageDots({ count, active, onLight = false }: PageDotsProps) {
  useStyles();
  return (
    <div className="kmr-pagedots" data-on-light={onLight ? 'true' : 'false'} role="tablist" aria-label="Pages">
      {Array.from({ length: count }, (_, i) => (
        <span key={i} className="kmr-pagedot" data-active={i === active ? 'true' : 'false'} role="tab" aria-selected={i === active} />
      ))}
    </div>
  );
}
