# Camada de Gateway Híbrida e Colaboração em Tempo Real

[cite_start]Para maximizar a eficiência de rede, o padrão BFF (Backend For Frontend) foi dividido em duas tecnologias distintas por tipo de protocolo e caso de uso[cite: 23].

## 1. API Gateway em Go (Golang)
[cite_start]Responsável por gerenciar todo o fluxo tradicional síncrono de *Request-Response* (HTTP/REST)[cite: 24].
* [cite_start]**Atribuições:** Atua como um proxy ultra-rápido de baixa latência[cite: 25].
* [cite_start]**Segurança e Tráfego:** Valida o token JWT no cabeçalho `Authorization` e aplica políticas de *Rate Limiting*[cite: 25].
* [cite_start]**Tradução de Protocolos:** Traduz as requisições HTTP recebidas do navegador para chamadas gRPC internas direcionadas aos microsserviços[cite: 25].

## 2. Motor de Colaboração em Elixir (Phoenix Channels)
[cite_start]Responsável estritamente pelas conexões persistentes bidirecionais em tempo real via WebSockets[cite: 26, 77].

### Gerenciamento de Estado em Memória (BEAM)
* [cite_start]**Concorrência Isolada:** Para cada ecrã sob edição ativa no painel de construção, a Erlang VM (BEAM) instancia um processo leve isolado (`GenServer`), mantendo a árvore de componentes em memória viva[cite: 27, 78].
* [cite_start]**Replicação e Sincronização:** As mutações leves enviadas por um usuário são processadas pelo `GenServer`, replicadas em *snapshots* de segurança numa instância global do Redis e propagadas instantaneamente via *broadcast* para os demais cocriadores[cite: 79].

### Estratégia de Persistência Baseada em Debounce (Write-Behind)
[cite_start]Para blindar a base de dados relacional contra exaustão de escrita induzida por micro-movimentos na UI, adota-se o seguinte fluxo[cite: 81]:
1. [cite_start]As alterações acumulam-se temporariamente apenas na memória do processo BEAM e no Redis[cite: 82].
2. [cite_start]Quando o `GenServer` detecta um período de inatividade de rede (ex: 5 segundos de silêncio), ele consolida a árvore recursiva num único payload estável[cite: 83].
3. [cite_start]O processo Elixir dispara uma **única chamada gRPC otimizada em lote** para o *Design Engine*, persistindo os dados na coluna `JSONB` da base de dados relacional[cite: 84].

### Controle de Presença e Conflitos
* [cite_start]**Phoenix Presence:** O rastreamento de utilizadores online e renderização de cursores simultâneos opera de forma descentralizada via CRDTs (*Conflict-Free Replicated Data Types*)[cite: 86].
* [cite_start]**Bloqueios Otimistas:** Conflitos de edição direta no mesmo elemento são prevenidos através de bloqueios temporários por *Blind Index*, desativando os inputs dinamicamente nos navegadores concorrentes[cite: 87].