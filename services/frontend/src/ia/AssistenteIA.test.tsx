import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import "@testing-library/jest-dom";
import type { ApiClient } from "../api/client";
import { AssistenteIA } from "./AssistenteIA";

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

afterEach(() => {
  sessionStorage.clear();
});

describe("AssistenteIA", () => {
  it("começa fechado, com apenas o botão flutuante visível", () => {
    render(<AssistenteIA client={fakeClient({})} />);
    expect(screen.getByRole("button", { name: /abrir assistente de design/i })).toBeInTheDocument();
    expect(screen.queryByText("Assistente de Design")).not.toBeInTheDocument();
  });

  it("abre o painel com chips de sugestão em estado vazio", () => {
    render(<AssistenteIA client={fakeClient({})} />);
    fireEvent.click(screen.getByRole("button", { name: /abrir assistente de design/i }));

    expect(screen.getByText("Assistente de Design")).toBeInTheDocument();
    expect(screen.getByText("Modelar multi-tenancy para o meu sistema")).toBeInTheDocument();
  });

  it("exibe a pill do sistema em foco quando informado", () => {
    render(<AssistenteIA client={fakeClient({})} sistemaNome="Clínica Fácil" />);
    fireEvent.click(screen.getByRole("button", { name: /abrir assistente de design/i }));
    expect(screen.getByText(/Clínica Fácil/)).toBeInTheDocument();
  });

  it("envia uma sugestão clicada e mostra a resposta do assistente", async () => {
    const enviarMensagemChatIA = vi.fn().mockResolvedValue("recomendação de arquitetura");
    render(<AssistenteIA client={fakeClient({ enviarMensagemChatIA })} />);
    fireEvent.click(screen.getByRole("button", { name: /abrir assistente de design/i }));
    fireEvent.click(screen.getByText("Modelar multi-tenancy para o meu sistema"));

    expect(await screen.findByText("Modelar multi-tenancy para o meu sistema")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText("recomendação de arquitetura")).toBeInTheDocument());
  });

  it("envia a mensagem digitada ao pressionar Enter e mostra erro amigável em falha", async () => {
    const enviarMensagemChatIA = vi.fn().mockRejectedValue(new Error("indisponível"));
    render(<AssistenteIA client={fakeClient({ enviarMensagemChatIA })} />);
    fireEvent.click(screen.getByRole("button", { name: /abrir assistente de design/i }));

    const campo = screen.getByPlaceholderText(/descreva o foco do seu sistema/i);
    fireEvent.change(campo, { target: { value: "por onde eu começo?" } });
    fireEvent.keyDown(campo, { key: "Enter" });

    await waitFor(() =>
      expect(screen.getByText(/não consegui responder agora/i)).toBeInTheDocument(),
    );
    expect(enviarMensagemChatIA).toHaveBeenCalledWith(
      [{ papel: "usuario", conteudo: "por onde eu começo?" }],
      undefined,
    );
  });
});
