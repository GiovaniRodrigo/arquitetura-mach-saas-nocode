import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Batcher } from "./batcher";

describe("Batcher (RNF07)", () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it("agrupa mutações de uma rajada num único flush após 16ms", () => {
    const lotes: Map<string, string>[] = [];
    const b = new Batcher<string>((lote) => lotes.push(lote));

    b.enfileirar("bi-a", "1");
    b.enfileirar("bi-b", "2");
    b.enfileirar("bi-c", "3");

    // Antes do intervalo, nada foi aplicado.
    vi.advanceTimersByTime(15);
    expect(lotes).toHaveLength(0);

    // Ao completar 16ms, um único flush com as três mutações.
    vi.advanceTimersByTime(1);
    expect(lotes).toHaveLength(1);
    expect(lotes[0].size).toBe(3);
  });

  it("coalesce mutações ao mesmo blind_index (last-write-wins)", () => {
    const lotes: Map<string, string>[] = [];
    const b = new Batcher<string>((lote) => lotes.push(lote));

    b.enfileirar("bi-a", "v1");
    b.enfileirar("bi-a", "v2");
    b.enfileirar("bi-a", "v3");

    vi.advanceTimersByTime(16);
    expect(lotes).toHaveLength(1);
    expect(lotes[0].get("bi-a")).toBe("v3");
    expect(lotes[0].size).toBe(1);
  });

  it("uma nova rajada após o flush gera um segundo lote", () => {
    const lotes: Map<string, string>[] = [];
    const b = new Batcher<string>((lote) => lotes.push(lote));

    b.enfileirar("bi-a", "1");
    vi.advanceTimersByTime(16);
    b.enfileirar("bi-b", "2");
    vi.advanceTimersByTime(16);

    expect(lotes).toHaveLength(2);
    expect(lotes[1].get("bi-b")).toBe("2");
  });

  it("flush() aplica imediatamente e cancela o timer pendente", () => {
    const lotes: Map<string, string>[] = [];
    const b = new Batcher<string>((lote) => lotes.push(lote));

    b.enfileirar("bi-a", "1");
    b.flush();
    expect(lotes).toHaveLength(1);

    // Não deve haver um segundo flush pelo timer já cancelado.
    vi.advanceTimersByTime(32);
    expect(lotes).toHaveLength(1);
    expect(b.pendentesCount).toBe(0);
  });
});
