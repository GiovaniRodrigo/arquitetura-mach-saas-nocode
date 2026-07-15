import { describe, expect, it, vi } from "vitest";
import { ApiClient, ApiError } from "./client";

function respostaJSON(status: number, corpo: unknown): Response {
  return new Response(JSON.stringify(corpo), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("ApiClient (RF03/RN08)", () => {
  it("anexa o JWT como Bearer em toda chamada", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, { permissions: {} }));
    const client = new ApiClient("http://gw", "tok-123", fetchFn as unknown as typeof fetch);

    await client.permissoes("s1", ["bi-1"]);

    const [, init] = fetchFn.mock.calls[0];
    expect((init.headers as Record<string, string>).Authorization).toBe("Bearer tok-123");
  });

  it("monta a query de permissões com sistema_id e bi repetidos", async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(200, { permissions: { "bi-1": { view: true, click: false } } }),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const mapa = await client.permissoes("s1", ["bi-1", "bi-2"]);

    const [url] = fetchFn.mock.calls[0];
    expect(url).toContain("sistema_id=s1");
    expect(url).toContain("bi=bi-1");
    expect(url).toContain("bi=bi-2");
    expect(mapa["bi-1"].view).toBe(true);
  });

  it("422 de formulário devolve o mapa de erros (não lança)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(422, { sucesso: false, erros_validacao: { "bi-idade": "acima do máximo" }, mensagem_status: "Falha de validação" }),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const resp = await client.submeterFormulario("s1", { "bi-idade": "999" });
    expect(resp.sucesso).toBe(false);
    expect(resp.erros_validacao?.["bi-idade"]).toBe("acima do máximo");
  });

  it("erros de transporte viram ApiError", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(401, { codigo: "UNAUTHORIZED", mensagem: "sem token" }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await expect(client.versaoAtiva("s1")).rejects.toBeInstanceOf(ApiError);
  });

  it("listarSistemas desembrulha o campo sistemas", async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(200, { sistemas: [{ id: "a", nome: "Alfa" }, { id: "b", nome: "Beta" }] }),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const lista = await client.listarSistemas();
    const [url] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/sistemas");
    expect(lista).toHaveLength(2);
    expect(lista[0].nome).toBe("Alfa");
  });

  it("criarSistema faz POST com o nome e devolve o sistema", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(201, { id: "sis-1", nome: "Demo" }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const novo = await client.criarSistema("Demo");
    const [, init] = fetchFn.mock.calls[0];
    expect(init.method).toBe("POST");
    expect(JSON.parse(init.body as string)).toEqual({ nome: "Demo" });
    expect(novo.id).toBe("sis-1");
  });

  it("criarSistema propaga 403 como ApiError (cliente final)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(403, { codigo: "FORBIDDEN", mensagem: "sem permissão" }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await expect(client.criarSistema("Demo")).rejects.toBeInstanceOf(ApiError);
  });

  it("com o fetch padrão, invoca o fetch global sem quebrar o binding (this)", async () => {
    // Reproduz o "Illegal invocation": o fetch nativo exige this global. Sem
    // fetchFn injetado, o client deve chamá-lo como função livre (this global).
    const orig = globalThis.fetch;
    const espiao = vi.fn(function (this: unknown) {
      if (this !== undefined && this !== globalThis) {
        throw new TypeError("Failed to execute 'fetch' on 'Window': Illegal invocation");
      }
      return Promise.resolve(respostaJSON(200, { sistemas: [] }));
    });
    globalThis.fetch = espiao as unknown as typeof fetch;
    try {
      const client = new ApiClient("http://gw", "t"); // usa o fetch padrão
      await expect(client.listarSistemas()).resolves.toEqual([]);
      expect(espiao).toHaveBeenCalledTimes(1);
    } finally {
      globalThis.fetch = orig;
    }
  });
});
