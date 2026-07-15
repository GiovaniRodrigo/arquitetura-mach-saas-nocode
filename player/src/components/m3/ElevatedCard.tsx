import React from 'react';

interface ElevatedCardProps extends React.HTMLAttributes<HTMLDivElement> {}

export function ElevatedCard({ children, className = '', ...props }: ElevatedCardProps) {
  return (
    <div
      className={`bg-white text-slate-900 rounded-3xl p-6 shadow-sm hover:shadow-md transition-all duration-200 border border-slate-100 ${className}`}
      {...props}
    >
      {children}
    </div>
  );
}
