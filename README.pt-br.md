# Gem Bridge

[Read in English](./README.md)

Gem Bridge é uma ponte local para ferramentas de IA escrita em Go.

O objetivo deste projeto é fornecer um daemon local seguro que exponha ferramentas controladas de filesystem, Git e desenvolvimento para um assistente de IA executado em uma interface baseada em navegador.

Este projeto é desenhado em torno de três princípios centrais:

- **Execução local-first**: arquivos e comandos são manipulados na máquina do usuário.
- **Isolamento de workspace**: as ferramentas ficam restritas a um diretório de projeto autorizado.
- **Protocolo agnóstico de ferramenta**: o daemon deve poder ser usado por diferentes frontends de IA no futuro.

## Funcionalidades atuais

- Raiz de workspace configurável por `--workspace`
- Resolução segura de paths dentro do workspace
- Listagem de diretórios
- Leitura de arquivos
- Criação segura de arquivos por `writeFile`
- Inspeção segura de status Git por `gitStatus`
- Proteção contra paths vazios, paths absolutos, path traversal, paths Windows com unidade, paths UNC e escapes por symlink
- Comportamento conservador de escrita que recusa sobrescrever arquivos existentes
- Limite de tamanho para escrita de arquivos
- Respostas JSON consistentes para resultados de execução de ferramentas
- CI cross-platform em Linux, macOS e Windows

## Documentação

- [Arquitetura](./docs/ARCHITECTURE.pt-br.md)
- [Modelo de Segurança](./docs/SECURITY_MODEL.pt-br.md)
- [Contexto do Projeto](./docs/PROJECT_CONTEXT.md)

As versões em inglês da documentação pública também estão disponíveis:

- [README.md](./README.md)
- [ARCHITECTURE.md](./docs/ARCHITECTURE.md)
- [SECURITY_MODEL.md](./docs/SECURITY_MODEL.md)

## Estrutura do projeto

```text
.
├── .github/
│   └── workflows/
│       └── ci.yml
├── cmd/
│   └── gem-bridge/
│       └── main.go
├── docs/
│   ├── ARCHITECTURE.md
│   ├── ARCHITECTURE.pt-br.md
│   ├── PROJECT_CONTEXT.md
│   ├── SECURITY_MODEL.md
│   └── SECURITY_MODEL.pt-br.md
├── internal/
│   ├── security/
│   │   ├── workspace.go
│   │   └── workspace_test.go
│   └── tools/
│       ├── files.go
│       ├── files_test.go
│       ├── git.go
│       └── git_test.go
├── go.mod
├── README.md
└── README.pt-br.md
```

## Uso

Execute o projeto a partir da raiz do workspace.

Listar arquivos no workspace atual:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"listDirectory","path":"."}'
```

Ler um arquivo:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"readFile","path":"go.mod"}'
```

Criar um novo arquivo:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"writeFile","path":"notes.txt","content":"hello from gem bridge"}'
```

Ler o arquivo criado:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"readFile","path":"notes.txt"}'
```

Inspecionar status Git:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"gitStatus"}'
```

Resposta esperada para um repositório limpo:

```json
{
  "success": true,
  "data": []
}
```

Tentativas de sobrescrever um arquivo existente são bloqueadas na versão atual:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"writeFile","path":"notes.txt","content":"overwrite"}'
```

Resposta esperada:

```json
{
  "success": false,
  "error": "file already exists"
}
```

Tentativas de acessar arquivos fora do workspace são bloqueadas:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"readFile","path":"../../.ssh/id_rsa"}'
```

Resposta esperada:

```json
{
  "success": false,
  "error": "access outside the workspace is blocked"
}
```

Quando executadas via `go run`, requisições com falha também retornam `exit status 1`, o que é um comportamento esperado para a versão CLI atual.

## Modelo de segurança

Gem Bridge trata a raiz do workspace como uma fronteira de segurança.

Todos os paths fornecidos pelo usuário devem ser relativos e são resolvidos pela camada de segurança antes que qualquer operação de filesystem seja executada. Paths vazios, paths absolutos, tentativas de path traversal, paths Windows com unidade, paths UNC e escapes por symlink são bloqueados para impedir acesso fora do workspace autorizado.

Exemplos de paths bloqueados:

```text
/etc/passwd
/home/user/.ssh/id_rsa
../../.env
../../.config
C:\Users\user\.ssh\id_rsa
C:/Users/user/.ssh/id_rsa
\\server\share\secret.txt
```

A implementação atual de `writeFile` é intencionalmente conservadora. Ela cria apenas arquivos de texto novos, recusa sobrescrever arquivos existentes, limita o tamanho do conteúdo e resolve paths pela camada compartilhada de segurança do workspace antes de escrever no disco.

A implementação atual de `gitStatus` é intencionalmente restrita. Ela executa apenas uma operação Git explícita e permitida dentro do workspace autorizado. O Gem Bridge não expõe execução arbitrária de shell.

## Integração contínua

O repositório inclui um workflow do GitHub Actions que roda em pushes e pull requests para a `main`.

A CI valida:

```bash
go fmt ./...
go test ./...
```

em:

```text
ubuntu-latest
macos-latest
windows-latest
```

## Roadmap

- Adicionar ferramentas Git:
  - `gitDiff`
- Adicionar pacotes estruturados para request e response
- Adicionar transporte WebSocket para desenvolvimento local
- Adicionar integração com extensão de navegador
- Adicionar suporte a Native Messaging
- Adicionar fluxo de aprovação para operações sensíveis
- Adicionar logging estruturado
- Expandir regras de mutação segura de arquivos quando necessário

## Por que este projeto existe

Assistentes modernos de IA são cada vez mais úteis para desenvolvimento de software, mas interfaces de IA baseadas em navegador normalmente não têm acesso direto e seguro aos arquivos locais de projeto de um desenvolvedor.

Gem Bridge explora uma abordagem local-first em que um assistente de IA pode solicitar ações controladas por meio de um daemon pequeno e auditável executando na própria máquina do desenvolvedor.

O objetivo de longo prazo é criar uma ponte segura entre ferramentas de IA conversacionais e fluxos locais de desenvolvimento sem expor o filesystem inteiro e sem depender de acesso remoto inseguro.

## Licença

Este projeto está atualmente em desenvolvimento ativo. Uma licença será adicionada antes da primeira release pública.
