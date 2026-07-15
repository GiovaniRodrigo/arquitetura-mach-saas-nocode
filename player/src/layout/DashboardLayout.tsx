import { Outlet } from 'react-router-dom';
import { NavPill } from '../components/m3/NavPill';

export function DashboardLayout() {
  return (
    <div className="flex h-screen w-full bg-background text-foreground overflow-hidden font-sans">
      {/* Navigation Rail for Desktop */}
      <nav className="hidden md:flex flex-col items-center w-24 py-4 bg-card border-r border-border z-10">
        <div className="flex flex-col gap-4 w-full px-3 mt-4">
          <NavPill to="/dashboard" label="Home" icon="🏠" end />
          <NavPill to="/dashboard/projects" label="Projects" icon="📁" />
          <NavPill to="/dashboard/settings" label="Settings" icon="⚙️" />
        </div>
      </nav>

      <div className="flex flex-col flex-1 min-w-0 overflow-hidden relative pb-16 md:pb-0">
        {/* Top App Bar */}
        <header className="flex items-center justify-between px-6 py-4 bg-background/80 backdrop-blur-md sticky top-0 z-10 border-b border-border md:border-none">
          <h1 className="text-xl font-heading font-medium text-foreground">Welcome, User</h1>
          <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold overflow-hidden cursor-pointer hover:shadow-md transition-shadow">
            U
          </div>
        </header>

        {/* Main Content Area */}
        <main className="flex-1 overflow-y-auto p-4 md:p-8">
          <Outlet />
        </main>
      </div>

      {/* Bottom Navigation for Mobile */}
      <nav className="md:hidden flex items-center justify-around w-full py-2 bg-card border-t border-border fixed bottom-0 z-20">
          <NavPill to="/dashboard" label="Home" icon="🏠" end />
          <NavPill to="/dashboard/projects" label="Projects" icon="📁" />
          <NavPill to="/dashboard/settings" label="Settings" icon="⚙️" />
      </nav>
    </div>
  );
}
