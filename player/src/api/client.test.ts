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
});
