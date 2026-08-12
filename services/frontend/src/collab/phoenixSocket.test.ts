import { describe, expect, it, vi } from "vitest";
import { CollabClient, type ChannelLike, type SocketFactory, type SocketLike } from "./phoenixSocket";

// Fake do Socket/Channel Phoenix: captura handlers e pushes, e simula o join.
function fakeFactory(joinResp: unknown = {}) {
  const handlers: Record<string, (p: any) => void> = {};
  const pushes: { evento: string; payload: object }[] = [];
  let joinStatus: "ok" | "error" = "ok";
  let paramsRecebidos: object | undefined;

  const channel: ChannelLike = {
    join() {
      const push = {
        receive(status: string, cb: (r?: unknown) => void) {
          if (status === joinStatus) cb(joinResp);
          return push;
        },
      };
      return push;
    },
    on(event, cb) {
      handlers[event] = cb;
    },
    push(evento, payload) {
      pushes.push({ evento, payload });
      return { receive: () => ({}) } as any;
    },
    leave() {
      return { receive: () => ({}) } as any;
    },
  };

  const socket: SocketLike = {
    connect: vi.fn(),
    disconnect: vi.fn(),
    channel: (_topic, params) => {
      paramsRecebidos = params;
      return channel;
    },
  };

  const factory: SocketFactory = () => socket;
  return {
    factory,
    socket,
    emitir: (event: string, payload: unknown) => handlers[event]?.(payload),
    pushes,
    setJoinStatus: (s: "ok" | "error") => (joinStatus = s),
    paramsRecebidos: () => paramsRecebidos,
  };
}

describe("CollabClient (RF06)", () => {
  it("conecta e resolve ao juntar-se com sucesso", async () => {
    const f = fakeFactory();
    const c = new CollabClient("ws://x", "tok", f.factory);
    await expect(c.entrar("s1", {})).resolves.toBeUndefined();
    expect(f.socket.connect).toHaveBeenCalled();
  });

  it("encaminha mutações recebidas ao handler", async () => {
    const f = fakeFactory();
    const c = new CollabClient("ws://x", "tok", f.factory);
    const onMutation = vi.fn();
    await c.entrar("s1", { onMutation });

    f.emitir("mutation", { mutation: { tipo: "update_props" }, from: "userA" });
    expect(onMutation).toHaveBeenCalledWith({ tipo: "update_props" }, "userA");
  });

  it("mutar/bloquear enviam os eventos corretos", async () => {
    const f = fakeFactory();
    const c = new CollabClient("ws://x", "tok", f.factory);
    await c.entrar("s1", {});

    c.mutar({ tipo: "remove", blind_index: "bi-a" });
    c.bloquear("bi-a");

    expect(f.pushes[0]).toEqual({ evento: "mutate", payload: { mutation: { tipo: "remove", blind_index: "bi-a" } } });
    expect(f.pushes[1]).toEqual({ evento: "lock", payload: { blind_index: "bi-a" } });
  });

  it("repassa sistema_id/design_id/nome como params do join (semeia o ScreenServer)", async () => {
    const f = fakeFactory();
    const c = new CollabClient("ws://x", "tok", f.factory);
    await c.entrar("s1", {}, { sistema_id: "sis-1", design_id: "s1", nome: "Home" });

    expect(f.paramsRecebidos()).toEqual({ sistema_id: "sis-1", design_id: "s1", nome: "Home" });
  });

  it("onJoin recebe a árvore devolvida pelo join antes de resolver", async () => {
    const arvore = { blind_index: "root", tipo: "tela" };
    const f = fakeFactory({ tree: arvore });
    const c = new CollabClient("ws://x", "tok", f.factory);
    const onJoin = vi.fn();

    await c.entrar("s1", { onJoin });
    expect(onJoin).toHaveBeenCalledWith(arvore);
  });

  it("rejeita quando o join falha", async () => {
    const f = fakeFactory();
    f.setJoinStatus("error");
    const c = new CollabClient("ws://x", "tok", f.factory);
    await expect(c.entrar("s1", {})).rejects.toBeDefined();
  });
});
