import { ensureStyles } from '../styles/ensureStyles';

export interface ToggleProps {
  checked: boolean;
  disabled?: boolean;
  onChange?: (next: boolean) => void;
  'aria-label'?: string;
}
/** iOS-proportioned switch (51x31) tinted violet when on: a native checkbox drawn by CSS. Settings only. */
export function Toggle({ checked, disabled = false, onChange, ...rest }: ToggleProps) {
  ensureStyles();
  return (
    <input
      type="checkbox"
      // biome-ignore lint/a11y/useAriaPropsForRole: the native checkbox's checked state is authoritative; ARIA in HTML prohibits aria-checked here.
      role="switch"
      className="mmr-toggle mmr-pressable"
      checked={checked}
      disabled={disabled}
      onChange={(e) => onChange?.(e.currentTarget.checked)}
      {...rest}
    />
  );
}
