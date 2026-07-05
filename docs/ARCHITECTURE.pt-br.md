# Arquitetura

[Read in English](./ARCHITECTURE.md)

## Propósito

Gem Bridge é uma bridge local-first de ferramentas escrita em Go.

Seu propósito é expor ferramentas locais controladas de desenvolvimento para assistentes de IA baseados em navegador sem conceder acesso irrestrito à máquina do usuário. O daemon foi desenhado para rodar localmente, impor uma fronteira rígida de workspace e permanecer independente de qualquer provedor de IA ou frontend de navegador específico.

## Objetivos arquiteturais

Gem Bridge segue estes objetivos arquiteturais:

1. **Execução local-first**
   - Arquivos, operações Git e comandos de desenvolvimento rodam na própria máquina do usuário.
   - O daemon não deve depender de execução remota nem de acesso inseguro em nuvem para operar.

2. **Isolamento de workspace**
   - A raiz configurada do workspace é a principal fronteira de segurança.
   - Paths fornecidos pelo usuário devem ser relativos.
   - Paths absolutos, path traversal e escapes por symlink devem ser bloqueados antes que qualquer operação de filesystem execute.

3. **Protocolo agnóstico de ferramenta**
   - O daemon não deve depender de um provedor específico de IA.
   - Assistentes de navegador, extensões de navegador, scripts locais, clientes WebSocket ou hosts de Native Messaging devem poder reutilizar o mesmo modelo de ferramentas.

4. **Preparação cross-platform**
   - O projeto deve ser capaz de suportar Linux, macOS e Windows.
   - Comportamentos específicos de plataforma devem ser isolados quando isso se tornar necessário.
   - O modelo central de ferramentas e segurança deve permanecer compartilhado entre plataformas.

5. **Evolução incremental**
   - A arquitetura deve permanecer simples enquanto o projeto for pequeno.
   - Abstrações devem ser introduzidas quando houver necessidade real, não antes.
   - Comportamentos sensíveis de segurança devem ser documentados e testados conforme o projeto evolui.

## Arquitetura atual

A versão atual é um protótipo baseado em CLI.

```text
JSON request
    ↓
cmd/gem-bridge
    ↓
tool dispatcher
    ↓
internal/tools
    ↓
internal/security
    ↓
authorized workspace
```

A CLI recebe uma requisição JSON, cria um workspace restrito, despacha a ferramenta solicitada e retorna uma resposta JSON consistente.

Ferramentas atualmente suportadas:

- `listDirectory`
- `readFile`

Exemplo atual de requisição:

```json
{
  "tool": "readFile",
  "path": "go.mod"
}
```

Exemplo atual de resposta de sucesso:

```json
{
  "success": true,
  "data": "..."
}
```

Exemplo atual de resposta de erro:

```json
{
  "success": false,
  "error": "access outside the workspace is blocked"
}
```

## Limites de pacotes

O projeto deve evoluir em torno de responsabilidades claras por pacote.

```text
cmd/gem-bridge
    Ponto de entrada da CLI, parsing de argumentos, decodificação de requisições e codificação de respostas.

internal/security
    Imposição da fronteira de workspace, validação de paths, segurança de symlinks e futuros helpers de segurança para comandos.

internal/tools
    Ferramentas expostas a clientes de IA, como leitura de arquivos, listagem de diretórios, futura escrita de arquivos e futuras operações Git.

internal/config
    Futuro carregamento e persistência de configuração local.

internal/transport
    Futuros transportes como WebSocket, HTTP local e Native Messaging.

internal/platform
    Futuro comportamento específico de sistema operacional que não possa permanecer portável pela biblioteca padrão do Go.
```

O projeto deve preferir primeiro limites baseados em domínio. Pacotes específicos de plataforma só devem ser introduzidos quando uma funcionalidade realmente exigir comportamento diferente em Linux, macOS ou Windows.

## Modelo de segurança do workspace

A raiz do workspace é a principal fronteira de segurança.

Todos os paths controlados pelo usuário devem passar pela camada de segurança antes que qualquer operação de filesystem toque no disco.

O daemon deve rejeitar:

- Paths vazios
- Paths absolutos
- Tentativas de path traversal
- Paths que escapem do workspace por symlinks
- Futuros alvos inseguros de escrita
- Futuras requisições inseguras de execução de comandos

Exemplos de paths bloqueados:

```text
/etc/passwd
/home/user/.ssh/id_rsa
../../.env
../../.config
C:\Users\user\.ssh\id_rsa
\\server\share\secret.txt
```

A camada de segurança deve permanecer reutilizável por todos os transportes. Uma requisição vinda da CLI, WebSocket, Native Messaging ou qualquer frontend futuro deve receber a mesma validação de path.

## Estratégia cross-platform

Gem Bridge deve suportar Linux, macOS e Windows, mas diferenças de plataforma devem ser tratadas com cuidado e de forma incremental.

### Paths de filesystem

Tratamento de paths é sensível para segurança e deve ser testado entre plataformas.

Paths absolutos em Linux e macOS normalmente se parecem com:

```text
/home/user/project
/etc/passwd
```

Paths absolutos no Windows podem se parecer com:

```text
C:\Users\User\project
C:/Users/User/project
\\server\share
```

O código deve preferir o pacote `filepath` do Go para comportamento de paths consciente do sistema operacional. Testes cross-platform devem ser adicionados via CI para validar segurança de paths em Linux, macOS e Windows.

### Configuração local

Configuração futura deve usar o diretório padrão de configuração do sistema operacional por meio de `os.UserConfigDir()`.

