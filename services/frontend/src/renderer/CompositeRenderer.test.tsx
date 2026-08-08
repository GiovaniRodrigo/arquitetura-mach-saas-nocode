import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { Componente, MapaPermissoes } from "../api/types";
import { CompositeRenderer } from "./CompositeRenderer";

const arvore: Componente = {
  blind_index: "bi-root",
  tipo: "container",
  componente_filhos: [
    { blind_index: "bi-input", tipo: "input_texto", propriedades: { label: "Nome" } },
    { blind_index: "bi-btn", tipo: "botao", propriedades: { label: "Enviar" } },
    { blind_index: "bi-secreto", tipo: "texto", propriedades: { label: "confidencial" } },
  ],
};

describe("CompositeRenderer (RF01/RN03)", () => {
  it("renderiza recursivamente os nós visíveis e omite os sem view", () => {
    const permissoes: MapaPermissoes = {
      "bi-root": { view: true, click: false },
      "bi-input": { view: true, click: true },
      "bi-btn": { view: true, click: true },
      // bi-secreto sem view → não renderizado
    };
    render(<CompositeRenderer no={arvore} permissoes={permissoes} />);

    expect(screen.getByText("Nome")).toBeTruthy();
    expect(screen.getByText("Enviar")).toBeTruthy();
    expect(screen.queryByText("confidencial")).toBeNull();
  });

  it("desabilita componentes sem click (RN03)", () => {
    const permissoes: MapaPermissoes = {
      "bi-root": { view: true, click: false },
      "bi-btn": { view: true, click: false },
    };
    render(<CompositeRenderer no={arvore} permissoes={permissoes} />);
    const botao = screen.getByText("Enviar") as HTMLButtonElement;
    expect(botao.disabled).toBe(true);
  });

  it("dispara onClick apenas quando clicável", () => {
    const onClick = vi.fn();
    const permissoes: MapaPermissoes = {
      "bi-root": { view: true, click: false },
      "bi-btn": { view: true, click: true },
    };
    render(<CompositeRenderer no={arvore} permissoes={permissoes} onClick={onClick} />);
    (screen.getByText("Enviar") as HTMLButtonElement).click();
    expect(onClick).toHaveBeenCalledWith("bi-btn");
  });

  it("raiz sem view não renderiza nada", () => {
    const { container } = render(<CompositeRenderer no={arvore} permissoes={{}} />);
    expect(container.firstChild).toBeNull();
  });
});
