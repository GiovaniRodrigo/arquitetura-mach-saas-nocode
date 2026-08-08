// Aba Telas (RF09) — casca de navegação apenas: layout de 3 colunas
// (sidebar de telas, canvas central, painel de propriedades). O editor
// visual funcional (canvas infinito, drag-and-drop, seleção de componente)
// é um editor completo inexistente hoje em services/frontend/src e fica fora do escopo
// atômico desta demanda — ver plan.md §2.3 e Riscos (análogo ao que
// 001-construtor-sistemas-mach-v4 §8 já isolava como "demanda própria").
export function AbaTelas() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-[200px_1fr_240px] gap-4 min-h-[400px]">
      <aside aria-label="Telas" className="bg-card border border-border rounded-2xl p-4">
        <h3 className="text-sm font-heading font-bold text-muted-foreground uppercase tracking-wide mb-3">
          Telas
        </h3>
        <p className="text-sm text-muted-foreground">Nenhuma tela criada ainda.</p>
      </aside>

      <div className="bg-card/50 border border-dashed border-border rounded-2xl flex items-center justify-center p-8">
        <p className="text-sm text-muted-foreground text-center">
          O editor de canvas será disponibilizado em breve.
        </p>
      </div>

      <section aria-label="Propriedades" className="bg-card border border-border rounded-2xl p-4">
        <h3 className="text-sm font-heading font-bold text-muted-foreground uppercase tracking-wide mb-3">
          Propriedades
        </h3>
        <p className="text-sm text-muted-foreground">Selecione um componente para editar suas propriedades.</p>
      </section>
    </div>
  );
}
