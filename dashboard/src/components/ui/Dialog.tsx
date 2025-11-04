import React, { useEffect, useRef } from 'react';
import { X } from 'lucide-react';
import { cn } from '@/utils/cn';

interface DialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: React.ReactNode;
}

export function Dialog({ open, onOpenChange, children }: DialogProps) {
  const dialogRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && open) {
        onOpenChange(false);
      }
    };

    if (open) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [open, onOpenChange]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity animate-in fade-in duration-200"
        onClick={() => onOpenChange(false)}
      />

      {/* Dialog */}
      <div
        ref={dialogRef}
        className="relative z-50 animate-in fade-in zoom-in-95 duration-200"
        role="dialog"
        aria-modal="true"
      >
        {children}
      </div>
    </div>
  );
}

interface DialogContentProps {
  children: React.ReactNode;
  className?: string;
  showClose?: boolean;
  onClose?: () => void;
}

export function DialogContent({
  children,
  className,
  showClose = true,
  onClose,
}: DialogContentProps) {
  return (
    <div
      className={cn(
        'relative bg-white dark:bg-dark-secondary rounded-lg shadow-2xl',
        'w-full max-w-lg max-h-[90vh] overflow-hidden',
        'border border-gray-200 dark:border-dark-border',
        className
      )}
      onClick={(e) => e.stopPropagation()}
    >
      {showClose && onClose && (
        <button
          onClick={onClose}
          className="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-white dark:ring-offset-dark-secondary transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-gray-400 dark:focus:ring-gray-500 focus:ring-offset-2 disabled:pointer-events-none text-gray-600 dark:text-text-secondary"
          aria-label="Close"
        >
          <X className="h-4 w-4" />
        </button>
      )}
      {children}
    </div>
  );
}

interface DialogHeaderProps {
  children: React.ReactNode;
  className?: string;
}

export function DialogHeader({ children, className }: DialogHeaderProps) {
  return (
    <div className={cn('px-6 py-4 border-b border-gray-200 dark:border-dark-border', className)}>
      {children}
    </div>
  );
}

interface DialogTitleProps {
  children: React.ReactNode;
  className?: string;
}

export function DialogTitle({ children, className }: DialogTitleProps) {
  return (
    <h2
      className={cn(
        'text-lg font-semibold text-gray-900 dark:text-text-primary leading-none tracking-tight',
        className
      )}
    >
      {children}
    </h2>
  );
}

interface DialogDescriptionProps {
  children: React.ReactNode;
  className?: string;
}

export function DialogDescription({
  children,
  className,
}: DialogDescriptionProps) {
  return (
    <p className={cn('text-sm text-gray-500 dark:text-text-secondary mt-1.5', className)}>{children}</p>
  );
}

interface DialogBodyProps {
  children: React.ReactNode;
  className?: string;
}

export function DialogBody({ children, className }: DialogBodyProps) {
  return <div className={cn('px-6 py-4', className)}>{children}</div>;
}

interface DialogFooterProps {
  children: React.ReactNode;
  className?: string;
}

export function DialogFooter({ children, className }: DialogFooterProps) {
  return (
    <div
      className={cn(
        'px-6 py-4 border-t border-gray-200 dark:border-dark-border bg-gray-50 dark:bg-dark-base',
        'flex items-center justify-end gap-3',
        className
      )}
    >
      {children}
    </div>
  );
}

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'warning' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
  children: React.ReactNode;
}

export function Button({
  variant = 'primary',
  size = 'md',
  loading = false,
  className,
  children,
  disabled,
  ...props
}: ButtonProps) {
  const baseStyles =
    'inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:focus-visible:ring-offset-dark-secondary disabled:cursor-not-allowed disabled:bg-gray-200 dark:disabled:bg-dark-surface disabled:text-gray-400 dark:disabled:text-text-muted disabled:border-gray-200 dark:disabled:border-dark-border';

  const variants = {
    primary:
      'bg-blue-600 dark:bg-blue-700 text-white hover:bg-blue-700 dark:hover:bg-blue-500 active:bg-blue-800 dark:active:bg-blue-600 focus-visible:ring-blue-600 dark:focus-visible:ring-blue-500',
    secondary:
      'bg-white dark:bg-dark-surface text-gray-900 dark:text-text-primary border border-gray-300 dark:border-dark-border hover:bg-gray-50 dark:hover:bg-dark-hover active:bg-gray-100 dark:active:bg-dark-elevated focus-visible:ring-gray-400 dark:focus-visible:ring-gray-500',
    danger:
      'bg-red-600 dark:bg-red-700 text-white hover:bg-red-700 dark:hover:bg-red-500 active:bg-red-800 dark:active:bg-red-600 focus-visible:ring-red-600 dark:focus-visible:ring-red-500',
    warning:
      'bg-amber-600 dark:bg-amber-700 text-white hover:bg-amber-700 dark:hover:bg-amber-500 active:bg-amber-800 dark:active:bg-amber-600 focus-visible:ring-amber-600 dark:focus-visible:ring-amber-500',
    ghost:
      'text-gray-700 dark:text-text-secondary hover:bg-gray-100 dark:hover:bg-dark-hover active:bg-gray-200 dark:active:bg-dark-elevated focus-visible:ring-gray-400 dark:focus-visible:ring-gray-500',
  };

  const sizes = {
    sm: 'h-8 px-3 text-xs',
    md: 'h-10 px-4 text-sm',
    lg: 'h-11 px-6 text-base',
  };

  return (
    <button
      className={cn(
        baseStyles,
        variants[variant],
        sizes[size],
        loading && 'opacity-70 cursor-wait',
        className
      )}
      disabled={disabled || loading}
      {...props}
    >
      {loading && (
        <svg
          className="animate-spin -ml-1 mr-2 h-4 w-4"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12" cy="12" r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      )}
      {children}
    </button>
  );
}

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
}

export function Input({
  label,
  error,
  helperText,
  className,
  id,
  ...props
}: InputProps) {
  const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;

  return (
    <div className="w-full">
      {label && (
        <label
          htmlFor={inputId}
          className="block text-sm font-medium text-gray-700 dark:text-text-primary mb-1.5"
        >
          {label}
        </label>
      )}
      <input
        id={inputId}
        className={cn(
          'flex h-10 w-full rounded-md border border-gray-300 dark:border-dark-border bg-white dark:bg-dark-surface px-3 py-2 text-sm',
          'text-gray-900 dark:text-text-primary',
          'placeholder:text-gray-400 dark:placeholder:text-text-muted',
          'focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent',
          'disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-gray-100 dark:disabled:bg-dark-base',
          error && 'border-red-500 dark:border-red-400 focus:ring-red-500 dark:focus:ring-red-400',
          className
        )}
        {...props}
      />
      {error && <p className="mt-1.5 text-xs text-red-600 dark:text-red-400">{error}</p>}
      {helperText && !error && (
        <p className="mt-1.5 text-xs text-gray-500 dark:text-text-secondary">{helperText}</p>
      )}
    </div>
  );
}
