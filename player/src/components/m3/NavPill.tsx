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
            ? 'bg-blue-100 text-blue-900 font-semibold'
            : 'text-slate-600 hover:bg-slate-100'
        }`
      }
    >
      {icon && <span className="text-2xl leading-none">{icon}</span>}
      <span className="text-xs">{label}</span>
    </NavLink>
  );
}
