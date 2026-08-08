import React from 'react';
import { NavLink } from 'react-router-dom';

interface NavPillProps {
  to: string;
  icon?: React.ReactNode;
  label: string;
  end?: boolean;
}

export function NavPill({ to, icon, label, end = false }: NavPillProps) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `flex flex-col items-center justify-center w-full py-3 gap-1 rounded-full transition-all duration-200 active:scale-95 ${
          isActive
            ? 'bg-primary/15 text-primary font-semibold'
            : 'text-muted-foreground hover:bg-secondary'
        }`
      }
    >
      {icon && <span className="text-2xl leading-none">{icon}</span>}
      <span className="text-xs">{label}</span>
    </NavLink>
  );
}
