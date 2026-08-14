import { useState, useRef, useEffect } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { LayoutDashboard, Users, Settings, LogOut, User, Search, HelpCircle, Moon, Sun } from 'lucide-react';
import { TooltipProvider } from '@/components/ui/tooltip';
import { Switch } from '@/components/ui/switch';
import { encerrarSessao } from '@/auth/session';
import { useApp } from '@/app/AppContext';
import { useTheme } from '@/theme/ThemeProvider';
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
  SidebarFooter,
  SidebarInset,
  SidebarGroup,
  SidebarGroupContent,
} from '@/components/ui/sidebar';

function sair() {
  encerrarSessao();
  window.location.reload();
}

// Nome da página atual exibido no cabeçalho, derivado da rota — mesmo
// agrupamento usado nos links da Sidebar.
function tituloDaPagina(pathname: string): string {
  if (pathname === '/dashboard') return 'Dashboard';
  if (pathname.startsWith('/dashboard/clientes')) return 'Clientes';
  if (pathname.startsWith('/dashboard/configuracao')) return 'Configuração';
  if (pathname.startsWith('/dashboard/perfil')) return 'Cadastro/Perfil';
  if (pathname.startsWith('/dashboard/ajuda')) return 'Ajuda';
  return 'Dashboard';
}

export function DashboardLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { usuario } = useApp();
  const { tema, alternarTema } = useTheme();
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

  const titulo = tituloDaPagina(location.pathname);
  // A aba Telas é um editor de viewport fixo (como Figma/Webflow): cabeçalho
  // e abas ficam parados, e cada painel (esquerdo/canvas/direito) rola por
  // conta própria dentro de uma altura definida — em vez da página inteira
  // rolar (comportamento padrão das demais páginas do dashboard).
  const ehAbaTelas = location.pathname.endsWith('/telas');

  return (
    <TooltipProvider>
      <SidebarProvider>
        <Sidebar>
          <SidebarHeader className="p-4">
            <h2 className="text-lg font-bold font-heading">MAYS - Make Your SaaS</h2>
          </SidebarHeader>
          <SidebarContent>
            <SidebarGroup>
              <SidebarGroupContent>
                <SidebarMenu>
                  <SidebarMenuItem>
                    <SidebarMenuButton render={<Link to="/dashboard" />} isActive={location.pathname === '/dashboard'}>
                      <LayoutDashboard />
                      <span>Dashboard</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                  <SidebarMenuItem>
                    <SidebarMenuButton render={<Link to="/dashboard/clientes" />} isActive={location.pathname.startsWith('/dashboard/clientes')}>
                      <Users />
                      <span>Clientes</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                  <SidebarMenuItem>
                    <SidebarMenuButton render={<Link to="/dashboard/configuracao" />} isActive={location.pathname.startsWith('/dashboard/configuracao')}>
                      <Settings />
                      <span>Configuração</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </SidebarContent>
          <SidebarFooter>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton render={<Link to="/dashboard/ajuda" />} isActive={location.pathname.startsWith('/dashboard/ajuda')}>
                  <HelpCircle />
                  <span>Ajuda</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarFooter>
        </Sidebar>

        <SidebarInset>
          {/* Top App Bar */}
          <header className="flex h-16 shrink-0 items-center justify-between px-4 md:px-6 bg-background/80 backdrop-blur-md sticky top-0 z-20 border-b border-border">
            <div className="flex items-center gap-4">
              <SidebarTrigger className="-ml-2" />
              <h1 className="text-xl font-heading font-medium text-foreground hidden sm:block">{titulo}</h1>
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

              {/* Alternância de tema claro/escuro */}
              <label className="flex items-center gap-2 px-1 cursor-pointer" aria-label="Alternar tema claro/escuro">
                <Sun className="w-4 h-4 text-muted-foreground" />
                <Switch
                  checked={tema === 'escuro'}
                  onCheckedChange={alternarTema}
                />
                <Moon className="w-4 h-4 text-muted-foreground" />
              </label>

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
                      onClick={() => { setMenuAberto(false); navigate('/dashboard/perfil'); }}
                      className="w-full flex items-center gap-2 px-4 py-2 text-sm hover:bg-secondary transition-colors"
                    >
                      <User className="w-4 h-4" /> Perfil
                    </button>
                    <button
                      type="button"
                      role="menuitem"
                      onClick={() => { setMenuAberto(false); navigate('/dashboard/configuracao'); }}
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

              {/* Atalho fixo de Perfil/Cadastro, à direita do cabeçalho */}
              <button
                type="button"
                onClick={() => navigate('/dashboard/perfil')}
                aria-label="Perfil/Cadastro"
                className={`flex items-center px-4 py-2 text-sm font-medium border border-border rounded-full transition-colors ${
                  location.pathname.startsWith('/dashboard/perfil')
                    ? 'text-foreground bg-secondary'
                    : 'text-muted-foreground hover:text-foreground bg-card hover:bg-secondary'
                }`}
              >
                <User className="w-4 h-4 mr-2" />
                Perfil/Cadastro
              </button>
            </div>
          </header>

          {/* Main Content Area */}
          <main
            className={
              ehAbaTelas
                ? 'flex flex-col min-h-0 overflow-hidden p-4 md:p-8 h-[calc(100svh-4rem)]'
                : 'flex-1 overflow-y-auto p-4 md:p-8'
            }
          >
            <Outlet />
          </main>
        </SidebarInset>

        {/* Command palette global (Cmd/Ctrl+K) */}
        <CommandPalette />
      </SidebarProvider>
    </TooltipProvider>
  );
}
