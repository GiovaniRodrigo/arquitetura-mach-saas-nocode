import React from 'react';

interface TonalCardProps extends React.HTMLAttributes<HTMLDivElement> {}

export function TonalCard({ children, className = '', ...props }: TonalCardProps) {
  return (
    <div
      className={`bg-secondary text-secondary-foreground rounded-3xl p-6 transition-all duration-200 ${className}`}
      {...props}
    >
      {children}
    </div>
  );
}
