import { describe, expect, it, vi } from "vitest";
import { CollabClient, type ChannelLike, type SocketFactory, type SocketLike } from "./phoenixSocket";

// Fake do Socket/Channel Phoenix: captura handlers e pushes, e simula o join.
function fakeFactory() {
  const handlers: Record<string, (p: any) => void> = {};
  const pushes: { evento: string; payload: object }[] = [];
  let joinStatus: "ok" | "error" = "ok";

  const channel: ChannelLike = {
    join() {
      const push = {
        receive(status: string, cb: (r?: unknown) => void) {
          if (status === joinStatus) cb({});
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
    channel: () => channel,
  };

  const factory: SocketFactory = () => socket;
  return {
    factory,
    socket,
    emitir: (event: string, payload: unknown) => handlers[event]?.(payload),
    pushes,
    setJoinStatus: (s: "ok" | "error") => (joinStatus = s),
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

  it("rejeita quando o join falha", async () => {
    const f = fakeFactory();
    f.setJoinStatus("error");
    const c = new CollabClient("ws://x", "tok", f.factory);
    await expect(c.entrar("s1", {})).rejects.toBeDefined();
  });
});
