# Mapeamento dos Pilares MACH & Componentes Core

[cite_start]A arquitetura do sistema é totalmente baseada no acrónimo **MACH**: *Microservices, API-first, Cloud-native SaaS e Headless*[cite: 4].

## 1. 🧩 M – Microsserviços (Microservices)
[cite_start]O motor da plataforma é dividido em serviços especializados, independentes e com ciclos de vida isolados[cite: 16]:

* [cite_start]**Design Engine (UI):** Responsável pelo CRUD que guarda as definições e metadados da interface criada pelo utilizador, estruturado em formato de árvore recursiva[cite: 17].
* [cite_start]**Logic Engine (Regras):** Armazena e interpreta as regras de negócio baseadas em nós lógicos (árvores de decisão)[cite: 18].
* [cite_start]**IAM Service (Identity & Access Management):** Centraliza o controlo de acessos, autenticação, níveis de permissão e isolamento de Tenants[cite: 19].
* [cite_start]**Deploy Engine:** Controla os estados de publicação dos sistemas dos clientes, gerindo o versionamento por flags de status e orquestrando o provisionamento de recursos dinâmicos[cite: 20].
* [cite_start]**Export Engine:** Serviço isolado responsável por coordenar a recolha de grandes volumes de dados para exportação de pacotes completos de forma assíncrona[cite: 21].

## 2. 🔌 A – API-first & Comunicação Interna
* [cite_start]**Abordagem de Contratos:** Toda a comunicação interna é definida antes do desenvolvimento do código de negócio[cite: 148].
* [cite_start]**Protocolo gRPC:** A comunicação entre os microsserviços internos ocorre via gRPC sobre HTTP/2 com Protocol Buffers, garantindo alta performance e tipagem forte[cite: 25, 198, 199].

## 3. ☁️ C – Cloud-native SaaS & Eficiência de Custos
* [cite_start]**Elasticidade da Nuvem:** O sistema aproveita a escalabilidade sob demanda da infraestrutura em nuvem para mitigar gargalos[cite: 30, 154].
* [cite_start]**Shared Database Multi-tenancy:** Para otimizar o uso de RAM e CPU, múltiplos clientes compartilham a mesma instância de base de dados[cite: 31]. [cite_start]O isolamento é garantido logicamente pela coluna `tenant_id`[cite: 32].

## 4. 👤 H – Headless & Motor de Renderização (Headless Player)
[cite_start]A interface gráfica visualizada pelo utilizador final é completamente desacoplada do back-end[cite: 35]:
* [cite_start]**Estrutura de Árvore Recursiva:** O front-end atua como um renderizador inteligente universal que consome definições estruturais brutas (JSON) baseadas no padrão *Composite* (propriedade `componente_filhos`)[cite: 36, 37].
* [cite_start]**Navegação Dinâmica:** A alternância entre ecrãs ocorre no modelo SPA (Single Page Application) através de ações de `redirect` associadas a rotas dinâmicas[cite: 38].