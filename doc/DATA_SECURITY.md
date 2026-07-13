# Gestão de Dados, Segurança e Isolamento IAM

[cite_start]A segurança e a integridade dos dados na plataforma são aplicadas em múltiplas camadas para suportar um ecossistema multi-tenant seguro[cite: 65].

## 1. Gestão de Dados Dinâmicos (JSONB)
[cite_start]Como os utilizadores criam seus próprios campos e tabelas, a base de dados utiliza colunas do tipo `JSONB` integradas à mesma tabela física[cite: 32]. 

## 2. Mapeamento Expandido por Blind Index
[cite_start]Para realizar validações e indexações sem comprometer a privacidade ou expor dados brutos[cite: 33]:
* [cite_start]As definições dos campos dinâmicos usam um modelo baseado no **Blind Index** (índice cego criptográfico) dos componentes como chave[cite: 33].
* [cite_start]Essa estrutura define o tipo, obrigatoriedade e limites (como valor mínimo ou tamanho máximo de string) garantindo validações seguras e anónimas[cite: 33, 319].

## 3. Segurança e Propagação de Identidade (IAM)
* [cite_start]**Tokens JWT e gRPC Metadata:** O token JWT interceptado pelo API Gateway em Go é extraído e injetado nativamente como **Metadados (Metadata) binários do gRPC**[cite: 66, 67]. [cite_start]Isso propaga o contexto de identidade e o `tenant_id` de forma segura para todos os microsserviços internos[cite: 67].
* [cite_start]**Avaliação de Condições no Servidor:** Para evitar fraudes visuais que burlem o código client-side, o *IAM Service* processa todas as condições dinâmicas estritamente no back-end[cite: 68]. [cite_start]O servidor avalia as regras complexas e devolve ao front-end apenas um mapa simples de booleanos indexado pelo *Blind Index* dos componentes afetados[cite: 69].

### Exemplo de Payload de Permissões enviado ao Headless Player
```json
{  
  "permissions": {    
    "8f3b2a1...": { "view": true, "click": false },    
    "4a9e2d3...": { "view": false, "click": false }  
  }
}
```

## 4. Isolamento de Dados na Base
[cite_start]O microsserviço de base de dados aplica cláusulas automáticas de filtragem (`WHERE tenant_id = :id`) baseadas no contexto extraído do gRPC, prevenindo nativamente o vazamento de dados entre clientes[cite: 69].