// Landing pública de apresentação do produto (RF01): acessível sem sessão,
// montada fora do AppProvider/DashboardLayout. Os CTAs (RF02) levam ao login
// social existente (/login); não há fluxo de trial dedicado nesta demanda.

import { Link } from 'react-router-dom';

export function Home() {
  return (
    <main className="min-h-screen w-full flex flex-col bg-background text-foreground font-sans">
      <header className="flex items-center justify-between px-6 md:px-12 py-6">
        <span className="font-heading text-lg font-bold">Plataforma MACH</span>
        <Link
          to="/login"
          className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
        >
          Entrar
        </Link>
      </header>

      <section className="flex-1 flex flex-col items-center justify-center text-center px-6 py-16 md:py-24">
        <h1 className="font-heading text-4xl md:text-6xl font-bold tracking-tight max-w-3xl">
          Construa sistemas de negócio sem escrever código
        </h1>
        <p className="text-muted-foreground text-lg md:text-xl mt-6 max-w-2xl">
          A arquitetura headless e no-code para o seu SaaS: telas, regras de negócio e
          publicação em um único lugar.
        </p>
        <div className="flex flex-col sm:flex-row gap-4 mt-10">
          <Link
            to="/login"
            className="px-6 py-3 bg-primary text-primary-foreground rounded-full font-semibold hover:bg-primary/90 active:scale-95 transition-all shadow-sm"
          >
            Testar grátis
          </Link>
          <Link
            to="/login"
            className="px-6 py-3 bg-secondary text-secondary-foreground rounded-full font-semibold hover:bg-secondary/80 active:scale-95 transition-all"
          >
            Entrar
          </Link>
        </div>
      </section>
    </main>
  );
}
