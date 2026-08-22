import { afterEach, describe, expect, it, vi } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import type { ApiClient } from "../api/client";
import { useAssistenteChat } from "./useAssistenteChat";

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

afterEach(() => {
  sessionStorage.clear();
});

describe("useAssistenteChat", () => {
  it("envia a mensagem do usuário e anexa a resposta do assistente", async () => {
    const enviarMensagemChatIA = vi.fn().mockResolvedValue("recomendação de arquitetura");
    const client = fakeClient({ enviarMensagemChatIA });

    const { result } = renderHook(() => useAssistenteChat(client, "Clínica Fácil"));

    await act(async () => {
      await result.current.enviar("como isolar dados entre clínicas?");
    });

    expect(result.current.historico).toEqual([
      { papel: "usuario", conteudo: "como isolar dados entre clínicas?" },
      { papel: "assistente", conteudo: "recomendação de arquitetura" },
    ]);
    expect(enviarMensagemChatIA).toHaveBeenCalledWith(
      [{ papel: "usuario", conteudo: "como isolar dados entre clínicas?" }],
      "Clínica Fácil",
    );
    expect(result.current.erro).toBeNull();
  });

  it("mantém a pergunta do usuário e expõe erro amigável quando a chamada falha", async () => {
    const enviarMensagemChatIA = vi.fn().mockRejectedValue(new Error("boom"));
    const client = fakeClient({ enviarMensagemChatIA });

    const { result } = renderHook(() => useAssistenteChat(client));
    await act(async () => {
      await result.current.enviar("oi");
    });

    await waitFor(() => expect(result.current.erro).not.toBeNull());
    expect(result.current.historico).toEqual([{ papel: "usuario", conteudo: "oi" }]);
  });

  it("ignora envio de mensagem vazia", async () => {
    const enviarMensagemChatIA = vi.fn();
    const client = fakeClient({ enviarMensagemChatIA });

    const { result } = renderHook(() => useAssistenteChat(client));
    await act(async () => {
      await result.current.enviar("   ");
    });

    expect(enviarMensagemChatIA).not.toHaveBeenCalled();
    expect(result.current.historico).toEqual([]);
  });

  it("limpar zera o histórico e o erro", async () => {
    const enviarMensagemChatIA = vi.fn().mockRejectedValue(new Error("boom"));
    const client = fakeClient({ enviarMensagemChatIA });

    const { result } = renderHook(() => useAssistenteChat(client));
    await act(async () => {
      await result.current.enviar("oi");
    });
    await waitFor(() => expect(result.current.erro).not.toBeNull());

    act(() => result.current.limpar());
    expect(result.current.historico).toEqual([]);
    expect(result.current.erro).toBeNull();
  });
});
