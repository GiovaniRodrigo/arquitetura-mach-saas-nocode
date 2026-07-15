import { useState, useRef, useEffect } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { Home, Folder, Settings, LogOut, User, Search } from 'lucide-react';
import { TooltipProvider } from '@/components/ui/tooltip';
import { encerrarSessao } from '@/auth/session';
import { useApp } from '@/app/AppContext';
import { CommandPalette } from '@/dashboard/CommandPalette';
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

function sair() {
  encerrarSessao();
  window.location.reload();
}

export function DashboardLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { usuario } = useApp();
  const [menuAberto, setMenuAberto] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Fecha o menu do avatar ao clicar fora (C7/RF14).
  useEffect(() => {
    if (!menuAberto) return;
    function onClickFora(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) setMenuAberto(false);
    }
    document.addEventListener('mousedown', onClickFora);
    return () => document.removeEventListener('mousedown', onClickFora);
  }, [menuAberto]);

  const saudacao = usuario.nome ? `Bem-vindo, ${usuario.nome}` : 'Bem-vindo';

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
              <h1 className="text-xl font-heading font-medium text-foreground hidden sm:block">{saudacao}</h1>
            </div>
            <div className="flex items-center gap-3">
              {/* Dica de busca — dispara o atalho Cmd/Ctrl+K do CommandPalette (RF10). */}
              <button
                type="button"
                aria-label="Buscar (Ctrl+K)"
                onClick={() =>
                  document.dispatchEvent(
                    new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true }),
                  )
                }
                className="hidden md:flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground bg-card hover:bg-secondary border border-border rounded-full transition-colors"
              >
                <Search className="w-4 h-4" />
                <span>Buscar…</span>
                <kbd className="text-[11px] font-mono border border-border rounded px-1.5 py-0.5">Ctrl K</kbd>
              </button>
              <button
                type="button"
                onClick={sair}
                className="flex items-center px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground bg-card hover:bg-secondary border border-border rounded-full transition-colors"
              >
                <LogOut className="w-4 h-4 mr-2" />
                Sair
              </button>

              {/* Menu do usuário (C7/RF14) */}
              <div className="relative" ref={menuRef}>
                <button
                  type="button"
                  aria-haspopup="menu"
                  aria-expanded={menuAberto}
                  aria-label="Menu do usuário"
                  onClick={() => setMenuAberto((v) => !v)}
                  className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold overflow-hidden cursor-pointer hover:shadow-md transition-shadow"
                >
                  {usuario.iniciais}
                </button>
                {menuAberto && (
                  <div
                    role="menu"
                    className="absolute right-0 mt-2 w-52 bg-popover text-popover-foreground border border-border rounded-2xl shadow-lg py-2 z-30"
                  >
                    {usuario.nome && (
                      <div className="px-4 py-2 border-b border-border">
                        <p className="text-sm font-medium truncate">{usuario.nome}</p>
                        {usuario.email && (
                          <p className="text-xs text-muted-foreground truncate">{usuario.email}</p>
                        )}
                      </div>
                    )}
                    <button
                      type="button"
                      role="menuitem"
                      onClick={() => { setMenuAberto(false); navigate('/dashboard/settings'); }}
                      className="w-full flex items-center gap-2 px-4 py-2 text-sm hover:bg-secondary transition-colors"
                    >
                      <User className="w-4 h-4" /> Perfil
                    </button>
                    <button
                      type="button"
                      role="menuitem"
                      onClick={() => { setMenuAberto(false); navigate('/dashboard/settings'); }}
                      className="w-full flex items-center gap-2 px-4 py-2 text-sm hover:bg-secondary transition-colors"
                    >
                      <Settings className="w-4 h-4" /> Configurações
                    </button>
                    <button
                      type="button"
                      role="menuitem"
                      onClick={sair}
                      className="w-full flex items-center gap-2 px-4 py-2 text-sm text-destructive hover:bg-secondary transition-colors"
                    >
                      <LogOut className="w-4 h-4" /> Sair
                    </button>
                  </div>
                )}
              </div>
            </div>
          </header>

          {/* Main Content Area */}
          <main className="flex-1 overflow-y-auto p-4 md:p-8">
            <Outlet />
          </main>
        </SidebarInset>

        {/* Command palette global (Cmd/Ctrl+K) */}
        <CommandPalette />
      </SidebarProvider>
    </TooltipProvider>
  );
}
