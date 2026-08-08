// Tela de login exibida quando não há sessão. Encaminha ao fluxo OAuth do
// Gateway (Google/GitHub); o token volta no fragmento e é capturado em session.ts.

import { urlLogin } from "./session";

export function Login() {
  return (
    <main className="min-h-screen w-full flex flex-col md:flex-row bg-background text-foreground font-sans">
      <div className="md:w-1/2 bg-zinc-900 flex flex-col justify-center items-center p-12 text-zinc-50 border-r border-zinc-800">
        <div className="max-w-md">
          <h1 className="font-heading text-4xl md:text-5xl font-bold tracking-tight mb-4 text-white">
            Plataforma MACH
          </h1>
          <p className="text-zinc-400 text-lg md:text-xl">
            A arquitetura headless e no-code para o seu SaaS.
          </p>
        </div>
      </div>
      <div className="md:w-1/2 flex flex-col justify-center items-center p-8 md:p-12">
        <div className="w-full max-w-sm space-y-8">
          <div className="text-center">
            <h2 className="font-heading text-3xl font-bold tracking-tight">Bem-vindo de volta</h2>
            <p className="text-muted-foreground mt-2">Entre para acessar seus sistemas.</p>
          </div>
          
          <div className="space-y-4">
            <a href={urlLogin("google")} className="flex items-center justify-center w-full px-4 py-3 bg-zinc-800 hover:bg-zinc-700 text-white rounded-full font-medium transition-colors border border-zinc-700">
              <svg className="w-5 h-5 mr-3" viewBox="0 0 24 24" fill="currentColor"><path d="M12.545,10.239v3.821h5.445c-0.712,2.315-2.647,3.972-5.445,3.972c-3.332,0-6.033-2.701-6.033-6.032s2.701-6.032,6.033-6.032c1.498,0,2.866,0.549,3.921,1.453l2.814-2.814C17.503,2.988,15.139,2,12.545,2C7.021,2,2.543,6.477,2.543,12s4.478,10,10.002,10c8.396,0,10.249-7.85,9.426-11.748L12.545,10.239z"/></svg>
              Entrar com Google
            </a>
            <a href={urlLogin("github")} className="flex items-center justify-center w-full px-4 py-3 bg-zinc-800 hover:bg-zinc-700 text-white rounded-full font-medium transition-colors border border-zinc-700">
              <svg className="w-5 h-5 mr-3" viewBox="0 0 24 24" fill="currentColor"><path d="M12,2C6.477,2,2,6.477,2,12c0,4.419,2.865,8.166,6.839,9.489c0.5,0.09,0.682-0.218,0.682-0.484c0-0.236-0.009-0.866-0.014-1.699c-2.782,0.602-3.369-1.34-3.369-1.34c-0.454-1.156-1.11-1.462-1.11-1.462c-0.908-0.62,0.069-0.608,0.069-0.608c1.004,0.071,1.532,1.03,1.532,1.03c0.892,1.529,2.341,1.087,2.91,0.831c0.091-0.645,0.35-1.085,0.635-1.334c-2.22-0.253-4.555-1.11-4.555-4.943c0-1.091,0.39-1.984,1.03-2.682C6.546,8.54,6.202,7.524,6.746,6.148c0,0,0.84-0.269,2.75,1.025C10.295,6.95,11.15,6.84,12,6.836c0.85,0.004,1.705,0.114,2.504,0.337c1.909-1.294,2.747-1.025,2.747-1.025c0.546,1.376,0.202,2.394,0.1,2.646c0.64,0.699,1.026,1.591,1.026,2.682c0,3.841-2.337,4.687-4.565,4.935c0.359,0.307,0.679,0.915,0.679,1.846c0,1.332-0.012,2.407-0.012,2.734c0,0.269,0.18,0.58,0.688,0.482C19.138,20.161,22,16.416,22,12C22,6.477,17.523,2,12,2z"/></svg>
              Entrar com GitHub
            </a>
          </div>
        </div>
      </div>
    </main>
  );
}
