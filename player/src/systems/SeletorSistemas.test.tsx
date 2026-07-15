import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import type { ApiClient } from "../api/client";
import { ApiError } from "../api/client";
import type { Sistema } from "../api/types";
import { SeletorSistemas } from "./SeletorSistemas";

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

describe("SeletorSistemas (RN01)", () => {
  it("lista os sistemas do tenant", async () => {
    const sistemas: Sistema[] = [
      { id: "a", nome: "Alfa" },
      { id: "b", nome: "Beta" },
    ];
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue(sistemas) });

    render(<SeletorSistemas client={client} />);

    expect(await screen.findByText("Alfa")).toBeTruthy();
    expect(screen.getByText("Beta")).toBeTruthy();
  });

  it("mostra empty state quando não há sistemas", async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    render(<SeletorSistemas client={client} />);
    expect(await screen.findByText(/Nenhum sistema ainda/)).toBeTruthy();
  });

  it("mostra erro com retry quando a listagem falha", async () => {
    const listar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, "INTERNAL", "boom"))
      .mockResolvedValueOnce([{ id: "a", nome: "Alfa" }]);
    const client = fakeClient({ listarSistemas: listar });

    render(<SeletorSistemas client={client} />);

    const retry = await screen.findByText("Tentar novamente");
    fireEvent.click(retry);

    expect(await screen.findByText("Alfa")).toBeTruthy();
    expect(listar).toHaveBeenCalledTimes(2);
  });
});
