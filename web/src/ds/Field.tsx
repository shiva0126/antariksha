import { cloneElement, isValidElement, useId, type InputHTMLAttributes, type ReactElement, type ReactNode, type SelectHTMLAttributes, type TextareaHTMLAttributes } from 'react';

/** Label above, hint below, error in red. Wraps exactly one control. */
export function Field({ label, hint, error, optional, children, className, span }: { label: ReactNode; hint?: ReactNode; error?: ReactNode; optional?: boolean; children: ReactElement; className?: string; span?: boolean }) {
  const id = useId();
  const control = isValidElement(children) ? cloneElement(children as ReactElement<{ id?: string; 'aria-invalid'?: boolean; 'aria-describedby'?: string }>, { id: (children.props as { id?: string }).id ?? id, 'aria-invalid': error ? true : undefined, 'aria-describedby': hint || error ? id + '-d' : undefined }) : children;
  return (
    <div className={['ds-field', error && 'ds-field--invalid', span && 'ds-span-2', className].filter(Boolean).join(' ')}>
      <label className="ds-field__label" htmlFor={(children.props as { id?: string }).id ?? id}>{label}{optional && <em> (optional)</em>}</label>
      {control}
      {(error || hint) && <span id={id + '-d'} className={error ? 'ds-field__error' : 'ds-field__hint'} role={error ? 'alert' : undefined}>{error || hint}</span>}
    </div>
  );
}

export const Input = (p: InputHTMLAttributes<HTMLInputElement>) => <input {...p} />;
export const Textarea = (p: TextareaHTMLAttributes<HTMLTextAreaElement>) => <textarea rows={3} {...p} />;

export function Select({ options, placeholder, ...p }: SelectHTMLAttributes<HTMLSelectElement> & { options: readonly (readonly [string, string])[]; placeholder?: string }) {
  return <select {...p}>{placeholder !== undefined && <option value="">{placeholder}</option>}{options.map(([v, l]) => <option key={v} value={v}>{l}</option>)}</select>;
}

export function Checkbox({ label, ...p }: InputHTMLAttributes<HTMLInputElement> & { label: ReactNode }) {
  return <label className="ds-check"><input type="checkbox" {...p} /><span>{label}</span></label>;
}

export function FormGrid({ children }: { children: ReactNode }) { return <div className="ds-form-grid">{children}</div>; }