Exemplos esperados:

```text
Linux:  ~/.config/gem-bridge/config.json
macOS:  ~/Library/Application Support/gem-bridge/config.json
Windows: %APPDATA%\gem-bridge\config.json
```

### Native Messaging

Suporte futuro a extensão de navegador pode usar Native Messaging.

O registro de host Native Messaging é específico por plataforma, então deve ser isolado em um pacote dedicado quando for implementado.

Possível pacote futuro:

```text
internal/platform/nativehost
```

### Execução de comandos

Gem Bridge não deve expor execução irrestrita de shell.

Em vez de aceitar comandos arbitrários, o daemon deve expor ferramentas explícitas como:

- `gitStatus`
- `gitDiff`
- `goTest`
- `goFmt`

Cada ferramenta deve chamar binários conhecidos com argumentos controlados. Isso mantém o comportamento mais seguro e mais fácil de suportar em Linux, macOS e Windows.

## Estratégia de transporte

O transporte atual é baseado em CLI.

Transportes futuros devem reutilizar as mesmas camadas de requisição, resposta, ferramentas e segurança.

Evolução esperada de transporte:

```text
CLI prototype
    ↓
WebSocket transport for local development
    ↓
Browser extension integration
    ↓
Native Messaging support
```

Uma requisição de ferramenta que falha não deve derrubar um futuro daemon de longa duração. No modo CLI, retornar exit status `1` em requisições com falha é aceitável. No modo daemon, erros devem ser retornados como JSON estruturado enquanto o processo continua rodando.

## Princípios de design das ferramentas

Ferramentas expostas a clientes de IA devem ser pequenas, explícitas e auditáveis.

Uma boa ferramenta:

- Tem um nome claro
- Tem um propósito estreito
- Recebe entrada estruturada
- Valida entrada antes de agir
- Usa a camada de segurança para acesso ao filesystem
- Retorna JSON estruturado
- Evita efeitos colaterais ocultos

Uma ferramenta ruim:

- Aceita comandos arbitrários de shell
- Aceita paths absolutos do cliente
- Realiza acesso amplo ao filesystem
- Altera arquivos sem regras claras
- Mistura transporte, lógica de negócio e checagens de segurança no mesmo lugar

## Regras futuras de escrita de arquivos

Escrita de arquivos será mais perigosa do que leitura e deve ter regras explícitas de segurança.

Operações futuras de escrita devem:

- Exigir paths relativos
- Resolver por meio da camada de segurança do workspace
- Bloquear path traversal e escapes por symlink
- Recusar escrita fora do workspace
- Evitar sobrescrever arquivos a menos que isso seja explicitamente permitido
- Considerar limites de tamanho
- Retornar erros JSON claros
- Eventualmente suportar fluxo de aprovação para operações sensíveis

## Regras futuras para ferramentas Git

Ferramentas Git devem ser explícitas e controladas.

Ferramentas Git iniciais mais seguras:

- `gitStatus`
- `gitDiff`

Ferramentas Git futuras mais sensíveis:

- `gitAdd`
- `gitCommit`
- `gitCheckout`
- `gitMerge`
- `gitPush`

Operações Git sensíveis devem exigir validação adicional e possivelmente aprovação do usuário.

## Estratégia de testes

Comportamentos sensíveis de segurança devem ser cobertos por testes automatizados.

Áreas prioritárias atuais:

- Criação da raiz do workspace
- Resolução de paths relativos
- Rejeição de paths vazios
- Rejeição de paths absolutos
- Rejeição de path traversal
- Rejeição de escape por symlink

Áreas prioritárias futuras:

- Comportamento cross-platform de paths em Linux, macOS e Windows
- Regras seguras de escrita de arquivos
- Comportamento das ferramentas Git
- Tratamento de requisições no nível de transporte
- Helpers de registro de Native Messaging
- Comportamento de allowlist de comandos

O projeto deve eventualmente rodar CI em:

```text
ubuntu-latest
macos-latest
windows-latest
```

## Não objetivos para o estágio atual

O estágio atual deve evitar:

- Execução arbitrária de shell
- Automação completa de GUI
- Exposição remota do daemon
- Sistemas complexos de plugins
- Abstrações prematuras de adapter de sistema operacional
- Acesso irrestrito ao filesystem
- Operações sensíveis sem regras explícitas

## Decisão de design: evitar abstração prematura de sistema operacional

Gem Bridge deve estar preparado para cross-platform, mas não deve começar com uma grande camada genérica de abstração de sistema operacional.

Evite introduzir interfaces amplas como:

```go
type OSAdapter interface {
    ResolvePath(path string) (string, error)
    RunCommand(name string, args ...string) ([]byte, error)
    ConfigDir() (string, error)
    RegisterNativeHost() error
}
```

Esse tipo de abstração seria amplo demais para o estágio atual e provavelmente esconderia decisões de segurança atrás de métodos vagos de plataforma.

Em vez disso, o projeto deve introduzir abstrações focadas somente quando necessidades reais específicas de plataforma aparecerem.

## Visão de longo prazo

A arquitetura de longo prazo é:

```text
Browser AI assistant
    ↓
Browser extension
    ↓
Gem Bridge local daemon
    ↓
controlled tools
    ↓
authorized workspace
```

Gem Bridge deve se tornar uma bridge local-first pequena, auditável e segura que permite que assistentes de IA interajam com workspaces de desenvolvimento de forma segura, sem expor todo o filesystem e sem exigir que os usuários migrem totalmente para uma IDE ou ferramenta de IA específica.
