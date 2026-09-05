import { useStyles } from '../styles/useStyles';

export interface ToggleProps {
  checked: boolean;
  disabled?: boolean;
  onChange?: (next: boolean) => void;
  'aria-label'?: string;
}
/** iOS-proportioned switch (51x31) tinted violet when on. Settings only. */
export function Toggle({ checked, disabled = false, onChange, ...rest }: ToggleProps) {
  useStyles();
  return (
    <button
      type="button" role="switch" aria-checked={checked} disabled={disabled}
      className="kmr-toggle kmr-pressable" data-checked={checked ? 'true' : 'false'}
      onClick={() => onChange?.(!checked)} {...rest}
    >
      <span className="kmr-toggle-knob" />
    </button>
  );
}
