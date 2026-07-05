# Gem Bridge

[Read in English](./README.md)

Gem Bridge é uma bridge local de ferramentas de IA escrita em Go.

O objetivo deste projeto é fornecer um daemon local seguro que exponha ferramentas controladas de filesystem e desenvolvimento para um assistente de IA em uma interface baseada em navegador.

Este projeto é desenhado em torno de três princípios centrais:

- **Execução local-first**: arquivos e comandos são tratados na máquina do usuário.
- **Isolamento de workspace**: as ferramentas ficam restritas a um diretório autorizado do projeto.
- **Protocolo agnóstico de ferramenta**: o daemon deve poder ser usado por diferentes frontends de IA no futuro.

## Funcionalidades atuais

- Resolução segura de paths dentro do workspace
- Listagem de diretórios
- Leitura de arquivos
- Proteção contra paths absolutos e ataques de path traversal
- Respostas JSON consistentes para resultados de execução de ferramentas

## Estrutura do projeto

```text
.
├── cmd/
│   └── gem-bridge/
│       └── main.go
├── internal/
│   ├── security/
│   │   └── workspace.go
│   └── tools/
│       └── files.go
├── go.mod
└── README.md
```

## Uso

Execute o projeto a partir da raiz do workspace.

Listar arquivos no workspace atual:

```bash
go run ./cmd/gem-bridge '{"tool":"listDirectory","path":"."}'
```

Ler um arquivo:

```bash
go run ./cmd/gem-bridge '{"tool":"readFile","path":"go.mod"}'
```

Tentativas de acessar arquivos fora do workspace são bloqueadas:

```bash
go run ./cmd/gem-bridge '{"tool":"readFile","path":"../../.ssh/id_rsa"}'
```

Resposta esperada:

```json
{
  "success": false,
  "error": "access outside the workspace is blocked"
}
```

Quando executada via `go run`, essa requisição com falha também retorna `exit status 1`, o que é um comportamento esperado para a versão CLI atual.

## Modelo de segurança

Gem Bridge trata a raiz do workspace como uma fronteira de segurança.

Todos os paths fornecidos pelo usuário devem ser relativos e são resolvidos pela camada de segurança antes que qualquer operação de filesystem seja executada. Paths absolutos e tentativas de path traversal são bloqueados para impedir acesso a arquivos fora do workspace autorizado.

Exemplos de paths bloqueados:

```text
/etc/passwd
/home/user/.ssh/id_rsa
../../.env
../../.config
```

## Roadmap

- Adicionar raiz de workspace configurável
- Adicionar escrita de arquivos com regras explícitas de segurança
- Adicionar ferramentas Git
- Adicionar transporte WebSocket para desenvolvimento local
- Adicionar integração com extensão de navegador
- Adicionar suporte a Native Messaging
- Adicionar fluxo de aprovação para operações sensíveis
- Adicionar testes automatizados para segurança e comportamento das ferramentas
- Adicionar logging estruturado

## Por que este projeto existe

Assistentes modernos de IA são cada vez mais úteis para desenvolvimento de software, mas interfaces de IA baseadas em navegador normalmente não têm acesso direto e seguro aos arquivos locais de projeto de um desenvolvedor.

Gem Bridge explora uma abordagem local-first em que um assistente de IA pode solicitar ações controladas por meio de um daemon pequeno e auditável rodando na própria máquina do desenvolvedor.

O objetivo de longo prazo é criar uma ponte segura entre ferramentas conversacionais de IA e fluxos locais de desenvolvimento sem expor todo o filesystem nem depender de acesso remoto inseguro.

## Licença

Este projeto está em desenvolvimento ativo. Uma licença será adicionada antes da primeira versão pública.
