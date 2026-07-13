# Construtor de Sistemas MACH V4 — Plataforma Low-Code / No-Code

## 1. Visão Geral
[cite_start]Este repositório contém a especificação técnica de arquitetura unificada para a plataforma **MACH V4**. [cite_start]O sistema foi projetado sob os princípios modernos de arquitetura distribuída para permitir que utilizadores criem as suas próprias aplicações digitais através de uma interface inteiramente visual[cite: 4, 6].

[cite_start]A versão V4 consolida os contratos gRPC via Protocol Buffers, o mapeamento de payloads dinâmicos por mapas chave-valor, o controlo de acessos refinado por componente (IAM) e integra camadas de mensajaria assíncrona escalável e sincronização em tempo real[cite: 5].

## 2. Requisitos Principais da Plataforma
* [cite_start]**CRUD de UI:** Criação e salvamento dos metadados de construção de interfaces visuais de utilizador[cite: 8].
* [cite_start]**CRUD de Regras de Negócio:** Registo de regras operacionais, controlo de acessos e requisitos funcionais[cite: 9].
* [cite_start]**Publicação Instantânea:** Mecanismo de deploy imediato baseado em uma abordagem interpretada[cite: 10].
* [cite_start]**Multi-tenancy Hierárquico:** Isolamento lógico estruturado para Donos, Parceiros e Clientes Finais[cite: 11].
* [cite_start]**Exportação Assíncrona:** Extração completa de dados operacionais e metadados em segundo plano[cite: 12].
* [cite_start]**Colaboração em Tempo Real:** Edição simultânea de múltiplos utilizadores no painel de construção[cite: 13].

## 3. Guia de Leitura da Documentação
Para entender detalhadamente cada componente da infraestrutura, consulte os arquivos especializados:
* [`ARCHITECTURE_PILLARS.md`](./ARCHITECTURE_PILLARS.md): Mapeamento dos 4 pilares MACH e divisão dos Microsserviços.
* [`GATEWAY_COLLABORATION.md`](./GATEWAY_COLLABORATION.md): Detalhes do Gateway híbrido (Go & Elixir) e sincronização via WebSockets.
* [`DATA_SECURITY.md`](./DATA_SECURITY.md): Gestão de dados dinâmicos, Blind Index e políticas de IAM por componente.
* [`ASYNC_OBSERVABILITY.md`](./ASYNC_OBSERVABILITY.md): Mensageria distribuída com KEDA e Rastreamento com OpenTelemetry.
* [`CONTRACTS_PERFORMANCE.md`](./CONTRACTS_PERFORMANCE.md): Contratos `.proto` oficiais, estratégias de renderização e plano de evolução técnica.