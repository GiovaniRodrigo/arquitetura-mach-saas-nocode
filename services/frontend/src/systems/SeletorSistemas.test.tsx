import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import type { ApiClient } from "../api/client";
import { ApiError } from "../api/client";
import type { Sistema } from "../api/types";
import type { UsuarioAutenticado } from "../auth/jwt";
import { SeletorSistemas } from "./SeletorSistemas";

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

const usuarioDono: UsuarioAutenticado = {
  nome: "Ana",
  iniciais: "A",
  podeCriarSistema: true,
  tenantId: "tenant-1",
};

const usuarioCliente: UsuarioAutenticado = {
  nome: "Bruno",
  iniciais: "B",
  podeCriarSistema: false,
  tenantId: "tenant-1",
};

// Renderiza dentro de um Router (useNavigate exige contexto de Router) e expõe
// a rota do construtor (destino da navegação, RF09-RF12) para asserção.
function renderSeletor(client: ApiClient, usuario: UsuarioAutenticado) {
  return render(
    <MemoryRouter initialEntries={["/sistemas"]}>
      <Routes>
        <Route path="/sistemas" element={<SeletorSistemas client={client} usuario={usuario} />} />
        <Route
          path="/dashboard/clientes/:tenantId/sistemas/:sistemaId/telas"
          element={<div>Construtor aberto</div>}
        />
      </Routes>
    </MemoryRouter>,
  );
}

describe("SeletorSistemas (RN01)", () => {
  it("lista os sistemas do tenant", async () => {
    const sistemas: Sistema[] = [
      { id: "a", nome: "Alfa" },
      { id: "b", nome: "Beta" },
    ];
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue(sistemas) });

    renderSeletor(client, usuarioDono);

    expect(await screen.findByText("Alfa")).toBeTruthy();
    expect(screen.getByText("Beta")).toBeTruthy();
  });

  it("mostra empty state quando não há sistemas", async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    renderSeletor(client, usuarioDono);
    expect(await screen.findByText(/Nenhum sistema ainda/)).toBeTruthy();
  });

  it("mostra erro com retry quando a listagem falha", async () => {
    const listar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, "INTERNAL", "boom"))
      .mockResolvedValueOnce([{ id: "a", nome: "Alfa" }]);
    const client = fakeClient({ listarSistemas: listar });

    renderSeletor(client, usuarioDono);

    const retry = await screen.findByText("Tentar novamente");
    fireEvent.click(retry);

    expect(await screen.findByText("Alfa")).toBeTruthy();
    expect(listar).toHaveBeenCalledTimes(2);
  });

  it("clicar num sistema abre o construtor (Telas/Regras/Versão) no tenant do usuário", async () => {
    const client = fakeClient({
      listarSistemas: vi.fn().mockResolvedValue([{ id: "sys-1", nome: "Alfa" }]),
    });
    renderSeletor(client, usuarioDono);

    fireEvent.click(await screen.findByText("Alfa"));

    expect(await screen.findByText("Construtor aberto")).toBeTruthy();
  });

  it("criar um sistema novo já abre o construtor dele", async () => {
    const client = fakeClient({
      listarSistemas: vi.fn().mockResolvedValue([]),
      criarSistema: vi.fn().mockResolvedValue({ id: "sys-novo", nome: "Gama" }),
    });
    renderSeletor(client, usuarioDono);

    fireEvent.change(await screen.findByLabelText("Criar novo sistema"), {
      target: { value: "Gama" },
    });
    fireEvent.click(screen.getByText("Criar"));

    expect(await screen.findByText("Construtor aberto")).toBeTruthy();
  });
});

describe("SeletorSistemas — visibilidade do formulário de criação (RN10)", () => {
  it("usuário dono/parceiro vê o formulário de criação", async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    renderSeletor(client, usuarioDono);
    expect(await screen.findByLabelText("Criar novo sistema")).toBeTruthy();
  });

  it("usuário cliente-final não vê o formulário de criação", async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    renderSeletor(client, usuarioCliente);
    await screen.findByText(/Nenhum sistema ainda/);
    expect(screen.queryByLabelText("Criar novo sistema")).toBeNull();
  });
});
