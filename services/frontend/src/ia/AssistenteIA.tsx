// Assistente de Design (chat de IA/RAG): gatilho global (FAB) + painel
// docked (Sheet), montado uma única vez em DashboardLayout — disponível em
// qualquer aba do Dashboard, incluindo o editor Canvas de viewport fixo
// (spec chat-ia-rag). Ver .claude/planos/chat-ia-rag/ui/ para a análise de
// UI que fundamenta o posicionamento e os estados abaixo.

import { useEffect, useRef, useState } from "react";
import { Send, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import type { ApiClient } from "../api/client";
import { useAssistenteChat } from "./useAssistenteChat";

const SUGESTOES = [
  "Modelar multi-tenancy para o meu sistema",
  "Revisar as regras de negócio do sistema atual",
  "Sugerir a estrutura de telas para começar",
];

export interface AssistenteIAProps {
  client: ApiClient;
  /** Nome do sistema em foco na rota atual, quando houver (herdado automaticamente). */
  sistemaNome?: string;
}

export function AssistenteIA({ client, sistemaNome }: AssistenteIAProps) {
  const [aberto, setAberto] = useState(false);
  const [rascunho, setRascunho] = useState("");
  const { historico, enviando, erro, enviar } = useAssistenteChat(client, sistemaNome);
  const fimDaListaRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fimDaListaRef.current?.scrollIntoView?.({ block: "end" });
  }, [historico, enviando, aberto]);

  async function enviarRascunho(texto: string) {
    setRascunho("");
    await enviar(texto);
  }

  function aoSubmeter(e: React.FormEvent) {
    e.preventDefault();
    void enviarRascunho(rascunho);
  }

  function aoTeclar(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      void enviarRascunho(rascunho);
    }
  }

  return (
    <>
      <Button
        type="button"
        onClick={() => setAberto(true)}
        aria-label="Abrir assistente de design (IA)"
        title="Assistente de Design"
        className="fixed right-6 bottom-6 z-40 size-14 rounded-full shadow-lg [&_svg]:size-6"
      >
        <Sparkles />
      </Button>

      <Sheet open={aberto} onOpenChange={setAberto}>
        <SheetContent side="right" className="w-full sm:max-w-md p-0">
          <SheetHeader className="border-b border-border flex-row items-center gap-2.5 space-y-0">
            <span className="flex size-8 items-center justify-center rounded-full bg-primary/10 text-primary">
              <Sparkles className="size-4" />
            </span>
            <SheetTitle>Assistente de Design</SheetTitle>
          </SheetHeader>

          {sistemaNome && (
            <span className="mx-4 -mt-2 w-fit rounded-full bg-accent/10 px-3 py-1 text-xs font-medium text-accent">
              ● Sistema atual: {sistemaNome}
            </span>
          )}

          <div className="flex-1 min-h-0 overflow-y-auto px-4 py-2 flex flex-col gap-3">
            {historico.length === 0 && (
              <div className="text-center px-2 py-6">
                <p className="text-sm text-muted-foreground mb-4">
                  Descreva o foco do seu sistema ou peça uma recomendação de arquitetura.
                </p>
                <div className="flex flex-wrap justify-center gap-2">
                  {SUGESTOES.map((sugestao) => (
                    <button
                      key={sugestao}
                      type="button"
                      onClick={() => void enviarRascunho(sugestao)}
                      className="rounded-full border border-border bg-card px-3.5 py-2 text-xs text-foreground hover:bg-secondary transition-colors"
                    >
                      {sugestao}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {historico.map((msg, i) => (
              <div
                key={i}
                className={`max-w-[85%] rounded-xl px-3.5 py-2.5 text-sm leading-relaxed whitespace-pre-wrap ${
                  msg.papel === "usuario"
                    ? "self-end bg-primary/10 rounded-br-sm"
                    : "self-start bg-secondary rounded-bl-sm"
                }`}
              >
                {msg.conteudo}
              </div>
            ))}

            {enviando && (
              <div
                role="status"
                aria-label="Assistente digitando"
                className="self-start bg-secondary rounded-xl rounded-bl-sm px-3.5 py-3 flex flex-col gap-1.5 w-40"
              >
                <span className="h-2 rounded bg-muted-foreground/20 animate-pulse w-[90%]" />
                <span className="h-2 rounded bg-muted-foreground/20 animate-pulse w-[60%]" />
                <span className="h-2 rounded bg-muted-foreground/20 animate-pulse w-[40%]" />
              </div>
            )}

            {erro && (
              <div className="self-start rounded-xl rounded-bl-sm bg-destructive/10 text-destructive px-3.5 py-2.5 text-sm">
                {erro}
              </div>
            )}

            <div ref={fimDaListaRef} />
          </div>

          <form onSubmit={aoSubmeter} className="flex-shrink-0 border-t border-border p-3 flex items-end gap-2">
            <textarea
              value={rascunho}
              onChange={(e) => setRascunho(e.target.value)}
              onKeyDown={aoTeclar}
              placeholder="Descreva o foco do seu sistema ou pergunte sobre a arquitetura…"
              rows={1}
              className="flex-1 resize-none rounded-lg border border-border bg-background px-3 py-2.5 text-sm min-h-10 max-h-32 focus:outline-none focus:ring-2 focus:ring-ring/50"
            />
            <Button
              type="submit"
              size="icon"
              disabled={enviando || rascunho.trim() === ""}
              aria-label="Enviar mensagem"
              className="shrink-0"
            >
              <Send className="size-4" />
            </Button>
          </form>
        </SheetContent>
      </Sheet>
    </>
  );
}
