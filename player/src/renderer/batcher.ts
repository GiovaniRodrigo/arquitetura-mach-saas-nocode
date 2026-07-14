// Batcher de 16ms (RNF07): agrupa mutações de UI recebidas em rajada e aplica um
// único diff por quadro (~60Hz), evitando jank. Mutações ao mesmo blind_index no
// mesmo lote são coalescidas (last-write-wins), de modo que o consumidor recebe um
// só valor por componente por flush.

export type Agendador = (cb: () => void, ms: number) => ReturnType<typeof setTimeout>;
export type Cancelador = (id: ReturnType<typeof setTimeout>) => void;

export interface BatcherOpts {
  /** Intervalo do lote em ms (padrão 16 → ~60Hz). */
  intervaloMs?: number;
  /** Agendador injetável (para testes com timers falsos). */
  agendar?: Agendador;
  cancelar?: Cancelador;
}

export class Batcher<T> {
  private pendentes = new Map<string, T>();
  private timer: ReturnType<typeof setTimeout> | null = null;
  private readonly intervaloMs: number;
  private readonly agendar: Agendador;
  private readonly cancelar: Cancelador;

  constructor(
    private readonly aplicar: (lote: Map<string, T>) => void,
    opts: BatcherOpts = {},
  ) {
    this.intervaloMs = opts.intervaloMs ?? 16;
    this.agendar = opts.agendar ?? ((cb, ms) => setTimeout(cb, ms));
    this.cancelar = opts.cancelar ?? ((id) => clearTimeout(id));
  }

  /** Enfileira uma mutação; agenda o flush se ainda não houver um pendente. */
  enfileirar(blindIndex: string, valor: T): void {
    this.pendentes.set(blindIndex, valor);
    if (this.timer === null) {
      this.timer = this.agendar(() => this.flush(), this.intervaloMs);
    }
  }

  /** Aplica imediatamente o lote pendente (se houver), cancelando o timer. */
  flush(): void {
    if (this.timer !== null) {
      this.cancelar(this.timer);
      this.timer = null;
    }
    if (this.pendentes.size === 0) return;
    const lote = this.pendentes;
    this.pendentes = new Map();
    this.aplicar(lote);
  }

  /** Nº de mutações ainda não aplicadas (diagnóstico/teste). */
  get pendentesCount(): number {
    return this.pendentes.size;
  }
}
