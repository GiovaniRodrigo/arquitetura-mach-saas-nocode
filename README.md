# MACH V4 System Builder — Plataforma Low-Code / No-Code

## 1. Visão Geral

Este repositório contém a especificação de arquitetura técnica unificada da plataforma **MACH V4**. O sistema foi desenhado sob princípios modernos de arquitetura distribuída para permitir que os utilizadores construam as suas próprias aplicações digitais através de uma interface totalmente visual.

A versão V4 consolida contratos gRPC via Protocol Buffers, mapeamento dinâmico de payloads usando mapas chave-valor, controlo de acesso granular ao nível de componente (IAM) e integra camadas de mensageria assíncrona escalável com sincronização em tempo real.

## 2. Requisitos Centrais da Plataforma

* **CRUD de UI:** Criação e gravação de metadados para a construção de interfaces visuais.
* **CRUD de Regras de Negócio:** Registo de regras operacionais, controlo de acessos e requisitos funcionais.
* **Publicação Instantânea:** Mecanismo de deploy imediato baseado numa abordagem interpretada.
* **Multi-Tenancy Hierárquico:** Isolamento lógico estruturado para Owners, Partners e Clientes Finais.
* **Exportação Assíncrona:** Extração completa em segundo plano de dados operacionais e metadados.
* **Colaboração em Tempo Real:** Edição simultânea multi-utilizador dentro do canvas do builder.

## 3. Guia de Leitura da Documentação

Para obter uma compreensão detalhada de cada componente da infraestrutura, consulte os ficheiros especializados:

* [`ARCHITECTURE_PILLARS.md`](doc/ARCHITECTURE_PILLARS.md): Mapeamento dos 4 pilares MACH e decomposição em microsserviços.
* [`GATEWAY_COLLABORATION.md`](doc/GATEWAY_COLLABORATION.md): Detalhes do Gateway híbrido (Go & Elixir) e sincronização via WebSocket.
* [`DATA_SECURITY.md`](doc/DATA_SECURITY.md): Gestão dinâmica de dados, Blind Index e políticas de IAM ao nível de componente.
* [`ASYNC_OBSERVABILITY.md`](doc/ASYNC_OBSERVABILITY.md): Mensageria distribuída com KEDA e tracing via OpenTelemetry.
* [`CONTRACTS_PERFORMANCE.md`](doc/CONTRACTS_PERFORMANCE.md): Contratos `.proto` oficiais, estratégias de renderização e roadmap de evolução técnica.
