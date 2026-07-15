import { Outlet, Link, useLocation } from 'react-router-dom';
import { Home, Folder, Settings } from 'lucide-react';
import { TooltipProvider } from '@/components/ui/tooltip';
import {
  SidebarProvider,
  Sidebar,
  SidebarContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarTrigger,
  SidebarHeader,
  SidebarInset,
  SidebarGroup,
  SidebarGroupContent,
} from '@/components/ui/sidebar';

export function DashboardLayout() {
  const location = useLocation();

  return (
    <TooltipProvider>
      <SidebarProvider>
        <Sidebar>
          <SidebarHeader className="p-4">
            <h2 className="text-lg font-bold font-heading">SaaS NoCode</h2>
          </SidebarHeader>
          <SidebarContent>
            <SidebarGroup>
              <SidebarGroupContent>
                <SidebarMenu>
                  <SidebarMenuItem>
                    <SidebarMenuButton render={<Link to="/dashboard" />} isActive={location.pathname === '/dashboard'}>
                      <Home />
                      <span>Home</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                  <SidebarMenuItem>
                    <SidebarMenuButton render={<Link to="/dashboard/projects" />} isActive={location.pathname.startsWith('/dashboard/projects')}>
                      <Folder />
                      <span>Projects</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                  <SidebarMenuItem>
                    <SidebarMenuButton render={<Link to="/dashboard/settings" />} isActive={location.pathname.startsWith('/dashboard/settings')}>
                      <Settings />
                      <span>Settings</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </SidebarContent>
        </Sidebar>

        <SidebarInset>
          {/* Top App Bar */}
          <header className="flex h-16 shrink-0 items-center justify-between px-4 md:px-6 bg-background/80 backdrop-blur-md sticky top-0 z-20 border-b border-border">
            <div className="flex items-center gap-4">
              <SidebarTrigger className="-ml-2" />
              <h1 className="text-xl font-heading font-medium text-foreground hidden sm:block">Welcome, User</h1>
            </div>
            <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold overflow-hidden cursor-pointer hover:shadow-md transition-shadow">
              U
            </div>
          </header>

          {/* Main Content Area */}
          <main className="flex-1 overflow-y-auto p-4 md:p-8">
            <Outlet />
          </main>
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  );
}
