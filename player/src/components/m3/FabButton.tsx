import React from 'react';

interface FabButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  icon?: React.ReactNode;
}

export function FabButton({ icon, children, className = '', ...props }: FabButtonProps) {
  return (
    <button
      className={`fixed bottom-6 right-6 flex items-center justify-center gap-2 bg-blue-200 text-blue-900 rounded-2xl px-5 py-4 shadow-md hover:shadow-lg active:scale-95 transition-all duration-200 font-medium z-50 ${className}`}
      {...props}
    >
      {icon && <span className="text-2xl leading-none">{icon}</span>}
      {children && <span>{children}</span>}
    </button>
  );
}
