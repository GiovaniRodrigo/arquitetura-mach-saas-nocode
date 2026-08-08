import { describe, expect, it, vi } from "vitest";
import { ApiClient, ApiError } from "./client";

function respostaJSON(status: number, corpo: unknown): Response {
  return new Response(JSON.stringify(corpo), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function respostaTexto(status: number, texto: string): Response {
  return new Response(texto, { status, headers: { "Content-Type": "text/plain" } });
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

  it("corpo não-JSON em erro vira ApiError legível (não SyntaxError)", async () => {
    // Reproduz o caso real: proxy devolve página de erro em texto puro. Antes
    // vazava «Unexpected token 'T', "The server"... is not valid JSON».
    const fetchFn = vi.fn().mockResolvedValue(
      respostaTexto(502, "The server encountered a temporary error and could not complete your request."),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const erro = await client.criarSistema("TEste").catch((e) => e);
    expect(erro).toBeInstanceOf(ApiError);
    expect(erro.status).toBe(502);
    expect(erro.message).toContain("The server");
    expect(erro.message).not.toContain("is not valid JSON");
  });

  it("corpo não-JSON em resposta 2xx vira ApiError de resposta inválida", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaTexto(200, "<html>não sou json</html>"));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const erro = await client.listarSistemas().catch((e) => e);
    expect(erro).toBeInstanceOf(ApiError);
    expect(erro.codigo).toBe("RESPOSTA_INVALIDA");
  });

  it("truncamento: mensagem de erro longa não estoura", async () => {
    const longo = "The server ".repeat(50);
    const fetchFn = vi.fn().mockResolvedValue(respostaTexto(503, longo));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const erro = await client.versaoAtiva("s1").catch((e) => e);
    expect(erro).toBeInstanceOf(ApiError);
    expect(erro.message.length).toBeLessThanOrEqual(141);
    expect(erro.message.endsWith("…")).toBe(true);
  });

  it("atualizarPerfil faz PATCH com nome e foto_url (RF17)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, {}));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.atualizarPerfil({ nome: "Ana Silva", foto_url: "http://x/a.png" });
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/conta/perfil");
    expect(init.method).toBe("PATCH");
    expect(JSON.parse(init.body as string)).toEqual({ nome: "Ana Silva", foto_url: "http://x/a.png" });
  });

  it("solicitarTrocaEmail faz POST sem efetivar o e-mail (RF18/RN08)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(202, {}));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.solicitarTrocaEmail("novo@x.com");
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/conta/email");
    expect(init.method).toBe("POST");
    expect(JSON.parse(init.body as string)).toEqual({ novo_email: "novo@x.com" });
  });

  it("confirmarTrocaEmail envia o token recebido por e-mail (RN08)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, {}));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.confirmarTrocaEmail("tok-conf-1");
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/conta/email/confirmar");
    expect(JSON.parse(init.body as string)).toEqual({ token: "tok-conf-1" });
  });

  it("atualizarWhiteLabel faz PUT e sinaliza validação de domínio pendente em 202 (RF13/RNF03)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(202, {}));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const resultado = await client.atualizarWhiteLabel({
      logo_url: "http://x/l.png",
      cor_primaria: "#111111",
      cor_secundaria: "#222222",
      dominio_proprio: "app.parceiro.com",
    });
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/configuracao/white-label");
    expect(init.method).toBe("PUT");
    expect(resultado.validandoDominio).toBe(true);
  });

  it("atualizarSenha faz PUT com senha atual e nova (RF14)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, {}));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.atualizarSenha("atual123", "nova123");
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/conta/senha");
    expect(JSON.parse(init.body as string)).toEqual({ senha_atual: "atual123", senha_nova: "nova123" });
  });

  it("ativarMfa devolve o URI do QR code TOTP (RF15/RNF01)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(200, { segredo_otp_auth_uri: "otpauth://totp/x" }),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const { segredoOtpAuthUri } = await client.ativarMfa();
    expect(segredoOtpAuthUri).toBe("otpauth://totp/x");
    expect(fetchFn.mock.calls[0][0]).toBe("http://gw/api/v1/conta/mfa/ativar");
  });

  it("confirmarMfa envia o código digitado (RF15)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, {}));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.confirmarMfa("123456");
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/conta/mfa/confirmar");
    expect(JSON.parse(init.body as string)).toEqual({ codigo: "123456" });
  });

  it("desativarMfa faz DELETE (RF15)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, {}));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.desativarMfa();
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/conta/mfa");
    expect(init.method).toBe("DELETE");
  });

  it("excluirConta propaga 409 TENANT_ATIVO_VINCULADO como ApiError (RF16/RN07)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(409, { codigo: "TENANT_ATIVO_VINCULADO", mensagem: "Existem tenants ativos vinculados a esta conta." }),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const erro = await client.excluirConta().catch((e) => e);
    expect(erro).toBeInstanceOf(ApiError);
    expect(erro.codigo).toBe("TENANT_ATIVO_VINCULADO");
  });

  it("excluirConta faz DELETE em /conta quando não há bloqueio (RF16)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, {}));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.excluirConta();
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/conta");
    expect(init.method).toBe("DELETE");
  });

  it("listarUltimosAcessos desembrulha o campo eventos (RF04/RN02)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(200, {
        eventos: [{ usuario_nome: "Ana", tenant_nome: "Acme", criado_em: "2026-08-06T12:00:00Z" }],
      }),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const eventos = await client.listarUltimosAcessos();
    expect(fetchFn.mock.calls[0][0]).toBe("http://gw/api/v1/dashboard/ultimos-acessos");
    expect(eventos).toHaveLength(1);
    expect(eventos[0].usuario_nome).toBe("Ana");
  });

  it("listarFeedback aplica o filtro de status na query (RF05/RN03)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, { itens: [] }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.listarFeedback("pendente");
    expect(fetchFn.mock.calls[0][0]).toBe("http://gw/api/v1/dashboard/feedback?status=pendente");
  });

  it("listarFeedback sem filtro não envia query (RF05)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, { itens: [] }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.listarFeedback();
    expect(fetchFn.mock.calls[0][0]).toBe("http://gw/api/v1/dashboard/feedback");
  });

  it("atualizarStatusFeedback faz PATCH para respondido (RN03)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(200, { id: "f1", tenant_nome: "Acme", mensagem: "x", status: "respondido", criado_em: "2026-08-06T12:00:00Z" }),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const atualizado = await client.atualizarStatusFeedback("f1", "respondido");
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/dashboard/feedback/f1");
    expect(init.method).toBe("PATCH");
    expect(JSON.parse(init.body as string)).toEqual({ status: "respondido" });
    expect(atualizado.status).toBe("respondido");
  });

  it("resumoFinanceiro devolve a receita agregada (RF06/RN04)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(200, { receita_total_centavos: 150000, moeda: "BRL", competencia: "2026-08" }),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const resumo = await client.resumoFinanceiro();
    expect(fetchFn.mock.calls[0][0]).toBe("http://gw/api/v1/dashboard/resumo-financeiro");
    expect(resumo.receita_total_centavos).toBe(150000);
  });

  it("listarTenants desembrulha o campo tenants (RF07)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, { tenants: [{ id: "t1", nome: "Acme" }] }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const tenants = await client.listarTenants();
    expect(fetchFn.mock.calls[0][0]).toBe("http://gw/api/v1/tenants");
    expect(tenants).toEqual([{ id: "t1", nome: "Acme" }]);
  });

  it("criarTenant faz POST com o nome e devolve o tenant (RF07)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(201, { id: "t-1", nome: "Acme" }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const novo = await client.criarTenant("Acme");
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/tenants");
    expect(init.method).toBe("POST");
    expect(JSON.parse(init.body as string)).toEqual({ nome: "Acme" });
    expect(novo.id).toBe("t-1");
  });

  it("criarTenant propaga 403 como ApiError (cliente final)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(403, { codigo: "FORBIDDEN", mensagem: "sem permissão" }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await expect(client.criarTenant("Acme")).rejects.toBeInstanceOf(ApiError);
  });

  it("obterTenant faz GET em /tenants/{id} e devolve o tenant (RF07)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, { id: "t1", nome: "Acme" }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const t = await client.obterTenant("t1");
    expect(fetchFn.mock.calls[0][0]).toBe("http://gw/api/v1/tenants/t1");
    expect(t).toEqual({ id: "t1", nome: "Acme" });
  });

  it("obterTenant propaga 404 como ApiError (fora da hierarquia ou inexistente)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(404, { codigo: "NOT_FOUND", mensagem: "cliente não encontrado" }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await expect(client.obterTenant("t1")).rejects.toBeInstanceOf(ApiError);
  });

  it("atualizarTenant faz PATCH com o novo nome e devolve o tenant (RF07)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, { id: "t1", nome: "Novo Nome" }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const atualizado = await client.atualizarTenant("t1", "Novo Nome");
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/tenants/t1");
    expect(init.method).toBe("PATCH");
    expect(JSON.parse(init.body as string)).toEqual({ nome: "Novo Nome" });
    expect(atualizado.nome).toBe("Novo Nome");
  });

  it("excluirTenant faz DELETE em /tenants/{id} (RF07)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, {}));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.excluirTenant("t1");
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/tenants/t1");
    expect(init.method).toBe("DELETE");
  });

  it("listarSistemas aceita filtro opcional de tenant_id (RF08)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, { sistemas: [] }));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.listarSistemas("t1");
    expect(fetchFn.mock.calls[0][0]).toBe("http://gw/api/v1/sistemas?tenant_id=t1");
  });

  it("listarRegrasNegocio desembrulha as regras do sistema (RF10)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(200, { regras: [{ id: "r1", blind_indexes: ["bi-cpf"], tipo: "tamanho", parametros: { max: 11 } }] }),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const regras = await client.listarRegrasNegocio("s1");
    expect(fetchFn.mock.calls[0][0]).toBe("http://gw/api/v1/sistemas/s1/regras-negocio");
    expect(regras).toHaveLength(1);
  });

  it("criarRegraNegocio faz POST com blind_indexes/tipo/parametros (RF10/RN06)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(201, { id: "r1", blind_indexes: ["bi-cpf"], tipo: "tamanho", parametros: { max: 11 } }),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const nova = await client.criarRegraNegocio("s1", { blind_indexes: ["bi-cpf"], tipo: "tamanho", parametros: { max: 11 } });
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/sistemas/s1/regras-negocio");
    expect(init.method).toBe("POST");
    expect(nova.id).toBe("r1");
  });

  it("listarVersoes desembrulha as versões do sistema (RF12)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(200, { versoes: [{ id: "v1", numero: 1, ativa: true, criado_em: "2026-08-06T12:00:00Z" }] }),
    );
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    const versoes = await client.listarVersoes("s1");
    expect(fetchFn.mock.calls[0][0]).toBe("http://gw/api/v1/sistemas/s1/versoes");
    expect(versoes[0].ativa).toBe(true);
  });

  it("publicarVersao faz POST no endpoint de publicação (RF12)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, {}));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.publicarVersao("s1", "v1");
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/sistemas/s1/versoes/v1/publicar");
    expect(init.method).toBe("POST");
  });

  it("reverterVersao faz POST no endpoint de reversão (RF12)", async () => {
    const fetchFn = vi.fn().mockResolvedValue(respostaJSON(200, {}));
    const client = new ApiClient("http://gw", "t", fetchFn as unknown as typeof fetch);

    await client.reverterVersao("s1", "v0");
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe("http://gw/api/v1/sistemas/s1/versoes/v0/reverter");
    expect(init.method).toBe("POST");
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
