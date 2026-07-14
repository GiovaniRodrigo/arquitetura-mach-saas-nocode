// Cliente WebSocket Phoenix para a colaboração no modo builder-preview (RF06):
// junta-se ao canal do ecrã, recebe mutações/locks/presença dos outros
// utilizadores e envia as próprias mutações. O JWT autentica no handshake.
//
// A dependência do Socket é injetável (SocketFactory) para permitir testes sem uma
// ligação WebSocket real.

import { Socket } from "phoenix";

export interface PushLike {
  receive(status: string, cb: (resp?: unknown) => void): PushLike;
}

export interface ChannelLike {
  join(): PushLike;
  on(event: string, cb: (payload: any) => void): unknown;
  push(event: string, payload: object): PushLike;
  leave(): PushLike;
}

export interface SocketLike {
  connect(): void;
  disconnect(cb?: () => void): void;
  channel(topic: string, params?: object): ChannelLike;
}

export type SocketFactory = (url: string, opts: { params: { token: string } }) => SocketLike;

export interface CollabHandlers {
  onMutation?: (mutation: unknown, from: string) => void;
  onLock?: (blindIndex: string, userId: string) => void;
  onUnlock?: (blindIndex: string) => void;
  onPresence?: (state: unknown) => void;
}

const factoryPadrao: SocketFactory = (url, opts) => new Socket(url, opts) as unknown as SocketLike;

export class CollabClient {
  private readonly socket: SocketLike;
  private channel: ChannelLike | null = null;

  constructor(url: string, token: string, factory: SocketFactory = factoryPadrao) {
    this.socket = factory(url, { params: { token } });
  }

  /** Conecta e junta-se ao canal do ecrã, ligando os handlers. */
  entrar(screenId: string, handlers: CollabHandlers = {}): Promise<void> {
    this.socket.connect();
    const ch = this.socket.channel(`screen:${screenId}`, {});

    ch.on("mutation", (p) => handlers.onMutation?.(p.mutation, p.from));
    ch.on("lock", (p) => handlers.onLock?.(p.blind_index, p.user_id));
    ch.on("unlock", (p) => handlers.onUnlock?.(p.blind_index));
    ch.on("presence_state", (p) => handlers.onPresence?.(p));

    this.channel = ch;

    return new Promise((resolve, reject) => {
      ch.join()
        .receive("ok", () => resolve())
        .receive("error", (e) => reject(e))
        .receive("timeout", () => reject(new Error("timeout ao juntar-se ao canal")));
    });
  }

  /** Envia uma mutação da árvore ao servidor de colaboração. */
  mutar(mutation: unknown): void {
    this.channel?.push("mutate", { mutation });
  }

  /** Adquire o bloqueio otimista de um componente (RN07). */
  bloquear(blindIndex: string): void {
    this.channel?.push("lock", { blind_index: blindIndex });
  }

  /** Libera o bloqueio de um componente. */
  desbloquear(blindIndex: string): void {
    this.channel?.push("unlock", { blind_index: blindIndex });
  }

  /** Atualiza a posição do cursor (presença). */
  cursor(x: number, y: number): void {
    this.channel?.push("cursor", { x, y });
  }

  /** Sai do canal e desconecta o socket. */
  sair(): void {
    this.channel?.leave();
    this.channel = null;
    this.socket.disconnect();
  }
}
