import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
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
};

const usuarioCliente: UsuarioAutenticado = {
  nome: "Bruno",
  iniciais: "B",
  podeCriarSistema: false,
};

describe("SeletorSistemas (RN01)", () => {
  it("lista os sistemas do tenant", async () => {
    const sistemas: Sistema[] = [
      { id: "a", nome: "Alfa" },
      { id: "b", nome: "Beta" },
    ];
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue(sistemas) });

    render(<SeletorSistemas client={client} usuario={usuarioDono} />);

    expect(await screen.findByText("Alfa")).toBeTruthy();
    expect(screen.getByText("Beta")).toBeTruthy();
  });

  it("mostra empty state quando não há sistemas", async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    render(<SeletorSistemas client={client} usuario={usuarioDono} />);
    expect(await screen.findByText(/Nenhum sistema ainda/)).toBeTruthy();
  });

  it("mostra erro com retry quando a listagem falha", async () => {
    const listar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, "INTERNAL", "boom"))
      .mockResolvedValueOnce([{ id: "a", nome: "Alfa" }]);
    const client = fakeClient({ listarSistemas: listar });

    render(<SeletorSistemas client={client} usuario={usuarioDono} />);

    const retry = await screen.findByText("Tentar novamente");
    fireEvent.click(retry);

    expect(await screen.findByText("Alfa")).toBeTruthy();
    expect(listar).toHaveBeenCalledTimes(2);
  });
});

describe("SeletorSistemas — visibilidade do formulário de criação (RN10)", () => {
  it("usuário dono/parceiro vê o formulário de criação", async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    render(<SeletorSistemas client={client} usuario={usuarioDono} />);
    expect(await screen.findByLabelText("Criar novo sistema")).toBeTruthy();
  });

  it("usuário cliente-final não vê o formulário de criação", async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    render(<SeletorSistemas client={client} usuario={usuarioCliente} />);
    await screen.findByText(/Nenhum sistema ainda/);
    expect(screen.queryByLabelText("Criar novo sistema")).toBeNull();
  });
});
