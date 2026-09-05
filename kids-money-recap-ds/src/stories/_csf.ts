/* Minimal CSF3 typing so stories compile without @storybook/react installed.
   If you add Storybook, replace these with `Meta` / `StoryObj` from '@storybook/react'. */
import type { ComponentType } from 'react';
export interface Meta<P = Record<string, unknown>> {
  title: string;
  component?: ComponentType<P>;
  parameters?: Record<string, unknown>;
  args?: Partial<P>;
  argTypes?: Record<string, unknown>;
  decorators?: Array<(Story: ComponentType) => JSX.Element>;
}
export interface StoryObj<P = Record<string, unknown>> {
  name?: string;
  args?: Partial<P>;
  render?: (args: P) => JSX.Element;
  parameters?: Record<string, unknown>;
}
